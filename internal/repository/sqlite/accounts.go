package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
)

type accountRepo struct {
	db *sql.DB
}

func (r *accountRepo) Create(ctx context.Context, a *repository.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, name, institution, account_type, source, plaid_account_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Name, a.Institution, a.AccountType, a.Source, a.ExternalID,
		a.CreatedAt.Format(timeFormat), a.UpdatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *accountRepo) GetByID(ctx context.Context, id string) (*repository.Account, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, institution, account_type, source, plaid_account_id, created_at, updated_at
		 FROM accounts WHERE id = ?`, id)
	return scanAccount(row)
}

func (r *accountRepo) GetByName(ctx context.Context, name string) (*repository.Account, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, institution, account_type, source, plaid_account_id, created_at, updated_at
		 FROM accounts WHERE name = ?`, name)
	return scanAccount(row)
}

func (r *accountRepo) List(ctx context.Context) ([]repository.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, institution, account_type, source, plaid_account_id, created_at, updated_at
		 FROM accounts ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var accounts []repository.Account
	for rows.Next() {
		a, err := scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *a)
	}
	return accounts, rows.Err()
}

func (r *accountRepo) Update(ctx context.Context, a *repository.Account) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET name = ?, institution = ?, account_type = ?, source = ?,
		 plaid_account_id = ?, updated_at = ? WHERE id = ?`,
		a.Name, a.Institution, a.AccountType, a.Source, a.ExternalID,
		a.UpdatedAt.Format(timeFormat), a.ID,
	)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	return nil
}

func (r *accountRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

func scanAccount(row *sql.Row) (*repository.Account, error) {
	var a repository.Account
	var createdAt, updatedAt string
	err := row.Scan(&a.ID, &a.Name, &a.Institution, &a.AccountType, &a.Source,
		&a.ExternalID, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan account: %w", err)
	}
	a.CreatedAt = parseTime(createdAt)
	a.UpdatedAt = parseTime(updatedAt)
	return &a, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccountRow(row rowScanner) (*repository.Account, error) {
	var a repository.Account
	var createdAt, updatedAt string
	err := row.Scan(&a.ID, &a.Name, &a.Institution, &a.AccountType, &a.Source,
		&a.ExternalID, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan account row: %w", err)
	}
	a.CreatedAt = parseTime(createdAt)
	a.UpdatedAt = parseTime(updatedAt)
	return &a, nil
}
