.PHONY: build test lint clean help

BINARY_NAME=mockctl
MAIN_PATH=./cmd/mockctl

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the CLI binary
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

test: ## Run unit tests with race detector
	go test -race -v ./...

test-integration: ## Run slow integration tests
	go test -race -tags=integration -v ./...

lint: ## Run golangci-lint
	golangci-lint run

clean: ## Remove build artifacts
	rm -rf bin/
	go clean -testcache
