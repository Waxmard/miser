#!/usr/bin/env bash
set -euo pipefail

JOB="${1:?usage: run.sh <job-name>}"
DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${MISER_CRON_LOG_DIR:-$HOME/.miser-cron}"
mkdir -p "$LOG_DIR"
LOG="$LOG_DIR/${JOB}-$(date +%Y%m%d-%H%M%S).log"

PROMPT_FILE="$DIR/$JOB.md"
[[ -f "$PROMPT_FILE" ]] || { echo "unknown job: $JOB" >&2; exit 2; }

case "$JOB" in
  budget-suggestions) PROMPT="$(cat "$PROMPT_FILE")" ;;
  *)                  PROMPT="$(cat "$DIR/_header.md" "$PROMPT_FILE")" ;;
esac

command -v claude >/dev/null || { echo "claude CLI not found" >&2; exit 127; }
command -v miser  >/dev/null || { echo "miser CLI not found"  >&2; exit 127; }

exec timeout 300 claude -p "$PROMPT" \
  --model sonnet \
  --allowedTools "Bash,Read,Write" \
  --verbose 2>&1 | tee "$LOG"
