# Configuration Reference

Complete reference for all configuration options.

---

## Configuration File Structure

```yaml
optimization:
  cache: { ... }
  invalidation: { ... }
  http2: { ... }
  circuit_breaker: { ... }
  monitoring: { ... }
  warmup: { ... }
```

---

## Cache Configuration

### Basic Settings

```yaml
cache:
  enabled: true
  max_memory_mb: 500
  default_ttl: "10m"
  cleanup_interval: "30s"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable caching |
| `max_memory_mb` | int | `500` | Maximum cache memory (MB) |
| `default_ttl` | duration | `"10m"` | Default TTL for cache entries |
| `cleanup_interval` | duration | `"30s"` | Background cleanup frequency |

### Advanced Settings

```yaml
cache:
  gc_threshold_percent: 0.8
  eviction_batch_size: 50
  enable_memory_tracker: true
  enable_leak_detection: true
  pressure_threshold: 0.85
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `gc_threshold_percent` | float | `0.8` | Memory % to trigger GC |
| `eviction_batch_size` | int | `50` | Entries to evict at once |
| `enable_memory_tracker` | bool | `true` | Enable memory tracking |
| `enable_leak_detection` | bool | `true` | Enable leak detection |
| `pressure_threshold` | float | `0.85` | Memory pressure threshold |

---

## Cache Invalidation

```yaml
invalidation:
  enable_tag_based: true
  enable_pattern_matching: true
  enable_dependency_tracking: true
  enable_version_based: true
  async_invalidation: true
  max_dependency_depth: 5
  invalidation_batch_size: 100
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enable_tag_based` | bool | `true` | Tag-based invalidation |
| `enable_pattern_matching` | bool | `true` | Pattern/regex invalidation |
| `enable_dependency_tracking` | bool | `true` | Dependency tracking |
| `enable_version_based` | bool | `true` | Version-based invalidation |
| `async_invalidation` | bool | `true` | Async invalidation |
| `max_dependency_depth` | int | `5` | Max dependency depth |
| `invalidation_batch_size` | int | `100` | Batch invalidation size |

---

## HTTP/2 Configuration

```yaml
http2:
  enabled: true
  max_connections_per_host: 20
  max_idle_conns: 100
  idle_timeout: "90s"
  tls_timeout: "10s"
  dial_timeout: "10s"
  keep_alive_timeout: "30s"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable HTTP/2 |
| `max_connections_per_host` | int | `20` | Connections per host |
| `max_idle_conns` | int | `100` | Total idle connections |
| `idle_timeout` | duration | `"90s"` | Idle connection timeout |
| `tls_timeout` | duration | `"10s"` | TLS handshake timeout |
| `dial_timeout` | duration | `"10s"` | Connection dial timeout |
| `keep_alive_timeout` | duration | `"30s"` | Keep-alive timeout |

---

## Circuit Breaker

```yaml
circuit_breaker:
  failure_threshold: 5
  failure_rate: 0.5
  minimum_requests: 10
  open_timeout: "30s"
  half_open_timeout: "10s"
  half_open_max_requests: 3
  exponential_backoff: true
  max_backoff_time: "5m"
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `failure_threshold` | int | `5` | Failures before opening |
| `failure_rate` | float | `0.5` | Failure rate threshold |
| `minimum_requests` | int | `10` | Min requests before evaluation |
| `open_timeout` | duration | `"30s"` | Time circuit stays open |
| `half_open_timeout` | duration | `"10s"` | Half-open state timeout |
| `half_open_max_requests` | int | `3` | Requests in half-open |
| `exponential_backoff` | bool | `true` | Use exponential backoff |
| `max_backoff_time` | duration | `"5m"` | Maximum backoff time |

---

## Monitoring

```yaml
monitoring:
  enabled: true
  dashboard_port: 8080
  metrics_interval: "5s"
  metrics_retention: "24h"
  high_resolution_metrics: true
  alerting_enabled: true
  prometheus_enabled: true
  prometheus_port: 9090
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable monitoring |
| `dashboard_port` | int | `8080` | Dashboard port |
| `metrics_interval` | duration | `"5s"` | Metrics collection interval |
| `metrics_retention` | duration | `"24h"` | Metrics retention period |
| `high_resolution_metrics` | bool | `true` | High-res metrics |
| `alerting_enabled` | bool | `true` | Enable alerting |
| `prometheus_enabled` | bool | `true` | Enable Prometheus export |
| `prometheus_port` | int | `9090` | Prometheus port |

---

## Warmup Configuration

```yaml
warmup:
  enabled: true
  urls:
    - "https://api.example.com/endpoint1"
    - "https://api.example.com/endpoint2"
  timeout: "30s"
  concurrent_warmup: 5
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable cache warmup |
| `urls` | []string | `[]` | URLs to warmup |
| `timeout` | duration | `"30s"` | Warmup timeout |
| `concurrent_warmup` | int | `5` | Concurrent warmup requests |

---

## Performance Targets

```yaml
targets:
  target_latency: "100ms"
  min_cache_hit_ratio: 0.6
  min_connection_reuse: 0.9
  max_memory_mb: 500
  max_error_rate: 0.01
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `target_latency` | duration | `"100ms"` | Target latency |
| `min_cache_hit_ratio` | float | `0.6` | Min cache hit ratio |
| `min_connection_reuse` | float | `0.9` | Min connection reuse |
| `max_memory_mb` | int | `500` | Max memory usage |
| `max_error_rate` | float | `0.01` | Max error rate (1%) |

---

## Environment Variables

```bash
# Override config file location
export API_OPTIMIZER_CONFIG=/path/to/config.yaml

# Enable debug logging
export API_OPTIMIZER_DEBUG=true

# Set log level
export API_OPTIMIZER_LOG_LEVEL=debug

# Alert webhook
export ALERT_WEBHOOK_URL=https://alerts.example.com/webhook
```

---

## Complete Example

See [config/production_config.yaml](../config/production_config.yaml) for complete example.
