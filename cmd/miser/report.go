package main

import (
	"fmt"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show the latest Claude narrative report",
	RunE:  runReport,
}

func init() {
	trendsCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, _ []string) error {
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

	report, err := repo.Reports().GetLatest(ctx)
	if err != nil {
		fmt.Println("No reports yet")
		return nil
	}

	fmt.Printf("%s %d Report\n\n", monthName(report.Month), report.Year)
	fmt.Println(report.Narrative)

	return nil
}

func monthName(m int) string {
	months := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	if m >= 1 && m <= 12 {
		return months[m]
	}
	return fmt.Sprintf("Month %d", m)
}
