package main

import (
	"fmt"
	"os"

	"github.com/Waxmard/miser/internal/process"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Print data as JSON for Claude Code cron jobs",
}

var processEmailsCmd = &cobra.Command{
	Use:   "emails",
	Short: "Print pending raw emails as JSON",
	RunE:  runProcessEmails,
}

var processCategorizeCmd = &cobra.Command{
	Use:   "categorize",
	Short: "Print uncategorized transactions as JSON",
	RunE:  runProcessCategorize,
}

var processTrendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Print monthly spending data as JSON",
	RunE:  runProcessTrends,
}

var processBudgetsCmd = &cobra.Command{
	Use:   "budgets",
	Short: "Print multi-month spending data for budget analysis as JSON",
	RunE:  runProcessBudgets,
}

func init() {
	processCmd.AddCommand(processEmailsCmd, processCategorizeCmd, processTrendsCmd, processBudgetsCmd)
	rootCmd.AddCommand(processCmd)
}

func runProcessEmails(cmd *cobra.Command, _ []string) error {
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

	return process.PrintPendingEmails(ctx, repo, cfg.Email.AccountName, os.Stdout)
}

func runProcessCategorize(cmd *cobra.Command, _ []string) error {
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

	return process.PrintUncategorized(ctx, repo, os.Stdout)
}

func runProcessTrends(cmd *cobra.Command, _ []string) error {
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

	return process.PrintTrends(ctx, repo, os.Stdout)
}

func runProcessBudgets(cmd *cobra.Command, _ []string) error {
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

	return process.PrintBudgetData(ctx, repo, os.Stdout)
}
