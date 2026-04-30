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

// ReportInput is the JSON format Claude writes for monthly reports.
type ReportInput struct {
	Year     int                        `json:"year"`
	Month    int                        `json:"month"`
	Sections []repository.ReportSection `json:"sections"`
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
		txn.Status = "pending_review"
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
		Narrative: "",
		Sections:  input.Sections,
		Data:      string(data),
		CreatedAt: time.Now().UTC(),
	}

	return repo.Reports().Create(ctx, report)
}

// BudgetsInput is the JSON format Claude writes after analyzing spending and suggesting budgets.
type BudgetsInput struct {
	Budgets []BudgetSuggestion `json:"budgets"`
	Remove  []string           `json:"remove,omitempty"` // category IDs whose budgets should be deleted
}

// BudgetSuggestion is a single budget recommendation from Claude.
type BudgetSuggestion struct {
	CategoryID string  `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Reasoning  string  `json:"reasoning"`
}

// WriteBudgetsResult holds counts from a WriteBudgets operation.
type WriteBudgetsResult struct {
	Set     int
	Removed int
}

// WriteBudgets reads Claude's budget suggestions from a JSON file, upserts budgets,
// and removes any budgets listed in the Remove field.
func WriteBudgets(ctx context.Context, repo repository.Repository, jsonPath string) (*WriteBudgetsResult, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var input BudgetsInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	result := &WriteBudgetsResult{}

	for i := range input.Budgets {
		s := &input.Budgets[i]

		// Validate that the category exists.
		if _, err := repo.Categories().GetByID(ctx, s.CategoryID); err != nil {
			return result, fmt.Errorf("category %q (%s) not found: %w", s.Category, s.CategoryID, err)
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
			return result, fmt.Errorf("set budget for %q: %w", s.Category, err)
		}
		result.Set++
	}

	// Remove budgets for categories the user chose to skip.
	for _, categoryID := range input.Remove {
		existing, err := repo.Budgets().GetByCategoryID(ctx, categoryID)
		if err != nil {
			continue // no budget to remove
		}
		if err := repo.Budgets().Delete(ctx, existing.ID); err != nil {
			return result, fmt.Errorf("delete budget for category %s: %w", categoryID, err)
		}
		result.Removed++
	}

	return result, nil
}

// ReviewInput is the JSON format Claude writes after reviewing pending transactions.
type ReviewInput struct {
	Results []ReviewResult `json:"results"`
}

type ReviewResult struct {
	TransactionID string `json:"transaction_id"`
	Action        string `json:"action"`             // "approve" or "change"
	Category      string `json:"category,omitempty"` // required when action is "change"
}

// WriteReview reads Claude's review decisions from a JSON file and finalizes transactions.
// Returns the number of transactions resolved.
func WriteReview(ctx context.Context, repo repository.Repository, jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var input ReviewInput
	if err := json.Unmarshal(data, &input); err != nil {
		return 0, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	resolved := 0

	for i := range input.Results {
		r := &input.Results[i]

		txn, err := repo.Transactions().GetByID(ctx, r.TransactionID)
		if err != nil {
			return resolved, fmt.Errorf("transaction %s not found: %w", r.TransactionID, err)
		}

		switch r.Action {
		case "approve":
			txn.Status = "categorized"
			// categorized_by stays as "claude"
		case "change":
			cat, err := repo.Categories().GetByName(ctx, r.Category)
			if err != nil {
				return resolved, fmt.Errorf("category %q not found: %w", r.Category, err)
			}
			txn.CategoryID = &cat.ID
			txn.Status = "categorized"
			categorizedBy := "manual"
			txn.CategorizedBy = &categorizedBy
		default:
			return resolved, fmt.Errorf("unknown action %q for transaction %s", r.Action, r.TransactionID)
		}

		txn.UpdatedAt = now
		if err := repo.Transactions().Update(ctx, txn); err != nil {
			return resolved, fmt.Errorf("update transaction %s: %w", r.TransactionID, err)
		}
		resolved++
	}

	return resolved, nil
}

// HierarchyInput is the JSON format Claude writes to organize categories into groups.
type HierarchyInput struct {
	Groups []HierarchyGroup `json:"groups"`
}

// HierarchyGroup defines a parent category and which existing categories belong under it.
type HierarchyGroup struct {
	Name     string   `json:"name"`     // Parent category name (may not yet exist)
	Children []string `json:"children"` // Existing category names to nest under it
}

// WriteHierarchy reads Claude's hierarchy suggestions from a JSON file and applies them.
// Parent categories are created if they don't exist. Children have their parent_id set.
// Returns the number of child categories organized.
func WriteHierarchy(ctx context.Context, repo repository.Repository, jsonPath string) (int, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var input HierarchyInput
	if err := json.Unmarshal(data, &input); err != nil {
		return 0, fmt.Errorf("parse JSON: %w", err)
	}

	now := time.Now().UTC()
	organized := 0

	for i := range input.Groups {
		g := &input.Groups[i]

		parent, err := repo.Categories().GetByName(ctx, g.Name)
		if err != nil {
			parent = &repository.Category{
				ID:        ulid.Make().String(),
				Name:      g.Name,
				CreatedAt: now,
			}
			if err := repo.Categories().Create(ctx, parent); err != nil {
				return organized, fmt.Errorf("create parent category %q: %w", g.Name, err)
			}
		}

		for _, childName := range g.Children {
			child, err := repo.Categories().GetByName(ctx, childName)
			if err != nil {
				return organized, fmt.Errorf("child category %q not found: %w", childName, err)
			}
			child.ParentID = &parent.ID
			if err := repo.Categories().Update(ctx, child); err != nil {
				return organized, fmt.Errorf("update category %q: %w", childName, err)
			}
			organized++
		}
	}

	return organized, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
