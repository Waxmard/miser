package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

func normalizeMerchantName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

type merchantIconRepo struct {
	db *sql.DB
}

func (r *merchantIconRepo) List(ctx context.Context) ([]repository.MerchantIcon, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT merchant_name, icon_slug, updated_at FROM merchant_icons ORDER BY merchant_name`)
	if err != nil {
		return nil, fmt.Errorf("list merchant icons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var icons []repository.MerchantIcon
	for rows.Next() {
		var m repository.MerchantIcon
		var updatedAt string
		if err := rows.Scan(&m.MerchantName, &m.IconSlug, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan merchant icon: %w", err)
		}
		m.UpdatedAt = parseTime(updatedAt)
		icons = append(icons, m)
	}
	return icons, rows.Err()
}

func (r *merchantIconRepo) Set(ctx context.Context, m *repository.MerchantIcon) error {
	m.MerchantName = normalizeMerchantName(m.MerchantName)
	m.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO merchant_icons (merchant_name, icon_slug, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(merchant_name) DO UPDATE SET icon_slug = excluded.icon_slug, updated_at = excluded.updated_at`,
		m.MerchantName, m.IconSlug, m.UpdatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("upsert merchant icon: %w", err)
	}
	return nil
}

func (r *merchantIconRepo) Delete(ctx context.Context, merchantName string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM merchant_icons WHERE merchant_name = ?`, normalizeMerchantName(merchantName))
	if err != nil {
		return fmt.Errorf("delete merchant icon: %w", err)
	}
	return nil
}
