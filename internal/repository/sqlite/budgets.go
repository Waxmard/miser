package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
)

type budgetRepo struct {
	db *sql.DB
}

func (r *budgetRepo) Set(ctx context.Context, b *repository.Budget) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO budgets (id, category_id, monthly_amount, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET monthly_amount = ?, updated_at = ?`,
		b.ID, b.CategoryID, b.MonthlyAmount,
		b.CreatedAt.Format(timeFormat), b.UpdatedAt.Format(timeFormat),
		b.MonthlyAmount, b.UpdatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("set budget: %w", err)
	}
	return nil
}

func (r *budgetRepo) GetByCategoryID(ctx context.Context, categoryID string) (*repository.Budget, error) {
	var b repository.Budget
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT b.id, b.category_id, b.monthly_amount, b.created_at, b.updated_at,
		 COALESCE(c.name, '') AS category_name
		 FROM budgets b LEFT JOIN categories c ON b.category_id = c.id
		 WHERE b.category_id = ?`, categoryID,
	).Scan(&b.ID, &b.CategoryID, &b.MonthlyAmount, &createdAt, &updatedAt, &b.CategoryName)
	if err != nil {
		return nil, fmt.Errorf("get budget by category: %w", err)
	}
	b.CreatedAt = parseTime(createdAt)
	b.UpdatedAt = parseTime(updatedAt)
	return &b, nil
}

func (r *budgetRepo) List(ctx context.Context) ([]repository.Budget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT b.id, b.category_id, b.monthly_amount, b.created_at, b.updated_at,
		 COALESCE(c.name, '') AS category_name
		 FROM budgets b LEFT JOIN categories c ON b.category_id = c.id
		 ORDER BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []repository.Budget
	for rows.Next() {
		var b repository.Budget
		var createdAt, updatedAt string
		if err := rows.Scan(&b.ID, &b.CategoryID, &b.MonthlyAmount, &createdAt, &updatedAt, &b.CategoryName); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		b.CreatedAt = parseTime(createdAt)
		b.UpdatedAt = parseTime(updatedAt)
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

func (r *budgetRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM budgets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
	return nil
}
