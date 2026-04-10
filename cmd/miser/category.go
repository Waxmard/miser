package main

import (
	"fmt"
	"time"

	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage categories",
}

var categoryMoveCmd = &cobra.Command{
	Use:   "move <name> <parent>",
	Short: "Move a category under a different parent group",
	Args:  cobra.ExactArgs(2),
	RunE:  runCategoryMove,
}

var categoryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new category",
	Args:  cobra.ExactArgs(1),
	RunE:  runCategoryCreate,
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a category and clear its transactions",
	Args:  cobra.ExactArgs(1),
	RunE:  runCategoryDelete,
}

func init() {
	categoryCreateCmd.Flags().String("parent", "", "Parent category name")
	categoryCmd.AddCommand(categoryMoveCmd, categoryCreateCmd, categoryDeleteCmd)
	rootCmd.AddCommand(categoryCmd)
}

func runCategoryMove(cmd *cobra.Command, args []string) error {
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

	name, parentName := args[0], args[1]

	cat, err := repo.Categories().GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("category %q not found", name)
	}

	parent, err := repo.Categories().GetByName(ctx, parentName)
	if err != nil {
		return fmt.Errorf("parent category %q not found", parentName)
	}

	cat.ParentID = &parent.ID
	if err := repo.Categories().Update(ctx, cat); err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	fmt.Printf("Moved %q under %q\n", name, parentName)
	return nil
}

func runCategoryCreate(cmd *cobra.Command, args []string) error {
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

	name := args[0]
	parentName, _ := cmd.Flags().GetString("parent")

	cat := &repository.Category{
		ID:        ulid.Make().String(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	if parentName != "" {
		parent, err := repo.Categories().GetByName(ctx, parentName)
		if err != nil {
			return fmt.Errorf("parent category %q not found", parentName)
		}
		cat.ParentID = &parent.ID
	}

	if err := repo.Categories().Create(ctx, cat); err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	if parentName != "" {
		fmt.Printf("Created category %q under %q\n", name, parentName)
	} else {
		fmt.Printf("Created category %q\n", name)
	}
	return nil
}

func runCategoryDelete(cmd *cobra.Command, args []string) error {
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

	name := args[0]

	cat, err := repo.Categories().GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("category %q not found", name)
	}

	// Ungroup any children.
	allCats, err := repo.Categories().List(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}
	ungrouped := 0
	for i := range allCats {
		c := &allCats[i]
		if c.ParentID != nil && *c.ParentID == cat.ID {
			c.ParentID = nil
			if err := repo.Categories().Update(ctx, c); err != nil {
				return fmt.Errorf("ungroup %q: %w", c.Name, err)
			}
			fmt.Printf("Warning: ungrouped child category %q\n", c.Name)
			ungrouped++
		}
	}

	// Clear transactions assigned to this category.
	txns, err := repo.Transactions().List(ctx, &repository.TransactionFilters{CategoryID: &cat.ID})
	if err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}
	now := time.Now().UTC()
	for i := range txns {
		t := &txns[i]
		t.CategoryID = nil
		t.Status = "uncategorized"
		t.UpdatedAt = now
		if err := repo.Transactions().Update(ctx, t); err != nil {
			return fmt.Errorf("clear transaction %s: %w", t.ID, err)
		}
	}

	// Delete rules for this category.
	rules, err := repo.Rules().List(ctx)
	if err != nil {
		return fmt.Errorf("list rules: %w", err)
	}
	deletedRules := 0
	for i := range rules {
		if rules[i].CategoryID == cat.ID {
			if err := repo.Rules().Delete(ctx, rules[i].ID); err != nil {
				return fmt.Errorf("delete rule %s: %w", rules[i].ID, err)
			}
			deletedRules++
		}
	}

	// Delete budget if any.
	if budget, err := repo.Budgets().GetByCategoryID(ctx, cat.ID); err == nil {
		_ = repo.Budgets().Delete(ctx, budget.ID)
	}

	if err := repo.Categories().Delete(ctx, cat.ID); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	fmt.Printf("Deleted %q. Cleared %d transaction(s). Deleted %d rule(s). Ungrouped %d child(ren).\n",
		name, len(txns), deletedRules, ungrouped)
	return nil
}
