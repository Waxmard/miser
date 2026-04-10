package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Waxmard/miser/internal/repository"
)

// HierarchyOutput is the JSON printed for Claude to suggest category groupings.
// It includes the current hierarchy state so Claude can build on existing groups.
type HierarchyOutput struct {
	CategoryCount int              `json:"category_count"`
	CurrentGroups []HierarchyGroup `json:"current_groups,omitempty"`
	Ungrouped     []FlatCategory   `json:"ungrouped,omitempty"`
}

// FlatCategory is a category without hierarchy context.
type FlatCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PrintHierarchy writes the category hierarchy state for Claude to build on.
// Current groups (parent + children) are listed separately from ungrouped categories.
func PrintHierarchy(ctx context.Context, repo repository.Repository, w io.Writer) error {
	cats, err := repo.Categories().List(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	// Build a map of parentID -> child names and track which IDs are children.
	childrenOf := make(map[string][]string) // parentID -> child names
	isChild := make(map[string]bool)
	byID := make(map[string]repository.Category)

	for i := range cats {
		c := &cats[i]
		byID[c.ID] = *c
		if c.ParentID != nil {
			childrenOf[*c.ParentID] = append(childrenOf[*c.ParentID], c.Name)
			isChild[c.ID] = true
		}
	}

	out := HierarchyOutput{
		CurrentGroups: []HierarchyGroup{},
		Ungrouped:     []FlatCategory{},
	}

	for i := range cats {
		c := &cats[i]
		if c.Name == "Uncategorized" {
			continue
		}
		if isChild[c.ID] {
			continue // printed under parent
		}
		if children, ok := childrenOf[c.ID]; ok {
			// This is a parent category.
			out.CurrentGroups = append(out.CurrentGroups, HierarchyGroup{
				Name:     c.Name,
				Children: children,
			})
		} else {
			// Ungrouped leaf category.
			out.Ungrouped = append(out.Ungrouped, FlatCategory{
				ID:   c.ID,
				Name: c.Name,
			})
		}
	}

	total := 0
	for _, g := range out.CurrentGroups {
		total += len(g.Children)
	}
	total += len(out.Ungrouped)
	out.CategoryCount = total

	if out.CategoryCount == 0 {
		out.CurrentGroups = nil
		out.Ungrouped = nil
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
