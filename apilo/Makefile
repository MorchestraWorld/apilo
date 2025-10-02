.PHONY: help build install clean test run dev lint fmt deps

# Variables
BINARY_NAME=apilo
VERSION=2.0.0
BUILD_DIR=bin
INSTALL_PATH=$(HOME)/go/bin

# Build flags
LDFLAGS=-ldflags="-w -s"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the CLI binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Build and install globally
	@echo "📦 Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@go install $(LDFLAGS)
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Run 'apilo --help' to get started"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Clean complete"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests complete"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

run: build ## Build and run locally
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode (no build)
	@echo "🚀 Running in dev mode..."
	@go run main.go

lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run || echo "⚠️  golangci-lint not installed. Run: brew install golangci-lint"

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

verify: lint test ## Verify code quality and tests
	@echo "✅ Verification complete"

release: clean verify build ## Build release version
	@echo "📦 Creating release v$(VERSION)..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_NAME) release/$(BINARY_NAME)-v$(VERSION)
	@tar -czf release/$(BINARY_NAME)-v$(VERSION).tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	@echo "✅ Release created: release/$(BINARY_NAME)-v$(VERSION).tar.gz"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "✅ Docker image built: $(BINARY_NAME):$(VERSION)"

all: clean deps build test install ## Clean, deps, build, test, and install
	@echo "✅ All tasks complete!"

# Default target
.DEFAULT_GOAL := help
