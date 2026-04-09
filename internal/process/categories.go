package process

import "github.com/Waxmard/miser/internal/repository"

// CategoryGroup is the JSON representation of a category with optional subcategories.
// Used in categorize and review process outputs so Claude understands the hierarchy.
type CategoryGroup struct {
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories,omitempty"`
}

// buildCategoryGroups organizes a flat category list into parent/child groups.
// Parents appear at the root level with their children nested inside.
// Standalone categories (no parent, no children) also appear at root level.
func buildCategoryGroups(cats []repository.Category) []CategoryGroup {
	byID := make(map[string]string, len(cats)) // id → name
	childrenOf := make(map[string][]string)    // parentID → []childName

	for i := range cats {
		c := &cats[i]
		byID[c.ID] = c.Name
		if c.ParentID != nil {
			childrenOf[*c.ParentID] = append(childrenOf[*c.ParentID], c.Name)
		}
	}

	var groups []CategoryGroup
	for i := range cats {
		c := &cats[i]
		if c.ParentID != nil {
			continue // nested under parent
		}
		group := CategoryGroup{Name: c.Name}
		group.Subcategories = childrenOf[c.ID] // nil if no children (omitempty handles it)
		groups = append(groups, group)
	}
	return groups
}
