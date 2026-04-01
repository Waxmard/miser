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
	defer repo.Close()

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

	// Build lookup maps.
	prevMap := make(map[string]float64)
	for i := range previousCats {
		prevMap[previousCats[i].Name] = previousCats[i].TotalAmount
	}
	budgetMap := make(map[string]float64)
	for i := range budgets {
		budgetMap[budgets[i].CategoryName] = budgets[i].MonthlyAmount
	}

	// Header.
	header := lipgloss.NewStyle().Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Fprintf(os.Stdout, "SPENDING TRENDS — %s vs %s\n\n",
		curStart.Format("January 2006"), prevStart.Format("January 2006"))

	fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
		header.Render(pad("CATEGORY", 24)),
		header.Render(padLeft(curStart.Format("Jan 2006"), 12)),
		header.Render(padLeft(prevStart.Format("Jan 2006"), 12)),
		header.Render(padLeft("CHANGE", 10)),
		header.Render(padLeft("BUDGET", 10)),
		header.Render(padLeft("STATUS", 10)),
	)

	for i := range currentCats {
		c := &currentCats[i]
		if c.TransactionCount == 0 {
			continue
		}

		category := truncate(c.Name, 24)
		curAmount := formatAmount(c.TotalAmount)
		prevAmount := prevMap[c.Name]
		prevStr := formatAmount(prevAmount)

		// Change calculation.
		var changeStr string
		if prevAmount != 0 {
			change := ((c.TotalAmount - prevAmount) / math.Abs(prevAmount)) * 100
			if change > 0 {
				changeStr = red.Render(fmt.Sprintf("+%.1f%%", change))
			} else {
				changeStr = green.Render(fmt.Sprintf("%.1f%%", change))
			}
		} else {
			changeStr = dim.Render("—")
		}

		// Budget status.
		budgetStr := dim.Render("—")
		statusStr := dim.Render("—")
		if b, ok := budgetMap[c.Name]; ok {
			budgetStr = formatAmount(b)
			// For expenses (negative amounts), compare absolute values.
			if math.Abs(c.TotalAmount) > b {
				statusStr = red.Render("OVER")
			} else {
				statusStr = green.Render("Under")
			}
		}

		fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
			pad(category, 24),
			padLeft(curAmount, 12),
			dim.Render(padLeft(prevStr, 12)),
			padLeft(changeStr, 10),
			padLeft(budgetStr, 10),
			padLeft(statusStr, 10),
		)
	}

	// Print latest narrative if available.
	report, err := repo.Reports().GetLatest(ctx)
	if err == nil && report != nil {
		fmt.Printf("\n%s\n", report.Narrative)
	}

	return nil
}
