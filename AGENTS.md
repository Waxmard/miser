# Repository Guidelines

## Project Structure & Module Organization

`cmd/miser/` contains the CLI entry point using cobra (includes `serve` and `daemon` commands). Application logic lives in `internal/`: `repository/` for database interfaces and implementations, `ingest/` for data sources (email, CSV, Monarch, SimpleFIN), `categorize/` for the rule engine, `process/` for Claude Code integration commands, `api/` for REST handlers serving the web UI, `report/` for reporting, and `config/` for TOML configuration loading. `web/` holds the Svelte + SvelteKit frontend (built with Bun) that is embedded into the Go binary via `web/embed.go`. Claude Code cron job documentation lives in `cron/`. CI workflows and Dependabot config are in `.github/`.

Keep command wiring in `cmd/`, business logic in `internal/`, and all SQL in `internal/repository/sqlite/`.

## Build, Test, and Development Commands

- `make build`: compile the binary to `bin/miser`.
- `make install`: install to `$GOPATH/bin`.
- `make web-build`: build the Svelte frontend with Bun.
- `make serve`: build frontend + binary and start the embedded web server.
- `make daemon`: run miser in daemon mode.
- `make sync`: run `miser sync` against all sources.
- `make test`: run all tests with verbose output.
- `make test-short`: run tests with `-short`.
- `make test-race`: run tests with the race detector.
- `make test-cover`: generate HTML coverage report.
- `make lint`: run golangci-lint using `.golangci.yml`.
- `make fmt`: format all Go files with goimports.
- `make vet`: run `go vet`.
- `make check`: run fmt, lint, vet, and test in sequence.
- `make tools`: install dev tooling (goimports, golangci-lint, lefthook) and set up git hooks.
- `make deps`: download and tidy Go module dependencies.
- `make docs`: regenerate command reference under `docs/commands/`.
- `make review` / `organize` / `weekly-report` / `monthly-report` / `budgets`: Claude Code cron tasks driven by prompts in `cron/`.

## Coding Style & Naming Conventions

Go standard formatting applies (tabs, gofmt). Use `goimports` for import organization. Linting rules are defined in `.golangci.yml` and enforced by lefthook pre-commit hooks and CI.

Use `PascalCase` for exported identifiers, `camelCase` for unexported. Package names are lowercase single words. Repository implementations go in sub-packages named after the driver (e.g., `sqlite`). Test files use `_test.go` suffix in the same package.

All entity IDs are ULIDs. Timestamps are ISO 8601 strings stored as TEXT. Amounts are float64 (negative = expense, positive = income).

## Testing Guidelines

Place test files alongside the code they test (`foo_test.go` next to `foo.go`). Use `testing.T` and table-driven tests. Repository tests should use SQLite `:memory:` databases for speed and isolation.

Run `make test` before opening a PR and `make test-cover` for larger refactors. CI runs `go test ./... -race` on every push and PR to main.

## Commit & Pull Request Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `perf:`, `test:`. Keep commit subjects short and imperative. Releases are automated via release-please.

PRs should explain the user-visible change, note any config or migration steps, and include example CLI output when behavior changes.

## Configuration Notes

Copy `config.example.toml` to `~/.miser/config.toml` and fill in credentials. The `config.toml` file is gitignored. Database defaults to `~/.miser/miser.db`.
