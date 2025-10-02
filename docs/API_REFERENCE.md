# API Reference

Programmatic usage of API Latency Optimizer.

---

## Quick Start

```go
import "github.com/yourorg/api-latency-optimizer/src"

func main() {
    // Create optimizer
    config := src.DefaultIntegratedConfig()
    optimizer, _ := src.NewIntegratedOptimizer(config)

    // Start
    optimizer.Start()
    defer optimizer.Stop()

    // Use optimized client
    client := optimizer.GetClient()
    resp, _ := client.Get("https://api.example.com")
}
```

---

## IntegratedOptimizer

### Constructor

```go
func NewIntegratedOptimizer(config *IntegratedConfig) (*IntegratedOptimizer, error)
```

### Methods

#### Start / Stop

```go
func (io *IntegratedOptimizer) Start() error
func (io *IntegratedOptimizer) Stop() error
```

#### Client Access

```go
func (io *IntegratedOptimizer) GetClient() *OptimizedClient
```

#### Statistics

```go
func (io *IntegratedOptimizer) GetCurrentStats() *IntegratedStats
func (io *IntegratedOptimizer) IsHealthy() bool
```

#### Benchmarking

```go
func (io *IntegratedOptimizer) RunBenchmark(config *BenchmarkRunConfig) (*BenchmarkResult, error)
func (io *IntegratedOptimizer) RunComparisonBenchmark(config *BenchmarkRunConfig) (*ComparisonResult, error)
```

---

## MemoryBoundedCache

### Constructor

```go
func NewMemoryBoundedCache(config *MemoryBoundedConfig) *MemoryBoundedCache
```

### Methods

```go
func (mbc *MemoryBoundedCache) Get(key string) (interface{}, bool)
func (mbc *MemoryBoundedCache) Set(key string, value interface{}, ttl time.Duration) error
func (mbc *MemoryBoundedCache) Delete(key string)
func (mbc *MemoryBoundedCache) Clear()
func (mbc *MemoryBoundedCache) Size() int
func (mbc *MemoryBoundedCache) GetMemoryStats() *MemoryStats
```

---

## Cache Invalidation

### AdvancedInvalidationManager

```go
func NewAdvancedInvalidationManager(config *InvalidationConfig) *AdvancedInvalidationManager
```

### Invalidation Methods

```go
// Tag-based
func (aim *AdvancedInvalidationManager) InvalidateByTag(tag string) error

// Pattern-based
func (aim *AdvancedInvalidationManager) InvalidateByPattern(pattern string) error

// Dependency-based
func (aim *AdvancedInvalidationManager) InvalidateByDependency(key string) error

// Version-based
func (aim *AdvancedInvalidationManager) InvalidateByVersion(version string) error
```

---

## Circuit Breaker

### Constructor

```go
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker
```

### Methods

```go
func (cb *CircuitBreaker) Execute(fn func() error) error
func (cb *CircuitBreaker) GetState() CircuitState
func (cb *CircuitBreaker) Reset()
func (cb *CircuitBreaker) GetMetrics() *CircuitBreakerMetrics
```

### States

```go
const (
    CircuitClosed    CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)
```

---

## Configuration Types

### IntegratedConfig

```go
type IntegratedConfig struct {
    ClientConfig        *OptimizedClientConfig
    BenchmarkConfig     *BenchmarkConfig
    MonitoringConfig    *MonitoringConfig
    WarmupEnabled       bool
    WarmupURLs          []string
    WarmupTimeout       time.Duration
    TargetLatency       time.Duration
    MinCacheHitRatio    float64
}
```

### MemoryBoundedConfig

```go
type MemoryBoundedConfig struct {
    MaxMemoryMB          int64
    GCThresholdPercent   float64
    GCInterval           time.Duration
    EnableGCOptimization bool
    EnableMemoryTracker  bool
}
```

---

## HTTP Endpoints

### Health Check

```
GET /health
Response: {"status":"healthy","uptime":"5m30s"}
```

### Metrics

```
GET /metrics
Response: JSON with all metrics
```

### Cache Statistics

```
GET /cache/stats
Response: {"hit_ratio":0.98,"size":1500,"memory_mb":245}
```

### Circuit Breaker Status

```
GET /circuit/status
Response: {"state":"closed","failures":0}
```

---

## Events & Callbacks

### Alert Callbacks

```go
alertManager.OnAlert(func(alert *Alert) {
    // Handle alert
    log.Printf("Alert: %s", alert.Message)
})

alertManager.OnResolve(func(alert *Alert) {
    // Handle resolution
    log.Printf("Resolved: %s", alert.Message)
})
```

---

## Examples

### Custom Configuration

```go
config := &src.IntegratedConfig{
    ClientConfig: &src.OptimizedClientConfig{
        CacheConfig: &src.MemoryBoundedConfig{
            MaxMemoryMB: 1000,
            DefaultTTL:  15 * time.Minute,
        },
    },
    WarmupEnabled: true,
    WarmupURLs: []string{
        "https://api.example.com/v1/users",
        "https://api.example.com/v1/posts",
    },
}

optimizer, _ := src.NewIntegratedOptimizer(config)
```

### Manual Cache Operations

```go
cache := optimizer.GetCache()

// Set with custom TTL
cache.Set("user:123", userData, 30*time.Minute)

// Get
if data, found := cache.Get("user:123"); found {
    // Use data
}

// Invalidate by tag
cache.InvalidateByTag("user:123")
```

### Circuit Breaker Usage

```go
breaker := src.NewCircuitBreaker(config)

err := breaker.Execute(func() error {
    return makeAPICall()
})

if err != nil {
    // Handle error or circuit open
}
```

---

See [Implementation Guide](../IMPLEMENTATION_GUIDE_AND_DRAWBACKS.md) for detailed usage.
