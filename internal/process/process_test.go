package process

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
)

func newTestRepo(t *testing.T) repository.Repository {
	t.Helper()
	repo, err := repository.New("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) error: %v", err)
	}
	t.Cleanup(func() { repo.Close() })

	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error: %v", err)
	}
	return repo
}

func TestPrintPendingEmails(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	email := &repository.RawEmail{
		ID:         "email_01",
		MessageID:  "<test@mail.com>",
		Subject:    "Transaction Alert",
		From:       "alerts@fidelity.com",
		Body:       "A transaction of $50.00 was made at STARBUCKS",
		ReceivedAt: now,
		Status:     "pending",
		CreatedAt:  now,
	}
	if err := repo.RawEmails().Create(ctx, email); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := PrintPendingEmails(ctx, repo, "Test Account", &buf); err != nil {
		t.Fatalf("PrintPendingEmails() error: %v", err)
	}

	var out PendingEmailsOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if out.PendingCount != 1 {
		t.Errorf("PendingCount = %d, want 1", out.PendingCount)
	}
	if out.AccountName != "Test Account" {
		t.Errorf("AccountName = %q, want %q", out.AccountName, "Test Account")
	}
	if len(out.Emails) != 1 {
		t.Fatalf("len(Emails) = %d, want 1", len(out.Emails))
	}
	if out.Emails[0].Subject != "Transaction Alert" {
		t.Errorf("Subject = %q, want %q", out.Emails[0].Subject, "Transaction Alert")
	}
}

func TestPrintUncategorized(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	acct := &repository.Account{
		ID: "acct_01", Name: "Checking", Institution: "test",
		AccountType: "checking", Source: "manual", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Accounts().Create(ctx, acct); err != nil {
		t.Fatal(err)
	}

	cat := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	if err := repo.Categories().Create(ctx, cat); err != nil {
		t.Fatal(err)
	}

	txn := &repository.Transaction{
		ID: "txn_01", AccountID: "acct_01", Amount: -25.00, Merchant: "New Place",
		Date: now, Source: "email", Status: "uncategorized", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Transactions().Create(ctx, txn); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := PrintUncategorized(ctx, repo, &buf); err != nil {
		t.Fatalf("PrintUncategorized() error: %v", err)
	}

	var out UncategorizedOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if out.UncategorizedCount != 1 {
		t.Errorf("UncategorizedCount = %d, want 1", out.UncategorizedCount)
	}
	if len(out.Categories) < 1 {
		t.Error("expected at least 1 category")
	}
}

func TestWriteParsedEmails(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	acct := &repository.Account{
		ID: "acct_01", Name: "Checking", Institution: "test",
		AccountType: "checking", Source: "manual", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Accounts().Create(ctx, acct); err != nil {
		t.Fatal(err)
	}

	email := &repository.RawEmail{
		ID: "email_01", MessageID: "<test@mail.com>", Subject: "Alert",
		From: "alerts@fidelity.com", Body: "test", ReceivedAt: now,
		Status: "pending", CreatedAt: now,
	}
	if err := repo.RawEmails().Create(ctx, email); err != nil {
		t.Fatal(err)
	}

	input := ParsedEmailsInput{
		Results: []ParsedEmailResult{
			{
				RawEmailID: "email_01",
				Parsed:     true,
				Transaction: &ParsedTransaction{
					Amount:        -50.00,
					Merchant:      "STARBUCKS #1234",
					MerchantClean: "Starbucks",
					Date:          "2026-03-28",
					Description:   "Debit card purchase",
				},
			},
		},
	}

	jsonPath := filepath.Join(t.TempDir(), "parsed.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	count, err := WriteParsedEmails(ctx, repo, jsonPath, "acct_01")
	if err != nil {
		t.Fatalf("WriteParsedEmails() error: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Verify the transaction was created (use list since we don't know the ULID).
	txns, err := repo.Transactions().List(ctx, &repository.TransactionFilters{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 1 {
		t.Fatalf("len(txns) = %d, want 1", len(txns))
	}
	if txns[0].Merchant != "STARBUCKS #1234" {
		t.Errorf("Merchant = %q, want %q", txns[0].Merchant, "STARBUCKS #1234")
	}
	if txns[0].Status != "uncategorized" {
		t.Errorf("Status = %q, want %q", txns[0].Status, "uncategorized")
	}
}

func TestWriteCategories(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	acct := &repository.Account{
		ID: "acct_01", Name: "Checking", Institution: "test",
		AccountType: "checking", Source: "manual", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Accounts().Create(ctx, acct); err != nil {
		t.Fatal(err)
	}

	cat := &repository.Category{ID: "cat_01", Name: "Liquids", CreatedAt: now}
	if err := repo.Categories().Create(ctx, cat); err != nil {
		t.Fatal(err)
	}

	txn := &repository.Transaction{
		ID: "txn_01", AccountID: "acct_01", Amount: -5.50, Merchant: "Starbucks",
		Date: now, Source: "email", Status: "uncategorized", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Transactions().Create(ctx, txn); err != nil {
		t.Fatal(err)
	}

	input := CategoriesInput{
		Results: []CategoryResult{
			{
				TransactionID: "txn_01",
				Category:      "Liquids",
				MerchantClean: "Starbucks",
				Confidence:    0.95,
				SuggestedRule: &SuggestedRule{
					Pattern:   "Starbucks",
					MatchType: "contains",
				},
			},
		},
	}

	jsonPath := filepath.Join(t.TempDir(), "categories.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	count, err := WriteCategories(ctx, repo, jsonPath)
	if err != nil {
		t.Fatalf("WriteCategories() error: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Verify the transaction was categorized.
	updated, err := repo.Transactions().GetByID(ctx, "txn_01")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "categorized" {
		t.Errorf("Status = %q, want %q", updated.Status, "categorized")
	}
	if updated.CategoryName != "Liquids" {
		t.Errorf("CategoryName = %q, want %q", updated.CategoryName, "Liquids")
	}

	// Verify a rule was auto-created (confidence >= 0.85).
	rule, err := repo.Rules().FindMatch(ctx, "Starbucks")
	if err != nil {
		t.Fatalf("FindMatch() error: %v", err)
	}
	if rule.Pattern != "Starbucks" {
		t.Errorf("Rule.Pattern = %q, want %q", rule.Pattern, "Starbucks")
	}
}

func TestWriteReport(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	input := ReportInput{
		Year:      2026,
		Month:     3,
		Narrative: "Spending was moderate this month.",
	}

	jsonPath := filepath.Join(t.TempDir(), "report.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteReport(ctx, repo, jsonPath); err != nil {
		t.Fatalf("WriteReport() error: %v", err)
	}

	report, err := repo.Reports().GetByMonth(ctx, 2026, 3)
	if err != nil {
		t.Fatalf("GetByMonth() error: %v", err)
	}
	if report.Narrative != "Spending was moderate this month." {
		t.Errorf("Narrative = %q, want %q", report.Narrative, "Spending was moderate this month.")
	}
}
