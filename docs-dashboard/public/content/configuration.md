# Configuration Guide

Complete configuration reference for the API Latency Optimizer.

---

## Configuration File Format

The optimizer uses YAML configuration files for all settings.

**Default Location:** `config/production_config.yaml`

**Example:**
```yaml
optimization:
  cache:
    enabled: true
    max_memory_mb: 500
    default_ttl: "10m"

  monitoring:
    enabled: true
    dashboard_port: 8080
```

---

## Cache Configuration

### Basic Settings

```yaml
optimization:
  cache:
    enabled: true              # Enable/disable caching
    max_memory_mb: 500         # Maximum cache memory (MB)
    default_ttl: "10m"         # Default time-to-live
    gc_threshold_percent: 0.8  # GC trigger threshold
    enable_memory_tracker: true # Memory tracking
```

**Parameters:**

- **enabled** (bool): Master switch for caching
- **max_memory_mb** (int): Hard memory limit for cache
- **default_ttl** (duration): Default expiration time for cache entries
- **gc_threshold_percent** (float): Trigger GC when cache reaches this % of max
- **enable_memory_tracker** (bool): Enable real-time memory tracking

### Advanced Cache Settings

```yaml
optimization:
  cache:
    eviction_policy: "lru"     # lru, lfu, fifo
    max_entries: 10000         # Maximum number of entries
    shard_count: 16            # Number of cache shards
    compression_enabled: true   # Compress large entries
    compression_threshold: 1024 # Compress if > 1KB
```

---

## Invalidation Configuration

```yaml
optimization:
  invalidation:
    enable_tag_based: true           # Tag-based invalidation
    enable_pattern_matching: true    # Pattern matching
    enable_dependency_tracking: true # Dependency tracking
    enable_version_based: true       # Version-based invalidation
    async_invalidation: true         # Non-blocking invalidation
    invalidation_batch_size: 100     # Batch size for async
```

---

## HTTP/2 Configuration

```yaml
optimization:
  http2:
    max_connections_per_host: 20  # Connection pool size
    idle_timeout: "90s"           # Keep-alive timeout
    tls_timeout: "10s"            # TLS handshake timeout
    max_concurrent_streams: 100   # HTTP/2 concurrent streams
    header_table_size: 4096       # HPACK table size
    initial_window_size: 65535    # Flow control window
```

---

## Circuit Breaker Configuration

```yaml
optimization:
  circuit_breaker:
    failure_threshold: 5          # Failures before opening
    open_timeout: "30s"           # Time in open state
    half_open_max_requests: 3     # Test requests in half-open
    failure_rate_threshold: 0.5   # 50% failure rate trigger
```

**States:**
- **Closed**: Normal operation
- **Open**: Fail fast (after threshold failures)
- **Half-Open**: Testing recovery

---

## Monitoring Configuration

```yaml
optimization:
  monitoring:
    enabled: true
    dashboard_port: 8080
    metrics_interval: "5s"
    alerting_enabled: true
    prometheus_enabled: true
    jaeger_enabled: true
    log_level: "info"           # debug, info, warn, error
```

---

## Alert Configuration

```yaml
optimization:
  alerts:
    # Latency alerts
    latency_p95_threshold_ms: 500
    latency_p99_threshold_ms: 1000

    # Memory alerts
    memory_usage_threshold_percent: 80
    memory_growth_threshold_mb: 100

    # Cache alerts
    cache_hit_ratio_threshold: 0.7
    cache_eviction_rate_threshold: 0.1

    # Error alerts
    error_rate_threshold_percent: 5

    # Cooldown
    cooldown_duration: "5m"

    # Channels
    email_enabled: true
    slack_webhook_url: "https://hooks.slack.com/..."
    pagerduty_enabled: false
```

---

## Production Configuration Example

Complete production-ready configuration:

```yaml
optimization:
  # Cache Configuration
  cache:
    enabled: true
    max_memory_mb: 1000
    default_ttl: "15m"
    gc_threshold_percent: 0.75
    enable_memory_tracker: true
    eviction_policy: "lru"
    max_entries: 50000
    shard_count: 32
    compression_enabled: true
    compression_threshold: 2048

  # Invalidation
  invalidation:
    enable_tag_based: true
    enable_pattern_matching: true
    enable_dependency_tracking: true
    enable_version_based: true
    async_invalidation: true
    invalidation_batch_size: 200

  # HTTP/2 Optimization
  http2:
    max_connections_per_host: 30
    idle_timeout: "120s"
    tls_timeout: "10s"
    max_concurrent_streams: 250
    header_table_size: 8192
    initial_window_size: 131072

  # Circuit Breaker
  circuit_breaker:
    failure_threshold: 3
    open_timeout: "20s"
    half_open_max_requests: 5
    failure_rate_threshold: 0.3

  # Monitoring
  monitoring:
    enabled: true
    dashboard_port: 8080
    metrics_interval: "10s"
    alerting_enabled: true
    prometheus_enabled: true
    jaeger_enabled: true
    log_level: "info"

  # Alerts
  alerts:
    latency_p95_threshold_ms: 200
    latency_p99_threshold_ms: 500
    memory_usage_threshold_percent: 85
    memory_growth_threshold_mb: 200
    cache_hit_ratio_threshold: 0.8
    cache_eviction_rate_threshold: 0.05
    error_rate_threshold_percent: 2
    cooldown_duration: "3m"
    email_enabled: true
    slack_webhook_url: "${SLACK_WEBHOOK_URL}"
```

---

## Environment Variables

Override configuration with environment variables:

```bash
# Cache
export OPTIMIZER_CACHE_MAX_MEMORY_MB=1000
export OPTIMIZER_CACHE_DEFAULT_TTL=15m

# Monitoring
export OPTIMIZER_DASHBOARD_PORT=8080
export OPTIMIZER_LOG_LEVEL=info

# Alerts
export OPTIMIZER_SLACK_WEBHOOK_URL=https://hooks.slack.com/...
```

---

## Command-Line Flags

Override settings via CLI:

```bash
./api-optimizer \
  --config config/production_config.yaml \
  --cache-max-memory 1000 \
  --dashboard-port 8080 \
  --log-level info
```

---

## Configuration Best Practices

### 1. Start Conservative
```yaml
cache:
  max_memory_mb: 250  # Start small
  default_ttl: "5m"   # Short TTL
```

### 2. Monitor and Adjust
- Watch cache hit ratio
- Adjust TTL based on data freshness needs
- Increase memory if hit ratio < 90%

### 3. Tune for Your Workload

**Read-Heavy:**
```yaml
cache:
  max_memory_mb: 1000
  default_ttl: "30m"
```

**Write-Heavy:**
```yaml
cache:
  max_memory_mb: 250
  default_ttl: "2m"
invalidation:
  async_invalidation: true
```

**Mixed:**
```yaml
cache:
  max_memory_mb: 500
  default_ttl: "10m"
```

---

## Validation

Validate configuration before deploying:

```bash
./api-optimizer --config config.yaml --validate
```

---

## Next Steps

- **[Quick Start](/docs/quickstart)** - Get started
- **[Features](/docs/features)** - Explore features
- **[Performance](/docs/performance)** - See results
