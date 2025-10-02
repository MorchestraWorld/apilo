# Configuration Reference

Complete configuration guide for API Latency Optimizer.

## Configuration File

Create `~/.apilo/config.yaml`:

```yaml
# API Latency Optimizer Configuration
# Version: 2.0

optimization:
  # Cache Configuration
  cache:
    enabled: true
    max_memory_mb: 500              # Maximum cache memory
    default_ttl: "10m"              # Default time-to-live
    gc_threshold_percent: 0.8       # GC trigger threshold
    enable_memory_tracker: true     # Enable memory tracking

  # Cache Invalidation
  invalidation:
    enable_tag_based: true          # Tag-based invalidation
    enable_pattern_matching: true   # Pattern matching
    enable_dependency_tracking: true # Dependency tracking
    enable_version_based: true      # Version-based
    async_invalidation: true        # Async processing

  # HTTP/2 Optimization
  http2:
    max_connections_per_host: 20    # Connection pool size
    idle_timeout: "90s"             # Idle connection timeout
    tls_timeout: "10s"              # TLS handshake timeout
    enable_push: true               # Server push support

  # Circuit Breaker
  circuit_breaker:
    failure_threshold: 5            # Failures before opening
    open_timeout: "30s"             # Time in open state
    half_open_max_requests: 3       # Test requests in half-open

  # Monitoring
  monitoring:
    enabled: true
    dashboard_port: 8080            # Dashboard HTTP port
    metrics_interval: "5s"          # Collection interval
    alerting_enabled: true          # Enable alerting
    prometheus_enabled: true        # Prometheus export

  # Alert Configuration
  alerts:
    latency_threshold_ms: 100       # Latency alert threshold
    error_rate_threshold: 0.01      # Error rate threshold (1%)
    memory_threshold_mb: 450        # Memory alert threshold
    cache_miss_threshold: 0.2       # Cache miss threshold (20%)
```

## Environment Variables

Override any configuration with environment variables:

```bash
# Cache Settings
export APILO_CACHE_ENABLED=true
export APILO_CACHE_MAX_MEMORY_MB=1000
export APILO_CACHE_DEFAULT_TTL=15m

# HTTP/2 Settings
export APILO_HTTP2_MAX_CONNECTIONS_PER_HOST=30
export APILO_HTTP2_IDLE_TIMEOUT=120s

# Monitoring Settings
export APILO_MONITORING_ENABLED=true
export APILO_DASHBOARD_PORT=9090
export APILO_PROMETHEUS_ENABLED=true

# Alert Settings
export APILO_ALERTS_LATENCY_THRESHOLD_MS=50
export APILO_ALERTS_ERROR_RATE_THRESHOLD=0.005
```

## Configuration Sections

### Cache Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | true | Enable caching |
| `max_memory_mb` | int | 500 | Maximum memory (MB) |
| `default_ttl` | duration | 10m | Default TTL |
| `gc_threshold_percent` | float | 0.8 | GC trigger (80%) |
| `enable_memory_tracker` | bool | true | Track memory usage |

### Invalidation Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enable_tag_based` | bool | true | Tag-based invalidation |
| `enable_pattern_matching` | bool | true | Pattern matching |
| `enable_dependency_tracking` | bool | true | Dependency tracking |
| `enable_version_based` | bool | true | Version-based |
| `async_invalidation` | bool | true | Async processing |

### HTTP/2 Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `max_connections_per_host` | int | 20 | Connection pool size |
| `idle_timeout` | duration | 90s | Idle timeout |
| `tls_timeout` | duration | 10s | TLS handshake timeout |
| `enable_push` | bool | true | Server push support |

### Circuit Breaker Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `failure_threshold` | int | 5 | Failures before open |
| `open_timeout` | duration | 30s | Open state duration |
| `half_open_max_requests` | int | 3 | Test requests |

### Monitoring Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | true | Enable monitoring |
| `dashboard_port` | int | 8080 | Dashboard port |
| `metrics_interval` | duration | 5s | Collection interval |
| `alerting_enabled` | bool | true | Enable alerts |
| `prometheus_enabled` | bool | true | Prometheus export |

### Alert Configuration

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `latency_threshold_ms` | int | 100 | Latency threshold (ms) |
| `error_rate_threshold` | float | 0.01 | Error rate (1%) |
| `memory_threshold_mb` | int | 450 | Memory threshold (MB) |
| `cache_miss_threshold` | float | 0.2 | Cache miss rate (20%) |

## Configuration Precedence

Configuration is loaded in the following order (last wins):

1. Default built-in configuration
2. Global config file: `/etc/apilo/config.yaml`
3. User config file: `~/.apilo/config.yaml`
4. Project config file: `./config/apilo.yaml`
5. Environment variables
6. Command-line flags

## Performance Tuning

### High-Traffic Scenarios

```yaml
optimization:
  cache:
    max_memory_mb: 1000
    default_ttl: "15m"
    gc_threshold_percent: 0.75

  http2:
    max_connections_per_host: 30
    idle_timeout: "120s"

  circuit_breaker:
    failure_threshold: 3
    open_timeout: "20s"
```

### Low-Latency Requirements

```yaml
optimization:
  cache:
    gc_threshold_percent: 0.75
    enable_memory_tracker: true

  http2:
    tls_timeout: "5s"
    enable_push: true

  circuit_breaker:
    failure_threshold: 3
    open_timeout: "10s"
```

### Memory-Constrained Environments

```yaml
optimization:
  cache:
    max_memory_mb: 250
    gc_threshold_percent: 0.7
    default_ttl: "5m"

  http2:
    max_connections_per_host: 10
```

## Validation

Validate your configuration:

```bash
# Check configuration syntax
apilo config validate

# Show current configuration
apilo config show

# Create default configuration
apilo config init
```

## Hot Reload

The optimizer supports dynamic configuration updates:

```bash
# Update configuration file
vim ~/.apilo/config.yaml

# Reload without restart (if supported)
kill -HUP $(pidof api-optimizer)
```

## Security Considerations

### Sensitive Data

Don't commit sensitive data to configuration files:

```yaml
# ❌ Bad - secrets in config
optimization:
  auth:
    api_key: "secret-key-12345"

# ✅ Good - use environment variables
optimization:
  auth:
    api_key: "${APILO_API_KEY}"
```

### File Permissions

Secure configuration files:

```bash
chmod 600 ~/.apilo/config.yaml
```

## Examples

### Development Configuration

```yaml
optimization:
  cache:
    enabled: true
    max_memory_mb: 100
    default_ttl: "1m"

  monitoring:
    enabled: true
    dashboard_port: 8080

  alerts:
    latency_threshold_ms: 500  # More lenient
```

### Staging Configuration

```yaml
optimization:
  cache:
    enabled: true
    max_memory_mb: 300
    default_ttl: "5m"

  monitoring:
    enabled: true
    alerting_enabled: true

  alerts:
    latency_threshold_ms: 200
```

### Production Configuration

```yaml
optimization:
  cache:
    enabled: true
    max_memory_mb: 500
    default_ttl: "10m"
    gc_threshold_percent: 0.8

  circuit_breaker:
    failure_threshold: 5
    open_timeout: "30s"

  monitoring:
    enabled: true
    alerting_enabled: true
    prometheus_enabled: true

  alerts:
    latency_threshold_ms: 100
    error_rate_threshold: 0.01
    memory_threshold_mb: 450
```

---

**Next Steps**:
- Performance Tuning: `apilo docs performance`
- Monitoring Setup: `apilo docs monitoring`
- Production Deploy: `apilo docs deployment`
