package repository

import (
	"context"
	"time"
)

type Repository interface {
	Accounts() AccountRepository
	Transactions() TransactionRepository
	Categories() CategoryRepository
	Rules() RuleRepository
	Budgets() BudgetRepository
	SyncState() SyncStateRepository
	RawEmails() RawEmailRepository
	Reports() ReportRepository
	Close() error
	Migrate(ctx context.Context) error
}

type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByName(ctx context.Context, name string) (*Account, error)
	List(ctx context.Context) ([]Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
}

type TransactionRepository interface {
	Create(ctx context.Context, txn *Transaction) error
	CreateBatch(ctx context.Context, txns []Transaction) (int, error)
	GetByID(ctx context.Context, id string) (*Transaction, error)
	FindBySourceID(ctx context.Context, source, sourceID string) (*Transaction, error)
	List(ctx context.Context, f *TransactionFilters) ([]Transaction, error)
	Update(ctx context.Context, txn *Transaction) error
	Delete(ctx context.Context, id string) error
	GetUncategorized(ctx context.Context, limit int) ([]Transaction, error)
	GetRecentCategorized(ctx context.Context, limit int) ([]Transaction, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, cat *Category) error
	GetByID(ctx context.Context, id string) (*Category, error)
	GetByName(ctx context.Context, name string) (*Category, error)
	List(ctx context.Context) ([]Category, error)
	ListWithCounts(ctx context.Context, from, to time.Time) ([]CategoryWithCount, error)
	Update(ctx context.Context, cat *Category) error
	Delete(ctx context.Context, id string) error
	Seed(ctx context.Context, categories []Category) error
}

type RuleRepository interface {
	Create(ctx context.Context, rule *CategoryRule) error
	List(ctx context.Context) ([]CategoryRule, error)
	FindMatch(ctx context.Context, merchant string) (*CategoryRule, error)
	IncrementHitCount(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type BudgetRepository interface {
	Set(ctx context.Context, budget *Budget) error
	GetByCategoryID(ctx context.Context, categoryID string) (*Budget, error)
	List(ctx context.Context) ([]Budget, error)
	Delete(ctx context.Context, id string) error
}

type SyncStateRepository interface {
	Get(ctx context.Context, source string) (*SyncState, error)
	Upsert(ctx context.Context, state *SyncState) error
}

type RawEmailRepository interface {
	Create(ctx context.Context, email *RawEmail) error
	GetPending(ctx context.Context, limit int) ([]RawEmail, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id, reason string) error
}

type ReportRepository interface {
	Create(ctx context.Context, report *Report) error
	GetLatest(ctx context.Context) (*Report, error)
	GetByMonth(ctx context.Context, year, month int) (*Report, error)
}

type TransactionFilters struct {
	AccountID  *string
	CategoryID *string
	Source     *string
	From       *time.Time
	To         *time.Time
	Merchant   *string
	Tag        *string
	Owner      *string
	MinAmount  *float64
	MaxAmount  *float64
	Limit      int
	Offset     int
}
