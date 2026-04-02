package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	"github.com/oklog/ulid/v2"
)

// MonarchRow represents a single row from the Monarch Money CSV export.
type MonarchRow struct {
	Date              string
	Merchant          string
	Category          string
	Account           string
	OriginalStatement string
	Notes             string
	Amount            float64
	Tags              string
	Owner             string
}

// MonarchResult holds the summary of a Monarch import.
type MonarchResult struct {
	Transactions int
	Accounts     int
	Categories   int
	Rules        int
	Skipped      int
}

// ImportMonarch reads a Monarch Money CSV and imports all data into the repository.
func ImportMonarch(ctx context.Context, repo repository.Repository, csvPath string) (*MonarchResult, error) {
	rows, err := parseMonarchCSV(csvPath)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}

	result := &MonarchResult{}

	// Filter out "Ignored" category rows.
	var kept []MonarchRow
	for i := range rows {
		if rows[i].Category == "Ignored" {
			result.Skipped++
			continue
		}
		kept = append(kept, rows[i])
	}

	// Create accounts.
	accountMap, err := ensureAccounts(ctx, repo, kept)
	if err != nil {
		return nil, fmt.Errorf("create accounts: %w", err)
	}
	result.Accounts = len(accountMap)

	// Create categories.
	categoryMap, err := ensureCategories(ctx, repo, kept)
	if err != nil {
		return nil, fmt.Errorf("create categories: %w", err)
	}
	result.Categories = len(categoryMap)

	// Import transactions.
	txns := buildTransactions(kept, accountMap, categoryMap)
	inserted, err := repo.Transactions().CreateBatch(ctx, txns)
	if err != nil {
		return nil, fmt.Errorf("insert transactions: %w", err)
	}
	result.Transactions = inserted

	// Extract and create rules.
	rules := extractRules(kept, categoryMap)
	for i := range rules {
		if err := repo.Rules().Create(ctx, &rules[i]); err != nil {
			return nil, fmt.Errorf("create rule %q: %w", rules[i].Pattern, err)
		}
	}
	result.Rules = len(rules)

	return result, nil
}

func parseMonarchCSV(path string) ([]MonarchRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Read and validate header.
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if len(header) < 9 || header[0] != "Date" {
		return nil, fmt.Errorf("unexpected CSV header: %v", header)
	}

	// Auto-detect format from second column.
	var parseRow func([]string) (*MonarchRow, error)
	switch header[1] {
	case "Merchant":
		// New format: Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags,Owner
		parseRow = parseNewFormat
	case "Original Date":
		// Old format: Date,Original Date,Account Type,Account Name,Account Number,
		//             Institution Name,Name,Custom Name,Amount,Description,Category,Note,Ignored From,Tax Deductible
		parseRow = parseOldFormat
	default:
		return nil, fmt.Errorf("unexpected CSV header: %v", header)
	}

	var rows []MonarchRow
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		row, err := parseRow(record)
		if err != nil {
			return nil, err
		}
		if row != nil {
			rows = append(rows, *row)
		}
	}

	return rows, nil
}

// New format: Date,Merchant,Category,Account,Original Statement,Notes,Amount,Tags,Owner
func parseNewFormat(record []string) (*MonarchRow, error) {
	if len(record) < 9 {
		return nil, nil
	}

	amount, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", record[6], err)
	}

	return &MonarchRow{
		Date:              record[0],
		Merchant:          record[1],
		Category:          record[2],
		Account:           record[3],
		OriginalStatement: record[4],
		Notes:             record[5],
		Amount:            amount,
		Tags:              record[7],
		Owner:             record[8],
	}, nil
}

// Old format: Date,Original Date,Account Type,Account Name,Account Number,
// Institution Name,Name,Custom Name,Amount,Description,Category,Note,Ignored From,Tax Deductible
func parseOldFormat(record []string) (*MonarchRow, error) {
	if len(record) < 11 {
		return nil, nil
	}

	// Skip rows with "Ignored From" populated (column 12, 0-indexed).
	if len(record) > 12 && record[12] != "" {
		return nil, nil
	}

	amount, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return nil, fmt.Errorf("parse amount %q: %w", record[8], err)
	}

	merchant := record[6] // Name
	if record[7] != "" {  // Custom Name takes precedence
		merchant = record[7]
	}

	return &MonarchRow{
		Date:              record[0],
		Merchant:          merchant,
		Category:          record[10],
		Account:           record[3], // Account Name
		OriginalStatement: record[9], // Description
		Notes:             record[11],
		Amount:            amount,
	}, nil
}

func ensureAccounts(ctx context.Context, repo repository.Repository, rows []MonarchRow) (map[string]string, error) {
	// Collect unique account names.
	seen := map[string]bool{}
	for i := range rows {
		seen[rows[i].Account] = true
	}

	accountMap := make(map[string]string) // name -> id
	now := time.Now().UTC()

	for name := range seen {
		// Check if account already exists.
		existing, err := repo.Accounts().GetByName(ctx, name)
		if err == nil {
			accountMap[name] = existing.ID
			continue
		}

		id := ulid.Make().String()
		acct := &repository.Account{
			ID:          id,
			Name:        name,
			Institution: guessInstitution(name),
			AccountType: guessAccountType(name),
			Source:      "monarch_import",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := repo.Accounts().Create(ctx, acct); err != nil {
			return nil, fmt.Errorf("create account %q: %w", name, err)
		}
		accountMap[name] = id
	}

	return accountMap, nil
}

func ensureCategories(ctx context.Context, repo repository.Repository, rows []MonarchRow) (map[string]string, error) {
	// Collect unique category names.
	seen := map[string]bool{}
	for i := range rows {
		seen[rows[i].Category] = true
	}

	categoryMap := make(map[string]string) // name -> id
	now := time.Now().UTC()

	for name := range seen {
		existing, err := repo.Categories().GetByName(ctx, name)
		if err == nil {
			categoryMap[name] = existing.ID
			continue
		}

		id := ulid.Make().String()
		cat := &repository.Category{
			ID:        id,
			Name:      name,
			CreatedAt: now,
		}
		if err := repo.Categories().Create(ctx, cat); err != nil {
			return nil, fmt.Errorf("create category %q: %w", name, err)
		}
		categoryMap[name] = id
	}

	return categoryMap, nil
}

func buildTransactions(rows []MonarchRow, accountMap, categoryMap map[string]string) []repository.Transaction {
	now := time.Now().UTC()
	txns := make([]repository.Transaction, 0, len(rows))

	for i := range rows {
		r := &rows[i]
		date, _ := time.Parse("2006-01-02", r.Date)
		sourceID := monarchSourceID(r)
		catID := categoryMap[r.Category]

		txn := repository.Transaction{
			ID:                ulid.Make().String(),
			AccountID:         accountMap[r.Account],
			CategoryID:        &catID,
			Amount:            r.Amount,
			Merchant:          r.Merchant,
			OriginalStatement: strPtr(r.OriginalStatement),
			Date:              date,
			Source:            "monarch_import",
			SourceID:          &sourceID,
			Status:            "categorized",
			CategorizedBy:     strPtr("monarch_import"),
			Tags:              nilIfEmpty(r.Tags),
			Owner:             nilIfEmpty(r.Owner),
			Notes:             nilIfEmpty(r.Notes),
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		txns = append(txns, txn)
	}

	return txns
}

func extractRules(rows []MonarchRow, categoryMap map[string]string) []repository.CategoryRule {
	// Count merchant -> category occurrences.
	type key struct{ merchant, category string }
	counts := map[key]int{}
	merchantTotal := map[string]int{}

	for i := range rows {
		k := key{merchant: rows[i].Merchant, category: rows[i].Category}
		counts[k]++
		merchantTotal[rows[i].Merchant]++
	}

	// Find merchants that map to the same category 90%+ of the time with 3+ transactions.
	now := time.Now().UTC()
	var rules []repository.CategoryRule
	seen := map[string]bool{}

	for k, count := range counts {
		total := merchantTotal[k.merchant]
		if total < 3 || seen[k.merchant] {
			continue
		}
		ratio := float64(count) / float64(total)
		if ratio < 0.9 {
			continue
		}

		seen[k.merchant] = true
		rules = append(rules, repository.CategoryRule{
			ID:         ulid.Make().String(),
			Pattern:    k.merchant,
			CategoryID: categoryMap[k.category],
			MatchType:  "exact",
			HitCount:   count,
			CreatedBy:  "monarch_import",
			CreatedAt:  now,
		})
	}

	return rules
}

func monarchSourceID(r *MonarchRow) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%.2f|%s", r.Date, r.Merchant, r.Amount, r.Account)))
	return fmt.Sprintf("monarch_%x", h[:8])
}

func guessInstitution(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "fidelity"):
		return "fidelity"
	case strings.Contains(lower, "capital one") || strings.Contains(lower, "360 checking") || strings.Contains(lower, "quicksilver"):
		return "capital_one"
	case strings.Contains(lower, "chase"):
		return "chase"
	case strings.Contains(lower, "amazon"):
		return "amazon"
	case strings.Contains(lower, "bilt"):
		return "bilt"
	case strings.Contains(lower, "verizon"):
		return "verizon"
	default:
		return "unknown"
	}
}

func guessAccountType(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "checking") || strings.Contains(lower, "cash management"):
		return "checking"
	case strings.Contains(lower, "savings"):
		return "savings"
	case strings.Contains(lower, "credit") || strings.Contains(lower, "card"):
		return "credit"
	case strings.Contains(lower, "ira") || strings.Contains(lower, "rollover"):
		return "investment"
	case strings.Contains(lower, "individual") || strings.Contains(lower, "joint") || strings.Contains(lower, "wros"):
		return "brokerage"
	default:
		return "checking"
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
