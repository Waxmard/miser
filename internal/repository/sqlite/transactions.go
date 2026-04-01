package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Waxmard/miser/internal/repository"
)

type transactionRepo struct {
	db *sql.DB
}

func (r *transactionRepo) Create(ctx context.Context, txn *repository.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions
		 (id, account_id, category_id, amount, merchant, merchant_clean, description,
		  original_statement, date, source, source_id, status, categorized_by, confidence,
		  tags, owner, notes, raw_data, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		txn.ID, txn.AccountID, txn.CategoryID, txn.Amount, txn.Merchant, txn.MerchantClean,
		txn.Description, txn.OriginalStatement, txn.Date.Format(timeFormat), txn.Source,
		txn.SourceID, txn.Status, txn.CategorizedBy, txn.Confidence, txn.Tags, txn.Owner,
		txn.Notes, txn.RawData, txn.CreatedAt.Format(timeFormat), txn.UpdatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *transactionRepo) CreateBatch(ctx context.Context, txns []repository.Transaction) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin batch tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO transactions
		 (id, account_id, category_id, amount, merchant, merchant_clean, description,
		  original_statement, date, source, source_id, status, categorized_by, confidence,
		  tags, owner, notes, raw_data, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare batch insert: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for i := range txns {
		t := &txns[i]
		res, err := stmt.ExecContext(ctx,
			t.ID, t.AccountID, t.CategoryID, t.Amount, t.Merchant, t.MerchantClean,
			t.Description, t.OriginalStatement, t.Date.Format(timeFormat), t.Source,
			t.SourceID, t.Status, t.CategorizedBy, t.Confidence, t.Tags, t.Owner,
			t.Notes, t.RawData, t.CreatedAt.Format(timeFormat), t.UpdatedAt.Format(timeFormat),
		)
		if err != nil {
			return inserted, fmt.Errorf("batch insert row %d: %w", i, err)
		}
		n, _ := res.RowsAffected()
		inserted += int(n)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit batch: %w", err)
	}
	return inserted, nil
}

func (r *transactionRepo) GetByID(ctx context.Context, id string) (*repository.Transaction, error) {
	row := r.db.QueryRowContext(ctx, baseTransactionQuery+` WHERE t.id = ?`, id)
	return scanTransaction(row)
}

func (r *transactionRepo) FindBySourceID(ctx context.Context, source, sourceID string) (*repository.Transaction, error) {
	row := r.db.QueryRowContext(ctx, baseTransactionQuery+` WHERE t.source = ? AND t.source_id = ?`, source, sourceID)
	return scanTransaction(row)
}

func (r *transactionRepo) List(ctx context.Context, f *repository.TransactionFilters) ([]repository.Transaction, error) {
	query := baseTransactionQuery
	var args []any
	var conditions []string

	if f.AccountID != nil {
		conditions = append(conditions, "t.account_id = ?")
		args = append(args, *f.AccountID)
	}
	if f.CategoryID != nil {
		conditions = append(conditions, "t.category_id = ?")
		args = append(args, *f.CategoryID)
	}
	if f.Source != nil {
		conditions = append(conditions, "t.source = ?")
		args = append(args, *f.Source)
	}
	if f.From != nil {
		conditions = append(conditions, "t.date >= ?")
		args = append(args, f.From.Format(timeFormat))
	}
	if f.To != nil {
		conditions = append(conditions, "t.date <= ?")
		args = append(args, f.To.Format(timeFormat))
	}
	if f.Merchant != nil {
		conditions = append(conditions, "t.merchant LIKE ?")
		args = append(args, "%"+*f.Merchant+"%")
	}
	if f.Tag != nil {
		conditions = append(conditions, "t.tags LIKE ?")
		args = append(args, "%"+*f.Tag+"%")
	}
	if f.Owner != nil {
		conditions = append(conditions, "t.owner = ?")
		args = append(args, *f.Owner)
	}
	if f.MinAmount != nil {
		conditions = append(conditions, "t.amount >= ?")
		args = append(args, *f.MinAmount)
	}
	if f.MaxAmount != nil {
		conditions = append(conditions, "t.amount <= ?")
		args = append(args, *f.MaxAmount)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.date DESC"

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", f.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txns []repository.Transaction
	for rows.Next() {
		t, err := scanTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, *t)
	}
	return txns, rows.Err()
}

func (r *transactionRepo) Update(ctx context.Context, txn *repository.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE transactions SET account_id = ?, category_id = ?, amount = ?, merchant = ?,
		 merchant_clean = ?, description = ?, original_statement = ?, date = ?, source = ?,
		 source_id = ?, status = ?, categorized_by = ?, confidence = ?, tags = ?, owner = ?,
		 notes = ?, raw_data = ?, updated_at = ? WHERE id = ?`,
		txn.AccountID, txn.CategoryID, txn.Amount, txn.Merchant, txn.MerchantClean,
		txn.Description, txn.OriginalStatement, txn.Date.Format(timeFormat), txn.Source,
		txn.SourceID, txn.Status, txn.CategorizedBy, txn.Confidence, txn.Tags, txn.Owner,
		txn.Notes, txn.RawData, txn.UpdatedAt.Format(timeFormat), txn.ID,
	)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}
	return nil
}

func (r *transactionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM transactions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

func (r *transactionRepo) GetUncategorized(ctx context.Context, limit int) ([]repository.Transaction, error) {
	query := baseTransactionQuery + ` WHERE t.status = 'uncategorized' ORDER BY t.date DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get uncategorized: %w", err)
	}
	defer rows.Close()

	var txns []repository.Transaction
	for rows.Next() {
		t, err := scanTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, *t)
	}
	return txns, rows.Err()
}

func (r *transactionRepo) GetRecentCategorized(ctx context.Context, limit int) ([]repository.Transaction, error) {
	query := baseTransactionQuery + ` WHERE t.status = 'categorized' ORDER BY t.date DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get recent categorized: %w", err)
	}
	defer rows.Close()

	var txns []repository.Transaction
	for rows.Next() {
		t, err := scanTransactionRow(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, *t)
	}
	return txns, rows.Err()
}

const baseTransactionQuery = `SELECT t.id, t.account_id, t.category_id, t.amount, t.merchant,
	t.merchant_clean, t.description, t.original_statement, t.date, t.source, t.source_id,
	t.status, t.categorized_by, t.confidence, t.tags, t.owner, t.notes, t.raw_data,
	t.created_at, t.updated_at,
	COALESCE(c.name, '') AS category_name, COALESCE(a.name, '') AS account_name
	FROM transactions t
	LEFT JOIN categories c ON t.category_id = c.id
	LEFT JOIN accounts a ON t.account_id = a.id`

func scanTransaction(row *sql.Row) (*repository.Transaction, error) {
	var t repository.Transaction
	var date, createdAt, updatedAt string
	err := row.Scan(
		&t.ID, &t.AccountID, &t.CategoryID, &t.Amount, &t.Merchant,
		&t.MerchantClean, &t.Description, &t.OriginalStatement, &date, &t.Source,
		&t.SourceID, &t.Status, &t.CategorizedBy, &t.Confidence, &t.Tags, &t.Owner,
		&t.Notes, &t.RawData, &createdAt, &updatedAt,
		&t.CategoryName, &t.AccountName,
	)
	if err != nil {
		return nil, fmt.Errorf("scan transaction: %w", err)
	}
	t.Date = parseTime(date)
	t.CreatedAt = parseTime(createdAt)
	t.UpdatedAt = parseTime(updatedAt)
	return &t, nil
}

func scanTransactionRow(row rowScanner) (*repository.Transaction, error) {
	var t repository.Transaction
	var date, createdAt, updatedAt string
	err := row.Scan(
		&t.ID, &t.AccountID, &t.CategoryID, &t.Amount, &t.Merchant,
		&t.MerchantClean, &t.Description, &t.OriginalStatement, &date, &t.Source,
		&t.SourceID, &t.Status, &t.CategorizedBy, &t.Confidence, &t.Tags, &t.Owner,
		&t.Notes, &t.RawData, &createdAt, &updatedAt,
		&t.CategoryName, &t.AccountName,
	)
	if err != nil {
		return nil, fmt.Errorf("scan transaction row: %w", err)
	}
	t.Date = parseTime(date)
	t.CreatedAt = parseTime(createdAt)
	t.UpdatedAt = parseTime(updatedAt)
	return &t, nil
}
