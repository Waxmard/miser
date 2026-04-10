package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var trendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Monthly spending trends",
	RunE:  runTrends,
}

func init() {
	rootCmd.AddCommand(trendsCmd)
}

func runTrends(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = repo.Close() }()

	now := time.Now().UTC()
	curYear, curMonth, _ := now.Date()
	curStart := time.Date(curYear, curMonth, 1, 0, 0, 0, 0, time.UTC)
	curEnd := curStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	prevStart := curStart.AddDate(0, -1, 0)
	prevEnd := curStart.Add(-time.Second)

	currentCats, err := repo.Categories().ListWithCounts(ctx, curStart, curEnd)
	if err != nil {
		return err
	}

	previousCats, err := repo.Categories().ListWithCounts(ctx, prevStart, prevEnd)
	if err != nil {
		return err
	}

	budgets, err := repo.Budgets().List(ctx)
	if err != nil {
		return err
	}

	// Roll up child totals into parents for current and previous months.
	rollUp := func(cats []repository.CategoryWithCount) {
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
	rollUp(currentCats)
	rollUp(previousCats)

	// Build lookup maps keyed by category name.
	prevMap := make(map[string]float64)
	for i := range previousCats {
		prevMap[previousCats[i].Name] = previousCats[i].TotalAmount
	}
	budgetMap := make(map[string]float64)
	for i := range budgets {
		budgetMap[budgets[i].CategoryName] = budgets[i].MonthlyAmount
	}

	// Build children map for hierarchy rendering.
	childrenOf := make(map[string][]int) // parentID → indices into currentCats
	for i := range currentCats {
		if currentCats[i].ParentID != nil {
			childrenOf[*currentCats[i].ParentID] = append(childrenOf[*currentCats[i].ParentID], i)
		}
	}

	// Header.
	header := lipgloss.NewStyle().Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	_, _ = fmt.Fprintf(os.Stdout, "SPENDING TRENDS — %s vs %s\n\n",
		curStart.Format("January 2006"), prevStart.Format("January 2006"))

	_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
		header.Render(pad("CATEGORY", 24)),
		header.Render(padLeft(curStart.Format("Jan 2006"), 12)),
		header.Render(padLeft(prevStart.Format("Jan 2006"), 12)),
		header.Render(padLeft("CHANGE", 10)),
		header.Render(padLeft("BUDGET", 10)),
		header.Render(padLeft("STATUS", 10)),
	)

	printTrendRow := func(name string, curTotal float64, indent string) {
		prevAmount := prevMap[name]

		var changeStr string
		if prevAmount != 0 {
			change := ((curTotal - prevAmount) / math.Abs(prevAmount)) * 100
			if change > 0 {
				changeStr = red.Render(fmt.Sprintf("+%.1f%%", change))
			} else {
				changeStr = green.Render(fmt.Sprintf("%.1f%%", change))
			}
		} else {
			changeStr = dim.Render("—")
		}

		budgetStr := dim.Render("—")
		statusStr := dim.Render("—")
		if b, ok := budgetMap[name]; ok {
			budgetStr = formatAmount(b)
			if math.Abs(curTotal) > b {
				statusStr = red.Render("OVER")
			} else {
				statusStr = green.Render("Under")
			}
		}

		label := indent + truncate(name, 24-len(indent))
		_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
			pad(label, 24),
			padLeft(formatAmount(curTotal), 12),
			dim.Render(padLeft(formatAmount(prevAmount), 12)),
			padLeft(changeStr, 10),
			padLeft(budgetStr, 10),
			padLeft(statusStr, 10),
		)
	}

	for i := range currentCats {
		c := &currentCats[i]
		if c.ParentID != nil || c.TransactionCount == 0 {
			continue
		}
		printTrendRow(c.Name, c.TotalAmount, "")
		for _, childIdx := range childrenOf[c.ID] {
			child := &currentCats[childIdx]
			if child.TransactionCount == 0 {
				continue
			}
			printTrendRow(child.Name, child.TotalAmount, "  ")
		}
	}

	// Print latest narrative if available.
	report, err := repo.Reports().GetLatest(ctx)
	if err == nil && report != nil {
		fmt.Printf("\n%s\n", report.Narrative)
	}

	return nil
}
