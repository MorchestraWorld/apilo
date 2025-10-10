.PHONY: help build install clean test run daemon-build daemon-install daemon-hooks daemon-clean

# Variables
BINARY_NAME=api-optimizer
CLI_NAME=apilo
VERSION=2.0.0
BUILD_DIR=bin
INSTALL_PATH=$(HOME)/go/bin
SRC_DIR=./src
CLI_DIR=./apilo
HOOKS_DIR=$(CLI_DIR)/hooks
CLAUDE_HOOKS_DIR=$(HOME)/.claude/hooks
APILO_CONFIG_DIR=$(HOME)/.apilo

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

# Daemon-specific build flags
DAEMON_LDFLAGS=-ldflags="-w -s \
	-X 'apilo/internal/build.Version=$(VERSION)' \
	-X 'apilo/internal/build.BuildTime=$(BUILD_TIME)' \
	-X 'apilo/internal/build.Commit=$(GIT_COMMIT)' \
	-X 'apilo/internal/build.SourceDir=$(SOURCE_DIR)'"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the api-optimizer binary
	@echo "🔨 Building $(BINARY_NAME) with embedded source path..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "📁 Source path embedded: $(SOURCE_DIR)"

daemon-build: ## Build the apilo CLI with daemon support
	@echo "🔨 Building $(CLI_NAME) CLI with daemon support..."
	@mkdir -p $(BUILD_DIR)
	@cd $(CLI_DIR) && go build $(DAEMON_LDFLAGS) -o ../$(BUILD_DIR)/$(CLI_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(CLI_NAME)"
	@echo "📁 Daemon support: enabled"
	@echo "📋 Features: daemon, claude, install, benchmark, monitor"

install: build ## Build and install api-optimizer globally
	@echo "📦 Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Run '$(BINARY_NAME) --help' to get started"

daemon-install: daemon-build ## Build and install apilo CLI globally
	@echo "📦 Installing $(CLI_NAME) to $(INSTALL_PATH)..."
	@cd $(CLI_DIR) && go install $(DAEMON_LDFLAGS)
	@echo "✅ Installation complete!"
	@echo ""
	@echo "Available commands:"
	@echo "  apilo daemon start   - Start background daemon"
	@echo "  apilo daemon stop    - Stop daemon"
	@echo "  apilo daemon status  - Check daemon status"
	@echo "  apilo claude setup   - Setup Claude Code integration"
	@echo "  apilo install        - Install globally"
	@echo ""
	@echo "Next: make daemon-hooks (to install Claude Code hooks)"

daemon-hooks: ## Install Claude Code hooks for automatic optimization
	@echo "🪝 Installing Claude Code hooks..."
	@mkdir -p $(CLAUDE_HOOKS_DIR)
	@mkdir -p $(APILO_CONFIG_DIR)/logs
	@cp $(HOOKS_DIR)/apilo-optimizer.sh $(CLAUDE_HOOKS_DIR)/
	@chmod +x $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh
	@echo "✅ Hook installed: $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh"
	@echo ""
	@echo "📝 Hook enables automatic API optimization in Claude Code"
	@echo ""
	@echo "To activate:"
	@echo "  1. Start daemon: apilo daemon start"
	@echo "  2. Use Claude Code normally - API calls auto-optimized"
	@echo "  3. Check status: apilo daemon status"
	@echo ""
	@echo "Hook detects and optimizes:"
	@echo "  • API calls to https://api.*"
	@echo "  • Endpoints with /api/, /v1/, etc."
	@echo "  • Common API URL patterns"

daemon-status: ## Check if daemon is running and show metrics
	@echo "📊 Checking daemon status..."
	@apilo daemon status || echo "⚠️  Daemon not running. Start with: apilo daemon start"

daemon-start: daemon-install ## Install and start the daemon
	@echo "🚀 Starting apilo daemon..."
	@apilo daemon start
	@sleep 2
	@make daemon-status

daemon-stop: ## Stop the running daemon
	@echo "🛑 Stopping apilo daemon..."
	@apilo daemon stop

daemon-restart: ## Restart the daemon
	@echo "🔄 Restarting apilo daemon..."
	@apilo daemon restart

daemon-logs: ## View daemon logs
	@apilo daemon logs

daemon-test: ## Test daemon functionality
	@echo "🧪 Testing daemon functionality..."
	@echo ""
	@echo "1. Building daemon..."
	@make daemon-build
	@echo ""
	@echo "2. Checking daemon status..."
	@$(BUILD_DIR)/$(CLI_NAME) daemon status || echo "✓ Daemon not running (expected)"
	@echo ""
	@echo "3. Testing daemon commands..."
	@$(BUILD_DIR)/$(CLI_NAME) daemon --help | head -10
	@echo ""
	@echo "✅ Daemon commands functional"
	@echo ""
	@echo "To test fully:"
	@echo "  make daemon-start    # Start daemon"
	@echo "  make daemon-status   # Check status"
	@echo "  make daemon-stop     # Stop daemon"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

daemon-clean: clean ## Clean all daemon artifacts including PID and logs
	@echo "🧹 Cleaning daemon artifacts..."
	@rm -f $(APILO_CONFIG_DIR)/daemon.pid
	@rm -rf $(APILO_CONFIG_DIR)/logs/*
	@echo "✅ Daemon artifacts cleaned"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v $(SRC_DIR)/...
	@echo "✅ Tests complete"

daemon-unit-test: ## Run daemon unit tests (TODO)
	@echo "🧪 Running daemon unit tests..."
	@cd $(CLI_DIR) && go test -v ./internal/daemon/...
	@echo "✅ Daemon tests complete"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out $(SRC_DIR)/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

run: build ## Build and run api-optimizer locally
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

daemon-run: daemon-build ## Build and run apilo CLI
	@echo "🚀 Running $(CLI_NAME)..."
	@./$(BUILD_DIR)/$(CLI_NAME)

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt $(SRC_DIR)/...
	@cd $(CLI_DIR) && go fmt ./...
	@echo "✅ Format complete"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@cd $(CLI_DIR) && go mod download && go mod tidy
	@echo "✅ Dependencies updated"

all: clean deps build test install ## Clean, deps, build, test, and install
	@echo "✅ All tasks complete!"

daemon-all: daemon-clean deps daemon-build daemon-install daemon-hooks ## Complete daemon setup
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║           APILO DAEMON INSTALLATION COMPLETE                      ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "✅ Daemon binary installed: $(INSTALL_PATH)/$(CLI_NAME)"
	@echo "✅ Claude Code hooks installed: $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh"
	@echo "✅ Configuration directory: $(APILO_CONFIG_DIR)"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  1. apilo daemon start         # Start background daemon"
	@echo "  2. apilo daemon status        # Verify it's running"
	@echo "  3. Use Claude Code normally   # API calls auto-optimized"
	@echo ""
	@echo "📊 Monitoring:"
	@echo "  apilo daemon logs             # View daemon logs"
	@echo "  apilo daemon status           # Check performance metrics"
	@echo "  curl http://localhost:9876/metrics  # Detailed metrics"
	@echo ""
	@echo "📚 Documentation:"
	@echo "  apilo daemon --help           # Daemon commands"
	@echo "  apilo claude setup            # Claude Code integration"
	@echo "  cat apilo/DAEMON.md           # Full daemon documentation"
	@echo ""

daemon-uninstall: daemon-stop ## Uninstall daemon and hooks
	@echo "🗑️  Uninstalling apilo daemon..."
	@rm -f $(INSTALL_PATH)/$(CLI_NAME)
	@rm -f $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh
	@rm -f $(APILO_CONFIG_DIR)/daemon.pid
	@echo "✅ Uninstall complete"
	@echo ""
	@echo "Configuration preserved in: $(APILO_CONFIG_DIR)"
	@echo "To fully remove: rm -rf $(APILO_CONFIG_DIR)"

# Default target
.DEFAULT_GOAL := help
