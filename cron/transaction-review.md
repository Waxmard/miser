# Transaction Review

Claude Code workflow to review and approve pending transaction categorizations. Run manually whenever needed, or set up a bi-weekly cron.

## Schedule

**Manual invocation (recommended):**

```bash
claude -p "$(cat /path/to/miser/cron/transaction-review.md)" --model sonnet --allowedTools "Bash,Read,Write"
```

**Or bi-weekly cron (every other Monday at 9am):**

```crontab
# Every other Monday at 9am (bi-weekly)
0 9 1-7,15-21 * 1 claude -p "Follow the instructions below exactly. Execute each step in order. Do not ask questions. $(cat /path/to/miser/cron/transaction-review.md)" --model sonnet --allowedTools "Bash,Read,Write"
```

Flags:
- `--model sonnet` — Sonnet follows multi-step instructions reliably; Haiku tends to ask clarifying questions instead of executing
- `--allowedTools "Bash,Read,Write"` — pre-approves tools so there are no interactive permission prompts

Note: `--bare` is not used because it strips authentication context needed for the CLI.

## Prompt

You are reviewing transaction categorizations for a personal finance tracker. Your job is to approve correct categorizations and fix any that are wrong.

### Step 1: Gather pending transactions

```bash
miser internal process review
```

This returns JSON with the structure:

```json
{
  "pending_count": 8,
  "transactions": [
    {
      "id": "01HXY...",
      "merchant": "LYFT *RIDE",
      "merchant_clean": "Lyft",
      "amount": -24.50,
      "date": "2026-04-08",
      "category": "Transportation",
      "confidence": 0.72,
      "description": "LYFT *RIDE SAN FRANCISCO"
    }
  ],
  "categories": ["Groceries", "Transportation", "Dining", "Housing", ...]
}
```

If `pending_count` is 0, there is nothing to review — stop here.

### Step 2: Review each transaction

For each transaction, decide:

- **approve** — the assigned category is correct
- **change** — the category is wrong; pick the correct one from the `categories` list

Use these guidelines:
- Trust high-confidence categorizations (≥ 0.85) unless obviously wrong
- Pay close attention to low-confidence ones (< 0.70) — these are most likely to be miscategorized
- If a transaction could fit multiple categories, pick the most specific one
- Amounts are negative for expenses

### Step 3: Write your decisions

Write a JSON file to `/tmp/miser-review.json` with this exact structure:

```json
{
  "results": [
    { "transaction_id": "01HXY...", "action": "approve" },
    { "transaction_id": "01HXZ...", "action": "change", "category": "Dining" }
  ]
}
```

Every transaction from Step 1 must have an entry — either `"approve"` or `"change"`.

### Step 4: Save the decisions

```bash
miser internal write review /tmp/miser-review.json
```

Verify no transactions remain pending:

```bash
miser internal process review
```

The `pending_count` should be 0.
