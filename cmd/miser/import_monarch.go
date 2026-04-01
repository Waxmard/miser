package main

import (
	"fmt"

	"github.com/Waxmard/miser/internal/config"
	"github.com/Waxmard/miser/internal/ingest"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var importMonarchCmd = &cobra.Command{
	Use:   "import-monarch <csv-file>",
	Short: "Import transactions from Monarch Money CSV export",
	Args:  cobra.ExactArgs(1),
	RunE:  runImportMonarch,
}

func init() {
	rootCmd.AddCommand(importMonarchCmd)
}

func runImportMonarch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	csvPath := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	fmt.Println("Parsing Monarch CSV...")
	result, err := ingest.ImportMonarch(ctx, repo, csvPath)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	fmt.Printf("Created %d accounts\n", result.Accounts)
	fmt.Printf("Imported %d transactions\n", result.Transactions)
	fmt.Printf("Created %d categories\n", result.Categories)
	fmt.Printf("Extracted %d categorization rules\n", result.Rules)
	if result.Skipped > 0 {
		fmt.Printf("Skipped %d \"Ignored\" transactions\n", result.Skipped)
	}

	return nil
}

// loadConfig loads config from the default path, falling back to defaults.
func loadConfig() (*config.Config, error) {
	path, err := config.DefaultPath()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return config.Default()
	}
	return cfg, nil
}
