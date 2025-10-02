# Monitoring Framework Guide

## Overview

The API Latency Optimizer includes a comprehensive monitoring framework that provides real-time performance insights, alerting, and metrics export capabilities. This guide covers all aspects of the monitoring system.

## Table of Contents

1. [Features](#features)
2. [Quick Start](#quick-start)
3. [Components](#components)
4. [Configuration](#configuration)
5. [Dashboard](#dashboard)
6. [Alerting](#alerting)
7. [Prometheus Integration](#prometheus-integration)
8. [API Reference](#api-reference)
9. [Examples](#examples)
10. [Troubleshooting](#troubleshooting)

## Features

### Real-Time Monitoring
- **Live Dashboard**: Web-based interface with auto-refreshing metrics
- **Performance Graphs**: Interactive charts for latency, cache hit ratio, and throughput
- **Historical Tracking**: Snapshot-based historical data retention

### Comprehensive Metrics
- **Cache Performance**: Hit ratio, memory usage, eviction rates
- **Latency Statistics**: P50, P95, P99, mean, min, max latencies
- **TTFB Metrics**: Time to first byte analysis
- **Throughput**: Requests per second, bytes per second
- **Error Tracking**: Error rates and failed request counts
- **Connection Pool**: Connection reuse rates

### Alerting System
- **Configurable Rules**: Define custom alert thresholds
- **Severity Levels**: INFO, WARNING, CRITICAL
- **Cooldown Periods**: Prevent alert fatigue
- **Alert History**: Track all triggered alerts
- **Callbacks**: Custom alert handlers (webhook-ready)

### Prometheus Export
- **Standard Format**: Full Prometheus metrics export
- **Custom Metrics**: Extensible metric definitions
- **Sample Dashboards**: Grafana dashboard templates
- **Alert Rules**: Pre-configured Prometheus alert rules

## Quick Start

### Basic Monitoring

Enable monitoring with the `--monitor` flag:

```bash
./api-latency-optimizer --monitor --url https://api.example.com
```

Access the dashboard at: `http://localhost:8080`

### With Alerting

Enable both monitoring and alerting:

```bash
./api-latency-optimizer --monitor --alerts --url https://api.example.com
```

### Custom Ports

Specify custom dashboard and Prometheus ports:

```bash
./api-latency-optimizer \
  --monitor \
  --dashboard-port 9000 \
  --prometheus-port 9091 \
  --url https://api.example.com
```

### With Configuration File

Use a configuration file for advanced settings:

```bash
./api-latency-optimizer \
  --monitor \
  --monitoring-config config/monitoring_config.yaml \
  --url https://api.example.com
```

## Components

### 1. Monitoring System (`monitoring.go`)

Core orchestration component that manages all monitoring subsystems.

**Key Features:**
- Lifecycle management (Start/Stop)
- Component coordination
- Background task scheduling
- Snapshot capture
- Report generation

**Configuration:**
```go
config := MonitoringConfig{
    MetricsInterval:    5 * time.Second,
    SnapshotInterval:   30 * time.Second,
    DashboardEnabled:   true,
    AlertingEnabled:    true,
    PrometheusEnabled:  true,
}
```

### 2. Metrics Collector (`metrics_collector.go`)

Centralized metrics aggregation and storage.

**Collected Metrics:**
- Cache statistics (hit ratio, size, memory)
- Latency percentiles (P50, P95, P99)
- Throughput metrics
- Error rates
- Performance scores

**Features:**
- Snapshot capture
- Trend analysis
- Historical data retention
- Configurable retention periods

### 3. Dashboard (`dashboard.go`)

Real-time web interface for monitoring.

**Endpoints:**
- `/` - Main dashboard HTML
- `/api/current` - Current metrics snapshot (JSON)
- `/api/snapshots` - Historical snapshots (JSON)
- `/api/summary` - Metrics summary (JSON)
- `/api/trends` - Trend analysis (JSON)

**Features:**
- Auto-refreshing charts
- Real-time metric updates
- Historical trend visualization
- Responsive design

### 4. Alert Manager (`alerts.go`)

Performance alerting and notification system.

**Alert Types:**
- Latency alerts
- Cache hit ratio alerts
- Memory usage alerts
- Error rate alerts
- Throughput alerts

**Features:**
- Configurable thresholds
- Severity levels
- Cooldown periods
- Alert history
- Acknowledgment support

### 5. Prometheus Exporter (`prometheus_exporter.go`)

Metrics export in Prometheus format.

**Exported Metrics:**
- All cache metrics
- All latency metrics
- Throughput metrics
- Error metrics
- Performance scores

**Features:**
- Standard Prometheus format
- Auto-discovery support
- Custom metric labels
- Sample alert rules

## Configuration

### Configuration File Structure

Located at: `config/monitoring_config.yaml`

```yaml
name: "production_monitoring"
description: "Production monitoring configuration"

# Metrics Collection
metrics:
  collection_interval: 5s
  snapshot_interval: 30s
  cleanup_interval: 5m
  retention_period: 24h
  max_snapshots: 1000

# Dashboard
dashboard:
  enabled: true
  port: 8080
  refresh_interval: 2s

# Alerting
alerting:
  enabled: true
  check_interval: 10s
  rules:
    - name: high_latency_p95
      type: latency
      threshold: 500.0
      comparator: gt
      severity: WARNING
      cooldown: 5m
      enabled: true

# Prometheus
prometheus:
  enabled: true
  port: 9090
  path: /metrics
```

### Alert Rule Configuration

```yaml
rules:
  - name: high_latency
    description: "P95 latency exceeds 500ms"
    type: latency
    threshold: 500.0       # milliseconds
    comparator: gt         # gt, gte, lt, lte, eq
    severity: WARNING      # INFO, WARNING, CRITICAL
    cooldown: 5m
    enabled: true

  - name: low_cache_hit_ratio
    description: "Cache hit ratio below 60%"
    type: cache_hit_ratio
    threshold: 0.60
    comparator: lt
    severity: WARNING
    cooldown: 10m
    enabled: true
```

### Performance Targets

Configure expected performance baselines:

```yaml
targets:
  cache_hit_ratio_min: 0.70
  latency_p95_max: 300.0
  latency_p99_max: 500.0
  error_rate_max: 0.01
  throughput_min: 50.0
```

## Dashboard

### Accessing the Dashboard

Default URL: `http://localhost:8080`

### Dashboard Sections

#### 1. Cache Performance Card
- Hit Ratio (with color coding)
- Size / Capacity
- Memory Usage
- Total Gets
- Evictions

#### 2. Latency Statistics Card
- P50 (Median)
- P95
- P99
- Mean
- Max

#### 3. Throughput & Reliability Card
- Requests/sec
- Bytes/sec
- Error Rate
- Connection Reuse
- Uptime

#### 4. Performance Grade Card
- Letter Grade (A-F)
- Numeric Score (0-100)

#### 5. Latency Trends Chart
Real-time line chart showing:
- P50 latency (green)
- P95 latency (orange)
- P99 latency (red)

#### 6. Cache Performance Chart
Dual-axis chart showing:
- Hit Ratio (0-1 scale)
- Memory Usage (MB)

### Dashboard Features

- **Auto-Refresh**: Configurable refresh interval (default: 2s)
- **Color Coding**: Green (good), Orange (warning), Red (critical)
- **Historical Data**: Up to 60 data points on charts
- **Responsive Design**: Works on mobile and desktop
- **Real-Time Updates**: WebSocket-free polling architecture

## Alerting

### Alert Types

| Type | Description | Typical Threshold |
|------|-------------|-------------------|
| `latency` | Average access latency | 500ms |
| `ttfb` | Time to first byte | 300ms |
| `cache_hit_ratio` | Cache effectiveness | 0.60 (60%) |
| `cache_memory` | Memory consumption | 500MB |
| `error_rate` | Request failures | 0.05 (5%) |
| `throughput` | Requests per second | 10 req/s |

### Severity Levels

- **INFO**: Informational alerts, low priority
- **WARNING**: Issues requiring attention
- **CRITICAL**: Urgent issues requiring immediate action

### Alert Lifecycle

1. **Triggered**: Alert condition met
2. **Active**: Alert is ongoing
3. **Acknowledged**: Alert acknowledged by user
4. **Resolved**: Condition no longer met
5. **Cooldown**: Waiting period before re-triggering

### Programmatic Alert Management

```go
// Add a custom alert rule
rule := AlertRule{
    Name:        "custom_alert",
    Description: "Custom performance threshold",
    Type:        AlertTypeLatency,
    Threshold:   750.0,
    Comparator:  "gt",
    Severity:    AlertSeverityWarning,
    Cooldown:    10 * time.Minute,
    Enabled:     true,
}
alertManager.AddRule(rule)

// Set alert callback
alertManager.SetOnAlert(func(alert *Alert) {
    log.Printf("ALERT: %s - %s", alert.Severity, alert.Message)
    // Send to webhook, email, Slack, etc.
})
```

## Prometheus Integration

### Metrics Endpoint

Default URL: `http://localhost:9090/metrics`

### Sample Prometheus Configuration

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'api_latency_optimizer'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

Run Prometheus:

```bash
prometheus --config.file=prometheus.yml
```

### Available Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| `api_latency_optimizer_cache_hit_ratio` | gauge | Cache hit ratio (0-1) |
| `api_latency_optimizer_cache_size` | gauge | Current cache entries |
| `api_latency_optimizer_cache_memory_usage_bytes` | gauge | Memory usage |
| `api_latency_optimizer_latency_p50_milliseconds` | gauge | P50 latency |
| `api_latency_optimizer_latency_p95_milliseconds` | gauge | P95 latency |
| `api_latency_optimizer_latency_p99_milliseconds` | gauge | P99 latency |
| `api_latency_optimizer_requests_per_second` | gauge | Throughput |
| `api_latency_optimizer_error_rate` | gauge | Error rate (0-1) |
| `api_latency_optimizer_performance_score` | gauge | Overall score (0-100) |

### Prometheus Alert Rules

Sample alert rules (`alerts.yml`):

```yaml
groups:
  - name: latency_alerts
    interval: 30s
    rules:
      - alert: HighLatencyP95
        expr: api_latency_optimizer_latency_p95_milliseconds > 500
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High P95 latency detected"
          description: "P95 latency is {{ $value }}ms"
```

### Grafana Dashboard

Import the sample Grafana dashboard:

1. Create a Prometheus data source in Grafana
2. Import dashboard JSON (see `docs/grafana_dashboard.json`)
3. Customize panels as needed

Sample panels:
- Latency over time (line chart)
- Cache hit ratio (gauge)
- Error rate (graph)
- Throughput (stat panel)

## API Reference

### REST API Endpoints

#### GET /api/current

Returns current metrics snapshot.

**Response:**
```json
{
  "timestamp": "2025-10-02T12:00:00Z",
  "cache_hit_ratio": 0.85,
  "latency_p95_ms": 245.5,
  "latency_p99_ms": 512.3,
  "requests_per_second": 125.4,
  "error_rate": 0.002,
  "performance_score": 92,
  "performance_grade": "A"
}
```

#### GET /api/snapshots?duration=1h

Returns historical snapshots.

**Parameters:**
- `duration` (optional): Time range (e.g., "1h", "24h")

**Response:**
```json
[
  {
    "timestamp": "2025-10-02T12:00:00Z",
    "cache_hit_ratio": 0.85,
    ...
  }
]
```

#### GET /api/summary

Returns aggregated metrics summary.

#### GET /api/trends?duration=1h

Returns trend analysis for specified duration.

## Examples

### Example 1: Basic Monitoring

```bash
./api-latency-optimizer \
  --monitor \
  --url https://api.example.com \
  --requests 1000 \
  --concurrency 50
```

### Example 2: Production Monitoring

```bash
./api-latency-optimizer \
  --monitor \
  --alerts \
  --monitoring-config config/production_monitoring.yaml \
  --config config/benchmark_config.yaml \
  --output ./monitoring/results
```

### Example 3: Prometheus + Grafana

```bash
# Terminal 1: Start API Latency Optimizer
./api-latency-optimizer \
  --monitor \
  --prometheus-port 9090 \
  --url https://api.example.com

# Terminal 2: Start Prometheus
prometheus --config.file=prometheus.yml

# Terminal 3: Start Grafana (assuming installed)
grafana-server
```

### Example 4: Continuous Monitoring

```bash
# Run benchmark with monitoring in loop
while true; do
  ./api-latency-optimizer \
    --monitor \
    --alerts \
    --requests 500 \
    --url https://api.example.com
  sleep 60
done
```

## Troubleshooting

### Dashboard Not Loading

**Problem:** Cannot access dashboard at http://localhost:8080

**Solutions:**
1. Check if monitoring is enabled: `--monitor`
2. Verify port is not in use: `lsof -i :8080`
3. Try different port: `--dashboard-port 9000`
4. Check firewall settings

### No Metrics Displayed

**Problem:** Dashboard shows "No metrics available"

**Solutions:**
1. Ensure benchmark has run at least once
2. Check metrics collection interval
3. Verify collector is attached to benchmark
4. Check browser console for API errors

### Alerts Not Triggering

**Problem:** Expected alerts not appearing

**Solutions:**
1. Verify alerting is enabled: `--alerts`
2. Check alert rule configuration
3. Verify thresholds are correct
4. Check alert cooldown periods
5. Review alert manager logs

### Prometheus Scrape Errors

**Problem:** Prometheus cannot scrape metrics

**Solutions:**
1. Verify Prometheus exporter is enabled
2. Check Prometheus configuration
3. Test metrics endpoint: `curl http://localhost:9090/metrics`
4. Verify network connectivity
5. Check Prometheus logs

### High Memory Usage

**Problem:** Monitoring system consuming too much memory

**Solutions:**
1. Reduce `max_snapshots` in configuration
2. Decrease `retention_period`
3. Enable snapshot compression
4. Increase `cleanup_interval`
5. Limit metrics buffer size

## Performance Considerations

### Minimal Overhead

The monitoring system is designed for <1% latency impact:

- Async metric collection
- Lock-free counters where possible
- Efficient snapshot storage
- Configurable collection intervals

### Resource Usage

Typical resource consumption:

- **CPU**: <2% on average
- **Memory**: 50-200MB (depending on snapshot retention)
- **Network**: Minimal (dashboard polling only)

### Scaling Recommendations

For high-volume scenarios:

1. Increase `collection_interval` to reduce overhead
2. Limit `max_snapshots` to control memory
3. Use external time-series database for long-term storage
4. Consider separate monitoring instances for large deployments

## Best Practices

### 1. Alert Configuration
- Start with conservative thresholds
- Adjust based on baseline performance
- Use appropriate cooldown periods
- Monitor alert frequency

### 2. Snapshot Retention
- Balance history depth with memory usage
- Use 24h retention for most cases
- Export to external storage for long-term analysis

### 3. Dashboard Usage
- Keep refresh interval ≥ 2 seconds
- Use trend charts for pattern identification
- Export snapshots for detailed analysis

### 4. Production Deployment
- Enable Prometheus export
- Configure external alerting (PagerDuty, Slack)
- Set up Grafana for advanced visualization
- Implement alert escalation policies

## Advanced Topics

### Custom Metrics

Extend the metrics collector with custom metrics:

```go
// Add custom metric collection
collector.AddCustomMetric("custom_latency", func() float64 {
    return calculateCustomLatency()
})
```

### Webhook Integration

Configure alert webhooks:

```yaml
notifications:
  enabled: true
  webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

### Distributed Monitoring

For multi-instance deployments:

1. Use Prometheus federation
2. Centralize metric storage
3. Implement instance labels
4. Use Grafana for unified dashboards

## Summary

The monitoring framework provides production-ready observability for the API Latency Optimizer with:

- **Real-time visibility** through web dashboard
- **Proactive alerting** for performance issues
- **Industry-standard export** via Prometheus
- **Comprehensive metrics** covering all performance aspects
- **Minimal overhead** (<1% latency impact)

For additional support, see the main [README](/Users/joshkornreich/Documents/Projects/Orchestra/api-latency-optimizer/docs/README.md) or submit an issue.
