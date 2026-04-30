# Miser Claude Code Jobs

This directory contains prompt files invoked by [Claude Code](https://docs.anthropic.com/claude-code) — some scheduled via cron, others run manually.

Each prompt file is **pure task content**. Autonomous jobs prepend [`_header.md`](./_header.md) at invocation time so the model treats the file as instructions to execute (not documentation about a cron job).

## Jobs

| Job | File | Mode | Schedule |
|-----|------|------|----------|
| Monthly report | [`monthly-report.md`](./monthly-report.md) | autonomous | 1st of every month at 9am |
| Weekly report | [`weekly-report.md`](./weekly-report.md) | autonomous | Every Monday at 9am |
| Transaction review | [`transaction-review.md`](./transaction-review.md) | autonomous | Bi-weekly Monday at 9am, or manual |
| Category hierarchy | [`category-hierarchy.md`](./category-hierarchy.md) | autonomous | Manual (after Monarch import) |
| Budget suggestions | [`budget-suggestions.md`](./budget-suggestions.md) | interactive | Manual |

## Invocation

All jobs run through [`run.sh`](./run.sh), which assembles the prompt (`_header.md` + `<job>.md` for autonomous, raw file for `budget-suggestions`), pre-approves tools, applies a 5-minute timeout, and tees output to a log file.

```bash
cron/run.sh monthly-report
cron/run.sh weekly-report
cron/run.sh transaction-review
cron/run.sh budget-suggestions
cron/run.sh category-hierarchy
```

Logs land at `~/.miser-cron/<job>-<timestamp>.log`. Override the directory with `MISER_CRON_LOG_DIR`.

## Crontab

```crontab
# 1st of every month at 9am — monthly spending report
0 9 1 * * /path/to/miser/cron/run.sh monthly-report

# Every Monday at 9am — weekly snapshot
0 9 * * 1 /path/to/miser/cron/run.sh weekly-report

# Every other Monday at 9am — transaction review
0 9 1-7,15-21 * 1 /path/to/miser/cron/run.sh transaction-review
```

Replace `/path/to/miser` with the absolute path to your checkout.

## Flags (set inside `run.sh`)

- `--model sonnet` — Sonnet follows multi-step instructions reliably; Haiku tends to ask clarifying questions instead of executing
- `--allowedTools "Bash,Read,Write"` — pre-approves tools so there are no interactive permission prompts
- `--verbose` — surfaces each tool call in the log
- `timeout 300` — kills the job after 5 minutes if Claude hangs

## Adding a new job

1. Drop a `your-job.md` file into this directory containing only the task body (start with `# Task: ...`, no schedule/doc prose at the top).
2. If interactive, add the file name to the `case` block in `run.sh` so it skips `_header.md`.
3. Add a row to the table above and a crontab line if scheduled.
