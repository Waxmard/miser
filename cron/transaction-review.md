# Task: Review and finalize pending transaction categorizations

You are reviewing transaction categorizations for a personal finance tracker. Approve correct categorizations and fix any that are wrong.

## Step 1: Gather pending transactions

```bash
miser internal process review
```

Returns JSON:

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
  "categories": [
    { "name": "Housing", "subcategories": ["Rent", "Parking", "Utilities"] },
    { "name": "Food", "subcategories": ["Groceries", "Restaurants"] },
    { "name": "Transportation" },
    { "name": "Flexible", "subcategories": ["Bars", "Entertainment"] }
  ]
}
```

If `pending_count` is 0, there is nothing to review — stop here and report "no pending transactions".

## Step 2: Review each transaction

For each transaction, decide:

- **approve** — the assigned category is correct
- **change** — the category is wrong; pick the correct one from the `categories` list

Guidelines:

- Trust high-confidence categorizations (≥ 0.85) unless obviously wrong
- Pay close attention to low-confidence ones (< 0.70) — these are most likely to be miscategorized
- Prefer the most specific subcategory (e.g. "Rent" over "Housing"), but use the parent if the subcategory is ambiguous
- Amounts are negative for expenses

## Step 3: Write your decisions

Write `/tmp/miser-review.json` matching this schema exactly. Every transaction from Step 1 must have an entry.

<output_schema>
{
  "results": [
    { "transaction_id": "01HXY...", "action": "approve" },
    { "transaction_id": "01HXZ...", "action": "change", "category": "Dining" }
  ]
}
</output_schema>

## Step 4: Persist and verify

```bash
miser internal write review /tmp/miser-review.json
miser internal process review
```

The verify command must show `pending_count` is 0. If it does not, stop and report the error.
