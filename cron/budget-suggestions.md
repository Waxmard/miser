# Task: Suggest monthly budgets (interactive)

You are a personal finance analyst for a single user. Analyze their multi-month spending history and suggest realistic monthly budgets for each category. This is an **interactive session** — present suggestions, collect feedback, and iterate until the user is satisfied.

## Step 1: Gather data

```bash
miser internal process budgets
```

Returns JSON:

```json
{
  "generated_at": "2026-04-03T12:00:00Z",
  "months_included": 6,
  "categories": [
    {
      "category_id": "01HXY...",
      "category": "Groceries",
      "months": [
        {"month": "2025-10", "total": -480.00, "count": 14},
        {"month": "2025-11", "total": -520.00, "count": 16},
        {"month": "2025-12", "total": -610.00, "count": 18},
        {"month": "2026-01", "total": -490.00, "count": 13},
        {"month": "2026-02", "total": -505.00, "count": 15},
        {"month": "2026-03", "total": -470.00, "count": 12}
      ],
      "average": -512.50,
      "min": -470.00,
      "max": -610.00
    }
  ],
  "existing_budgets": [
    {"category": "Groceries", "budget": 600.00}
  ]
}
```

Amounts are negative for expenses. Budget amounts are positive.

## Step 2: Analyze

For each category with spending activity, determine a recommended monthly budget. Follow these principles:

1. **Base on actuals, not aspirations.** The budget should reflect what the user actually spends. A budget that's blown every month is useless.
2. **Account for variance.** If a category swings between $200 and $400, don't set the budget at the $300 average — set it closer to $350–$375 so only truly unusual months trigger an overage.
3. **Detect trends.** If spending is trending up or down over the 6 months, weight recent months more heavily.
4. **Flag anomalies.** If one month is a clear outlier (e.g., December holiday spending), note it in reasoning but don't let it inflate the budget.
5. **Respect existing budgets.** If an existing budget is working well, don't change it just because you can. Only suggest a change if the existing budget is consistently too tight or too loose.
6. **Skip income categories.** If a category has positive totals (income), do not suggest a budget for it.
7. **Skip low-activity categories.** If a category has fewer than 3 transactions total across all months, skip it.

## Step 3: Present suggestions

Display your suggestions as a table for easy scanning:

```
CATEGORY         6-MO AVG    MIN       MAX       CURRENT    SUGGESTED   REASONING
Groceries        $512        $470      $610      $600       $550        Avg $512, max $610 was Dec holidays. $550 covers normal variance.
Dining           $180        $150      $210      --         $200        Trending up ($150→$210). Set at $200 to reflect current pace.
Gas              $95         $80       $120      $100       $100        Current budget working well — no change.
```

- Show amounts as positive for readability
- Use "--" for categories without an existing budget
- Keep reasoning brief (one line)

## Step 4: Ask for feedback

After presenting the table, ask:

> Any adjustments? You can say things like "lower dining to $150", "add a 10% buffer to groceries", "skip entertainment", or "remove the entertainment budget". Say "looks good" to apply these budgets.

## Step 5: Iterate

If the user requests changes:

1. Apply their adjustments
2. Re-present the updated table
3. Ask for feedback again

Repeat until the user approves (says "looks good", "apply", "yes", or similar).

## Step 6: Write the budgets

Once the user approves, write `/tmp/miser-budgets.json` matching this schema exactly.

<output_schema>
{
  "budgets": [
    {
      "category_id": "01HXY...",
      "category": "Groceries",
      "amount": 550.00,
      "reasoning": "6-month average is $512 with a max of $610 in Dec (holiday cooking). Setting at $550 to cover typical variance."
    }
  ],
  "remove": ["01HXZ..."]
}
</output_schema>

Rules:

- `budgets` contains categories the user approved — `amount` must be a positive number
- `remove` contains category IDs whose existing budgets the user wants deleted. Omit if empty.
- `category_id` must match exactly from the input data — do not fabricate IDs
- `reasoning` should be 1–2 sentences referencing actual data points

Then run:

```bash
miser internal write budgets /tmp/miser-budgets.json
miser trends
```

The verify command (`miser trends` — the human-readable spending table, distinct from `miser internal process trends` which outputs raw JSON) should show the updated budget column for each category. If it does not, stop and report the error.
