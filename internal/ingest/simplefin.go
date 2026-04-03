package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Waxmard/miser/internal/categorize"
	"github.com/Waxmard/miser/internal/config"
	"github.com/Waxmard/miser/internal/repository"
	"github.com/oklog/ulid/v2"
)

// SimpleFinSyncResult holds the summary of a SimpleFIN sync.
type SimpleFinSyncResult struct {
	AccountsSynced int
	Found          int
	Stored         int
	Categorized    int
}

// simpleFIN API response types.
type sfAccountSet struct {
	Errors   []string    `json:"errors"`
	Accounts []sfAccount `json:"accounts"`
}

type sfAccount struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Currency     string          `json:"currency"`
	Balance      string          `json:"balance"`
	BalanceDate  int64           `json:"balance-date"`
	Transactions []sfTransaction `json:"transactions"`
}

type sfTransaction struct {
	ID          string `json:"id"`
	Posted      int64  `json:"posted"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Pending     bool   `json:"pending"`
}

// SyncSimpleFIN fetches accounts and transactions from SimpleFIN and stores
// them in the repository. It runs the rule engine on any new uncategorized
// transactions afterward.
func SyncSimpleFIN(ctx context.Context, repo repository.Repository, cfg *config.SimpleFinConfig) (*SimpleFinSyncResult, error) {
	result := &SimpleFinSyncResult{}

	// Determine start date: last sync or 90 days ago.
	startDate := time.Now().AddDate(0, 0, -90)
	syncState, err := repo.SyncState().Get(ctx, "simplefin")
	if err == nil {
		startDate = syncState.LastSyncAt.AddDate(0, 0, -1) // overlap by 1 day
	}

	accountSet, err := fetchSimpleFinAccounts(ctx, cfg.AccessURL, startDate)
	if err != nil {
		return nil, fmt.Errorf("fetch simplefin: %w", err)
	}

	for _, errMsg := range accountSet.Errors {
		fmt.Printf("SimpleFIN warning: %s\n", errMsg)
	}

	// Map SimpleFIN accounts to miser accounts.
	accountMap, err := ensureSimpleFinAccounts(ctx, repo, accountSet.Accounts)
	if err != nil {
		return nil, fmt.Errorf("ensure accounts: %w", err)
	}
	result.AccountsSynced = len(accountMap)

	// Build and insert transactions.
	txns := buildSimpleFinTransactions(accountSet.Accounts, accountMap)
	result.Found = len(txns)

	if len(txns) > 0 {
		inserted, err := repo.Transactions().CreateBatch(ctx, txns)
		if err != nil {
			return nil, fmt.Errorf("insert transactions: %w", err)
		}
		result.Stored = inserted

		// Run rule engine on new uncategorized transactions.
		if inserted > 0 {
			ruleResult, err := categorize.RunRules(ctx, repo)
			if err != nil {
				fmt.Printf("Warning: rule engine error: %v\n", err)
			} else {
				result.Categorized = ruleResult.Categorized
			}
		}
	}

	// Update sync state.
	if err := repo.SyncState().Upsert(ctx, &repository.SyncState{
		Source:     "simplefin",
		LastSyncAt: time.Now().UTC(),
	}); err != nil {
		return result, fmt.Errorf("update sync state: %w", err)
	}

	return result, nil
}

func fetchSimpleFinAccounts(ctx context.Context, accessURL string, startDate time.Time) (*sfAccountSet, error) {
	parsed, err := url.Parse(accessURL)
	if err != nil {
		return nil, fmt.Errorf("parse access URL: %w", err)
	}

	// Build the /accounts endpoint URL.
	endpoint := fmt.Sprintf("%s://%s@%s%s/accounts",
		parsed.Scheme, parsed.User.String(), parsed.Host, parsed.Path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add start-date query parameter.
	q := req.URL.Query()
	q.Set("start-date", strconv.FormatInt(startDate.Unix(), 10))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("access denied — access URL may be revoked (run: miser setup simplefin <new-token>)")
	}
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, fmt.Errorf("SimpleFIN subscription expired — renew at https://beta-bridge.simplefin.org")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var accountSet sfAccountSet
	if err := json.NewDecoder(resp.Body).Decode(&accountSet); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &accountSet, nil
}

func ensureSimpleFinAccounts(ctx context.Context, repo repository.Repository, accounts []sfAccount) (map[string]string, error) {
	accountMap := make(map[string]string) // simplefin account id -> miser account id
	now := time.Now().UTC()

	for i := range accounts {
		sfAcct := &accounts[i]

		// Try to find by name first (may already exist from Monarch import).
		existing, err := repo.Accounts().GetByName(ctx, sfAcct.Name)
		if err == nil {
			accountMap[sfAcct.ID] = existing.ID
			continue
		}

		id := ulid.Make().String()
		acct := &repository.Account{
			ID:          id,
			Name:        sfAcct.Name,
			Institution: guessInstitution(sfAcct.Name),
			AccountType: guessAccountType(sfAcct.Name),
			Source:      "simplefin",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := repo.Accounts().Create(ctx, acct); err != nil {
			return nil, fmt.Errorf("create account %q: %w", sfAcct.Name, err)
		}
		accountMap[sfAcct.ID] = id
	}

	return accountMap, nil
}

func buildSimpleFinTransactions(accounts []sfAccount, accountMap map[string]string) []repository.Transaction {
	now := time.Now().UTC()
	var txns []repository.Transaction

	for i := range accounts {
		sfAcct := &accounts[i]
		miserAccountID := accountMap[sfAcct.ID]

		for j := range sfAcct.Transactions {
			sfTxn := &sfAcct.Transactions[j]

			// Skip pending transactions.
			if sfTxn.Pending {
				continue
			}

			amount, err := strconv.ParseFloat(sfTxn.Amount, 64)
			if err != nil {
				continue
			}

			date := time.Unix(sfTxn.Posted, 0).UTC()
			sourceID := fmt.Sprintf("simplefin_%s", sfTxn.ID)

			txn := repository.Transaction{
				ID:                ulid.Make().String(),
				AccountID:         miserAccountID,
				Amount:            amount,
				Merchant:          cleanMerchant(sfTxn.Description),
				OriginalStatement: strPtr(sfTxn.Description),
				Date:              date,
				Source:            "simplefin",
				SourceID:          &sourceID,
				Status:            "uncategorized",
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			txns = append(txns, txn)
		}
	}

	return txns
}

// cleanMerchant does basic cleanup of a bank transaction description.
func cleanMerchant(desc string) string {
	s := strings.TrimSpace(desc)
	// Remove common transaction prefixes.
	for _, prefix := range []string{"POS ", "ACH ", "DEBIT ", "CREDIT "} {
		s = strings.TrimPrefix(s, prefix)
	}
	return s
}
