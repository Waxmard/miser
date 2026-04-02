package repository

import "time"

type Account struct {
	ID          string
	Name        string
	Institution string  // "fidelity", "capital_one", "chase", etc.
	AccountType string  // "checking", "savings", "credit", "investment", "brokerage"
	Source      string  // "email", "simplefin", "csv", "monarch_import"
	ExternalID  *string // provider-specific account ID (DB column: plaid_account_id)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Transaction struct {
	ID                string
	AccountID         string
	CategoryID        *string
	Amount            float64 // negative = expense, positive = income
	Merchant          string  // raw merchant string
	MerchantClean     *string // cleaned/normalized name
	Description       *string
	OriginalStatement *string // raw bank statement text
	Date              time.Time
	Source            string  // "email", "simplefin", "csv", "monarch_import"
	SourceID          *string // dedup key
	Status            string  // "uncategorized", "categorized"
	CategorizedBy     *string // "claude", "rule", "manual", "monarch_import"
	Confidence        *float64
	Tags              *string // comma-separated: "Subscription,Retail Sync"
	Owner             *string // "Shared", or a person's name
	Notes             *string
	RawData           *string // original email body or CSV row as JSON
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Joined fields (populated by queries with JOINs)
	CategoryName string
	AccountName  string
}

type RawEmail struct {
	ID         string
	MessageID  string
	Subject    string
	From       string
	Body       string
	ReceivedAt time.Time
	Status     string // "pending", "processed", "failed"
	Error      *string
	CreatedAt  time.Time
}

type Category struct {
	ID        string
	Name      string
	ParentID  *string
	Icon      *string
	CreatedAt time.Time
}

type CategoryWithCount struct {
	Category
	TransactionCount int
	TotalAmount      float64
}

type CategoryRule struct {
	ID           string
	Pattern      string
	CategoryID   string
	MatchType    string // "contains", "exact", "regex"
	Priority     int
	HitCount     int
	CreatedBy    string // "claude", "manual", "monarch_import"
	CreatedAt    time.Time
	CategoryName string // joined field
}

type Budget struct {
	ID            string
	CategoryID    string
	MonthlyAmount float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CategoryName  string // joined field
}

type SyncState struct {
	Source         string
	LastSyncAt     time.Time
	LastMessageUID *uint32
	Metadata       *string
}

type Report struct {
	ID        string
	Year      int
	Month     int
	Narrative string
	Data      string
	CreatedAt time.Time
}
