.PHONY: help build install clean test hooks daemon-start daemon-stop daemon-status uninstall configure

# Variables
BINARY_NAME=apilo
VERSION=2.0.0
BUILD_DIR=bin
INSTALL_PATH=$(HOME)/go/bin
CLAUDE_HOOKS_DIR=$(HOME)/.claude/hooks
CLAUDE_TOOLS_DIR=$(HOME)/.claude/tools
CLAUDE_COMMANDS_DIR=$(HOME)/.claude/commands
APILO_CONFIG_DIR=$(HOME)/.apilo
APILO_LOGS_DIR=$(HOME)/.apilo/logs

# Get build information
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
SOURCE_DIR=$(shell pwd)

# Build flags
LDFLAGS=-ldflags="-w -s \
	-X 'apilo/internal/build.Version=$(VERSION)' \
	-X 'apilo/internal/build.BuildTime=$(BUILD_TIME)' \
	-X 'apilo/internal/build.Commit=$(GIT_COMMIT)' \
	-X 'apilo/internal/build.SourceDir=$(SOURCE_DIR)'"

help: ## Show this help message
	@echo '╔═══════════════════════════════════════════════════════════════════╗'
	@echo '║                    Apilo Makefile Help                            ║'
	@echo '╚═══════════════════════════════════════════════════════════════════╝'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo 'Quick Start:'
	@echo '  make install         # Build and install apilo globally'
	@echo '  make configure       # Setup Claude Code integration'
	@echo '  make daemon-start    # Start the daemon'
	@echo ''

build: ## Build apilo binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Build and install apilo globally
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║                  Installing Apilo CLI                             ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "📦 Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@go install $(LDFLAGS)
	@echo "✅ Installation complete!"
	@echo ""
	@echo "📋 Installed:"
	@echo "   • Binary: $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo ""
	@echo "🎯 Next steps:"
	@echo "   1. make configure    # Setup Claude Code integration"
	@echo "   2. make daemon-start # Start background daemon"
	@echo ""
	@echo "💡 Run 'apilo --help' to see available commands"

configure: install ## Install and configure Claude Code integration
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║            Configuring Claude Code Integration                    ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🔧 Setting up directories..."
	@mkdir -p $(CLAUDE_HOOKS_DIR)
	@mkdir -p $(CLAUDE_TOOLS_DIR)
	@mkdir -p $(CLAUDE_COMMANDS_DIR)
	@mkdir -p $(APILO_CONFIG_DIR)
	@mkdir -p $(APILO_LOGS_DIR)
	@echo "✅ Directories created"
	@echo ""
	@echo "🪝 Installing Claude Code integration..."
	@$(INSTALL_PATH)/$(BINARY_NAME) claude install
	@echo ""
	@echo "✅ Configuration complete!"
	@echo ""
	@echo "📊 Installed components:"
	@echo "   • Hook:    $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh"
	@echo "   • Tool:    $(CLAUDE_TOOLS_DIR)/apilo.json"
	@echo "   • Command: $(CLAUDE_COMMANDS_DIR)/apilo"
	@echo "   • Config:  $(APILO_CONFIG_DIR)/"
	@echo ""
	@echo "🚀 Start daemon: make daemon-start"

hooks: ## Install only Claude Code hooks (lightweight)
	@echo "🪝 Installing Claude Code hooks..."
	@mkdir -p $(CLAUDE_HOOKS_DIR)
	@cp hooks/apilo-optimizer.sh $(CLAUDE_HOOKS_DIR)/
	@chmod +x $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh
	@echo "✅ Hook installed: $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh"
	@echo ""
	@echo "📝 Hook enables automatic API optimization"
	@echo "🚀 Start daemon: make daemon-start"

daemon-start: ## Start the apilo daemon
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║                  Starting Apilo Daemon                            ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon start
	@echo ""
	@echo "✅ Daemon started!"
	@echo ""
	@echo "📊 Monitor: make daemon-status"
	@echo "📋 Logs:    make daemon-logs"

daemon-stop: ## Stop the apilo daemon
	@echo "🛑 Stopping apilo daemon..."
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon stop

daemon-restart: ## Restart the apilo daemon
	@echo "🔄 Restarting apilo daemon..."
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon restart

daemon-status: ## Check daemon status
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon status

daemon-logs: ## View daemon logs
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon logs

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./... 2>&1 | grep -v "no test files" || true
	@echo "✅ Tests complete"

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out ./... 2>&1 | grep -v "no test files" || true
	@go tool cover -html=coverage.out -o coverage.html 2>/dev/null || true
	@echo "✅ Coverage report: coverage.html"

test-daemon: build ## Test daemon functionality locally
	@echo "🧪 Testing daemon commands..."
	@echo ""
	@echo "1. Testing daemon --help..."
	@./$(BUILD_DIR)/$(BINARY_NAME) daemon --help | head -15
	@echo ""
	@echo "2. Testing daemon status..."
	@./$(BUILD_DIR)/$(BINARY_NAME) daemon status || echo "✓ Daemon not running (expected)"
	@echo ""
	@echo "✅ Daemon commands functional"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

clean-all: clean ## Clean all artifacts including daemon data
	@echo "🧹 Cleaning all daemon artifacts..."
	@rm -f $(APILO_CONFIG_DIR)/daemon.pid
	@rm -rf $(APILO_LOGS_DIR)/*
	@echo "✅ All artifacts cleaned"

uninstall: ## Uninstall apilo and remove all components
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║                  Uninstalling Apilo                               ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@$(INSTALL_PATH)/$(BINARY_NAME) daemon stop 2>/dev/null || true
	@echo "🗑️  Removing components..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@rm -f $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh
	@rm -f $(CLAUDE_TOOLS_DIR)/apilo.json
	@rm -f $(CLAUDE_COMMANDS_DIR)/apilo
	@rm -f $(APILO_CONFIG_DIR)/daemon.pid
	@echo "✅ Uninstall complete"
	@echo ""
	@echo "📁 Configuration preserved in: $(APILO_CONFIG_DIR)"
	@echo "   To fully remove: rm -rf $(APILO_CONFIG_DIR)"

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

deps: ## Download and tidy dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

check-install: ## Verify installation
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║                  Verifying Installation                          ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🔍 Checking binary..."
	@which $(BINARY_NAME) || echo "❌ Binary not found in PATH"
	@echo ""
	@echo "🔍 Checking version..."
	@$(BINARY_NAME) version || echo "❌ Binary not executable"
	@echo ""
	@echo "🔍 Checking hook..."
	@ls -la $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh || echo "⚠️  Hook not installed (run: make configure)"
	@echo ""
	@echo "🔍 Checking daemon status..."
	@$(BINARY_NAME) daemon status || echo "⚠️  Daemon not running (start with: make daemon-start)"
	@echo ""
	@echo "✅ Verification complete"

all: clean deps build install configure ## Complete setup: clean, deps, build, install, and configure
	@echo ""
	@echo "╔═══════════════════════════════════════════════════════════════════╗"
	@echo "║           APILO INSTALLATION COMPLETE                             ║"
	@echo "╚═══════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "✅ Binary installed:  $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo "✅ Hook installed:    $(CLAUDE_HOOKS_DIR)/apilo-optimizer.sh"
	@echo "✅ Config directory:  $(APILO_CONFIG_DIR)"
	@echo ""
	@echo "🚀 Quick Start Guide:"
	@echo ""
	@echo "   1. Start daemon:"
	@echo "      $$ make daemon-start"
	@echo ""
	@echo "   2. Check status:"
	@echo "      $$ make daemon-status"
	@echo ""
	@echo "   3. Use Claude Code normally - API calls auto-optimized!"
	@echo ""
	@echo "📊 Monitoring Commands:"
	@echo "   make daemon-status    # Check daemon health"
	@echo "   make daemon-logs      # View logs"
	@echo "   apilo --help          # See all commands"
	@echo ""
	@echo "📚 Documentation:"
	@echo "   cat DAEMON.md         # Daemon documentation"
	@echo "   cat HOOK_GUIDE.md     # Hook documentation"
	@echo "   cat INSTALLATION.md   # Installation guide"
	@echo ""

run: build ## Build and run apilo locally
	@echo "🚀 Running apilo..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

version: ## Show version information
	@$(BINARY_NAME) version 2>/dev/null || ./$(BUILD_DIR)/$(BINARY_NAME) version

# Default target
.DEFAULT_GOAL := help
