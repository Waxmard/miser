.PHONY: build run dev dev-web dev-api test lint fmt clean init sync help review organize weekly-report monthly-report budgets

build:                          ## Build the miser binary
	go build -o bin/miser ./cmd/miser

run: build                      ## Build and run (ARGS="sync email")
	./bin/miser $(ARGS)

install:                        ## Install to $GOPATH/bin
	go install ./cmd/miser

init: build                     ## First-time setup
	./bin/miser init

sync: build                     ## Sync all sources
	./bin/miser sync

dev:                            ## Hot-reload dev (run `make dev-web` and `make dev-api` in 2 terminals)
	@echo "Run 'make dev-web' and 'make dev-api' in separate terminals"

dev-web:                        ## Vite dev server with HMR (port 5173)
	cd web && bun run dev

dev-api:                        ## Go API with air hot reload (port 8090)
	air

daemon: build                   ## Daemon mode
	./bin/miser daemon

fmt:                            ## Format Go files
	goimports -w .

lint:                           ## Lint
	golangci-lint run ./cmd/... ./internal/...

vet:                            ## Vet
	go vet ./...

check: fmt lint vet test        ## All checks

test:                           ## Tests
	go test ./... -v

test-short:                     ## Short tests
	go test ./... -short

test-race:                      ## Tests + race detector
	go test ./... -race -count=1

test-cover:                     ## Coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

deps:                           ## Download deps
	go mod download && go mod tidy

tools:                          ## Install dev tools
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/evilmartians/lefthook@latest
	go install github.com/air-verse/air@latest
	lefthook install

review:                         ## Review pending transaction categorizations with Claude
	claude -p "$$(cat cron/transaction-review.md)" --model sonnet --allowedTools "Bash,Read,Write"

organize:                       ## Organize categories into a hierarchy with Claude
	claude -p "Follow the instructions below exactly. Execute each step in order. Do not ask questions. $$(cat cron/category-hierarchy.md)" --model sonnet --allowedTools "Bash,Read,Write"

weekly-report:                  ## Generate weekly spending report with Claude
	claude -p "$$(cat cron/weekly-report.md)" --model sonnet --allowedTools "Bash,Read,Write"

monthly-report:                 ## Generate monthly spending report with Claude
	claude -p "$$(cat cron/monthly-report.md)" --model sonnet --allowedTools "Bash,Read,Write"

budgets:                        ## Generate budget suggestions with Claude
	claude -p "$$(cat cron/budget-suggestions.md)" --model sonnet --allowedTools "Bash,Read,Write"

docs: build                     ## Generate command reference docs
	./bin/miser gen-docs docs/commands/

clean:                          ## Clean
	rm -rf bin/ coverage.out coverage.html

help:                           ## Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
