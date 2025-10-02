# Troubleshooting Guide

Common issues and solutions for API Latency Optimizer.

---

## High Memory Usage

### Symptoms
- Memory usage exceeds configured limits
- GC running frequently
- System slowdown

### Diagnosis
```bash
# Check current memory usage
curl http://localhost:8080/metrics | grep memory_usage

# Check cache size
curl http://localhost:8080/cache/stats
```

### Solutions

**1. Reduce cache size**
```yaml
cache:
  max_memory_mb: 250  # Lower limit
```

**2. Adjust GC threshold**
```yaml
cache:
  gc_threshold_percent: 0.7  # Trigger GC earlier
```

**3. Enable emergency cleanup**
```yaml
cache:
  enable_emergency_cleanup: true
```

---

## Low Cache Hit Ratio

### Symptoms
- Cache hit ratio <70%
- High latency despite caching
- Frequent cache misses

### Diagnosis
```bash
# Check cache statistics
curl http://localhost:8080/cache/stats

# Review invalidation patterns
curl http://localhost:8080/cache/invalidations
```

### Solutions

**1. Increase TTL**
```yaml
cache:
  default_ttl: "15m"  # Longer cache lifetime
```

**2. Increase memory limit**
```yaml
cache:
  max_memory_mb: 1000  # More cache space
```

**3. Review invalidation strategy**
- Check if invalidation is too aggressive
- Adjust tag-based invalidation patterns
- Review dependency relationships

---

## Circuit Breaker Tripping

### Symptoms
- Requests failing with circuit breaker error
- Circuit state shows "OPEN"
- Frequent failover events

### Diagnosis
```bash
# Check circuit breaker state
curl http://localhost:8080/circuit/status

# Review failure logs
tail -f /var/log/api-optimizer/errors.log
```

### Solutions

**1. Adjust failure threshold**
```yaml
circuit_breaker:
  failure_threshold: 10  # Less sensitive
  open_timeout: "60s"    # Longer recovery time
```

**2. Check backend health**
- Verify target API is responsive
- Check network connectivity
- Review timeout settings

**3. Enable failover**
```yaml
failover:
  enable_failover: true
  backup_services:
    - url: "https://backup-api.example.com"
```

---

## Slow Performance

### Symptoms
- Latency not improving
- P95 latency >100ms
- Low throughput

### Diagnosis
```bash
# Check performance metrics
curl http://localhost:8080/metrics

# Run benchmark
go test ./tests/... -bench=BenchmarkLatency
```

### Solutions

**1. Verify cache is enabled**
```yaml
cache:
  enabled: true
```

**2. Check HTTP/2 configuration**
```yaml
http2:
  enabled: true
  max_connections_per_host: 30
```

**3. Review connection pooling**
```yaml
http2:
  idle_timeout: "120s"  # Keep connections longer
```

---

## Alert Storm

### Symptoms
- Too many alerts firing
- Alert fatigue
- Same alerts repeating

### Diagnosis
```bash
# Check alert frequency
curl http://localhost:8080/alerts/history

# Review alert rules
cat /etc/api-optimizer/config.yaml | grep alerts
```

### Solutions

**1. Adjust thresholds**
```yaml
alerts:
  latency_threshold_ms: 200  # Less sensitive
  cooldown_period: "10m"     # Longer cooldown
```

**2. Enable alert cooldown**
```yaml
alerts:
  enable_cooldown: true
  min_alert_interval: "5m"
```

---

## Memory Leaks

### Symptoms
- Gradual memory growth over time
- Memory not released after GC
- Eventually hits memory limit

### Diagnosis
```bash
# Check memory trend
curl http://localhost:8080/memory/trend

# Review leak detection
curl http://localhost:8080/cache/leak-status
```

### Solutions

**1. Enable leak detection**
```yaml
cache:
  enable_memory_tracker: true
  enable_leak_detection: true
```

**2. Force emergency cleanup**
```bash
curl -X POST http://localhost:8080/cache/emergency-cleanup
```

**3. Restart service if needed**
```bash
systemctl restart api-optimizer
```

---

## Connection Errors

### Symptoms
- Connection timeouts
- TLS handshake failures
- Network errors

### Diagnosis
```bash
# Test connectivity
curl -v https://target-api.example.com

# Check connection stats
curl http://localhost:8080/connections/stats
```

### Solutions

**1. Increase timeouts**
```yaml
http2:
  tls_timeout: "15s"
  dial_timeout: "10s"
```

**2. Adjust connection limits**
```yaml
http2:
  max_idle_conns: 100
  max_conns_per_host: 30
```

---

## Getting Help

### Debug Mode

```bash
# Enable debug logging
./api-optimizer --config config.yaml --debug=true

# View detailed logs
tail -f /var/log/api-optimizer/debug.log
```

### Collect Diagnostics

```bash
# Export diagnostics
curl http://localhost:8080/diagnostics > diagnostics.json

# Generate health report
curl http://localhost:8080/health/detailed > health-report.txt
```

### Report Issues

Include in your report:
- Version: `./api-optimizer --version`
- Configuration (redact secrets)
- Error logs
- Diagnostics export
- Steps to reproduce

---

See [Production Runbook](../PRODUCTION_RUNBOOK.md) for operations procedures.
