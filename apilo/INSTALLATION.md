# Apilo Installation Guide

## Quick Install (Recommended)

### One-Command Setup
```bash
# From project root
make daemon-all
```

This installs:
- ✅ Apilo CLI globally (`~/go/bin/apilo`)
- ✅ Claude Code hooks (`~/.claude/hooks/apilo-optimizer.sh`)
- ✅ Configuration directory (`~/.apilo/`)

### Or Install Just Apilo CLI
```bash
cd apilo
go install
```

## Installation Methods

### Method 1: Using Makefile (Complete Setup)
```bash
cd /path/to/api-latency-optimizer
make daemon-all
```

**What it does:**
1. Cleans old builds
2. Downloads dependencies
3. Builds apilo CLI
4. Installs globally to `~/go/bin/apilo`
5. Installs Claude Code hook to `~/.claude/hooks/`
6. Creates config directory `~/.apilo/`

### Method 2: Using apilo CLI (Hook Only)
```bash
# Install apilo first
go install ./apilo

# Then install Claude Code integration
apilo claude install
```

**What it does:**
1. Creates tool config: `~/.claude/tools/apilo.json`
2. Creates slash command: `~/.claude/commands/apilo`
3. **Installs hook script: `~/.claude/hooks/apilo-optimizer.sh`** ✅

### Method 3: Manual Installation
```bash
# 1. Install apilo
cd apilo
go install

# 2. Copy hook manually
cp apilo/hooks/apilo-optimizer.sh ~/.claude/hooks/
chmod +x ~/.claude/hooks/apilo-optimizer.sh

# 3. Create config directory
mkdir -p ~/.apilo/logs
```

## Verification

### 1. Check Apilo Installation
```bash
which apilo
# Should show: /Users/yourname/go/bin/apilo

apilo --version
# Shows version and build info
```

### 2. Check Hook Installation
```bash
ls -la ~/.claude/hooks/apilo-optimizer.sh
# Should show: -rwxr-xr-x ... apilo-optimizer.sh

# Test hook
echo "test https://api.example.com" | ~/.claude/hooks/apilo-optimizer.sh
```

### 3. Check Daemon
```bash
apilo daemon status
# Should show: Status: Not Running (before starting)

apilo daemon start
# Starts daemon on port 9876

apilo daemon status
# Should show: Status: Running
```

## Hook Installation - The Answer to Your Question

**Q: How does the hook get installed without make?**

**A: The hook is now embedded in the `apilo claude install` command!**

When you run:
```bash
apilo claude install
```

The CLI automatically:
1. Creates `~/.claude/hooks/` directory
2. Writes the hook script with proper permissions (755)
3. Reports: "✅ Installed optimization hook"

**No Makefile needed** - the hook script is embedded directly in the Go binary at `apilo/cmd/claude.go:252-363`.

### How It Works

```go
// In apilo/cmd/claude.go
func installClaudeTool(...) {
    // ... other setup ...

    // Install optimization hook script
    hooksDir := filepath.Join(homeDir, ".claude", "hooks")
    os.MkdirAll(hooksDir, 0755)

    hookScript := `#!/bin/bash
    # Full hook script embedded here
    ...
    `

    hookPath := filepath.Join(hooksDir, "apilo-optimizer.sh")
    os.WriteFile(hookPath, []byte(hookScript), 0755)

    // ✅ Hook installed!
}
```

## Post-Installation

### Start the Daemon
```bash
apilo daemon start
```

### Use Claude Code Normally
The hook automatically detects and optimizes API calls:

```
You: "Fetch data from https://api.github.com/users/octocat"

Behind the scenes:
1. Hook detects API URL
2. Sends to daemon (localhost:9876)
3. Daemon checks cache
4. Returns optimized response
5. You see result instantly

Second request: <10ms (98% faster!)
```

### Monitor Performance
```bash
# Check status
apilo daemon status

# View logs
apilo daemon logs

# Get metrics
curl http://localhost:9876/metrics | jq
```

## Troubleshooting

### Hook Not Found After `apilo claude install`
```bash
# Re-run installation
apilo claude install

# Verify it exists
ls -la ~/.claude/hooks/apilo-optimizer.sh

# Should see:
# ✅ Installed optimization hook
#    📁 Hook: ~/.claude/hooks/apilo-optimizer.sh
```

### Daemon Won't Start
```bash
# Check if already running
apilo daemon status

# Remove stale PID
rm ~/.apilo/daemon.pid

# Try again
apilo daemon start
```

### Hook Not Working
```bash
# 1. Check daemon is running
apilo daemon status

# 2. Test daemon health
curl http://localhost:9876/health

# 3. Test hook manually
echo "https://api.example.com" | ~/.claude/hooks/apilo-optimizer.sh

# 4. Check hook permissions
chmod +x ~/.claude/hooks/apilo-optimizer.sh
```

## Uninstallation

### Remove Everything
```bash
# Using Makefile
make daemon-uninstall

# Or manually
rm ~/go/bin/apilo
rm ~/.claude/hooks/apilo-optimizer.sh
rm ~/.claude/tools/apilo.json
rm ~/.claude/commands/apilo
rm -rf ~/.apilo/
```

### Keep Configuration
```bash
# Remove binaries but keep config
rm ~/go/bin/apilo
rm ~/.claude/hooks/apilo-optimizer.sh

# Config preserved in ~/.apilo/
```

## Summary

**The hook gets installed 3 ways:**

1. **`make daemon-all`** - Copies from `apilo/hooks/apilo-optimizer.sh`
2. **`apilo claude install`** - Embeds hook in CLI, writes to `~/.claude/hooks/`
3. **Manual** - Copy file yourself

**Recommended**: Just run `apilo claude install` after installing apilo CLI. The hook is embedded in the binary and automatically installed with proper permissions.

No Makefile required! ✅
