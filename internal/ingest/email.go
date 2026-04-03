package ingest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	"github.com/Waxmard/miser/internal/config"
	"github.com/Waxmard/miser/internal/repository"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	gomessage "github.com/emersion/go-message/mail"
	"github.com/oklog/ulid/v2"
)

// EmailSyncResult holds the summary of an email sync.
type EmailSyncResult struct {
	Found  int
	Stored int
}

// SyncEmail connects to Gmail via IMAP, fetches new emails from the configured
// label, and stores them in raw_emails. It tracks the last seen UID in sync_state
// to avoid re-fetching.
func SyncEmail(ctx context.Context, repo repository.Repository, cfg *config.EmailConfig) (*EmailSyncResult, error) {
	result := &EmailSyncResult{}

	opts := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	}

	addr := fmt.Sprintf("%s:%d", cfg.IMAPServer, cfg.IMAPPort)
	client, err := imapclient.DialTLS(addr, opts)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Login(cfg.Username, cfg.AppPassword).Wait(); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	// Select the mailbox (Gmail label as IMAP folder).
	if _, err := client.Select(cfg.Label, nil).Wait(); err != nil {
		return nil, fmt.Errorf("select %q: %w", cfg.Label, err)
	}

	// Determine search start: use last synced UID or fall back to 30 days ago.
	var sinceDate time.Time
	syncState, err := repo.SyncState().Get(ctx, "email")
	if err != nil {
		sinceDate = time.Now().AddDate(0, 0, -30)
	} else {
		sinceDate = syncState.LastSyncAt.AddDate(0, 0, -1) // overlap by 1 day for safety
	}

	// Search for messages since the date.
	criteria := &imap.SearchCriteria{
		Since: sinceDate,
	}
	searchCmd := client.Search(criteria, nil)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		return result, nil
	}

	// Fetch matching messages.
	seqSet := imap.SeqSetNum(seqNums...)
	fetchOpts := &imap.FetchOptions{
		UID:      true,
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierText},
		},
	}

	fetchCmd := client.Fetch(seqSet, fetchOpts)
	defer func() { _ = fetchCmd.Close() }()

	var maxUID imap.UID
	now := time.Now().UTC()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		buf, err := msg.Collect()
		if err != nil {
			continue // skip individual failures
		}

		result.Found++

		if buf.UID > maxUID {
			maxUID = buf.UID
		}

		// Skip if we've already seen this UID.
		if syncState != nil && syncState.LastMessageUID != nil && buf.UID <= imap.UID(*syncState.LastMessageUID) {
			continue
		}

		messageID := ""
		subject := ""
		from := ""
		var receivedAt time.Time

		if buf.Envelope != nil {
			messageID = buf.Envelope.MessageID
			subject = buf.Envelope.Subject
			if len(buf.Envelope.From) > 0 {
				from = buf.Envelope.From[0].Addr()
			}
			receivedAt = buf.Envelope.Date
		}
		if receivedAt.IsZero() {
			receivedAt = buf.InternalDate
		}

		// Extract body text.
		body := extractBody(buf)

		if messageID == "" {
			messageID = fmt.Sprintf("uid-%d-%d", buf.UID, receivedAt.Unix())
		}

		email := &repository.RawEmail{
			ID:         ulid.Make().String(),
			MessageID:  messageID,
			Subject:    subject,
			From:       from,
			Body:       body,
			ReceivedAt: receivedAt,
			Status:     "pending",
			CreatedAt:  now,
		}

		if err := repo.RawEmails().Create(ctx, email); err != nil {
			// Likely a duplicate (unique constraint on message_id), skip.
			continue
		}
		result.Stored++
	}

	if err := fetchCmd.Close(); err != nil {
		return result, fmt.Errorf("fetch close: %w", err)
	}

	// Update sync state.
	uid := uint32(maxUID)
	if err := repo.SyncState().Upsert(ctx, &repository.SyncState{
		Source:         "email",
		LastSyncAt:     now,
		LastMessageUID: &uid,
	}); err != nil {
		return result, fmt.Errorf("update sync state: %w", err)
	}

	if err := client.Logout().Wait(); err != nil {
		// Non-fatal, connection will be closed anyway.
		_ = err
	}

	return result, nil
}

func extractBody(buf *imapclient.FetchMessageBuffer) string {
	for _, section := range buf.BodySection {
		if len(section.Bytes) == 0 {
			continue
		}

		// Try to parse as a MIME message to get the text part.
		reader, err := gomessage.CreateReader(bytes.NewReader(section.Bytes))
		if err != nil {
			// Not a valid MIME message, return raw text.
			return string(section.Bytes)
		}

		for {
			part, err := reader.NextPart()
			if err != nil {
				break
			}
			ct := part.Header.Get("Content-Type")
			if strings.HasPrefix(ct, "text/plain") || strings.HasPrefix(ct, "text/html") {
				body, err := io.ReadAll(part.Body)
				if err != nil {
					continue
				}
				return string(body)
			}
		}

		// Fallback: return raw bytes.
		return string(section.Bytes)
	}
	return ""
}
