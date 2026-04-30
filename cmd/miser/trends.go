package main

import (
	"fmt"
	"math"
	"os"

	"github.com/Waxmard/miser/internal/process"
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

	out, err := process.GetTrends(ctx, repo)
	if err != nil {
		return err
	}

	header := lipgloss.NewStyle().Bold(true)
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	_, _ = fmt.Fprintf(os.Stdout, "SPENDING TRENDS — %s vs %s (MTD)\n\n", out.CurrentMonth, out.PreviousMonth)

	_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
		header.Render(pad("CATEGORY", 24)),
		header.Render(padLeft("CURRENT", 12)),
		header.Render(padLeft("PREVIOUS", 12)),
		header.Render(padLeft("CHANGE", 10)),
		header.Render(padLeft("BUDGET", 10)),
		header.Render(padLeft("PACING", 10)),
	)

	printRow := func(c process.CategoryTrend, indent string) {
		var changeStr string
		if c.Previous != 0 {
			if c.DeltaPct > 0 {
				changeStr = green.Render(fmt.Sprintf("+%.1f%%", c.DeltaPct))
			} else {
				changeStr = red.Render(fmt.Sprintf("%.1f%%", c.DeltaPct))
			}
		} else {
			changeStr = dim.Render("—")
		}

		budgetStr := dim.Render("—")
		pacingStr := dim.Render("—")
		if c.Budget > 0 {
			budgetStr = formatAmount(c.Budget)
			switch c.Pacing {
			case "over":
				pacingStr = red.Render("OVER")
			case "ahead":
				pacingStr = red.Render("ahead")
			case "behind":
				pacingStr = green.Render("behind")
			case "on_track":
				pacingStr = green.Render("on track")
			}
		}

		label := indent + truncate(c.Category, 24-len(indent))
		_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s  %s\n",
			pad(label, 24),
			padLeft(formatAmount(c.Current), 12),
			dim.Render(padLeft(formatAmount(c.Previous), 12)),
			padLeft(changeStr, 10),
			padLeft(budgetStr, 10),
			padLeft(pacingStr, 10),
		)
	}

	for i := range out.Categories {
		c := out.Categories[i]
		// Skip parents whose only contribution is rolled-up children with no own activity.
		if math.Abs(c.Current)+math.Abs(c.Previous) == 0 {
			continue
		}
		printRow(c, "")
		for _, child := range c.Subcategories {
			printRow(child, "  ")
		}
	}

	report, err := repo.Reports().GetLatest(ctx)
	if err == nil && report != nil {
		fmt.Printf("\n%s\n", report.Narrative)
	}

	return nil
}
