# Installation Guide

Complete installation instructions for API Latency Optimizer.

---

## Prerequisites

### System Requirements
- **OS**: Linux, macOS, or Windows
- **Go**: 1.21 or higher
- **Memory**: 1GB+ RAM recommended
- **Disk**: 100MB+ for installation and logs

### Check Go Version

```bash
go version
# Should show: go version go1.21 or higher
```

---

## Installation Methods

### Method 1: Build from Source (Recommended)

```bash
# Clone repository
git clone https://github.com/yourorg/api-latency-optimizer.git
cd api-latency-optimizer

# Install dependencies
go mod download

# Build binary
go build -ldflags="-w -s" -o bin/api-optimizer ./src

# Verify installation
./bin/api-optimizer --version
```

### Method 2: Go Install

```bash
go install github.com/yourorg/api-latency-optimizer/src@latest
```

### Method 3: Download Pre-built Binary

```bash
# Download latest release
curl -LO https://github.com/yourorg/api-latency-optimizer/releases/latest/download/api-optimizer-linux-amd64

# Make executable
chmod +x api-optimizer-linux-amd64

# Move to PATH
sudo mv api-optimizer-linux-amd64 /usr/local/bin/api-optimizer
```

---

## Configuration

### 1. Create Configuration Directory

```bash
sudo mkdir -p /etc/api-optimizer
sudo chown $USER /etc/api-optimizer
```

### 2. Copy Configuration Files

```bash
# Copy production config template
cp config/production_config.yaml /etc/api-optimizer/config.yaml

# Copy additional configs
cp config/*.yaml /etc/api-optimizer/
```

### 3. Edit Configuration

```bash
nano /etc/api-optimizer/config.yaml
```

Key settings to review:
- `cache.max_memory_mb`: Adjust based on available RAM
- `http2.max_connections_per_host`: Adjust for target API
- `monitoring.dashboard_port`: Change if port 8080 is in use

---

## Verification

### Test Installation

```bash
# Check version
./bin/api-optimizer --version

# Test with help
./bin/api-optimizer --help

# Run health check
./bin/api-optimizer --config /etc/api-optimizer/config.yaml --check
```

### Run Quick Test

```bash
# Start optimizer
./bin/api-optimizer --config /etc/api-optimizer/config.yaml &

# Wait for startup
sleep 5

# Check health
curl http://localhost:8080/health

# Stop
pkill api-optimizer
```

---

## Optional: System Service

### Create Systemd Service (Linux)

```bash
# Create service file
sudo nano /etc/systemd/system/api-optimizer.service
```

```ini
[Unit]
Description=API Latency Optimizer
After=network.target

[Service]
Type=simple
User=api-optimizer
ExecStart=/usr/local/bin/api-optimizer --config /etc/api-optimizer/config.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl enable api-optimizer
sudo systemctl start api-optimizer

# Check status
sudo systemctl status api-optimizer
```

---

## Troubleshooting Installation

### Go Version Too Old

```bash
# Install latest Go from golang.org
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Build Errors

```bash
# Clean and retry
go clean -cache
go mod tidy
go build ./src
```

### Permission Errors

```bash
# Fix permissions
chmod +x bin/api-optimizer
sudo chown $USER:$USER bin/api-optimizer
```

---

## Next Steps

1. Review [Configuration Reference](CONFIGURATION.md)
2. Follow [Deployment Guide](DEPLOYMENT.md)
3. Set up [Monitoring](MONITORING_GUIDE.md)

---

See [Quick Start Guide](../QUICK_START.md) to begin using the optimizer.
