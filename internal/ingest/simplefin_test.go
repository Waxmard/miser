package ingest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSimpleFinAccounts(t *testing.T) { //nolint:paralleltest // uses shared httptest server
	accountSet := sfAccountSet{
		Accounts: []sfAccount{
			{
				ID:       "acct-1",
				Name:     "360 Checking",
				Currency: "USD",
				Balance:  "1500.00",
				Transactions: []sfTransaction{
					{
						ID:          "txn-1",
						Posted:      1700000000,
						Amount:      "-42.50",
						Description: "POS TRADER JOES #123",
					},
					{
						ID:          "txn-2",
						Posted:      1700086400,
						Amount:      "2500.00",
						Description: "ACH PAYROLL ACME CORP",
					},
					{
						ID:          "txn-pending",
						Posted:      0,
						Amount:      "-10.00",
						Description: "PENDING CHARGE",
						Pending:     true,
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("start-date") == "" {
			t.Error("expected start-date query parameter")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accountSet) //nolint:errcheck // test
	}))
	defer server.Close()

	accessURL := server.URL

	result, err := fetchSimpleFinAccounts(t.Context(), accessURL, time.Time{})
	if err != nil {
		t.Fatalf("fetchSimpleFinAccounts() error: %v", err)
	}

	if len(result.Accounts) != 1 {
		t.Fatalf("got %d accounts, want 1", len(result.Accounts))
	}
	if result.Accounts[0].Name != "360 Checking" {
		t.Errorf("account name = %q, want %q", result.Accounts[0].Name, "360 Checking")
	}
	if len(result.Accounts[0].Transactions) != 3 {
		t.Errorf("got %d transactions, want 3", len(result.Accounts[0].Transactions))
	}
}

func TestBuildSimpleFinTransactions(t *testing.T) {
	accounts := []sfAccount{
		{
			ID:   "acct-1",
			Name: "Test Account",
			Transactions: []sfTransaction{
				{
					ID:          "txn-1",
					Posted:      1700000000,
					Amount:      "-42.50",
					Description: "POS TRADER JOES #123",
				},
				{
					ID:          "txn-2",
					Posted:      1700086400,
					Amount:      "2500.00",
					Description: "ACH PAYROLL ACME CORP",
				},
				{
					ID:          "txn-pending",
					Posted:      0,
					Amount:      "-10.00",
					Description: "PENDING",
					Pending:     true,
				},
			},
		},
	}

	accountMap := map[string]string{"acct-1": "miser-acct-id"}
	txns := buildSimpleFinTransactions(accounts, accountMap)

	if len(txns) != 2 {
		t.Fatalf("got %d transactions, want 2 (pending should be skipped)", len(txns))
	}

	// First transaction: expense.
	if txns[0].Amount != -42.50 {
		t.Errorf("txn[0].Amount = %f, want -42.50", txns[0].Amount)
	}
	if txns[0].Merchant != "TRADER JOES #123" {
		t.Errorf("txn[0].Merchant = %q, want %q", txns[0].Merchant, "TRADER JOES #123")
	}
	if txns[0].Source != "simplefin" {
		t.Errorf("txn[0].Source = %q, want %q", txns[0].Source, "simplefin")
	}
	if *txns[0].SourceID != "simplefin_txn-1" {
		t.Errorf("txn[0].SourceID = %q, want %q", *txns[0].SourceID, "simplefin_txn-1")
	}
	if txns[0].AccountID != "miser-acct-id" {
		t.Errorf("txn[0].AccountID = %q, want %q", txns[0].AccountID, "miser-acct-id")
	}

	// Second transaction: income.
	if txns[1].Amount != 2500.00 {
		t.Errorf("txn[1].Amount = %f, want 2500.00", txns[1].Amount)
	}
	if txns[1].Merchant != "PAYROLL ACME CORP" {
		t.Errorf("txn[1].Merchant = %q, want %q", txns[1].Merchant, "PAYROLL ACME CORP")
	}
}

func TestCleanMerchant(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"POS TRADER JOES", "TRADER JOES"},
		{"ACH PAYROLL ACME", "PAYROLL ACME"},
		{"DEBIT CARD PURCHASE", "CARD PURCHASE"},
		{"CREDIT REFUND", "REFUND"},
		{"WALMART STORE 1234", "WALMART STORE 1234"},
		{"  POS STARBUCKS  ", "STARBUCKS"},
	}

	for _, tt := range tests {
		got := cleanMerchant(tt.input)
		if got != tt.want {
			t.Errorf("cleanMerchant(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
