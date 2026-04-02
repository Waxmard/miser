package main

import (
	"fmt"

	"github.com/Waxmard/miser/internal/ingest"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run all enabled sync sources",
	RunE:  runSyncAll,
}

var syncEmailCmd = &cobra.Command{
	Use:   "email",
	Short: "Poll Gmail for new transaction emails",
	RunE:  runSyncEmail,
}

var syncSimpleFinCmd = &cobra.Command{
	Use:   "simplefin",
	Short: "Pull transactions from SimpleFIN",
	RunE:  runSyncSimpleFin,
}

func init() {
	syncCmd.AddCommand(syncEmailCmd)
	syncCmd.AddCommand(syncSimpleFinCmd)
	rootCmd.AddCommand(syncCmd)
}

func runSyncAll(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if cfg.Email.Enabled {
		if err := runSyncEmail(cmd, nil); err != nil {
			return err
		}
	}

	if cfg.SimpleFin.Enabled {
		if err := runSyncSimpleFin(cmd, nil); err != nil {
			return err
		}
	}

	return nil
}

func runSyncEmail(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.Email.Enabled {
		fmt.Println("Email sync is disabled in config")
		return nil
	}

	if cfg.Email.AppPassword == "" {
		return fmt.Errorf("email app_password not configured — set it in ~/.miser/config.toml")
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	fmt.Printf("Connecting to %s...\n", cfg.Email.IMAPServer)
	result, err := ingest.SyncEmail(ctx, repo, &cfg.Email)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d emails in %s\n", result.Found, cfg.Email.Label)
	fmt.Printf("Stored %d new raw emails (status: pending)\n", result.Stored)
	return nil
}

func runSyncSimpleFin(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.SimpleFin.Enabled {
		fmt.Println("SimpleFIN sync is disabled in config")
		return nil
	}

	if cfg.SimpleFin.AccessURL == "" {
		return fmt.Errorf("simplefin access_url not configured — run: miser setup simplefin <setup-token>")
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer repo.Close()

	fmt.Println("Syncing from SimpleFIN...")
	result, err := ingest.SyncSimpleFIN(ctx, repo, &cfg.SimpleFin)
	if err != nil {
		return err
	}

	fmt.Printf("Synced %d accounts\n", result.AccountsSynced)
	fmt.Printf("Found %d transactions, stored %d new\n", result.Found, result.Stored)
	if result.Categorized > 0 {
		fmt.Printf("Auto-categorized %d transactions via rules\n", result.Categorized)
	}
	return nil
}
