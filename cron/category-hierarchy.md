# Task: Organize categories into a parent/child hierarchy

You are organizing personal finance categories into a logical hierarchy for a spending tracker. Group flat categories into parent/child relationships.

## Step 1: Get the current categories

```bash
miser internal process hierarchy
```

Returns JSON:

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

## Step 2: Design a hierarchy

Start from the existing groups in `current_groups` and place any `ungrouped` categories into appropriate groups. You may also reorganize existing groups if a better structure makes sense.

Guidelines:

- Aim for **4–8 parent groups** covering all categories
- Parent group names should not be leaf category names
- Every category should appear as a child of exactly one group, or be left ungrouped if it truly doesn't fit
- Do **not** include "Uncategorized" in any group
- Common useful groups: Housing, Food & Drink, Transportation, Health & Wellness, Entertainment, Finance, Shopping, Income, Personal

## Step 3: Write your hierarchy

Write `/tmp/miser-hierarchy.json` with the **complete** hierarchy (all groups and all children, not just changes). Treat the schema as a contract.

<output_schema>
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
</output_schema>

Only include groups that have at least one child. Do not list a category as a child if it doesn't exist in the categories list from Step 1.

## Step 4: Persist and verify

```bash
miser internal write hierarchy /tmp/miser-hierarchy.json
miser categories
```

Categories should now appear grouped under their parents. If they do not, stop and report the error.
