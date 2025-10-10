# API Latency Optimizer - Benchmark Results

## Wikipedia Test Run (2025-10-03)

### Configuration
- **Target**: https://www.wikipedia.org/
- **Total Requests**: 600 (3 iterations × 200 requests)
- **Concurrency**: 20
- **Iterations**: 3

### Latency Metrics
- **P50 (Median)**: 25.97 ms
- **P95**: 102.30 ms
- **P99**: 112.80 ms
- **Mean**: 33.85 ms
- **Range**: 20.68 ms - 117.71 ms

### Throughput
- **RPS (Avg)**: 541.31 req/s
- **RPS (Peak)**: 567.96 req/s
- **RPS (Min)**: 506.49 req/s
- **Bandwidth**: 52,252 bytes/s (51.03 KB/s)

### Reliability
- **Total Requests**: 600
- **Successful**: 600 (100.00%)
- **Failed**: 0 (0.00% error rate)
- **Connection Reuse**: 90.00%

### Cache Performance
- **Hit Ratio**: 0.00% (no cached content for first runs)
- **Memory Usage**: 0.00 MB
- **Peak Memory**: 0.00 MB

### Performance Score
- **Overall Grade**: 63/100
- **Test Duration**: 5.65 seconds
- **Uptime**: 100%

### Token Usage
*Note: Token metrics not captured in this benchmark run*

---
*Results saved to: benchmarks/results/quick_benchmark_20251003_024444/*
