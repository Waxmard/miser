# Task: Generate the weekly spending snapshot

You are a personal finance analyst for a single user. Produce a brief, actionable weekly spending snapshot.

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
  "month_progress": 0.43,
  "categories": [
    {
      "category": "Groceries",
      "current": -260.00, "previous": -240.00,
      "delta_abs": -20.00, "delta_pct": 8.3,
      "budget": 600.00, "budget_used_pct": 43.3, "pacing": "on_track",
      "txn_count": 8
    }
  ],
  "top_movers": [{"category": "Dining", "current": -180, "previous": -90, "delta_abs": -90, "delta_pct": 100.0}],
  "anomalies": [{"merchant": "Delta", "category": "Travel", "amount": -892, "date": "2026-04-12", "reason": "5.2x category median ($170)"}],
  "budgets": [{"category": "Groceries", "budget": 600.00}]
}
```

Also fetch this week's transactions for the merchant-level callout:

```bash
miser transactions --from $(date -v-7d +%Y-%m-%d) --to $(date -v-1d +%Y-%m-%d) --limit 50
```

## Step 2: Analyze

Read fields directly from the JSON — do not recompute math:

1. **Top spending categories** — sort `categories` by `|current|` descending, take the top 3–5
2. **Budget pacing** — for each category with `pacing` set, report it directly (`on_track` / `ahead` / `behind` / `over`). Cite `budget_used_pct` and `month_progress`
3. **Notable transactions** — list `anomalies[]` if present; otherwise pick the largest entries from this week's transactions
4. **Month-over-month context** — use `top_movers[]` and per-category `delta_abs` / `delta_pct` for MoM (current MTD vs prior month MTD-clamped)

## Step 3: Write the report

Write `/tmp/miser-report.json` matching this schema exactly. Treat the schema as a contract, not as example prose.

<output_schema>
{
  "year": 2026,
  "month": 4,
  "narrative": "your narrative here"
}
</output_schema>

Use the current year and month. The narrative should be ~200 words in this format:

```
**Week of Apr 7–13, 2026**

- Top spending: Groceries ($125, 4 txns), Dining ($85, 3 txns), Gas ($45, 1 txn)
- Budget pacing: Groceries at 65% of $600 budget with 40% of month remaining — slightly ahead. Dining on track.
- Notable: $85 charge at [merchant] is higher than your typical dining transaction
- vs. last month: Total spending is running 12% lower than March at this point

**Action:** Keep an eye on grocery spending — you're trending $50 over last month's pace.
```

Rules for the narrative:

- Lead with the most important insight, not a generic summary
- Use actual numbers — don't say "spending increased", say "spending increased $45 (+12%)"
- Only mention categories that have meaningful activity
- The action item must be specific and based on the data, not generic advice
- Amounts are negative for expenses in the raw data — present them as positive in the narrative for readability

## Step 4: Persist and verify

```bash
miser internal write report /tmp/miser-report.json
miser trends report
```

The verify command must print the narrative you just wrote. If it does not, stop and report the error.
