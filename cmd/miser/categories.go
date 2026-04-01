package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List categories with transaction counts",
	RunE:  runCategories,
}

func init() {
	rootCmd.AddCommand(categoriesCmd)
}

func runCategories(cmd *cobra.Command, _ []string) error {
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
	year, month, _ := now.Date()
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	cats, err := repo.Categories().ListWithCounts(ctx, from, to)
	if err != nil {
		return err
	}

	header := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Fprintf(os.Stdout, "%s  %s  %s\n",
		header.Render(pad("CATEGORY", 28)),
		header.Render(padLeft("COUNT", 8)),
		header.Render(padLeft("TOTAL", 12)),
	)

	for i := range cats {
		c := &cats[i]
		countStr := fmt.Sprintf("%d", c.TransactionCount)
		totalStr := formatAmount(c.TotalAmount)
		if c.TransactionCount == 0 {
			countStr = dim.Render(countStr)
			totalStr = dim.Render(totalStr)
		}

		fmt.Fprintf(os.Stdout, "%s  %s  %s\n",
			pad(truncate(c.Name, 28), 28),
			padLeft(countStr, 8),
			padLeft(totalStr, 12),
		)
	}

	return nil
}
