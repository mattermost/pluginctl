# pluginctl Makefile
# Based on common Go project patterns

# Build information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build variables
GO_VERSION ?= $(shell awk '/^go / {print $$2}' go.mod)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Tool versions
GOLANGCI_LINT_VERSION ?= v1.64.8
GORELEASER_VERSION ?= v2.10.2

# Project variables
BINARY_NAME = pluginctl
MAIN_PACKAGE = ./cmd/pluginctl
DIST_DIR = ./dist
BIN_DIR = ./build/bin

# Build flags
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Default target
.PHONY: all
all: clean lint test build

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Clean build artifacts
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@rm -f $(BINARY_NAME)

# Install dependencies
.PHONY: deps
deps: ## Install/update dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run tests
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint code
.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@$(BIN_DIR)/golangci-lint run

# Fix linting issues
.PHONY: lint-fix
lint-fix: ## Fix linting issues
	@echo "Fixing linting issues..."
	@$(BIN_DIR)/golangci-lint run --fix

# Install binary
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PACKAGE)

# Run the application
.PHONY: run
run: ## Run the application
	@go run $(MAIN_PACKAGE) $(ARGS)

# Format code
.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

# Generate code
.PHONY: generate
generate: ## Generate code
	@echo "Generating code..."
	@go generate ./...

# Check for updates
.PHONY: check-updates
check-updates: ## Check for dependency updates
	@echo "Checking for dependency updates..."
	@go list -u -m all

# Release (requires goreleaser)
.PHONY: release
release: ## Create a release
	@echo "Creating release..."
	@$(BIN_DIR)/goreleaser release --clean

# Snapshot release (for testing)
.PHONY: snapshot
snapshot: ## Create a snapshot release
	@echo "Creating snapshot release..."
	@$(BIN_DIR)/goreleaser release --snapshot --clean

# Development setup
.PHONY: dev-setup
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@mkdir -p $(BIN_DIR)
	@if [ ! -f "$(BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION); \
		mv $(BIN_DIR)/golangci-lint $(BIN_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION); \
	else \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) already installed"; \
	fi
	@ln -sf golangci-lint-$(GOLANGCI_LINT_VERSION) $(BIN_DIR)/golangci-lint
	@if [ ! -f "$(BIN_DIR)/goreleaser-$(GORELEASER_VERSION)" ]; then \
		echo "Installing goreleaser $(GORELEASER_VERSION)..."; \
		GOBIN=$(PWD)/$(BIN_DIR) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION); \
		mv $(BIN_DIR)/goreleaser $(BIN_DIR)/goreleaser-$(GORELEASER_VERSION); \
	else \
		echo "goreleaser $(GORELEASER_VERSION) already installed"; \
	fi
	@ln -sf goreleaser-$(GORELEASER_VERSION) $(BIN_DIR)/goreleaser
	@echo "Development tools installed in $(BIN_DIR)"

# Verify build
.PHONY: verify
verify: clean lint test snapshot ## Verify build (clean, lint, test, snapshot)
	@echo "Build verification complete"

# Quick development build
.PHONY: dev
dev: fmt lint snapshot ## Quick development build (fmt, lint, snapshot)

# Check changes target
.PHONY: check-changes
check-changes: lint test ## Check changes (lint, test)
	@echo "All checks passed!"

# CI target
.PHONY: ci
ci: deps verify ## CI target (deps, verify)

# Print build info
.PHONY: version
version: ## Print version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "Tool Versions:"
	@echo "  golangci-lint: $(GOLANGCI_LINT_VERSION)"
	@echo "  goreleaser: $(GORELEASER_VERSION)"
