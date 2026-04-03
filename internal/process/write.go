package process

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	"github.com/oklog/ulid/v2"
)

// ParsedEmailsInput is the JSON format Claude writes after parsing emails.
type ParsedEmailsInput struct {
	Results []ParsedEmailResult `json:"results"`
}

type ParsedEmailResult struct {
	RawEmailID  string             `json:"raw_email_id"`
	Parsed      bool               `json:"parsed"`
	Reason      string             `json:"reason,omitempty"`
	Transaction *ParsedTransaction `json:"transaction,omitempty"`
}

type ParsedTransaction struct {
	Amount          float64 `json:"amount"`
	Merchant        string  `json:"merchant"`
	MerchantClean   string  `json:"merchant_clean,omitempty"`
	Date            string  `json:"date"`
	Time            string  `json:"time,omitempty"`
	TransactionType string  `json:"transaction_type,omitempty"`
	Description     string  `json:"description,omitempty"`
}

// CategoriesInput is the JSON format Claude writes after categorizing transactions.
type CategoriesInput struct {
	Results []CategoryResult `json:"results"`
}

type CategoryResult struct {
	TransactionID string         `json:"transaction_id"`
	Category      string         `json:"category"`
	MerchantClean string         `json:"merchant_clean,omitempty"`
	Confidence    float64        `json:"confidence"`
	SuggestedRule *SuggestedRule `json:"suggested_rule,omitempty"`
}

type SuggestedRule struct {
	Pattern   string `json:"pattern"`
	MatchType string `json:"match_type"`
}

// ReportInput is the JSON format Claude writes for weekly reports.
type ReportInput struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Narrative string `json:"narrative"`
}

// WriteParsedEmails reads Claude's parse results from a JSON file and writes transactions to the DB.
// Returns the number of transactions created.
func WriteParsedEmails(ctx context.Context, repo repository.Repository, jsonPath, accountID string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var input ParsedEmailsInput
	if err := json.Unmarshal(data, &input); err != nil {
		return 0, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	created := 0

	for i := range input.Results {
		r := &input.Results[i]

		if !r.Parsed {
			if err := repo.RawEmails().MarkFailed(ctx, r.RawEmailID, r.Reason); err != nil {
				return created, fmt.Errorf("mark failed %s: %w", r.RawEmailID, err)
			}
			continue
		}

		if r.Transaction == nil {
			continue
		}

		pt := r.Transaction
		date, _ := time.Parse("2006-01-02", pt.Date)
		sourceID := r.RawEmailID

		txn := &repository.Transaction{
			ID:            ulid.Make().String(),
			AccountID:     accountID,
			Amount:        pt.Amount,
			Merchant:      pt.Merchant,
			MerchantClean: nilIfEmpty(pt.MerchantClean),
			Description:   nilIfEmpty(pt.Description),
			Date:          date,
			Source:        "email",
			SourceID:      &sourceID,
			Status:        "uncategorized",
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := repo.Transactions().Create(ctx, txn); err != nil {
			return created, fmt.Errorf("create transaction: %w", err)
		}
		created++

		if err := repo.RawEmails().MarkProcessed(ctx, r.RawEmailID); err != nil {
			return created, fmt.Errorf("mark processed %s: %w", r.RawEmailID, err)
		}
	}

	return created, nil
}

// WriteCategories reads Claude's categorization results from a JSON file and updates transactions.
// Returns the number of transactions categorized.
func WriteCategories(ctx context.Context, repo repository.Repository, jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var input CategoriesInput
	if err := json.Unmarshal(data, &input); err != nil {
		return 0, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	categorized := 0

	for i := range input.Results {
		r := &input.Results[i]

		cat, err := repo.Categories().GetByName(ctx, r.Category)
		if err != nil {
			return categorized, fmt.Errorf("category %q not found: %w", r.Category, err)
		}

		txn, err := repo.Transactions().GetByID(ctx, r.TransactionID)
		if err != nil {
			return categorized, fmt.Errorf("transaction %s not found: %w", r.TransactionID, err)
		}

		txn.CategoryID = &cat.ID
		txn.Status = "categorized"
		categorizedBy := "claude"
		txn.CategorizedBy = &categorizedBy
		txn.Confidence = &r.Confidence
		txn.MerchantClean = nilIfEmpty(r.MerchantClean)
		txn.UpdatedAt = now

		if err := repo.Transactions().Update(ctx, txn); err != nil {
			return categorized, fmt.Errorf("update transaction %s: %w", r.TransactionID, err)
		}
		categorized++

		// Auto-create rule if confidence is high enough and a rule is suggested.
		if r.Confidence >= 0.85 && r.SuggestedRule != nil {
			rule := &repository.CategoryRule{
				ID:         ulid.Make().String(),
				Pattern:    r.SuggestedRule.Pattern,
				CategoryID: cat.ID,
				MatchType:  r.SuggestedRule.MatchType,
				HitCount:   1,
				CreatedBy:  "claude",
				CreatedAt:  now,
			}
			// Ignore errors from duplicate rules.
			_ = repo.Rules().Create(ctx, rule)
		}
	}

	return categorized, nil
}

// WriteReport reads Claude's report from a JSON file and stores it.
func WriteReport(ctx context.Context, repo repository.Repository, jsonPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var input ReportInput
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	report := &repository.Report{
		ID:        ulid.Make().String(),
		Year:      input.Year,
		Month:     input.Month,
		Narrative: input.Narrative,
		Data:      string(data),
		CreatedAt: time.Now().UTC(),
	}

	return repo.Reports().Create(ctx, report)
}

// BudgetsInput is the JSON format Claude writes after analyzing spending and suggesting budgets.
type BudgetsInput struct {
	Budgets []BudgetSuggestion `json:"budgets"`
}

// BudgetSuggestion is a single budget recommendation from Claude.
type BudgetSuggestion struct {
	CategoryID string  `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Reasoning  string  `json:"reasoning"`
}

// WriteBudgets reads Claude's budget suggestions from a JSON file and upserts them.
// Returns the number of budgets set.
func WriteBudgets(ctx context.Context, repo repository.Repository, jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var input BudgetsInput
	if err := json.Unmarshal(data, &input); err != nil {
		return 0, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	count := 0

	for i := range input.Budgets {
		s := &input.Budgets[i]

		// Validate that the category exists.
		if _, err := repo.Categories().GetByID(ctx, s.CategoryID); err != nil {
			return count, fmt.Errorf("category %q (%s) not found: %w", s.Category, s.CategoryID, err)
		}

		// Reuse existing budget ID if one exists for this category.
		var budgetID string
		var createdAt time.Time
		existing, err := repo.Budgets().GetByCategoryID(ctx, s.CategoryID)
		if err == nil && existing != nil {
			budgetID = existing.ID
			createdAt = existing.CreatedAt
		} else {
			budgetID = ulid.Make().String()
			createdAt = now
		}

		budget := &repository.Budget{
			ID:            budgetID,
			CategoryID:    s.CategoryID,
			MonthlyAmount: s.Amount,
			CreatedAt:     createdAt,
			UpdatedAt:     now,
		}

		if err := repo.Budgets().Set(ctx, budget); err != nil {
			return count, fmt.Errorf("set budget for %q: %w", s.Category, err)
		}
		count++
	}

	return count, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
