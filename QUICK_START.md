# API Latency Optimizer - Quick Start Guide

## Installation

```bash
cd /Users/joshkornreich/Documents/Projects/Orchestra/api-latency-optimizer
go build -o bin/api-latency-optimizer ./src
```

## Quick Commands

### Basic Benchmark
```bash
./bin/api-latency-optimizer -url https://api.anthropic.com -requests 100 -concurrency 10
```

### Production Benchmark (Statistical Rigor)
```bash
./bin/api-latency-optimizer \
  -url https://api.anthropic.com \
  -requests 1000 \
  -concurrency 50 \
  -iterations 5 \
  -warmup 2
```

### With Baseline Comparison
```bash
./bin/api-latency-optimizer \
  -url https://api.anthropic.com \
  -requests 500 \
  -concurrency 20 \
  -iterations 3 \
  -compare benchmarks/baseline_v1_official.json
```

### From Configuration File
```bash
./bin/api-latency-optimizer -config config/benchmark_config.yaml
```

## Common Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `-url` | https://api.anthropic.com | Target endpoint |
| `-requests` | 100 | Total requests |
| `-concurrency` | 10 | Parallel requests |
| `-iterations` | 3 | Benchmark runs |
| `-warmup` | 1 | Warmup iterations |
| `-timeout` | 30s | Request timeout |
| `-keepalive` | true | HTTP keep-alive |
| `-output` | ./benchmarks/results | Output directory |
| `-compare` | - | Baseline JSON path |
| `-quiet` | false | Suppress output |

## Understanding Results

### Key Metrics

- **P50**: Half of requests are faster (median)
- **P95**: 95% of requests are faster (SLA metric)
- **P99**: 99% of requests are faster (tail latency)
- **RPS**: Requests processed per second
- **TTFB**: Time to first byte

### Performance Targets

- **Good**: P95 < 200ms, P99 < 500ms
- **Acceptable**: P95 < 500ms, P99 < 1000ms
- **Poor**: P95 > 500ms requires optimization

## File Locations

### Results
```
benchmarks/results/[suite_name]_[timestamp]/
├── benchmark.json         # Detailed metrics
├── suite_results.json     # Complete suite
├── SUMMARY.md            # Human-readable
└── COMPARISON.md         # Baseline comparison (if -compare used)
```

### Baseline
```
benchmarks/baseline/quick_benchmark_20251002_013441/
└── suite_results.json    # Save as baseline_v1_official.json
```

## Workflow

### 1. Establish Baseline
```bash
./bin/api-latency-optimizer -requests 1000 -concurrency 50 -iterations 5
cp benchmarks/results/*/suite_results.json benchmarks/baseline_v1.json
```

### 2. Make Optimization
```bash
# Edit code, configuration, or infrastructure
```

### 3. Re-Benchmark
```bash
./bin/api-latency-optimizer \
  -requests 1000 \
  -concurrency 50 \
  -iterations 5 \
  -compare benchmarks/baseline_v1.json
```

### 4. Analyze
```bash
# Check COMPARISON.md for improvements/regressions
cat benchmarks/results/*/COMPARISON.md
```

## Baseline Performance

**Anthropic API** (as of 2025-10-02):
- RPS: 50.19 req/sec
- P50: 172.65 ms
- P95: 333.91 ms
- P99: 526.87 ms

## Testing

```bash
# Run test suite
go test ./tests/... -v

# Run benchmarks
go test ./tests/... -bench=. -benchmem
```

## Troubleshooting

### High Variance
- Increase iterations: `-iterations 5`
- Add warmup: `-warmup 2`
- Check network stability

### Timeouts
- Increase timeout: `-timeout 60s`
- Reduce concurrency: `-concurrency 5`
- Check target availability

### Memory Issues
- Reduce requests: `-requests 100`
- Lower concurrency: `-concurrency 10`
- Disable raw metrics

## Documentation

- **Comprehensive Guide**: `docs/README.md`
- **Baseline Report**: `BASELINE_REPORT.md`
- **Configuration**: `config/benchmark_config.yaml`

## Version

Current: 1.0.0
Agent: Benchmarker-Performance-2025-09-04
