# Category Hierarchy

Claude Code workflow to organize flat categories into a parent/child hierarchy. Run once after the initial Monarch import (or any time you want to reorganize).

## When to run

- After `miser import-monarch` — you'll have many flat categories from Monarch data
- When new categories have accumulated and you want to group them

## Usage

```bash
claude -p "$(cat /path/to/miser/cron/category-hierarchy.md)" --model sonnet --allowedTools "Bash,Read,Write"
```

## Prompt

You are organizing personal finance categories into a logical hierarchy for a spending tracker. Your job is to group flat categories into parent/child relationships.

### Step 1: Get the current categories

```bash
miser internal process hierarchy
```

This returns JSON showing the existing hierarchy state:

```json
{
  "category_count": 24,
  "current_groups": [
    {
      "name": "Food & Drink",
      "children": ["Groceries", "Restaurants", "Bars/Drinking"]
    },
    {
      "name": "Housing",
      "children": ["Rent", "Gas & Electric", "Internet"]
    }
  ],
  "ungrouped": [
    { "id": "01HX...", "name": "NewCategory" },
    { "id": "01HX...", "name": "AnotherNew" }
  ]
}
```

- `current_groups` — categories already organized into parent groups
- `ungrouped` — categories not yet assigned to any group
- `category_count` — total number of non-Uncategorized leaf categories

If `category_count` is 0, there is nothing to organize — stop here.

If `ungrouped` is empty, the hierarchy is already complete. Print a summary table of the current groups and their children, then stop — do not write any files or run any further commands.

### Step 2: Design a hierarchy

Start from the existing groups in `current_groups` and place any `ungrouped` categories into appropriate groups. You may also reorganize existing groups if a better structure makes sense.

Guidelines:

- Aim for **4–8 parent groups** covering all categories
- Parent group names should not be leaf category names
- Every category should appear as a child of exactly one group, or be left ungrouped if it truly doesn't fit
- Do **not** include "Uncategorized" in any group
- Common useful groups: Housing, Food & Drink, Transportation, Health & Wellness, Entertainment, Finance, Shopping, Income, Personal

### Step 3: Write your hierarchy

Write a JSON file to `/tmp/miser-hierarchy.json` with the **complete** hierarchy (all groups and all children, not just changes):

```json
{
  "groups": [
    {
      "name": "Housing",
      "children": ["Rent", "Parking", "Utilities", "Home Improvement"]
    },
    {
      "name": "Food & Drink",
      "children": ["Groceries", "Restaurants", "Coffee Shops"]
    },
    {
      "name": "Transportation",
      "children": ["Gas", "Parking Fees", "Rideshare", "Auto & Transport"]
    }
  ]
}
```

Only include groups that have at least one child. Do not list a category as a child if it doesn't exist in the categories list from Step 1.

### Step 4: Apply the hierarchy

```bash
miser internal write hierarchy /tmp/miser-hierarchy.json
```

### Step 5: Verify

```bash
miser categories
```

Categories should now appear grouped under their parents. Run again with `--from` and `--to` flags to see the hierarchy with transaction data.
