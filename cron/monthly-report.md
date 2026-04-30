# Task: Generate the monthly spending report

You are a personal finance analyst for a single user. Produce a structured monthly spending report for the month that just ended.

## Step 1: Gather data

Run these commands and read their output:

```bash
miser internal process trends
```

Returns JSON with pre-computed deltas, pacing, top movers, and anomalies:

```json
{
  "current_month": "2026-04",
  "previous_month": "2026-03",
  "month_progress": 0.03,
  "categories": [
    {
      "category": "Groceries",
      "current": -45.00, "previous": -520.00,
      "delta_abs": 475.00, "delta_pct": -91.3,
      "budget": 600.00, "budget_used_pct": 7.5, "pacing": "behind",
      "txn_count": 1
    }
  ],
  "top_movers": [{"category": "Dining", "current": -50, "previous": -420, "delta_abs": 370, "delta_pct": -88.1}],
  "anomalies": [{"merchant": "Delta", "category": "Travel", "amount": -892, "date": "2026-03-15", "reason": "5.2x category median ($170)"}],
  "budgets": [{"category": "Groceries", "budget": 600.00}]
}
```

Since this runs on the 1st, `current_month` is the new (nearly empty) month and `previous_month` is the completed month you are summarizing. **All `current`/`previous`/`delta_*` fields refer to the new month vs the completed month — they are NOT what you want for the report.**

For the completed month's own data, run:

```bash
miser categories --from $(date -v-1m -v1d +%Y-%m-%d) --to $(date -v-1d +%Y-%m-%d)
```

Compare it against two months ago for MoM context:

```bash
miser categories --from $(date -v-2m -v1d +%Y-%m-%d) --to $(date -v-2m -v-1d +%Y-%m-%d)
```

## Step 2: Analyze

From the data, determine:

1. **Total spending** — sum of all expense categories for the completed month, and the transaction count (from `miser categories` output)
2. **Month-over-month change** — completed month vs two months ago
3. **Budget scorecard** — for each category with a budget set, compute actual / budget / pct used / over-or-under for the completed month
4. **Biggest movers** — categories with the largest absolute dollar change between completed month and two months ago, with a short reason if apparent
5. **Notable transactions** — largest individual charges in the completed month worth calling out

## Step 3: Write the report

Write `/tmp/miser-report.json` matching this schema exactly. Treat the schema as a contract, not as example prose.

<output_schema>
{
  "year": 2026,
  "month": 3,
  "sections": [
    {
      "type": "stat",
      "title": "March Total",
      "value": "$3,241",
      "delta": "+8.9%",
      "sign": "negative",
      "note": "vs $2,980 in February • 87 transactions"
    },
    {
      "type": "scorecard",
      "title": "Budget Scorecard",
      "items": [
        { "label": "Dining", "value": "$420", "note": "$300 budget", "pct": 140, "sign": "negative" },
        { "label": "Groceries", "value": "$550", "note": "$600 budget", "pct": 91, "sign": "positive" }
      ]
    },
    {
      "type": "movers",
      "title": "Biggest Changes",
      "items": [
        { "label": "Dining", "value": "+$120", "note": "4 restaurant charges over $40", "sign": "negative" },
        { "label": "Shopping", "value": "-$200", "note": "back to normal after Feb purchases", "sign": "positive" }
      ]
    },
    {
      "type": "transactions",
      "title": "Notable Transactions",
      "items": [
        { "label": "Delta Airlines", "value": "$892", "note": "Apr 2", "sign": "negative" }
      ]
    },
    {
      "type": "takeaways",
      "title": "Takeaways",
      "items": [
        { "label": "Dining over budget 3 months in a row — consider a weekly dining cap of $75" },
        { "label": "Groceries consistent for 3 months — budget well-calibrated" },
        { "label": "March was highest-spend month this quarter — total trending up" }
      ]
    }
  ]
}
</output_schema>

Rules:

- Use the **completed** month's year and month (not the current date)
- `sign`: `"negative"` = bad/expense color, `"positive"` = good/income color, omit for neutral
- `delta` on the stat section: format as `"+X.X%"` or `"-X.X%"` — spending up = `"negative"` sign
- `scorecard` `pct`: integer percentage of budget used (can exceed 100 for over-budget)
- `movers` `value`: dollar change with sign, e.g. `"+$120"` or `"-$200"`
- `takeaways`: exactly 3 items, specific and actionable with actual numbers
- Amounts are negative for expenses in the raw data — present them as positive dollars in the output

## Step 4: Persist and verify

```bash
miser internal write report /tmp/miser-report.json
miser trends report
```

The verify command must print the report. If it does not, stop and report the error.
