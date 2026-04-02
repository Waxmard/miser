# SimpleFIN Setup

SimpleFIN pulls transactions from your bank accounts automatically via the SimpleFIN Bridge API. It supports 16,000+ US/Canadian financial institutions through MX.

## 1. Subscribe

Go to https://beta-bridge.simplefin.org and create an account.

Pricing: **$1.50/month** or **$15/year**.

## 2. Link your institutions

After subscribing, add each bank or card you want to sync. You can link up to 25 institutions.

Before linking, verify your institution is supported:
https://beta-bridge.simplefin.org/search-institutions

Institutions to link:

| Institution | Accounts |
|---|---|
| Capital One | 360 Checking, Quicksilver |
| Bilt | World Elite Mastercard |
| Synchrony (Amazon) | Amazon Store Card |
| Synchrony (Verizon) | Verizon Visa Signature |

Add any additional accounts as needed. Fidelity is handled separately via Gmail IMAP sync.

## 3. Get a setup token

In the SimpleFIN Bridge dashboard, click **"Create SimpleFIN Token"**. This generates a one-time-use token.

Copy the full token string (it's a long base64-encoded value).

## 4. Claim the token

```bash
miser setup simplefin <paste-token-here>
```

This exchanges the token for a permanent access URL and saves it to `~/.miser/config.toml`.

If you see "setup token already claimed or invalid", generate a new token in the Bridge dashboard. Tokens can only be claimed once.

## 5. Verify

Pull your initial transactions:

```bash
miser sync simplefin
```

You should see output like:

```
Syncing from SimpleFIN...
Synced 4 accounts
Found 120 transactions, stored 85 new
Auto-categorized 30 transactions via rules
```

Check that your accounts and transactions look right:

```bash
miser accounts
miser transactions --limit 20
miser categories
```

## 6. Ongoing usage

Run `miser sync` to pull from all enabled sources (SimpleFIN + email):

```bash
miser sync
```

Or sync just SimpleFIN:

```bash
miser sync simplefin
```

SimpleFIN updates once per day per institution. Running sync more frequently is fine (duplicates are ignored), but won't produce new data until the next daily refresh.

For automated daily syncs, add a cron job:

```bash
# Example: sync every day at 8am
0 8 * * * /path/to/miser sync >> ~/.miser/sync.log 2>&1
```

## Troubleshooting

| Error | Fix |
|---|---|
| "access denied" | Access URL may be revoked. Generate a new setup token and re-run `miser setup simplefin`. |
| "SimpleFIN subscription expired" | Renew at https://beta-bridge.simplefin.org |
| Missing accounts | Check that the institution is linked in the SimpleFIN Bridge dashboard. |
| Duplicate transactions | Safe to ignore. Miser deduplicates by transaction ID automatically. |
