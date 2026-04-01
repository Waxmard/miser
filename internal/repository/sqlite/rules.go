package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Waxmard/miser/internal/repository"
)

type ruleRepo struct {
	db *sql.DB
}

func (r *ruleRepo) Create(ctx context.Context, rule *repository.CategoryRule) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO category_rules (id, pattern, category_id, match_type, priority, hit_count, created_by, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Pattern, rule.CategoryID, rule.MatchType, rule.Priority,
		rule.HitCount, rule.CreatedBy, rule.CreatedAt.Format(timeFormat),
	)
	if err != nil {
		return fmt.Errorf("insert rule: %w", err)
	}
	return nil
}

func (r *ruleRepo) List(ctx context.Context) ([]repository.CategoryRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT cr.id, cr.pattern, cr.category_id, cr.match_type, cr.priority,
		 cr.hit_count, cr.created_by, cr.created_at, COALESCE(c.name, '') AS category_name
		 FROM category_rules cr
		 LEFT JOIN categories c ON cr.category_id = c.id
		 ORDER BY cr.priority DESC, cr.hit_count DESC`)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	defer rows.Close()

	var rules []repository.CategoryRule
	for rows.Next() {
		var rule repository.CategoryRule
		var createdAt string
		if err := rows.Scan(&rule.ID, &rule.Pattern, &rule.CategoryID, &rule.MatchType,
			&rule.Priority, &rule.HitCount, &rule.CreatedBy, &createdAt, &rule.CategoryName); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}
		rule.CreatedAt = parseTime(createdAt)
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *ruleRepo) FindMatch(ctx context.Context, merchant string) (*repository.CategoryRule, error) {
	// First try exact match.
	var rule repository.CategoryRule
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT cr.id, cr.pattern, cr.category_id, cr.match_type, cr.priority,
		 cr.hit_count, cr.created_by, cr.created_at, COALESCE(c.name, '') AS category_name
		 FROM category_rules cr
		 LEFT JOIN categories c ON cr.category_id = c.id
		 WHERE cr.match_type = 'exact' AND LOWER(cr.pattern) = LOWER(?)
		 ORDER BY cr.priority DESC, cr.hit_count DESC LIMIT 1`, merchant,
	).Scan(&rule.ID, &rule.Pattern, &rule.CategoryID, &rule.MatchType,
		&rule.Priority, &rule.HitCount, &rule.CreatedBy, &createdAt, &rule.CategoryName)
	if err == nil {
		rule.CreatedAt = parseTime(createdAt)
		return &rule, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("find exact match: %w", err)
	}

	// Then try contains match.
	rows, err := r.db.QueryContext(ctx,
		`SELECT cr.id, cr.pattern, cr.category_id, cr.match_type, cr.priority,
		 cr.hit_count, cr.created_by, cr.created_at, COALESCE(c.name, '') AS category_name
		 FROM category_rules cr
		 LEFT JOIN categories c ON cr.category_id = c.id
		 WHERE cr.match_type = 'contains'
		 ORDER BY cr.priority DESC, cr.hit_count DESC`)
	if err != nil {
		return nil, fmt.Errorf("query contains rules: %w", err)
	}
	defer rows.Close()

	merchantLower := strings.ToLower(merchant)
	for rows.Next() {
		var r repository.CategoryRule
		var ca string
		if err := rows.Scan(&r.ID, &r.Pattern, &r.CategoryID, &r.MatchType,
			&r.Priority, &r.HitCount, &r.CreatedBy, &ca, &r.CategoryName); err != nil {
			return nil, fmt.Errorf("scan contains rule: %w", err)
		}
		if strings.Contains(merchantLower, strings.ToLower(r.Pattern)) {
			r.CreatedAt = parseTime(ca)
			return &r, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nil, sql.ErrNoRows
}

func (r *ruleRepo) IncrementHitCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE category_rules SET hit_count = hit_count + 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("increment hit count: %w", err)
	}
	return nil
}

func (r *ruleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM category_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	return nil
}
