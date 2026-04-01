package main

import (
	"fmt"
	"os"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List categorization rules",
	RunE:  runRules,
}

func init() {
	rootCmd.AddCommand(rulesCmd)
}

func runRules(cmd *cobra.Command, _ []string) error {
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

	rules, err := repo.Rules().List(ctx)
	if err != nil {
		return err
	}

	if len(rules) == 0 {
		fmt.Println("No rules defined")
		return nil
	}

	header := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
		header.Render(pad("PATTERN", 30)),
		header.Render(pad("CATEGORY", 24)),
		header.Render(padLeft("HITS", 6)),
		header.Render(pad("SOURCE", 16)),
	)

	for i := range rules {
		r := &rules[i]
		fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
			pad(truncate(r.Pattern, 30), 30),
			pad(truncate(r.CategoryName, 24), 24),
			padLeft(fmt.Sprintf("%d", r.HitCount), 6),
			dim.Render(pad(r.CreatedBy, 16)),
		)
	}

	return nil
}
