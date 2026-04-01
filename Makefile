.PHONY: build run test lint fmt clean init sync help

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

daemon: build                   ## Daemon mode
	./bin/miser daemon

fmt:                            ## Format Go files
	goimports -w .

lint:                           ## Lint
	golangci-lint run ./...

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
	lefthook install

clean:                          ## Clean
	rm -rf bin/ coverage.out coverage.html

help:                           ## Help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
