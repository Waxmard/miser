package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Waxmard/miser/internal/repository"
)

type PendingReviewOutput struct {
	PendingCount int                `json:"pending_count"`
	Transactions []PendingReviewTxn `json:"transactions"`
	Categories   []CategoryGroup    `json:"categories"`
}

type PendingReviewTxn struct {
	ID            string   `json:"id"`
	Merchant      string   `json:"merchant"`
	MerchantClean string   `json:"merchant_clean,omitempty"`
	Amount        float64  `json:"amount"`
	Date          string   `json:"date"`
	Category      string   `json:"category"`
	Confidence    *float64 `json:"confidence,omitempty"`
	Description   string   `json:"description,omitempty"`
}

// PrintPendingReview writes transactions awaiting review as JSON to w.
func PrintPendingReview(ctx context.Context, repo repository.Repository, w io.Writer) error {
	txns, err := repo.Transactions().GetPendingReview(ctx, 0)
	if err != nil {
		return fmt.Errorf("get pending review: %w", err)
	}

	cats, err := repo.Categories().List(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	out := PendingReviewOutput{
		PendingCount: len(txns),
		Transactions: []PendingReviewTxn{},
		Categories:   buildCategoryGroups(cats),
	}

	for i := range txns {
		t := &txns[i]
		pt := PendingReviewTxn{
			ID:       t.ID,
			Merchant: t.Merchant,
			Amount:   t.Amount,
			Date:     t.Date.Format("2006-01-02"),
			Category: t.CategoryName,
		}
		if t.MerchantClean != nil {
			pt.MerchantClean = *t.MerchantClean
		}
		if t.Confidence != nil {
			pt.Confidence = t.Confidence
		}
		if t.Description != nil {
			pt.Description = *t.Description
		}
		out.Transactions = append(out.Transactions, pt)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
