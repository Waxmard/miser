package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

func newTestDB(t *testing.T) repository.Repository {
	t.Helper()
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) error: %v", err)
	}
	t.Cleanup(func() { repo.Close() })

	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error: %v", err)
	}
	return repo
}

func TestMigrate(t *testing.T) {
	repo := newTestDB(t)

	// Running migrate again should be idempotent.
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("second Migrate() error: %v", err)
	}
}

func TestAccountsCRUD(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	acct := &repository.Account{
		ID:          "acct_01",
		Name:        "Test Checking",
		Institution: "test_bank",
		AccountType: "checking",
		Source:      "manual",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := repo.Accounts().Create(ctx, acct); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := repo.Accounts().GetByID(ctx, "acct_01")
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if got.Name != "Test Checking" {
		t.Errorf("Name = %q, want %q", got.Name, "Test Checking")
	}

	got, err = repo.Accounts().GetByName(ctx, "Test Checking")
	if err != nil {
		t.Fatalf("GetByName() error: %v", err)
	}
	if got.ID != "acct_01" {
		t.Errorf("ID = %q, want %q", got.ID, "acct_01")
	}

	accounts, err := repo.Accounts().List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("List() returned %d accounts, want 1", len(accounts))
	}
}

func TestCategorySeed(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	cats := []repository.Category{
		{ID: "cat_01", Name: "Groceries", CreatedAt: now},
		{ID: "cat_02", Name: "Restaurants", CreatedAt: now},
		{ID: "cat_03", Name: "Bars/Drinking", CreatedAt: now},
	}

	if err := repo.Categories().Seed(ctx, cats); err != nil {
		t.Fatalf("Seed() error: %v", err)
	}

	list, err := repo.Categories().List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List() returned %d categories, want 3", len(list))
	}

	// Seed again should be idempotent (INSERT OR IGNORE).
	if err := repo.Categories().Seed(ctx, cats); err != nil {
		t.Fatalf("second Seed() error: %v", err)
	}
	list, err = repo.Categories().List(ctx)
	if err != nil {
		t.Fatalf("List() after re-seed error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List() after re-seed returned %d categories, want 3", len(list))
	}
}

func TestTransactionCreateAndQuery(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	// Create account and category first.
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

	catID := "cat_01"
	sourceID := "src_001"
	txn := &repository.Transaction{
		ID:         "txn_01",
		AccountID:  "acct_01",
		CategoryID: &catID,
		Amount:     -87.32,
		Merchant:   "Whole Foods",
		Date:       now,
		Source:     "email",
		SourceID:   &sourceID,
		Status:     "categorized",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repo.Transactions().Create(ctx, txn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := repo.Transactions().GetByID(ctx, "txn_01")
	if err != nil {
		t.Fatalf("GetByID() error: %v", err)
	}
	if got.Amount != -87.32 {
		t.Errorf("Amount = %f, want %f", got.Amount, -87.32)
	}
	if got.CategoryName != "Groceries" {
		t.Errorf("CategoryName = %q, want %q", got.CategoryName, "Groceries")
	}
	if got.AccountName != "Checking" {
		t.Errorf("AccountName = %q, want %q", got.AccountName, "Checking")
	}

	// Test List with filters.
	txns, err := repo.Transactions().List(ctx, &repository.TransactionFilters{Limit: 10})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(txns) != 1 {
		t.Errorf("List() returned %d, want 1", len(txns))
	}

	// Test FindBySourceID.
	got, err = repo.Transactions().FindBySourceID(ctx, "email", "src_001")
	if err != nil {
		t.Fatalf("FindBySourceID() error: %v", err)
	}
	if got.ID != "txn_01" {
		t.Errorf("ID = %q, want %q", got.ID, "txn_01")
	}
}

func TestRuleFindMatch(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	cat := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	if err := repo.Categories().Create(ctx, cat); err != nil {
		t.Fatal(err)
	}

	rule := &repository.CategoryRule{
		ID: "rule_01", Pattern: "Whole Foods", CategoryID: "cat_01",
		MatchType: "exact", CreatedBy: "manual", CreatedAt: now,
	}
	if err := repo.Rules().Create(ctx, rule); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Exact match (case-insensitive).
	got, err := repo.Rules().FindMatch(ctx, "Whole Foods")
	if err != nil {
		t.Fatalf("FindMatch() error: %v", err)
	}
	if got.CategoryName != "Groceries" {
		t.Errorf("CategoryName = %q, want %q", got.CategoryName, "Groceries")
	}

	// Contains match.
	containsRule := &repository.CategoryRule{
		ID: "rule_02", Pattern: "ChargePoint", CategoryID: "cat_01",
		MatchType: "contains", CreatedBy: "manual", CreatedAt: now,
	}
	if err := repo.Rules().Create(ctx, containsRule); err != nil {
		t.Fatal(err)
	}

	got, err = repo.Rules().FindMatch(ctx, "ChargePoint Station #1234")
	if err != nil {
		t.Fatalf("FindMatch(contains) error: %v", err)
	}
	if got.Pattern != "ChargePoint" {
		t.Errorf("Pattern = %q, want %q", got.Pattern, "ChargePoint")
	}
}

func TestBatchInsertDedup(t *testing.T) {
	repo := newTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	acct := &repository.Account{
		ID: "acct_01", Name: "Checking", Institution: "test",
		AccountType: "checking", Source: "manual", CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Accounts().Create(ctx, acct); err != nil {
		t.Fatal(err)
	}

	sourceID := "dup_001"
	txns := []repository.Transaction{
		{ID: "txn_01", AccountID: "acct_01", Amount: -10, Merchant: "Test", Date: now,
			Source: "email", SourceID: &sourceID, Status: "uncategorized", CreatedAt: now, UpdatedAt: now},
		{ID: "txn_02", AccountID: "acct_01", Amount: -10, Merchant: "Test", Date: now,
			Source: "email", SourceID: &sourceID, Status: "uncategorized", CreatedAt: now, UpdatedAt: now},
	}

	inserted, err := repo.Transactions().CreateBatch(ctx, txns)
	if err != nil {
		t.Fatalf("CreateBatch() error: %v", err)
	}
	if inserted != 1 {
		t.Errorf("CreateBatch() inserted %d, want 1 (dedup)", inserted)
	}
}
