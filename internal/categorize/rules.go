package categorize

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

// CategorizeResult holds the summary of a rule-based categorization run.
type CategorizeResult struct {
	Checked     int
	Categorized int
}

// RunRules checks all uncategorized transactions against category rules.
// Matched transactions are updated with the rule's category.
func RunRules(ctx context.Context, repo repository.Repository) (*CategorizeResult, error) {
	txns, err := repo.Transactions().GetUncategorized(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("get uncategorized: %w", err)
	}

	result := &CategorizeResult{Checked: len(txns)}
	now := time.Now().UTC()

	for i := range txns {
		t := &txns[i]

		rule, err := repo.Rules().FindMatch(ctx, t.Merchant)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return result, fmt.Errorf("find match for %q: %w", t.Merchant, err)
		}

		t.CategoryID = &rule.CategoryID
		t.Status = "categorized"
		categorizedBy := "rule"
		t.CategorizedBy = &categorizedBy
		confidence := 1.0
		t.Confidence = &confidence
		t.UpdatedAt = now

		if err := repo.Transactions().Update(ctx, t); err != nil {
			return result, fmt.Errorf("update transaction %s: %w", t.ID, err)
		}

		if err := repo.Rules().IncrementHitCount(ctx, rule.ID); err != nil {
			return result, fmt.Errorf("increment hit count %s: %w", rule.ID, err)
		}

		result.Categorized++
	}

	return result, nil
}
