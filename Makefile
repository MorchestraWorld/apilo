.PHONY: help build install clean test run

# Variables
BINARY_NAME=api-optimizer
VERSION=1.0.0
BUILD_DIR=bin
INSTALL_PATH=$(HOME)/go/bin
SRC_DIR=./src

# Get build information
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
SOURCE_DIR=$(shell pwd)

# Build flags with injected variables
LDFLAGS=-ldflags="-w -s \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.Commit=$(GIT_COMMIT)' \
	-X 'main.SourceDir=$(SOURCE_DIR)'"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the api-optimizer binary
	@echo "🔨 Building $(BINARY_NAME) with embedded source path..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "📁 Source path embedded: $(SOURCE_DIR)"

install: build ## Build and install globally
	@echo "📦 Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Run '$(BINARY_NAME) --help' to get started"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v $(SRC_DIR)/...
	@echo "✅ Tests complete"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out $(SRC_DIR)/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

run: build ## Build and run locally
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt $(SRC_DIR)/...
	@echo "✅ Format complete"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

all: clean deps build test install ## Clean, deps, build, test, and install
	@echo "✅ All tasks complete!"

# Default target
.DEFAULT_GOAL := help
