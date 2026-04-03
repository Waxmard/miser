# Weekly Spending Report

Claude Code cron job that generates a weekly spending snapshot every Monday.

## Schedule

```crontab
# Every Monday at 9am
0 9 * * 1 claude --bare -p "$(cat /path/to/miser/cron/weekly-report.md)" --model haiku --allowedTools "Bash,Read,Write"
```

Flags:
- `--bare` — skips hooks, MCP servers, CLAUDE.md, auto-memory for reliable unattended execution
- `--model haiku` — Haiku is sufficient for this task and much faster/cheaper than Opus
- `--allowedTools "Bash,Read,Write"` — pre-approves tools so there are no interactive permission prompts

## Prompt

You are a personal finance analyst for a single user. Your job is to generate a brief, actionable weekly spending snapshot.

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

Also run to get this week's transactions:

```bash
miser transactions --from $(date -v-7d +%Y-%m-%d) --to $(date -v-1d +%Y-%m-%d) --limit 50
```

### Step 2: Analyze

From the data, determine:

1. **Top spending categories this week** — which categories had the most activity in the last 7 days?
2. **Budget pacing** — for each category with a budget, are we on track for the month? (Compare current spending to `budget * (day_of_month / days_in_month)`)
3. **Notable transactions** — any unusually large purchases, new merchants, or unexpected charges?
4. **Month-over-month context** — how does this month's pace compare to last month?

### Step 3: Write the report

Write a JSON file to `/tmp/miser-report.json` with this exact structure:

```json
{
  "year": 2026,
  "month": 4,
  "narrative": "your narrative here"
}
```

Use the current year and month. The narrative should be ~200 words in this format:

```
**Week of Apr 7–13, 2026**

- Top spending: Groceries ($125, 4 txns), Dining ($85, 3 txns), Gas ($45, 1 txn)
- Budget pacing: Groceries at 65% of $600 budget with 40% of month remaining — slightly ahead. Dining on track.
- Notable: $85 charge at [merchant] is higher than your typical dining transaction
- vs. last month: Total spending is running 12% lower than March at this point

**Action:** Keep an eye on grocery spending — you're trending $50 over last month's pace.
```

Key rules for the narrative:
- Lead with the most important insight, not a generic summary
- Use actual numbers — don't say "spending increased", say "spending increased $45 (+12%)"
- Only mention categories that have meaningful activity
- The action item should be specific and based on the data, not generic advice
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
