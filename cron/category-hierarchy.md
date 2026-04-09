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

This returns JSON:

```json
{
  "category_count": 24,
  "categories": [
    { "id": "01HX...", "name": "Groceries" },
    { "id": "01HX...", "name": "Rent" },
    { "id": "01HX...", "name": "Netflix" },
    ...
  ]
}
```

If `category_count` is 0, there is nothing to organize — stop here.

### Step 2: Design a hierarchy

Group the categories into logical parent groups. Guidelines:

- Aim for **4–8 parent groups** covering all categories
- Parent group names should be new (not already in the category list)
- Every category in the list should appear as a child of exactly one group, or be left ungrouped if it doesn't fit
- Do **not** include "Uncategorized" in any group
- Common useful groups: Housing, Food, Transportation, Flexible, Subscriptions, Health, Income, Savings

### Step 3: Write your hierarchy

Write a JSON file to `/tmp/miser-hierarchy.json` with this structure:

```json
{
  "groups": [
    {
      "name": "Housing",
      "children": ["Rent", "Parking", "Utilities", "Home Improvement"]
    },
    {
      "name": "Food",
      "children": ["Groceries", "Restaurants", "Coffee Shops"]
    },
    {
      "name": "Subscriptions",
      "children": ["Netflix", "Spotify", "Amazon Prime", "iCloud"]
    },
    {
      "name": "Flexible",
      "children": ["Bars", "Entertainment", "Shopping", "Clothing"]
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
