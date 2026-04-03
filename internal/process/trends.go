package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

type TrendsOutput struct {
	CurrentMonth  string          `json:"current_month"`
	PreviousMonth string          `json:"previous_month"`
	Current       []CategoryTotal `json:"current"`
	Previous      []CategoryTotal `json:"previous"`
	Budgets       []BudgetEntry   `json:"budgets"`
}

type CategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// PrintTrends writes monthly spending data as JSON to w.
func PrintTrends(ctx context.Context, repo repository.Repository, w io.Writer) error {
	now := time.Now().UTC()
	curYear, curMonth, _ := now.Date()
	curStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)
	curEnd := curStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	prevStart := curStart.AddDate(0, -1, 0)
	prevEnd := curStart.Add(-time.Second)

	currentCats, err := repo.Categories().ListWithCounts(ctx, curStart, curEnd)
	if err != nil {
		return fmt.Errorf("current month categories: %w", err)
	}

	previousCats, err := repo.Categories().ListWithCounts(ctx, prevStart, prevEnd)
	if err != nil {
		return fmt.Errorf("previous month categories: %w", err)
	}

	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		return fmt.Errorf("list budgets: %w", err)
	}

	out := TrendsOutput{
		CurrentMonth:  curStart.Format("2006-01"),
		PreviousMonth: prevStart.Format("2006-01"),
	}

	for i := range currentCats {
		c := &currentCats[i]
		if c.TransactionCount > 0 {
			out.Current = append(out.Current, CategoryTotal{
				Category: c.Name,
				Total:    c.TotalAmount,
				Count:    c.TransactionCount,
			})
		}
	}

	for i := range previousCats {
		c := &previousCats[i]
		if c.TransactionCount > 0 {
			out.Previous = append(out.Previous, CategoryTotal{
				Category: c.Name,
				Total:    c.TotalAmount,
				Count:    c.TransactionCount,
			})
		}
	}

	for i := range budgets {
		out.Budgets = append(out.Budgets, BudgetEntry{
			Category: budgets[i].CategoryName,
			Budget:   budgets[i].MonthlyAmount,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
