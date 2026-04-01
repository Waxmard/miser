package main

import (
	"fmt"
	"os"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List accounts",
	RunE:  runAccounts,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}

func runAccounts(cmd *cobra.Command, _ []string) error {
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

	accounts, err := repo.Accounts().List(ctx)
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts")
		return nil
	}

	header := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
		header.Render(pad("NAME", 40)),
		header.Render(pad("INSTITUTION", 16)),
		header.Render(pad("TYPE", 12)),
		header.Render(pad("SOURCE", 16)),
	)

	for i := range accounts {
		a := &accounts[i]
		fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
			pad(truncate(a.Name, 40), 40),
			pad(a.Institution, 16),
			pad(a.AccountType, 12),
			dim.Render(pad(a.Source, 16)),
		)
	}

	return nil
}
