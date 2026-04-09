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
	defer func() { _ = repo.Close() }()

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

	_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s\n",
		header.Render(pad("CATEGORY", 28)),
		header.Render(padLeft("COUNT", 8)),
		header.Render(padLeft("TOTAL", 12)),
	)

	// Build child map for hierarchy rendering.
	type catRow struct {
		name  string
		count int
		total float64
	}
	childrenOf := make(map[string][]catRow)
	for i := range cats {
		c := &cats[i]
		if c.ParentID != nil {
			childrenOf[*c.ParentID] = append(childrenOf[*c.ParentID], catRow{c.Name, c.TransactionCount, c.TotalAmount})
		}
	}

	printRow := func(name string, count int, total float64, indent string) {
		countStr := fmt.Sprintf("%d", count)
		totalStr := formatAmount(total)
		if count == 0 {
			countStr = dim.Render(countStr)
			totalStr = dim.Render(totalStr)
		}
		label := indent + truncate(name, 28-len(indent))
		_, _ = fmt.Fprintf(os.Stdout, "%s  %s  %s\n",
			pad(label, 28),
			padLeft(countStr, 8),
			padLeft(totalStr, 12),
		)
	}

	for i := range cats {
		c := &cats[i]
		if c.ParentID != nil {
			continue // printed under parent
		}
		printRow(c.Name, c.TransactionCount, c.TotalAmount, "")
		for _, child := range childrenOf[c.ID] {
			printRow(child.name, child.count, child.total, "  ")
		}
	}

	return nil
}
