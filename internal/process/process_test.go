package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestPrintBudgetData(t *testing.T) {
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

	groceries := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	dining := &repository.Category{ID: "cat_02", Name: "Dining", CreatedAt: now}
	if err := repo.Categories().Create(ctx, groceries); err != nil {
		t.Fatal(err)
	}
	if err := repo.Categories().Create(ctx, dining); err != nil {
		t.Fatal(err)
	}

	// Create transactions across 3 of the last 6 months.
	curYear, curMonth, _ := now.Date()
	currentMonthStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)

	months := []time.Time{
		currentMonthStart.AddDate(0, -1, 5), // last month
		currentMonthStart.AddDate(0, -2, 5), // 2 months ago
		currentMonthStart.AddDate(0, -3, 5), // 3 months ago
	}

	catID := groceries.ID
	categorizedBy := "manual"
	for i, date := range months {
		txn := &repository.Transaction{
			ID: fmt.Sprintf("txn_g_%d", i), AccountID: "acct_01",
			Amount: -100.00 * float64(i+1), Merchant: "Grocery Store",
			Date: date, Source: "manual", Status: "categorized",
			CategoryID: &catID, CategorizedBy: &categorizedBy,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := repo.Transactions().Create(ctx, txn); err != nil {
			t.Fatal(err)
		}
	}

	diningID := dining.ID
	txn := &repository.Transaction{
		ID: "txn_d_0", AccountID: "acct_01",
		Amount: -50.00, Merchant: "Restaurant",
		Date: months[0], Source: "manual", Status: "categorized",
		CategoryID: &diningID, CategorizedBy: &categorizedBy,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Transactions().Create(ctx, txn); err != nil {
		t.Fatal(err)
	}

	// Set an existing budget for groceries.
	budget := &repository.Budget{
		ID: "bud_01", CategoryID: "cat_01", MonthlyAmount: 500.00,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Budgets().Set(ctx, budget); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := PrintBudgetData(ctx, repo, &buf); err != nil {
		t.Fatalf("PrintBudgetData() error: %v", err)
	}

	var out BudgetDataOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if out.MonthsIncluded != 6 {
		t.Errorf("MonthsIncluded = %d, want 6", out.MonthsIncluded)
	}
	if len(out.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(out.Categories))
	}

	// Categories should be sorted alphabetically.
	if out.Categories[0].Category != "Dining" {
		t.Errorf("Categories[0].Category = %q, want %q", out.Categories[0].Category, "Dining")
	}
	if out.Categories[1].Category != "Groceries" {
		t.Errorf("Categories[1].Category = %q, want %q", out.Categories[1].Category, "Groceries")
	}

	// Each category should have 6 months of data.
	for _, cat := range out.Categories {
		if len(cat.Months) != 6 {
			t.Errorf("Category %q: len(Months) = %d, want 6", cat.Category, len(cat.Months))
		}
	}

	// Existing budgets should be present.
	if len(out.ExistingBudgets) != 1 {
		t.Fatalf("len(ExistingBudgets) = %d, want 1", len(out.ExistingBudgets))
	}
	if out.ExistingBudgets[0].Category != "Groceries" {
		t.Errorf("ExistingBudgets[0].Category = %q, want %q", out.ExistingBudgets[0].Category, "Groceries")
	}
}

func TestPrintBudgetData_EmptyDB(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	var buf bytes.Buffer
	if err := PrintBudgetData(ctx, repo, &buf); err != nil {
		t.Fatalf("PrintBudgetData() error: %v", err)
	}

	var out BudgetDataOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if out.MonthsIncluded != 6 {
		t.Errorf("MonthsIncluded = %d, want 6", out.MonthsIncluded)
	}
	if len(out.Categories) != 0 {
		t.Errorf("len(Categories) = %d, want 0", len(out.Categories))
	}
}

func TestWriteBudgets(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	groceries := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	dining := &repository.Category{ID: "cat_02", Name: "Dining", CreatedAt: now}
	if err := repo.Categories().Create(ctx, groceries); err != nil {
		t.Fatal(err)
	}
	if err := repo.Categories().Create(ctx, dining); err != nil {
		t.Fatal(err)
	}

	// Create an existing budget for groceries.
	budget := &repository.Budget{
		ID: "bud_01", CategoryID: "cat_01", MonthlyAmount: 500.00,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Budgets().Set(ctx, budget); err != nil {
		t.Fatal(err)
	}

	input := BudgetsInput{
		Budgets: []BudgetSuggestion{
			{CategoryID: "cat_01", Category: "Groceries", Amount: 550.00, Reasoning: "test"},
			{CategoryID: "cat_02", Category: "Dining", Amount: 200.00, Reasoning: "test"},
		},
	}

	jsonPath := filepath.Join(t.TempDir(), "budgets.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := WriteBudgets(ctx, repo, jsonPath)
	if err != nil {
		t.Fatalf("WriteBudgets() error: %v", err)
	}
	if result.Set != 2 {
		t.Errorf("Set = %d, want 2", result.Set)
	}

	// Verify no duplicates -- should be exactly 2 budgets.
	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 2 {
		t.Fatalf("len(budgets) = %d, want 2", len(budgets))
	}

	// Groceries budget should have been updated (same ID).
	grocBudget, err := repo.Budgets().GetByCategoryID(ctx, "cat_01")
	if err != nil {
		t.Fatal(err)
	}
	if grocBudget.ID != "bud_01" {
		t.Errorf("Groceries budget ID = %q, want %q (should reuse existing)", grocBudget.ID, "bud_01")
	}
	if grocBudget.MonthlyAmount != 550.00 {
		t.Errorf("Groceries MonthlyAmount = %f, want 550.00", grocBudget.MonthlyAmount)
	}

	// Dining budget should have been newly created.
	dinBudget, err := repo.Budgets().GetByCategoryID(ctx, "cat_02")
	if err != nil {
		t.Fatal(err)
	}
	if dinBudget.MonthlyAmount != 200.00 {
		t.Errorf("Dining MonthlyAmount = %f, want 200.00", dinBudget.MonthlyAmount)
	}
}

func TestWriteBudgets_UpdateExisting(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	cat := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	if err := repo.Categories().Create(ctx, cat); err != nil {
		t.Fatal(err)
	}

	budget := &repository.Budget{
		ID: "bud_01", CategoryID: "cat_01", MonthlyAmount: 500.00,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Budgets().Set(ctx, budget); err != nil {
		t.Fatal(err)
	}

	input := BudgetsInput{
		Budgets: []BudgetSuggestion{
			{CategoryID: "cat_01", Category: "Groceries", Amount: 550.00, Reasoning: "test"},
		},
	}

	jsonPath := filepath.Join(t.TempDir(), "budgets.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := WriteBudgets(ctx, repo, jsonPath)
	if err != nil {
		t.Fatalf("WriteBudgets() error: %v", err)
	}
	if result.Set != 1 {
		t.Errorf("Set = %d, want 1", result.Set)
	}

	// Should be exactly 1 budget, not 2.
	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatalf("len(budgets) = %d, want 1", len(budgets))
	}
	if budgets[0].ID != "bud_01" {
		t.Errorf("Budget ID = %q, want %q", budgets[0].ID, "bud_01")
	}
	if budgets[0].MonthlyAmount != 550.00 {
		t.Errorf("MonthlyAmount = %f, want 550.00", budgets[0].MonthlyAmount)
	}
}

func TestWriteBudgets_Remove(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()

	groceries := &repository.Category{ID: "cat_01", Name: "Groceries", CreatedAt: now}
	dining := &repository.Category{ID: "cat_02", Name: "Dining", CreatedAt: now}
	if err := repo.Categories().Create(ctx, groceries); err != nil {
		t.Fatal(err)
	}
	if err := repo.Categories().Create(ctx, dining); err != nil {
		t.Fatal(err)
	}

	// Create budgets for both categories.
	for _, b := range []*repository.Budget{
		{ID: "bud_01", CategoryID: "cat_01", MonthlyAmount: 500.00, CreatedAt: now, UpdatedAt: now},
		{ID: "bud_02", CategoryID: "cat_02", MonthlyAmount: 200.00, CreatedAt: now, UpdatedAt: now},
	} {
		if err := repo.Budgets().Set(ctx, b); err != nil {
			t.Fatal(err)
		}
	}

	// Update groceries, remove dining.
	input := BudgetsInput{
		Budgets: []BudgetSuggestion{
			{CategoryID: "cat_01", Category: "Groceries", Amount: 550.00, Reasoning: "test"},
		},
		Remove: []string{"cat_02"},
	}

	jsonPath := filepath.Join(t.TempDir(), "budgets.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := WriteBudgets(ctx, repo, jsonPath)
	if err != nil {
		t.Fatalf("WriteBudgets() error: %v", err)
	}
	if result.Set != 1 {
		t.Errorf("Set = %d, want 1", result.Set)
	}
	if result.Removed != 1 {
		t.Errorf("Removed = %d, want 1", result.Removed)
	}

	// Should be exactly 1 budget remaining.
	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(budgets) != 1 {
		t.Fatalf("len(budgets) = %d, want 1", len(budgets))
	}
	if budgets[0].CategoryName != "Groceries" {
		t.Errorf("remaining budget category = %q, want %q", budgets[0].CategoryName, "Groceries")
	}
}

func TestWriteBudgets_InvalidCategoryID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	input := BudgetsInput{
		Budgets: []BudgetSuggestion{
			{CategoryID: "nonexistent", Category: "Fake", Amount: 100.00, Reasoning: "test"},
		},
	}

	jsonPath := filepath.Join(t.TempDir(), "budgets.json")
	data, _ := json.Marshal(input)
	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := WriteBudgets(ctx, repo, jsonPath)
	if err == nil {
		t.Fatal("WriteBudgets() expected error for invalid category ID, got nil")
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
