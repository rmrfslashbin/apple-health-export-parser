# Build information
BINARY := apple-health-export-parser
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: help build test lint lint-fix vet staticcheck vulncheck clean build-all check

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build binary (default)
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/
	@echo "Build complete!"

test: ## Run tests with coverage
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@command -v staticcheck >/dev/null 2>&1 || { echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest"; exit 1; }
	staticcheck ./...

vulncheck: ## Check for known vulnerabilities
	@echo "Running govulncheck..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }
	govulncheck ./...

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. See: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. See: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run --fix

check: vet staticcheck test ## Run all checks (vet, staticcheck, test)
	@echo "All checks passed!"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/ build/ coverage.out coverage.html
	@echo "Clean complete!"

.DEFAULT_GOAL := build