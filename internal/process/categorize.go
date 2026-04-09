package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Waxmard/miser/internal/repository"
)

type UncategorizedOutput struct {
	UncategorizedCount int                  `json:"uncategorized_count"`
	Transactions       []UncategorizedTxn   `json:"transactions"`
	Categories         []string             `json:"categories"`
	RecentExamples     []CategorizedExample `json:"recent_examples"`
}

type UncategorizedTxn struct {
	ID          string  `json:"id"`
	Merchant    string  `json:"merchant"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description,omitempty"`
}

type CategorizedExample struct {
	Merchant string `json:"merchant"`
	Category string `json:"category"`
}

// PrintUncategorized writes uncategorized transactions as JSON to w.
func PrintUncategorized(ctx context.Context, repo repository.Repository, w io.Writer) error {
	txns, err := repo.Transactions().GetUncategorized(ctx, 0)
	if err != nil {
		return fmt.Errorf("get uncategorized: %w", err)
	}

	cats, err := repo.Categories().List(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	recent, err := repo.Transactions().GetRecentCategorized(ctx, 20)
	if err != nil {
		return fmt.Errorf("get recent categorized: %w", err)
	}

	out := UncategorizedOutput{
		UncategorizedCount: len(txns),
		Transactions:       []UncategorizedTxn{},
		RecentExamples:     []CategorizedExample{},
	}

	for i := range txns {
		t := &txns[i]
		ut := UncategorizedTxn{
			ID:       t.ID,
			Merchant: t.Merchant,
			Amount:   t.Amount,
			Date:     t.Date.Format("2006-01-02"),
		}
		if t.Description != nil {
			ut.Description = *t.Description
		}
		out.Transactions = append(out.Transactions, ut)
	}

	for i := range cats {
		out.Categories = append(out.Categories, cats[i].Name)
	}

	for i := range recent {
		r := &recent[i]
		out.RecentExamples = append(out.RecentExamples, CategorizedExample{
			Merchant: r.Merchant,
			Category: r.CategoryName,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
