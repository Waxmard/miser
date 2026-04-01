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

func init() {
	syncCmd.AddCommand(syncEmailCmd)
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
