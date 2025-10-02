# Quick Start: Using API Optimizer in Claude Code

Get started with the `/api-optimize` command in 5 minutes.

---

## Step 1: Verify Installation

The command is already installed. Verify by typing:

```
/api-optimize
```

You should see the command help.

---

## Step 2: Run Your First Benchmark

Test against a public API:

```
/api-optimize https://httpbin.org/get
```

**Expected Output:**
```markdown
# API Latency Benchmark Results

**Target**: https://httpbin.org/get
**Requests**: 100
**Concurrency**: 10

## Performance Metrics

| Metric | Value |
|--------|-------|
| P50 Latency | 152.3ms |
| P95 Latency | 287.5ms |
| P99 Latency | 398.2ms |
| Throughput | 32.5 req/sec |
| Error Rate | 0.0% |
```

---

## Step 3: Enable Caching (See the Magic!)

Run the same test with caching enabled:

```
/api-optimize https://httpbin.org/get --enable-cache --requests 200
```

**Expected Output:**
```markdown
# API Optimization Results

**Status**: ✅ Optimization Active
**Cache Hit Ratio**: 99.5%
**Performance Improvement**: 94.2%

## Before vs After

| Metric | Baseline | Optimized | Improvement |
|--------|----------|-----------|-------------|
| P50 Latency | 152ms | 8ms | 94.7% |
| P95 Latency | 287ms | 12ms | 95.8% |
| Throughput | 32.5 RPS | 412.3 RPS | 12.7x |
```

---

## Step 4: Try Production Mode

Enable all optimizations:

```
/api-optimize https://httpbin.org/get --production
```

**What Happens:**
1. ✅ Memory-bounded caching enabled (500MB)
2. ✅ HTTP/2 optimization activated
3. ✅ Circuit breaker protection enabled
4. ✅ Monitoring dashboard started on port 8080

**Access Dashboard:**
Open browser to: http://localhost:8080/dashboard

---

## Step 5: Test Your Own API

Replace with your API endpoint:

```
/api-optimize https://api.yourdomain.com/v1/users --enable-cache --enable-http2
```

---

## Common Use Cases

### Use Case 1: Quick Performance Check

```
/api-optimize https://api.example.com/endpoint
```

**When**: Need quick performance metrics
**Result**: Latency percentiles and throughput

### Use Case 2: Optimization Testing

```
/api-optimize https://api.example.com/endpoint --enable-cache --save-baseline
```

**When**: Testing optimization impact
**Result**: Baseline saved, optimizations active

### Use Case 3: Production Deployment

```
/api-optimize https://api.example.com/endpoint --production --dashboard
```

**When**: Ready for production monitoring
**Result**: Full optimization + real-time dashboard

### Use Case 4: Comparison Analysis

```
# First, save baseline
/api-optimize https://api.example.com/endpoint --save-baseline

# Then compare with optimizations
/api-optimize https://api.example.com/endpoint --enable-cache --compare baseline.json
```

**When**: Need before/after comparison
**Result**: Detailed improvement metrics

---

## Tips for Best Results

### 1. Start Simple
```
/api-optimize https://your-api.com --requests 50
```
Quick test with 50 requests.

### 2. Increase Load Gradually
```
/api-optimize https://your-api.com --requests 500 --concurrency 20
```
More realistic load testing.

### 3. Enable Optimizations Incrementally
```
# Step 1: Cache only
/api-optimize https://your-api.com --enable-cache

# Step 2: Add HTTP/2
/api-optimize https://your-api.com --enable-cache --enable-http2

# Step 3: Full production
/api-optimize https://your-api.com --production
```

### 4. Monitor in Production
```
/api-optimize https://your-api.com --production --alerts
```
Get alerts when performance degrades.

---

## Understanding the Output

### Latency Metrics

- **P50 (Median)**: 50% of requests are faster than this
- **P95**: 95% of requests are faster (SLA metric)
- **P99**: 99% of requests are faster (tail latency)

### What's Good?

| Metric | Good | Acceptable | Needs Work |
|--------|------|------------|------------|
| P50 | <50ms | <200ms | >200ms |
| P95 | <100ms | <500ms | >500ms |
| P99 | <200ms | <1000ms | >1000ms |

### Cache Hit Ratio

- **98%+**: Excellent (most requests cached)
- **70-98%**: Good (reasonable caching)
- **<70%**: Poor (check TTL settings)

---

## Troubleshooting

### "Command not found"

**Solution**: Type `/` and look for `api-optimize` in list

### "Build failed"

**Solution**:
```bash
cd /Users/joshkornreich/Documents/Projects/api-latency-optimizer
go build ./src
```

### "Port 8080 already in use"

**Solution**:
```
/api-optimize https://api.example.com --dashboard-port 8081
```

---

## Next Steps

1. **Read Integration Guide**: [CLAUDE_CODE_INTEGRATION.md](CLAUDE_CODE_INTEGRATION.md)
2. **Explore Configuration**: [docs/CONFIGURATION.md](docs/CONFIGURATION.md)
3. **API Reference**: [docs/API_REFERENCE.md](docs/API_REFERENCE.md)
4. **Deployment Guide**: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

---

## Example Workflow

```
# 1. Quick test
/api-optimize https://api.myapp.com/users

# 2. Enable cache
/api-optimize https://api.myapp.com/users --enable-cache

# 3. Save baseline for comparison
/api-optimize https://api.myapp.com/users --enable-cache --save-baseline

# 4. Deploy to production with monitoring
/api-optimize https://api.myapp.com/users --production --dashboard

# 5. Access dashboard
# Open: http://localhost:8080/dashboard
```

---

**You're ready to optimize!** 🚀

Start with: `/api-optimize https://httpbin.org/get`
