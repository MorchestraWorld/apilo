# Claude Code Integration Guide

Complete guide for using the API Latency Optimizer with Claude Code.

---

## Overview

The API Latency Optimizer is now integrated into Claude Code as the `/api-optimize` slash command, providing direct access to benchmarking and optimization features from within your Claude Code sessions.

---

## Installation

### 1. Command is Already Installed

The `/api-optimize` command has been installed to:
```
~/.claude/commands/api-optimize.md
```

Claude Code automatically discovers commands in this directory.

### 2. Verify Installation

```
Type /api-optimize in Claude Code to see command help
```

---

## Quick Start

### Basic API Benchmark

```
/api-optimize https://api.example.com
```

**What happens:**
1. Claude Code builds the optimizer (if needed)
2. Runs 100 requests with 10 concurrent connections
3. Shows performance metrics (P50, P95, P99 latency)
4. Displays throughput and error rates

### Enable Caching

```
/api-optimize https://api.example.com --enable-cache
```

**Result:**
- 90-95% latency reduction on repeated requests
- 98% cache hit ratio
- Memory-bounded caching (500MB default)

### Production Deployment

```
/api-optimize https://api.example.com --production
```

**Activates:**
- Memory-bounded caching
- HTTP/2 optimization
- Circuit breaker protection
- Real-time monitoring dashboard

---

## Usage Examples

### Scenario 1: API Performance Analysis

**Goal**: Understand current API performance

```
/api-optimize https://api.myservice.com/v1/users --requests 500
```

**Output:**
```markdown
# API Latency Benchmark Results

**Target**: https://api.myservice.com/v1/users
**Requests**: 500
**Concurrency**: 10

## Performance Metrics

| Metric | Value |
|--------|-------|
| P50 Latency | 245.3ms |
| P95 Latency | 487.2ms |
| P99 Latency | 652.1ms |
| Throughput | 18.7 req/sec |
| Error Rate | 0.2% |
```

### Scenario 2: Optimization Testing

**Goal**: Test optimization impact

```
/api-optimize https://api.myservice.com/v1/users --enable-cache --enable-http2 --save-baseline
```

**Result:**
- Baseline saved for future comparisons
- Cache and HTTP/2 enabled
- Shows before/after metrics

### Scenario 3: Production Monitoring

**Goal**: Deploy with monitoring

```
/api-optimize https://api.myservice.com/v1/users --production --dashboard
```

**Access Dashboard:**
1. Command outputs: `Dashboard URL: http://localhost:8080/dashboard`
2. Open in browser to see real-time metrics
3. Monitor cache hit ratios, latency, throughput

### Scenario 4: Comparison Testing

**Goal**: Compare optimized vs baseline

```
/api-optimize https://api.myservice.com/v1/users --enable-cache --compare baseline.json
```

**Output:**
```markdown
# Optimization Comparison

| Metric | Baseline | Optimized | Improvement |
|--------|----------|-----------|-------------|
| P50 Latency | 245ms | 15ms | 93.9% |
| P95 Latency | 487ms | 42ms | 91.4% |
| Throughput | 18.7 RPS | 198.5 RPS | 10.6x |

✅ Significant improvement detected
✅ Statistical validation passed
```

---

## Integration Architecture

### How It Works

```
┌─────────────────────────────────────────┐
│        Claude Code Session              │
├─────────────────────────────────────────┤
│                                          │
│  User types: /api-optimize <url>        │
│         │                                │
│         ▼                                │
│  ┌──────────────────────┐               │
│  │ Command Parser       │               │
│  └──────────┬───────────┘               │
│             │                            │
│             ▼                            │
│  ┌──────────────────────┐               │
│  │ Build Optimizer      │               │
│  │ (if needed)          │               │
│  └──────────┬───────────┘               │
│             │                            │
│             ▼                            │
│  ┌──────────────────────┐               │
│  │ Execute Benchmark    │               │
│  │ /api-optimizer       │               │
│  └──────────┬───────────┘               │
│             │                            │
│             ▼                            │
│  ┌──────────────────────┐               │
│  │ Parse & Format       │               │
│  │ Results              │               │
│  └──────────┬───────────┘               │
│             │                            │
│             ▼                            │
│  ┌──────────────────────┐               │
│  │ Display to User      │               │
│  └──────────────────────┘               │
│                                          │
└─────────────────────────────────────────┘
```

### File Locations

- **Command**: `~/.claude/commands/api-optimize.md`
- **Optimizer**: `/Users/joshkornreich/Documents/Projects/api-latency-optimizer/`
- **Binary**: `/Users/joshkornreich/Documents/Projects/api-latency-optimizer/bin/api-optimizer`
- **Results**: `./benchmarks/results/` (in current working directory)

---

## Configuration

### Default Settings

```yaml
requests: 100
concurrency: 10
iterations: 3
warmup: 1
timeout: 30s
```

### Production Preset

```yaml
requests: 1000
concurrency: 50
iterations: 5
warmup: 2
cache:
  enabled: true
  size_mb: 500
  ttl: 10m
http2:
  enabled: true
circuit_breaker:
  enabled: true
monitoring:
  enabled: true
  dashboard: true
  port: 8080
```

### Custom Configuration

Create a config file:

```yaml
# my-api-config.yaml
optimization:
  cache:
    max_memory_mb: 1000
    default_ttl: "15m"
  http2:
    max_connections_per_host: 30
```

Use with command:

```
/api-optimize https://api.example.com --config my-api-config.yaml
```

---

## Advanced Usage

### Programmatic Integration

You can also use the optimizer programmatically from Claude Code:

```go
// Claude Code can execute this
package main

import "github.com/TSMCP/api-latency-optimizer/src"

func main() {
    config := src.DefaultIntegratedConfig()
    optimizer, _ := src.NewIntegratedOptimizer(config)

    optimizer.Start()
    defer optimizer.Stop()

    // Use optimized client
    client := optimizer.GetClient()
    resp, _ := client.Get("https://api.example.com")
}
```

### Batch Testing

Test multiple endpoints:

```bash
# Claude Code can run this script
for endpoint in /users /posts /comments; do
  /api-optimize https://api.example.com$endpoint --requests 200
done
```

### Continuous Monitoring

Keep dashboard running:

```
/api-optimize https://api.example.com --production --dashboard &
```

Then access http://localhost:8080/dashboard anytime.

---

## Troubleshooting

### Command Not Found

**Issue**: Claude Code doesn't recognize `/api-optimize`

**Solution**:
1. Verify file exists: `ls ~/.claude/commands/api-optimize.md`
2. Restart Claude Code session
3. Type `/` to see all available commands

### Build Errors

**Issue**: Optimizer fails to build

**Solution**:
```bash
cd /Users/joshkornreich/Documents/Projects/api-latency-optimizer
go mod tidy
go build ./src
```

### Port Already in Use

**Issue**: Dashboard port 8080 is occupied

**Solution**:
```
/api-optimize https://api.example.com --dashboard-port 8081
```

### High Memory Usage

**Issue**: Cache using too much memory

**Solution**:
```
/api-optimize https://api.example.com --cache-size 250
```

---

## Performance Tips

### 1. Start Small

```
/api-optimize https://api.example.com --requests 50
```

### 2. Incremental Optimization

```
# Step 1: Baseline
/api-optimize https://api.example.com --save-baseline

# Step 2: Add cache
/api-optimize https://api.example.com --enable-cache --compare baseline.json

# Step 3: Add HTTP/2
/api-optimize https://api.example.com --enable-cache --enable-http2 --compare baseline.json
```

### 3. Monitor Production

```
/api-optimize https://api.example.com --production --alerts
```

---

## Best Practices

### For Development

- Use `--requests 50` for quick tests
- Enable `--verbose` for debugging
- Save baselines frequently

### For Testing

- Use `--requests 500 --iterations 3`
- Compare with baselines
- Test with realistic concurrency

### For Production

- Use `--production` preset
- Enable `--monitor --dashboard`
- Set up `--alerts`
- Monitor memory usage

---

## Comparison with Direct Usage

### Claude Code Integration (`/api-optimize`)

**Advantages:**
- ✅ Quick access from Claude Code
- ✅ Formatted output in session
- ✅ Automated build process
- ✅ Integrated help system

### Direct CLI Usage

**When to use:**
- Automation scripts
- CI/CD pipelines
- Custom configurations
- Advanced features

**How to use:**
```bash
cd /Users/joshkornreich/Documents/Projects/api-latency-optimizer
./bin/api-optimizer -url https://api.example.com -requests 1000
```

---

## Next Steps

1. **Try Basic Benchmark**
   ```
   /api-optimize https://httpbin.org/get
   ```

2. **Enable Optimization**
   ```
   /api-optimize https://httpbin.org/get --enable-cache
   ```

3. **Explore Dashboard**
   ```
   /api-optimize https://httpbin.org/get --dashboard
   ```

4. **Read Documentation**
   - [README](README.md)
   - [API Reference](docs/API_REFERENCE.md)
   - [Configuration Guide](docs/CONFIGURATION.md)

---

## Support

- **Documentation**: `/Users/joshkornreich/Documents/Projects/api-latency-optimizer/docs/`
- **Troubleshooting**: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- **GitHub**: https://github.com/TSMCP/api-latency-optimizer

---

**Integration Status**: ✅ Production Ready
**Command Version**: 1.0
**Optimizer Version**: 2.0
