package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
	_ "modernc.org/sqlite"
)

func init() {
	repository.Register("sqlite", New)
}

type DB struct {
	db *sql.DB
}

func New(dsn string) (repository.Repository, error) {
	db, err := sql.Open("sqlite", dsn+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Accounts() repository.AccountRepository {
	return &accountRepo{db: d.db}
}

func (d *DB) Transactions() repository.TransactionRepository {
	return &transactionRepo{db: d.db}
}

func (d *DB) Categories() repository.CategoryRepository {
	return &categoryRepo{db: d.db}
}

func (d *DB) Rules() repository.RuleRepository {
	return &ruleRepo{db: d.db}
}

func (d *DB) Budgets() repository.BudgetRepository {
	return &budgetRepo{db: d.db}
}

func (d *DB) SyncState() repository.SyncStateRepository {
	return &syncStateRepo{db: d.db}
}

func (d *DB) RawEmails() repository.RawEmailRepository {
	return &rawEmailRepo{db: d.db}
}

func (d *DB) Reports() repository.ReportRepository {
	return &reportRepo{db: d.db}
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Migrate(ctx context.Context) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Create schema_version table if it doesn't exist.
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	var version int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return fmt.Errorf("get schema version: %w", err)
	}

	if version < 1 {
		if err := migrateV1(ctx, tx); err != nil {
			return fmt.Errorf("migrate v1: %w", err)
		}
	}

	return tx.Commit()
}

func migrateV1(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE accounts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			institution TEXT NOT NULL,
			account_type TEXT NOT NULL,
			source TEXT NOT NULL,
			plaid_account_id TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,

		`CREATE TABLE categories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			parent_id TEXT,
			icon TEXT,
			created_at TEXT NOT NULL
		)`,

		`CREATE TABLE transactions (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL REFERENCES accounts(id),
			category_id TEXT REFERENCES categories(id),
			amount REAL NOT NULL,
			merchant TEXT NOT NULL,
			merchant_clean TEXT,
			description TEXT,
			original_statement TEXT,
			date TEXT NOT NULL,
			source TEXT NOT NULL,
			source_id TEXT,
			status TEXT NOT NULL DEFAULT 'uncategorized',
			categorized_by TEXT,
			confidence REAL,
			tags TEXT,
			owner TEXT,
			notes TEXT,
			raw_data TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(source, source_id)
		)`,

		`CREATE INDEX idx_transactions_date ON transactions(date)`,
		`CREATE INDEX idx_transactions_account ON transactions(account_id)`,
		`CREATE INDEX idx_transactions_category ON transactions(category_id)`,
		`CREATE INDEX idx_transactions_source ON transactions(source, source_id)`,
		`CREATE INDEX idx_transactions_status ON transactions(status)`,
		`CREATE INDEX idx_transactions_tags ON transactions(tags)`,
		`CREATE INDEX idx_transactions_owner ON transactions(owner)`,

		`CREATE TABLE raw_emails (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL UNIQUE,
			subject TEXT NOT NULL,
			sender TEXT NOT NULL,
			body TEXT NOT NULL,
			received_at TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			error TEXT,
			created_at TEXT NOT NULL
		)`,

		`CREATE INDEX idx_raw_emails_status ON raw_emails(status)`,

		`CREATE TABLE category_rules (
			id TEXT PRIMARY KEY,
			pattern TEXT NOT NULL,
			category_id TEXT NOT NULL REFERENCES categories(id),
			match_type TEXT NOT NULL DEFAULT 'contains',
			priority INTEGER NOT NULL DEFAULT 0,
			hit_count INTEGER NOT NULL DEFAULT 0,
			created_by TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,

		`CREATE TABLE budgets (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL REFERENCES categories(id),
			monthly_amount REAL NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,

		`CREATE TABLE sync_state (
			source TEXT PRIMARY KEY,
			last_sync_at TEXT NOT NULL,
			last_message_uid INTEGER,
			metadata TEXT
		)`,

		`CREATE TABLE reports (
			id TEXT PRIMARY KEY,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			narrative TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(year, month)
		)`,

		`INSERT INTO schema_version (version) VALUES (1)`,
	}

	for _, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:40], err)
		}
	}

	return nil
}
