# miser

Personal finance CLI that replaces [Monarch Money](https://www.monarchmoney.com/). Aggregates transactions from multiple financial accounts, auto-categorizes them using [Claude Code](https://claude.ai/code) cron jobs, and surfaces spending trends.

## Features

- **Gmail IMAP ingestion** — polls for Fidelity transaction alert emails
- **Rule-based categorization** — 176 pre-seeded merchant-to-category rules from Monarch history
- **AI categorization** — Claude Code cron jobs handle unknown merchants (zero API cost with Claude Max)
- **Monarch Money migration** — one-time import of ~4,700 transactions, 54 categories, 12 accounts
- **Spending trends** — month-over-month comparisons with budget tracking
- **Weekly reports** — Claude-generated narrative summaries

## Quick start

```bash
# Install dev tools (golangci-lint, goimports, lefthook)
make tools

# Install to $GOPATH/bin (usually ~/go/bin)
go install ./cmd/miser

# Initialize config and database
miser init

# Import Monarch Money history (one-time)
./bin/miser import-monarch ~/Downloads/monarch-transactions.csv

# Configure Gmail (edit ~/.miser/config.toml with your app password)
# Then sync emails
./bin/miser sync email
```

## Commands

```
miser init                              Create config, DB, seed categories
miser import-monarch <csv>              One-time Monarch Money migration

miser sync                              Run all enabled sync sources
miser sync email                        Poll Gmail for new transaction emails

miser transactions [--from] [--to]      List transactions (aliases: txns, tx)
miser categories                        List categories with counts
miser rules                             List categorization rules
miser accounts                          List accounts
miser trends                            Monthly spending trends
miser report                            Latest Claude narrative report

miser process emails                    Print pending emails as JSON (for Claude Code)
miser process categorize                Print uncategorized txns as JSON
miser process trends                    Print monthly data as JSON
miser write-parsed <json>               Write Claude's parse results to DB
miser write-categories <json>           Write Claude's categorizations to DB
miser write-report <json>               Write Claude's report to DB
```

## How it works

The system has two halves:

1. **Go CLI** — deterministic operations: email polling, rule-based categorization, queries, imports
2. **Claude Code cron jobs** — AI-powered: email parsing, transaction categorization, trend narratives

They communicate through the shared SQLite database.

```
Email arrives
  -> miser sync email (stores raw email)
  -> Claude Code cron: parse email -> miser write-parsed
  -> Rule engine auto-categorizes known merchants
  -> Claude Code cron: categorize unknowns -> miser write-categories
  -> miser transactions (query)
```

## Configuration

Copy `config.example.toml` to `~/.miser/config.toml`:

```toml
[database]
driver = "sqlite"
sqlite_path = "~/.miser/miser.db"

[email]
enabled = true
imap_server = "imap.gmail.com"
imap_port = 993
username = "you@gmail.com"
app_password = ""           # Gmail App Password
label = "Finance/Fidelity"
poll_interval_minutes = 15
```

### Gmail setup

1. Enable IMAP in Gmail settings
2. Generate an App Password at https://myaccount.google.com/apppasswords
3. Create a Gmail filter: `from:(fidelity.com OR elanfinancial.com) subject:(transaction OR alert)` -> Apply label `Finance/Fidelity`, skip inbox
4. Add the app password to `~/.miser/config.toml`

## Tech stack

- **Go** with SQLite (designed for future Postgres migration)
- **Repository pattern** — all DB access through interfaces, swappable backends
- **Claude Code cron jobs** — zero-cost AI via Claude Max subscription
- **lipgloss** — terminal styling

## Development

```bash
make build          # Build binary to bin/
make test           # Run tests
make lint           # Run golangci-lint
make check          # fmt + lint + vet + test
make help           # Show all targets
```
