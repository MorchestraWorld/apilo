# Monitoring Guide

Complete guide to monitoring API Latency Optimizer in production.

## Dashboard Access

```bash
apilo monitor <url> --port 8080
```

Access at: http://localhost:8080/dashboard

## Available Metrics

- Latency percentiles (P50, P95, P99)
- Cache hit/miss ratios
- Memory usage
- CPU usage
- Request throughput
- Error rates

## Prometheus Integration

Metrics available at: http://localhost:8080/metrics

## Alert Configuration

Configure alerts in `~/.apilo/config.yaml`

---

**See Also**: `apilo docs configuration`
