package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

type categoryRepo struct {
	db *sql.DB
}

func (r *categoryRepo) Create(ctx context.Context, cat *repository.Category) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (id, name, parent_id, icon, created_at) VALUES (?, ?, ?, ?, ?)`,
		cat.ID, cat.Name, cat.ParentID, cat.Icon, cat.CreatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id string) (*repository.Category, error) {
	var cat repository.Category
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, parent_id, icon, created_at FROM categories WHERE id = ?`, id,
	).Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.Icon, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get category by id: %w", err)
	}
	cat.CreatedAt = parseTime(createdAt)
	return &cat, nil
}

func (r *categoryRepo) GetByName(ctx context.Context, name string) (*repository.Category, error) {
	var cat repository.Category
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, parent_id, icon, created_at FROM categories WHERE name = ?`, name,
	).Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.Icon, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get category by name: %w", err)
	}
	cat.CreatedAt = parseTime(createdAt)
	return &cat, nil
}

func (r *categoryRepo) List(ctx context.Context) ([]repository.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, parent_id, icon, created_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cats []repository.Category
	for rows.Next() {
		var cat repository.Category
		var createdAt string
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.Icon, &createdAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cat.CreatedAt = parseTime(createdAt)
		cats = append(cats, cat)
	}
	return cats, rows.Err()
}

func (r *categoryRepo) ListWithCounts(ctx context.Context, from, to time.Time) ([]repository.CategoryWithCount, error) {
	var rows *sql.Rows
	var err error

	if from.IsZero() && to.IsZero() {
		rows, err = r.db.QueryContext(ctx,
			`SELECT c.id, c.name, c.parent_id, c.icon, c.created_at,
			 COUNT(t.id) AS txn_count, COALESCE(SUM(t.amount), 0) AS total_amount
			 FROM categories c
			 LEFT JOIN transactions t ON c.id = t.category_id
			 GROUP BY c.id
			 ORDER BY total_amount ASC`)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT c.id, c.name, c.parent_id, c.icon, c.created_at,
			 COUNT(t.id) AS txn_count, COALESCE(SUM(t.amount), 0) AS total_amount
			 FROM categories c
			 LEFT JOIN transactions t ON c.id = t.category_id AND t.date >= ? AND t.date <= ?
			 GROUP BY c.id
			 ORDER BY total_amount ASC`,
			from.Format(timeFormat), to.Format(timeFormat),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list categories with counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cats []repository.CategoryWithCount
	for rows.Next() {
		var cwc repository.CategoryWithCount
		var createdAt string
		if err := rows.Scan(&cwc.ID, &cwc.Name, &cwc.ParentID, &cwc.Icon, &createdAt,
			&cwc.TransactionCount, &cwc.TotalAmount); err != nil {
			return nil, fmt.Errorf("scan category with count: %w", err)
		}
		cwc.CreatedAt = parseTime(createdAt)
		cats = append(cats, cwc)
	}
	return cats, rows.Err()
}

func (r *categoryRepo) Update(ctx context.Context, cat *repository.Category) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name = ?, parent_id = ?, icon = ? WHERE id = ?`,
		cat.Name, cat.ParentID, cat.Icon, cat.ID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (r *categoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (r *categoryRepo) Seed(ctx context.Context, categories []repository.Category) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO categories (id, name, parent_id, icon, created_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare seed: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for i := range categories {
		c := &categories[i]
		if _, err := stmt.ExecContext(ctx, c.ID, c.Name, c.ParentID, c.Icon, c.CreatedAt.Format(timeFormat)); err != nil {
			return fmt.Errorf("seed category %q: %w", c.Name, err)
		}
	}

	return tx.Commit()
}
