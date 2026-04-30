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

type TrendsOutput struct {
	CurrentMonth  string            `json:"current_month"`
	PreviousMonth string            `json:"previous_month"`
	MonthProgress float64           `json:"month_progress"`
	Categories    []CategoryTrend   `json:"categories"`
	TopMovers     []CategoryTrend   `json:"top_movers"`
	Anomalies     []TransactionFlag `json:"anomalies,omitempty"`
	Budgets       []BudgetEntry     `json:"budgets"`
}

type CategoryTrend struct {
	Category      string          `json:"category"`
	Current       float64         `json:"current"`
	Previous      float64         `json:"previous"`
	DeltaAbs      float64         `json:"delta_abs"`
	DeltaPct      float64         `json:"delta_pct,omitempty"`
	Budget        float64         `json:"budget,omitempty"`
	BudgetUsedPct float64         `json:"budget_used_pct,omitempty"`
	Pacing        string          `json:"pacing,omitempty"`
	TxnCount      int             `json:"txn_count"`
	Subcategories []CategoryTrend `json:"subcategories,omitempty"`
}

type TransactionFlag struct {
	TxnID    string  `json:"transaction_id"`
	Date     string  `json:"date"`
	Merchant string  `json:"merchant"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Reason   string  `json:"reason"`
}

const anomalyHistoryMonths = 6
const anomalyRatio = 3.0
const anomalyMaxResults = 5
const topMoverCount = 5

// mtdEnd returns the end-of-day timestamp for monthStart's month at the given
// day-of-month, clamped to the month's last day.
func mtdEnd(monthStart time.Time, day int) time.Time {
	lastDay := monthStart.AddDate(0, 1, -1).Day()
	if day > lastDay {
		day = lastDay
	}
	y, m, _ := monthStart.Date()
	return time.Date(y, m, day, 23, 59, 59, 0, time.UTC)
}

// rollUpChildren adds each child's TotalAmount and TransactionCount into its parent.
func rollUpChildren(cats []repository.CategoryWithCount) {
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
}

// classifyPacing returns a pacing label given budget usage and month progress.
// Both inputs are 0..1 ratios. Returns "" when no budget is set.
func classifyPacing(budgetUsed, monthProgress float64) string {
	if budgetUsed > 1.0 {
		return "over"
	}
	switch {
	case budgetUsed > monthProgress+0.10:
		return "ahead"
	case budgetUsed < monthProgress-0.10:
		return "behind"
	default:
		return "on_track"
	}
}

// buildTrend constructs a CategoryTrend for one category from current+previous totals + optional budget.
func buildTrend(name string, current, previous, budget, monthProgress float64, txnCount int) CategoryTrend {
	t := CategoryTrend{
		Category: name,
		Current:  current,
		Previous: previous,
		DeltaAbs: current - previous,
		TxnCount: txnCount,
	}
	if previous != 0 {
		t.DeltaPct = math.Round(((current-previous)/math.Abs(previous))*1000) / 10
	}
	if budget > 0 {
		t.Budget = budget
		used := math.Abs(current) / budget
		t.BudgetUsedPct = math.Round(used*1000) / 10
		t.Pacing = classifyPacing(used, monthProgress)
	}
	return t
}

// GetTrends returns hierarchical category trends for the current month vs the
// previous month clamped to today's day-of-month. Pacing, deltas, top movers,
// and anomalies are pre-computed so consumers don't need to re-derive them.
func GetTrends(ctx context.Context, repo repository.Repository) (*TrendsOutput, error) {
	now := time.Now().UTC()
	curYear, curMonth, curDay := now.Date()
	curStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)
	curEnd := curStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	daysInMonth := curStart.AddDate(0, 1, -1).Day()
	monthProgress := float64(curDay) / float64(daysInMonth)

	prevStart := curStart.AddDate(0, -1, 0)
	prevEnd := mtdEnd(prevStart, curDay)

	currentCats, err := repo.Categories().ListWithCounts(ctx, curStart, curEnd)
	if err != nil {
		return nil, fmt.Errorf("current month categories: %w", err)
	}
	previousCats, err := repo.Categories().ListWithCounts(ctx, prevStart, prevEnd)
	if err != nil {
		return nil, fmt.Errorf("previous month categories: %w", err)
	}
	rollUpChildren(currentCats)
	rollUpChildren(previousCats)

	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	budgetByName := make(map[string]float64, len(budgets))
	for i := range budgets {
		budgetByName[budgets[i].CategoryName] = budgets[i].MonthlyAmount
	}

	prevByName := make(map[string]float64, len(previousCats))
	for i := range previousCats {
		prevByName[previousCats[i].Name] = previousCats[i].TotalAmount
	}

	// Build hierarchical Categories. Skip categories with no activity in either month.
	childrenOf := make(map[string][]int, len(currentCats))
	for i := range currentCats {
		if currentCats[i].ParentID != nil {
			childrenOf[*currentCats[i].ParentID] = append(childrenOf[*currentCats[i].ParentID], i)
		}
	}

	var categories []CategoryTrend
	var leaves []CategoryTrend
	for i := range currentCats {
		c := &currentCats[i]
		if c.ParentID != nil {
			continue
		}
		prev := prevByName[c.Name]
		if c.TransactionCount == 0 && prev == 0 {
			continue
		}
		trend := buildTrend(c.Name, c.TotalAmount, prev, budgetByName[c.Name], monthProgress, c.TransactionCount)
		for _, childIdx := range childrenOf[c.ID] {
			child := &currentCats[childIdx]
			childPrev := prevByName[child.Name]
			if child.TransactionCount == 0 && childPrev == 0 {
				continue
			}
			ct := buildTrend(child.Name, child.TotalAmount, childPrev, budgetByName[child.Name], monthProgress, child.TransactionCount)
			trend.Subcategories = append(trend.Subcategories, ct)
			leaves = append(leaves, ct)
		}
		// If parent has no children with activity, treat parent itself as leaf for top-movers.
		if len(trend.Subcategories) == 0 {
			leaves = append(leaves, trend)
		}
		categories = append(categories, trend)
	}

	// Top movers: leaf-level, sorted by |delta_abs| desc.
	movers := make([]CategoryTrend, len(leaves))
	copy(movers, leaves)
	sort.Slice(movers, func(i, j int) bool {
		return math.Abs(movers[i].DeltaAbs) > math.Abs(movers[j].DeltaAbs)
	})
	if len(movers) > topMoverCount {
		movers = movers[:topMoverCount]
	}

	anomalies, err := detectAnomalies(ctx, repo, curStart, curEnd)
	if err != nil {
		return nil, fmt.Errorf("detect anomalies: %w", err)
	}

	out := &TrendsOutput{
		CurrentMonth:  curStart.Format("2006-01"),
		PreviousMonth: prevStart.Format("2006-01"),
		MonthProgress: math.Round(monthProgress*1000) / 1000,
		Categories:    categories,
		TopMovers:     movers,
		Anomalies:     anomalies,
	}
	for i := range budgets {
		out.Budgets = append(out.Budgets, BudgetEntry{
			Category: budgets[i].CategoryName,
			Budget:   budgets[i].MonthlyAmount,
		})
	}
	return out, nil
}

// detectAnomalies flags current-month transactions whose absolute amount exceeds
// anomalyRatio× the median absolute amount for that category over the prior
// anomalyHistoryMonths. Returns up to anomalyMaxResults flags ordered by ratio.
func detectAnomalies(ctx context.Context, repo repository.Repository, curStart, curEnd time.Time) ([]TransactionFlag, error) {
	historyStart := curStart.AddDate(0, -anomalyHistoryMonths, 0)
	historyEnd := curStart.Add(-time.Second)

	historyTxns, err := repo.Transactions().List(ctx, &repository.TransactionFilters{
		From: &historyStart,
		To:   &historyEnd,
	})
	if err != nil {
		return nil, err
	}
	currentTxns, err := repo.Transactions().List(ctx, &repository.TransactionFilters{
		From: &curStart,
		To:   &curEnd,
	})
	if err != nil {
		return nil, err
	}

	// Build per-category absolute-amount samples from history.
	byCategory := make(map[string][]float64)
	for i := range historyTxns {
		t := &historyTxns[i]
		if t.Amount >= 0 {
			continue // ignore income
		}
		byCategory[t.CategoryName] = append(byCategory[t.CategoryName], math.Abs(t.Amount))
	}
	medianByCategory := make(map[string]float64, len(byCategory))
	for name, samples := range byCategory {
		medianByCategory[name] = median(samples)
	}

	type scored struct {
		flag  TransactionFlag
		ratio float64
	}
	var found []scored
	for i := range currentTxns {
		t := &currentTxns[i]
		if t.Amount >= 0 {
			continue
		}
		med := medianByCategory[t.CategoryName]
		if med <= 0 {
			continue
		}
		amt := math.Abs(t.Amount)
		ratio := amt / med
		if ratio < anomalyRatio {
			continue
		}
		merchant := t.Merchant
		if t.MerchantClean != nil && *t.MerchantClean != "" {
			merchant = *t.MerchantClean
		}
		found = append(found, scored{
			flag: TransactionFlag{
				TxnID:    t.ID,
				Date:     t.Date.Format("2006-01-02"),
				Merchant: merchant,
				Category: t.CategoryName,
				Amount:   t.Amount,
				Reason:   fmt.Sprintf("%.1fx category median ($%.0f)", ratio, med),
			},
			ratio: ratio,
		})
	}

	sort.Slice(found, func(i, j int) bool { return found[i].ratio > found[j].ratio })
	if len(found) > anomalyMaxResults {
		found = found[:anomalyMaxResults]
	}
	out := make([]TransactionFlag, len(found))
	for i := range found {
		out[i] = found[i].flag
	}
	return out, nil
}

// median returns the median of a slice of floats. Mutates the slice (sorts it).
// Returns 0 for an empty slice.
func median(xs []float64) float64 {
	n := len(xs)
	if n == 0 {
		return 0
	}
	sort.Float64s(xs)
	if n%2 == 1 {
		return xs[n/2]
	}
	return (xs[n/2-1] + xs[n/2]) / 2
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
