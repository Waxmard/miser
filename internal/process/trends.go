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
	Category      string          `json:"category"`
	Total         float64         `json:"total"`
	Count         int             `json:"count"`
	Subcategories []CategoryTotal `json:"subcategories,omitempty"`
}

// buildHierarchicalTotals organizes a flat CategoryWithCount list into a tree.
// Parent totals are rolled up from their children, and children are nested inside
// the parent's Subcategories field. Categories with zero transactions are omitted.
func buildHierarchicalTotals(cats []repository.CategoryWithCount) []CategoryTotal {
	// Roll up child totals into parents.
	byID := make(map[string]*repository.CategoryWithCount, len(cats))
	for i := range cats {
		byID[cats[i].ID] = &cats[i]
	}
	for i := range cats {
		c := &cats[i]
		if c.ParentID != nil {
			if parent, ok := byID[*c.ParentID]; ok {
				parent.TotalAmount += c.TotalAmount
				parent.TransactionCount += c.TransactionCount
			}
		}
	}

	var result []CategoryTotal
	for i := range cats {
		c := &cats[i]
		if c.ParentID != nil {
			continue // appears nested under parent
		}
		if c.TransactionCount == 0 {
			continue
		}
		ct := CategoryTotal{
			Category: c.Name,
			Total:    c.TotalAmount,
			Count:    c.TransactionCount,
		}
		for j := range cats {
			child := &cats[j]
			if child.ParentID == nil || *child.ParentID != c.ID || child.TransactionCount == 0 {
				continue
			}
			ct.Subcategories = append(ct.Subcategories, CategoryTotal{
				Category: child.Name,
				Total:    child.TotalAmount,
				Count:    child.TransactionCount,
			})
		}
		result = append(result, ct)
	}
	return result
}

// GetTrends returns monthly spending data for the current and previous months.
func GetTrends(ctx context.Context, repo repository.Repository) (*TrendsOutput, error) {
	now := time.Now().UTC()
	curYear, curMonth, _ := now.Date()
	curStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)
	curEnd := curStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	prevStart := curStart.AddDate(0, -1, 0)
	prevEnd := curStart.Add(-time.Second)

	currentCats, err := repo.Categories().ListWithCounts(ctx, curStart, curEnd)
	if err != nil {
		return nil, fmt.Errorf("current month categories: %w", err)
	}

	previousCats, err := repo.Categories().ListWithCounts(ctx, prevStart, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous month categories: %w", err)
	}

	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	out := &TrendsOutput{
		CurrentMonth:  curStart.Format("2006-01"),
		PreviousMonth: prevStart.Format("2006-01"),
	}

	out.Current = buildHierarchicalTotals(currentCats)
	out.Previous = buildHierarchicalTotals(previousCats)

	for i := range budgets {
		out.Budgets = append(out.Budgets, BudgetEntry{
			Category: budgets[i].CategoryName,
			Budget:   budgets[i].MonthlyAmount,
		})
	}

	return out, nil
}

// PrintTrends writes monthly spending data as JSON to w.
func PrintTrends(ctx context.Context, repo repository.Repository, w io.Writer) error {
	out, err := GetTrends(ctx, repo)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
