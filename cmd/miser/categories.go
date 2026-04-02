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
	categoriesCmd.Flags().String("from", "", "Start date (YYYY-MM-DD), default: all time")
	categoriesCmd.Flags().String("to", "", "End date (YYYY-MM-DD), default: all time")
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

	var from, to time.Time
	if f, _ := cmd.Flags().GetString("from"); f != "" {
		from, _ = time.Parse("2006-01-02", f)
	}
	if t, _ := cmd.Flags().GetString("to"); t != "" {
		to, _ = time.Parse("2006-01-02", t)
	}

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
