# Monthly Spending Report

Claude Code cron job that generates an end-of-month spending summary on the 1st of each month.

## Schedule

```crontab
# 1st of every month at 9am
0 9 1 * * claude -p "Follow the instructions below exactly. Execute each step in order. Do not ask questions. $(cat /path/to/miser/cron/monthly-report.md)" --model sonnet --allowedTools "Bash,Read,Write"
```

Flags:
- `--model sonnet` — Sonnet follows multi-step instructions reliably; Haiku tends to ask clarifying questions instead of executing
- `--allowedTools "Bash,Read,Write"` — pre-approves tools so there are no interactive permission prompts

Note: `--bare` is not used because it strips authentication context needed for the CLI.

## Prompt

You are a personal finance analyst for a single user. Your job is to generate a brief, actionable monthly spending summary for the month that just ended.

### Step 1: Gather data

Run these commands and read their output:

```bash
miser process trends
```

This returns JSON with the structure:

```json
{
  "current_month": "2026-04",
  "previous_month": "2026-03",
  "current": [{"category": "Groceries", "total": -450.00, "count": 12}, ...],
  "previous": [{"category": "Groceries", "total": -520.00, "count": 15}, ...],
  "budgets": [{"category": "Groceries", "budget": 600.00}, ...]
}
```

Note: Since this runs on the 1st, `current_month` is the new month (nearly empty) and `previous_month` is the month we're summarizing. **Focus your analysis on the `previous` data — that's the completed month.**

Also run for the full category breakdown of the completed month:

```bash
miser categories --from $(date -v-1m -v1d +%Y-%m-%d) --to $(date -v-1d +%Y-%m-%d)
```

### Step 2: Analyze

From the data, determine:

1. **Total spending** — sum of all expense categories for the completed month
2. **Month-over-month change** — compare previous month totals to the month before that (the `current` field in trends data, which on the 1st will be nearly empty, is not useful — instead compare the `previous` data against what you can infer)
3. **Budget scorecard** — for each category with a budget, did spending come in under or over? By how much?
4. **Biggest movers** — which categories changed the most vs the prior month (both up and down)?
5. **Patterns** — any recurring charges, seasonal spending, or trends worth noting?

### Step 3: Write the report

Write a JSON file to `/tmp/miser-report.json` with this exact structure:

```json
{
  "year": 2026,
  "month": 3,
  "narrative": "your narrative here"
}
```

Use the **completed** month's year and month (not the current date). The narrative should be ~200 words in this format:

```
**March 2026 Summary**

- Total spending: $3,245 across 87 transactions (vs $2,980 in Feb, +8.9%)
- Budget scorecard: 4/6 budgets hit. Groceries under by $50. Dining over by $120.
- Biggest increases: Dining +$120 (+35%), driven by 4 restaurant charges over $40
- Biggest decreases: Shopping -$200 (-45%), back to normal after February's one-time purchases
- Pattern: Subscription charges totaled $85 across 6 services

**Top 3 takeaways:**
1. Dining is the main budget miss — consider a weekly dining cap of $75
2. Grocery spending has been consistent for 3 months — current budget is well-calibrated
3. Total spending is trending up — March was the highest month this quarter
```

Key rules for the narrative:
- Lead with total spending and the month-over-month trend
- Budget scorecard should be a quick pass/fail, not a detailed breakdown of every category
- "Biggest movers" = categories with the largest absolute dollar change, not percentage
- Takeaways should be specific and actionable — reference actual numbers
- Amounts are negative for expenses in the data — present them as positive in the narrative for readability

### Step 4: Save the report

```bash
miser write-report /tmp/miser-report.json
```

Verify it was saved:

```bash
miser report
```

This should display the narrative you just wrote.
