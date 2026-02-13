.PHONY: help build build-cli run test lint fmt vet clean install deps tidy docker-build docker-up docker-down

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=magento
CLI_BINARY_NAME=cli
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

help: ## Display this help message
	@echo "GoGento - Magento Go API"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the main server binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) magento.go

build-cli: ## Build the CLI binary
	@echo "Building $(CLI_BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CLI_BINARY_NAME) cli.go

build-all: build build-cli ## Build all binaries

run: ## Run the main server (development mode)
	@echo "Running $(BINARY_NAME)..."
	$(GO) run magento.go

run-cli: ## Run the CLI
	@echo "Running $(CLI_BINARY_NAME)..."
	$(GO) run cli.go

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

fmt: ## Format code with go fmt
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

check: fmt vet ## Run fmt and vet

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(CLI_BINARY_NAME)
	rm -f coverage.txt coverage.html
	rm -rf dist/ build/
	$(GO) clean

install: ## Install dependencies
	@echo "Installing dependencies..."
	$(GO) mod download

deps: install ## Alias for install

tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	$(GO) mod verify

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t gogento:latest .

docker-up: ## Start services with docker-compose
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down: ## Stop services with docker-compose
	@echo "Stopping Docker services..."
	docker-compose down

dev: ## Run development server with auto-reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install it with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

all: clean tidy check test build-all ## Run all checks and build everything
