package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Waxmard/miser/internal/repository"
)

type PendingEmailsOutput struct {
	PendingCount int           `json:"pending_count"`
	Emails       []EmailOutput `json:"emails"`
	AccountName  string        `json:"account_name"`
}

type EmailOutput struct {
	ID         string `json:"id"`
	MessageID  string `json:"message_id"`
	Subject    string `json:"subject"`
	From       string `json:"from"`
	Body       string `json:"body"`
	ReceivedAt string `json:"received_at"`
}

// PrintPendingEmails writes pending raw emails as JSON to w.
func PrintPendingEmails(ctx context.Context, repo repository.Repository, accountName string, w io.Writer) error {
	emails, err := repo.RawEmails().GetPending(ctx, 0)
	if err != nil {
		return fmt.Errorf("get pending emails: %w", err)
	}

	out := PendingEmailsOutput{
		PendingCount: len(emails),
		AccountName:  accountName,
	}

	for i := range emails {
		e := &emails[i]
		out.Emails = append(out.Emails, EmailOutput{
			ID:         e.ID,
			MessageID:  e.MessageID,
			Subject:    e.Subject,
			From:       e.From,
			Body:       e.Body,
			ReceivedAt: e.ReceivedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
