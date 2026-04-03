package main

import (
	"fmt"

	"github.com/Waxmard/miser/internal/categorize"
	"github.com/Waxmard/miser/internal/process"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var writeParsedCmd = &cobra.Command{
	Use:   "parsed <json-file>",
	Short: "Write Claude's email parse results to the database",
	Args:  cobra.ExactArgs(1),
	RunE:  runWriteParsed,
}

var writeCategoriesCmd = &cobra.Command{
	Use:   "categories <json-file>",
	Short: "Write Claude's categorization results to the database",
	Args:  cobra.ExactArgs(1),
	RunE:  runWriteCategories,
}

var writeReportCmd = &cobra.Command{
	Use:   "report <json-file>",
	Short: "Write Claude's narrative report to the database",
	Args:  cobra.ExactArgs(1),
	RunE:  runWriteReport,
}

var writeBudgetsCmd = &cobra.Command{
	Use:   "budgets <json-file>",
	Short: "Write Claude's budget suggestions to the database",
	Args:  cobra.ExactArgs(1),
	RunE:  runWriteBudgets,
}

func init() {
	internalWriteCmd.AddCommand(writeParsedCmd, writeCategoriesCmd, writeReportCmd, writeBudgetsCmd)
}

func runWriteParsed(cmd *cobra.Command, args []string) error {
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

	// Look up the account ID for the configured email account.
	acct, err := repo.Accounts().GetByName(ctx, cfg.Email.AccountName)
	if err != nil {
		return fmt.Errorf("account %q not found (run 'miser init' first): %w", cfg.Email.AccountName, err)
	}

	count, err := process.WriteParsedEmails(ctx, repo, args[0], acct.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Created %d transactions\n", count)

	// Run rule engine on new uncategorized transactions.
	if count > 0 {
		ruleResult, err := categorize.RunRules(ctx, repo)
		if err != nil {
			return fmt.Errorf("rule engine: %w", err)
		}
		if ruleResult.Categorized > 0 {
			fmt.Printf("Rule engine categorized %d/%d transactions\n", ruleResult.Categorized, ruleResult.Checked)
		}
	}

	return nil
}

func runWriteCategories(cmd *cobra.Command, args []string) error {
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

	count, err := process.WriteCategories(ctx, repo, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Categorized %d transactions\n", count)
	return nil
}

func runWriteBudgets(cmd *cobra.Command, args []string) error {
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

	result, err := process.WriteBudgets(ctx, repo, args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Set %d budgets\n", result.Set)
	if result.Removed > 0 {
		fmt.Printf("Removed %d budgets\n", result.Removed)
	}
	return nil
}

func runWriteReport(cmd *cobra.Command, args []string) error {
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

	if err := process.WriteReport(ctx, repo, args[0]); err != nil {
		return err
	}
	fmt.Println("Report saved")
	return nil
}
