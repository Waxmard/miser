package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	flagFrom     string
	flagTo       string
	flagCategory string
	flagTag      string
	flagOwner    string
	flagAccount  string
	flagLimit    int
)

var transactionsCmd = &cobra.Command{
	Use:     "transactions",
	Aliases: []string{"txns", "tx"},
	Short:   "List transactions",
	RunE:    runTransactions,
}

func init() {
	f := transactionsCmd.Flags()
	f.StringVar(&flagFrom, "from", "", "Start date (YYYY-MM-DD)")
	f.StringVar(&flagTo, "to", "", "End date (YYYY-MM-DD)")
	f.StringVar(&flagCategory, "category", "", "Filter by category name")
	f.StringVar(&flagTag, "tag", "", "Filter by tag")
	f.StringVar(&flagOwner, "owner", "", "Filter by owner")
	f.StringVar(&flagAccount, "account", "", "Filter by account name")
	f.IntVarP(&flagLimit, "limit", "n", 50, "Max transactions to show")
	rootCmd.AddCommand(transactionsCmd)
}

func runTransactions(cmd *cobra.Command, _ []string) error {
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

	filters := &repository.TransactionFilters{
		Limit: flagLimit,
	}

	if flagFrom != "" {
		t, err := time.Parse("2006-01-02", flagFrom)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		filters.From = &t
	} else {
		// Default: last 30 days.
		t := time.Now().AddDate(0, 0, -30)
		filters.From = &t
	}

	if flagTo != "" {
		t, err := time.Parse("2006-01-02", flagTo)
		if err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
		filters.To = &t
	}

	if flagCategory != "" {
		cat, err := repo.Categories().GetByName(ctx, flagCategory)
		if err != nil {
			return fmt.Errorf("category %q not found: %w", flagCategory, err)
		}
		filters.CategoryID = &cat.ID
	}

	if flagTag != "" {
		filters.Tag = &flagTag
	}
	if flagOwner != "" {
		filters.Owner = &flagOwner
	}
	if flagAccount != "" {
		acct, err := repo.Accounts().GetByName(ctx, flagAccount)
		if err != nil {
			return fmt.Errorf("account %q not found: %w", flagAccount, err)
		}
		filters.AccountID = &acct.ID
	}

	txns, err := repo.Transactions().List(ctx, filters)
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	if len(txns) == 0 {
		fmt.Println("No transactions found")
		return nil
	}

	printTransactionTable(txns)
	return nil
}

func printTransactionTable(txns []repository.Transaction) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	header := lipgloss.NewStyle().Bold(true)

	// Print header.
	fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s\n",
		header.Render(pad("DATE", 12)),
		header.Render(pad("MERCHANT", 24)),
		header.Render(pad("CATEGORY", 22)),
		header.Render(padLeft("AMOUNT", 10)),
		header.Render(pad("SOURCE", 10)),
	)

	for i := range txns {
		t := &txns[i]

		date := t.Date.Format("2006-01-02")
		merchant := truncate(t.Merchant, 24)
		category := truncate(t.CategoryName, 22)
		if category == "" {
			category = "Uncategorized"
		}

		amountStr := formatAmount(t.Amount)
		amountStyle := red
		if t.Amount > 0 {
			amountStyle = green
		}

		source := t.Source
		if t.CategorizedBy != nil {
			source += " (" + *t.CategorizedBy + ")"
		}
		source = truncate(source, 18)

		fmt.Fprintf(os.Stdout, "%s  %s  %s  %s  %s\n",
			dim.Render(pad(date, 12)),
			pad(merchant, 24),
			pad(category, 22),
			amountStyle.Render(padLeft(amountStr, 10)),
			dim.Render(pad(source, 10)),
		)
	}
}

func formatAmount(amount float64) string {
	abs := math.Abs(amount)
	if amount < 0 {
		return fmt.Sprintf("-$%.2f", abs)
	}
	return fmt.Sprintf("$%.2f", abs)
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-1] + "…"
}
