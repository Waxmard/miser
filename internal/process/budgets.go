package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"time"

	"github.com/Waxmard/miser/internal/repository"
)

// BudgetDataOutput is the JSON sent to Claude for budget analysis.
type BudgetDataOutput struct {
	GeneratedAt     string               `json:"generated_at"`
	MonthsIncluded  int                  `json:"months_included"`
	Categories      []BudgetCategoryData `json:"categories"`
	ExistingBudgets []BudgetEntry        `json:"existing_budgets"`
}

// BudgetCategoryData holds per-category spending across multiple months.
type BudgetCategoryData struct {
	CategoryID string            `json:"category_id"`
	Category   string            `json:"category"`
	Months     []MonthlySpending `json:"months"`
	Average    float64           `json:"average"`
	Min        float64           `json:"min"`
	Max        float64           `json:"max"`
}

// MonthlySpending is one month's spending for one category.
type MonthlySpending struct {
	Month string  `json:"month"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

const budgetMonths = 6

// PrintBudgetData writes multi-month spending history as JSON to w for budget analysis.
func PrintBudgetData(ctx context.Context, repo repository.Repository, w io.Writer) error {
	now := time.Now().UTC()
	curYear, curMonth, _ := now.Date()
	currentMonthStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)

	// Build month boundaries for 6 completed months (excluding current partial month).
	type monthRange struct {
		label string
		start time.Time
		end   time.Time
	}
	months := make([]monthRange, budgetMonths)
	for i := range budgetMonths {
		start := currentMonthStart.AddDate(0, -(i + 1), 0)
		end := start.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		months[budgetMonths-1-i] = monthRange{
			label: start.Format("2006-01"),
			start: start,
			end:   end,
		}
	}

	// Gather spending per category per month.
	type catKey struct {
		id   string
		name string
	}
	catMonths := make(map[catKey][]MonthlySpending)

	for _, m := range months {
		cats, err := repo.Categories().ListWithCounts(ctx, m.start, m.end)
		if err != nil {
			return fmt.Errorf("categories for %s: %w", m.label, err)
		}
		for i := range cats {
			c := &cats[i]
			key := catKey{id: c.ID, name: c.Name}
			catMonths[key] = append(catMonths[key], MonthlySpending{
				Month: m.label,
				Total: c.TotalAmount,
				Count: c.TransactionCount,
			})
		}
	}

	// Build output, only including categories with at least one transaction.
	var categories []BudgetCategoryData
	for key, spending := range catMonths {
		hasActivity := false
		for _, ms := range spending {
			if ms.Count > 0 {
				hasActivity = true
				break
			}
		}
		if !hasActivity {
			continue
		}

		var sum float64
		var count int
		minVal := math.Inf(1)
		maxVal := math.Inf(-1)
		for _, ms := range spending {
			if ms.Count > 0 {
				sum += ms.Total
				count++
				minVal = math.Min(minVal, ms.Total)
				maxVal = math.Max(maxVal, ms.Total)
			}
		}

		avg := 0.0
		if count > 0 {
			avg = sum / float64(count)
		}
		if math.IsInf(minVal, 0) {
			minVal = 0
		}
		if math.IsInf(maxVal, 0) {
			maxVal = 0
		}

		categories = append(categories, BudgetCategoryData{
			CategoryID: key.id,
			Category:   key.name,
			Months:     spending,
			Average:    math.Round(avg*100) / 100,
			Min:        minVal,
			Max:        maxVal,
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Category < categories[j].Category
	})

	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		return fmt.Errorf("list budgets: %w", err)
	}

	var existing []BudgetEntry
	for i := range budgets {
		existing = append(existing, BudgetEntry{
			Category: budgets[i].CategoryName,
			Budget:   budgets[i].MonthlyAmount,
		})
	}

	out := BudgetDataOutput{
		GeneratedAt:     now.Format(time.RFC3339),
		MonthsIncluded:  budgetMonths,
		Categories:      categories,
		ExistingBudgets: existing,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
