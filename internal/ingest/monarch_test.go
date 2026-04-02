package ingest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
)

const testCSV = `Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags,Owner
2025-03-01,Trader Joe's,Groceries,Cash Management (Individual) (...3002),DEBIT CARD PURCHASE TRADER JOE'S,,"-45.67",,Shared
2025-03-02,Trader Joe's,Groceries,Cash Management (Individual) (...3002),DEBIT CARD PURCHASE TRADER JOE'S,,"-32.10",,Shared
2025-03-03,Trader Joe's,Groceries,Cash Management (Individual) (...3002),DEBIT CARD PURCHASE TRADER JOE'S,,"-28.99",,
2025-03-04,DoorDash,Restaurants,CREDIT CARD (...2621),DOORDASH ORDER,,"-22.50",Subscription,
2025-03-05,DoorDash,Restaurants,CREDIT CARD (...2621),DOORDASH ORDER,,"-18.75",,
2025-03-06,DoorDash,Restaurants,CREDIT CARD (...2621),DOORDASH ORDER,,"-31.20",,
2025-03-07,ChargePoint,Charging,Cash Management (Individual) (...3002),CHARGEPOINT STATION,,"-8.40",,
2025-03-08,ChargePoint,Charging,Cash Management (Individual) (...3002),CHARGEPOINT STATION,,"-12.30",,
2025-03-09,ChargePoint,Charging,Cash Management (Individual) (...3002),CHARGEPOINT STATION,,"-9.10",,
2025-03-10,Random Place,Restaurants,CREDIT CARD (...2621),RANDOM PLACE,,"-15.00",,
2025-03-11,Ignored Txn,Ignored,Cash Management (Individual) (...3002),IGNORED,,"-5.00",,
`

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

func writeTestCSV(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "monarch.csv")
	if err := os.WriteFile(path, []byte(testCSV), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImportMonarch(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	csvPath := writeTestCSV(t)

	result, err := ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		t.Fatalf("ImportMonarch() error: %v", err)
	}

	if result.Transactions != 10 {
		t.Errorf("Transactions = %d, want 10", result.Transactions)
	}
	if result.Accounts != 2 {
		t.Errorf("Accounts = %d, want 2", result.Accounts)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	// Trader Joe's (3 txns, 100% Groceries), DoorDash (3 txns, 100% Restaurants),
	// ChargePoint (3 txns, 100% Charging) = 3 rules.
	// "Random Place" only has 1 txn, so no rule.
	if result.Rules != 3 {
		t.Errorf("Rules = %d, want 3", result.Rules)
	}
}

func TestImportMonarchIdempotent(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	csvPath := writeTestCSV(t)

	result1, err := ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		t.Fatalf("first ImportMonarch() error: %v", err)
	}

	// Second import should insert 0 new transactions (dedup by source_id).
	result2, err := ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		t.Fatalf("second ImportMonarch() error: %v", err)
	}
	if result2.Transactions != 0 {
		t.Errorf("second import Transactions = %d, want 0 (dedup)", result2.Transactions)
	}
	_ = result1
}

func TestImportMonarchCategories(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	csvPath := writeTestCSV(t)

	_, err := ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		t.Fatal(err)
	}

	// Should have created Groceries, Restaurants, Charging categories.
	cats, err := repo.Categories().List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	names := make(map[string]bool)
	for _, c := range cats {
		names[c.Name] = true
	}

	for _, want := range []string{"Groceries", "Restaurants", "Charging"} {
		if !names[want] {
			t.Errorf("missing category %q", want)
		}
	}
}

func TestImportMonarchRuleMatching(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	csvPath := writeTestCSV(t)

	_, err := ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		t.Fatal(err)
	}

	// Rule for "Trader Joe's" should exist and match.
	rule, err := repo.Rules().FindMatch(ctx, "Trader Joe's")
	if err != nil {
		t.Fatalf("FindMatch() error: %v", err)
	}
	if rule.Pattern != "Trader Joe's" {
		t.Errorf("Pattern = %q, want %q", rule.Pattern, "Trader Joe's")
	}
}

const testCSVOldFormat = `Date,Original Date,Account Type,Account Name,Account Number,Institution Name,Name,Custom Name,Amount,Description,Category,Note,Ignored From,Tax Deductible
2023-11-07,2023-11-07,Credit Card,Costco Visa,7686,Citibank,ExxonMobil,,66.92,EXXON T HIJAZI,Auto & Transport,,,
2023-11-15,2023-11-15,Credit Card,Costco Visa,7686,Citibank,Shell,,54.4,SHELL OIL 57543616205,Auto & Transport,,,
2023-11-17,2023-11-17,Credit Card,Quicksilver,2459,Capital One,EZPASSVA,,70,EZPASSVA AUTO REPLENIS,Auto & Transport,,,
2023-11-19,2023-11-19,Credit Card,Costco Visa,7686,Citibank,ExxonMobil,,52.44,EXXON RIGHT FIRST TIME,Auto & Transport,,,
2023-11-25,2023-11-25,Credit Card,Costco Visa,7686,Citibank,ExxonMobil,,48.10,EXXON MAIN STREET,Auto & Transport,,,
2024-01-05,2024-01-05,Credit Card,Costco Visa,7686,Citibank,Ignored Row,,10,SOMETHING,Auto & Transport,,Auto & Transport,
`

func TestImportMonarchOldFormat(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	path := filepath.Join(t.TempDir(), "old-format.csv")
	if err := os.WriteFile(path, []byte(testCSVOldFormat), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportMonarch(ctx, repo, path)
	if err != nil {
		t.Fatalf("ImportMonarch() error: %v", err)
	}

	// 6 data rows, 1 has "Ignored From" populated → 5 imported.
	if result.Transactions != 5 {
		t.Errorf("Transactions = %d, want 5", result.Transactions)
	}
	if result.Accounts != 2 {
		t.Errorf("Accounts = %d, want 2 (Costco Visa, Quicksilver)", result.Accounts)
	}

	// ExxonMobil appears 3 times → should create a rule.
	rules, err := repo.Rules().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i := range rules {
		if rules[i].Pattern == "ExxonMobil" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected rule for ExxonMobil")
	}
}

func TestGuessInstitution(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Cash Management (Individual) (...3002)", "unknown"},
		{"CREDIT CARD (...2621)", "unknown"},
		{"INDIVIDUAL - TOD (...8354)", "unknown"},
		{"Joint WROS (...4087)", "unknown"},
		{"Rollover IRA ...811", "unknown"},
		{"Roth Contributory IRA ...870", "unknown"},
		{"360 Checking (...8855)", "capital_one"},
		{"Quicksilver (...2459)", "capital_one"},
		{"Amazon Store Card (...5689)", "amazon"},
		{"BILT WORLD ELITE MASTERCARD (...6729)", "bilt"},
		{"Verizon Visa Signature Card (...2630)", "verizon"},
		{"Some Random Account", "unknown"},
	}

	for _, tt := range tests {
		got := guessInstitution(tt.name)
		if got != tt.want {
			t.Errorf("guessInstitution(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
