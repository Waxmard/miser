package categorize

import (
	"context"
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

func TestRunRulesCategorizesMatching(t *testing.T) {
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

	rule := &repository.CategoryRule{
		ID: "rule_01", Pattern: "Trader Joe's", CategoryID: "cat_01",
		MatchType: "exact", CreatedBy: "monarch_import", CreatedAt: now,
	}
	if err := repo.Rules().Create(ctx, rule); err != nil {
		t.Fatal(err)
	}

	// One matching, one not.
	txn1 := &repository.Transaction{
		ID: "txn_01", AccountID: "acct_01", Amount: -45.00, Merchant: "Trader Joe's",
		Date: now, Source: "email", Status: "uncategorized", CreatedAt: now, UpdatedAt: now,
	}
	txn2 := &repository.Transaction{
		ID: "txn_02", AccountID: "acct_01", Amount: -20.00, Merchant: "Unknown Store",
		Date: now, Source: "email", Status: "uncategorized", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Transactions().Create(ctx, txn1); err != nil {
		t.Fatal(err)
	}
	if err := repo.Transactions().Create(ctx, txn2); err != nil {
		t.Fatal(err)
	}

	result, err := RunRules(ctx, repo)
	if err != nil {
		t.Fatalf("RunRules() error: %v", err)
	}

	if result.Checked != 2 {
		t.Errorf("Checked = %d, want 2", result.Checked)
	}
	if result.Categorized != 1 {
		t.Errorf("Categorized = %d, want 1", result.Categorized)
	}

	// Verify the matched transaction.
	updated, err := repo.Transactions().GetByID(ctx, "txn_01")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "categorized" {
		t.Errorf("Status = %q, want %q", updated.Status, "categorized")
	}
	if updated.CategoryName != "Groceries" {
		t.Errorf("CategoryName = %q, want %q", updated.CategoryName, "Groceries")
	}
	if updated.CategorizedBy == nil || *updated.CategorizedBy != "rule" {
		t.Errorf("CategorizedBy = %v, want %q", updated.CategorizedBy, "rule")
	}

	// Verify the unmatched transaction is still uncategorized.
	unchanged, err := repo.Transactions().GetByID(ctx, "txn_02")
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.Status != "uncategorized" {
		t.Errorf("Status = %q, want %q", unchanged.Status, "uncategorized")
	}

	// Verify hit count was incremented.
	rules, err := repo.Rules().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(rules))
	}
	if rules[0].HitCount != 1 {
		t.Errorf("HitCount = %d, want 1", rules[0].HitCount)
	}
}

func TestRunRulesNoUncategorized(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	result, err := RunRules(ctx, repo)
	if err != nil {
		t.Fatalf("RunRules() error: %v", err)
	}
	if result.Checked != 0 {
		t.Errorf("Checked = %d, want 0", result.Checked)
	}
	if result.Categorized != 0 {
		t.Errorf("Categorized = %d, want 0", result.Categorized)
	}
}

func TestRunRulesContainsMatch(t *testing.T) {
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

	cat := &repository.Category{ID: "cat_01", Name: "Charging", CreatedAt: now}
	if err := repo.Categories().Create(ctx, cat); err != nil {
		t.Fatal(err)
	}

	rule := &repository.CategoryRule{
		ID: "rule_01", Pattern: "ChargePoint", CategoryID: "cat_01",
		MatchType: "contains", CreatedBy: "claude", CreatedAt: now,
	}
	if err := repo.Rules().Create(ctx, rule); err != nil {
		t.Fatal(err)
	}

	txn := &repository.Transaction{
		ID: "txn_01", AccountID: "acct_01", Amount: -12.40,
		Merchant: "ChargePoint Station #5678",
		Date:     now, Source: "email", Status: "uncategorized", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Transactions().Create(ctx, txn); err != nil {
		t.Fatal(err)
	}

	result, err := RunRules(ctx, repo)
	if err != nil {
		t.Fatalf("RunRules() error: %v", err)
	}
	if result.Categorized != 1 {
		t.Errorf("Categorized = %d, want 1", result.Categorized)
	}

	updated, err := repo.Transactions().GetByID(ctx, "txn_01")
	if err != nil {
		t.Fatal(err)
	}
	if updated.CategoryName != "Charging" {
		t.Errorf("CategoryName = %q, want %q", updated.CategoryName, "Charging")
	}
}
