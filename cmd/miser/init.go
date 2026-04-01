package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Waxmard/miser/internal/config"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create config, database, and seed categories",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	dir, err := config.DefaultDir()
	if err != nil {
		return err
	}

	// Create data directory.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	// Write default config if it doesn't exist.
	configPath := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		src, err := os.ReadFile("config.example.toml")
		if err != nil {
			// Fall back to writing a minimal config.
			cfg, err := config.Default()
			if err != nil {
				return err
			}
			_ = cfg // config.example.toml not found, skip config creation
			fmt.Fprintf(os.Stderr, "Warning: config.example.toml not found, skipping config creation\n")
			fmt.Fprintf(os.Stderr, "Copy config.example.toml to %s manually\n", configPath)
		} else {
			if err := os.WriteFile(configPath, src, 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
			fmt.Printf("Created %s\n", configPath)
		}
	} else {
		fmt.Printf("Config already exists: %s\n", configPath)
	}

	// Open database and run migrations.
	cfg, err := config.Default()
	if err != nil {
		return err
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	if err := repo.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	fmt.Printf("Created %s (sqlite)\n", cfg.Database.SQLitePath)

	// Seed categories.
	count, err := seedCategories(ctx, repo)
	if err != nil {
		return fmt.Errorf("seed categories: %w", err)
	}
	fmt.Printf("Seeded %d categories\n", count)

	return nil
}

func seedCategories(ctx context.Context, repo repository.Repository) (int, error) {
	now := time.Now().UTC()
	var cats []repository.Category
	for _, name := range defaultCategories {
		cats = append(cats, repository.Category{
			ID:        ulid.Make().String(),
			Name:      name,
			CreatedAt: now,
		})
	}
	if err := repo.Categories().Seed(ctx, cats); err != nil {
		return 0, err
	}
	return len(cats), nil
}

var defaultCategories = []string{
	"Auto Maintenance",
	"Auto Payment",
	"Balance Adjustments",
	"Bars/Drinking",
	"Books",
	"Career Growth",
	"Cash & ATM",
	"Charging",
	"Charity",
	"Chess",
	"Clothing",
	"Credit Card Payment",
	"Dentist",
	"Dividends & Capital Gains",
	"Electronics",
	"Entertainment & Recreation",
	"Financial & Legal Services",
	"Financial Fees",
	"Fitness",
	"For Work Reimbursement",
	"Furniture & Housewares",
	"Gas & Electric",
	"Gifts",
	"Groceries",
	"Household Necessities",
	"Ignored",
	"Insurance",
	"Interest",
	"Internet & Cable",
	"Jean",
	"Liquids",
	"Loan Repayment",
	"Medical",
	"Miscellaneous",
	"Nicotine",
	"Office Supplies & Expenses",
	"Other Income",
	"Parking & Tolls",
	"Paychecks",
	"Personal Care",
	"Phone",
	"Postage & Shipping",
	"Public Transit",
	"Rent",
	"Restaurants",
	"Sell",
	"Shopping",
	"Taxes",
	"Taxi & Ride Shares",
	"Transfer",
	"Travel & Vacation",
	"Uncategorized",
	"Water",
	"Zoe",
}
