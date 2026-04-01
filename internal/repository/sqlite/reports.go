package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
)

type reportRepo struct {
	db *sql.DB
}

func (r *reportRepo) Create(ctx context.Context, report *repository.Report) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO reports (id, year, month, narrative, data, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		report.ID, report.Year, report.Month, report.Narrative, report.Data,
		report.CreatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	return nil
}

func (r *reportRepo) GetLatest(ctx context.Context) (*repository.Report, error) {
	var rpt repository.Report
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, year, month, narrative, data, created_at FROM reports ORDER BY year DESC, month DESC LIMIT 1`,
	).Scan(&rpt.ID, &rpt.Year, &rpt.Month, &rpt.Narrative, &rpt.Data, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get latest report: %w", err)
	}
	rpt.CreatedAt = parseTime(createdAt)
	return &rpt, nil
}

func (r *reportRepo) GetByMonth(ctx context.Context, year, month int) (*repository.Report, error) {
	var rpt repository.Report
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, year, month, narrative, data, created_at FROM reports WHERE year = ? AND month = ?`,
		year, month,
	).Scan(&rpt.ID, &rpt.Year, &rpt.Month, &rpt.Narrative, &rpt.Data, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get report by month: %w", err)
	}
	rpt.CreatedAt = parseTime(createdAt)
	return &rpt, nil
}
