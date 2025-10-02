# Deployment Guide

**Version**: 2.0
**Status**: Production Ready

---

## Quick Deployment

### 1. Build Production Binary

```bash
# Build with optimizations
go build -ldflags="-w -s" -o api-optimizer ./src

# Verify build
./api-optimizer --version
```

### 2. Prepare Configuration

```bash
# Copy production config template
cp config/production_config.yaml /etc/api-optimizer/config.yaml

# Edit configuration
nano /etc/api-optimizer/config.yaml
```

### 3. Start Service

```bash
./api-optimizer \
  --config /etc/api-optimizer/config.yaml \
  --monitor=true \
  --dashboard=true \
  --port=8080
```

---

## Configuration Review

### Memory Settings
```yaml
cache:
  max_memory_mb: 500          # Adjust based on available RAM
  gc_threshold_percent: 0.8   # Trigger GC at 80% memory
```

### Cache TTL
```yaml
cache:
  default_ttl: "10m"  # Balance freshness vs performance
```

### HTTP/2 Optimization
```yaml
http2:
  max_connections_per_host: 20
  idle_timeout: "90s"
```

### Alert Thresholds
```yaml
monitoring:
  alert_thresholds:
    latency_warning_ms: 100
    latency_critical_ms: 500
    cache_hit_ratio_warning: 0.7
    memory_warning_percent: 0.8
```

---

## Health Checks

### Startup Verification

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","uptime":"5m30s"}
```

### Metrics Validation

```bash
# Check metrics
curl http://localhost:8080/metrics

# Verify cache is working
curl http://localhost:8080/cache/stats
```

---

## Monitoring Setup

### Dashboard Access

```bash
# Access monitoring dashboard
open http://localhost:8080/dashboard
```

### Alert Configuration

```bash
# Configure alert webhook (optional)
export ALERT_WEBHOOK_URL="https://your-alerting-service.com/webhook"
```

---

## Post-Deployment Checklist

- [ ] Service started successfully
- [ ] Health check returns 200 OK
- [ ] Dashboard accessible
- [ ] Cache warmup completed
- [ ] Memory usage within bounds
- [ ] Alert system operational
- [ ] Metrics being collected

---

## Rollback Procedure

If issues occur:

```bash
# Stop service
pkill api-optimizer

# Revert to previous version
cp /backup/api-optimizer ./api-optimizer

# Restart with previous config
./api-optimizer --config /etc/api-optimizer/config.yaml.bak
```

---

## Production Considerations

### Resource Requirements
- **CPU**: 2+ cores recommended
- **Memory**: 1GB+ (depends on cache size)
- **Disk**: 100MB+ for logs and metrics
- **Network**: Low latency connection to target APIs

### Security
- Enable TLS for dashboard
- Restrict dashboard access by IP
- Use environment variables for secrets
- Regular security updates

---

See [PRODUCTION_RUNBOOK.md](../PRODUCTION_RUNBOOK.md) for operations guide.
