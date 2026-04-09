package process

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Waxmard/miser/internal/repository"
)

// HierarchyOutput is the JSON printed for Claude to suggest category groupings.
type HierarchyOutput struct {
	CategoryCount int            `json:"category_count"`
	Categories    []FlatCategory `json:"categories"`
}

// FlatCategory is a category without hierarchy context, used as input for Claude's grouping.
type FlatCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PrintHierarchy writes the flat (non-Uncategorized) category list for Claude to group.
func PrintHierarchy(ctx context.Context, repo repository.Repository, w io.Writer) error {
	cats, err := repo.Categories().List(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	out := HierarchyOutput{
		Categories: []FlatCategory{},
	}

	for i := range cats {
		c := &cats[i]
		if c.Name == "Uncategorized" {
			continue
		}
		out.Categories = append(out.Categories, FlatCategory{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	out.CategoryCount = len(out.Categories)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
