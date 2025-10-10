package daemon

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// CPUStats tracks CPU usage
type CPUStats struct {
	lastSampleTime time.Time
	lastCPUTime    time.Duration
	cpuPercent     float64
	mu             sync.RWMutex
}

// Metrics tracks daemon performance metrics
type Metrics struct {
	totalRequests      int64
	cacheHits          int64
	cacheMisses        int64
	errors             int64
	totalLatency       int64
	latencyCount       int64
	claudeInputTokens  int64
	claudeOutputTokens int64
	claudeTotalCost    int64 // Cost in cents
	claudeRequests     int64
	cpuStats           *CPUStats
	mu                 sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		cpuStats: &CPUStats{
			lastSampleTime: time.Now(),
		},
	}
}

// IncrementRequests increments the total request counter
func (m *Metrics) IncrementRequests() {
	atomic.AddInt64(&m.totalRequests, 1)
}

// IncrementCacheHits increments the cache hit counter
func (m *Metrics) IncrementCacheHits() {
	atomic.AddInt64(&m.cacheHits, 1)
}

// IncrementCacheMisses increments the cache miss counter
func (m *Metrics) IncrementCacheMisses() {
	atomic.AddInt64(&m.cacheMisses, 1)
}

// IncrementErrors increments the error counter
func (m *Metrics) IncrementErrors() {
	atomic.AddInt64(&m.errors, 1)
}

// RecordLatency records a request latency
func (m *Metrics) RecordLatency(latency time.Duration) {
	atomic.AddInt64(&m.totalLatency, int64(latency))
	atomic.AddInt64(&m.latencyCount, 1)
}

// IncrementClaudeTokens records Claude token usage
func (m *Metrics) IncrementClaudeTokens(inputTokens, outputTokens int64) {
	atomic.AddInt64(&m.claudeInputTokens, inputTokens)
	atomic.AddInt64(&m.claudeOutputTokens, outputTokens)
	atomic.AddInt64(&m.claudeRequests, 1)
}

// RecordClaudeCost records Claude API cost in dollars
func (m *Metrics) RecordClaudeCost(costDollars float64) {
	costCents := int64(costDollars * 100)
	atomic.AddInt64(&m.claudeTotalCost, costCents)
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() *MetricsStats {
	totalReq := atomic.LoadInt64(&m.totalRequests)
	hits := atomic.LoadInt64(&m.cacheHits)
	misses := atomic.LoadInt64(&m.cacheMisses)
	errors := atomic.LoadInt64(&m.errors)
	totalLat := atomic.LoadInt64(&m.totalLatency)
	latCount := atomic.LoadInt64(&m.latencyCount)

	var cacheHitRatio float64
	if totalReq > 0 {
		cacheHitRatio = float64(hits) / float64(totalReq)
	}

	var avgLatency time.Duration
	if latCount > 0 {
		avgLatency = time.Duration(totalLat / latCount)
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memoryMB := float64(memStats.Alloc) / 1024 / 1024

	// Get CPU usage
	m.cpuStats.mu.RLock()
	cpuPercent := m.cpuStats.cpuPercent
	m.cpuStats.mu.RUnlock()

	// Get Claude metrics if any requests were made
	var claudeMetrics *ClaudeTokenMetrics
	claudeReq := atomic.LoadInt64(&m.claudeRequests)
	if claudeReq > 0 {
		inputTok := atomic.LoadInt64(&m.claudeInputTokens)
		outputTok := atomic.LoadInt64(&m.claudeOutputTokens)
		costCents := atomic.LoadInt64(&m.claudeTotalCost)

		claudeMetrics = &ClaudeTokenMetrics{
			InputTokens:   inputTok,
			OutputTokens:  outputTok,
			TotalTokens:   inputTok + outputTok,
			Cost:          float64(costCents) / 100.0,
			TotalRequests: claudeReq,
			Model:         "claude-sonnet-4-20250514",
		}
	}

	return &MetricsStats{
		TotalRequests: totalReq,
		CacheHits:     hits,
		CacheMisses:   misses,
		Errors:        errors,
		CacheHitRatio: cacheHitRatio,
		AvgLatency:    avgLatency,
		MemoryUsageMB: memoryMB,
		CPUPercent:    cpuPercent,
		ClaudeMetrics: claudeMetrics,
	}
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics() {
	// Update CPU usage
	m.updateCPUUsage()

	// Force garbage collection to get accurate memory stats
	runtime.GC()
}

// updateCPUUsage calculates CPU usage percentage
func (m *Metrics) updateCPUUsage() {
	m.cpuStats.mu.Lock()
	defer m.cpuStats.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(m.cpuStats.lastSampleTime).Seconds()

	if elapsed == 0 {
		return
	}

	// Get current goroutine count and estimate CPU time
	numGoroutine := float64(runtime.NumGoroutine())
	numCPU := float64(runtime.NumCPU())

	// Simple estimation based on goroutine count
	// This is an approximation since Go doesn't expose process CPU time directly
	// For more accurate tracking, would need to use OS-specific syscalls
	estimatedCPU := (numGoroutine / numCPU) * 2.0 // Rough estimate

	// Cap at 100%
	if estimatedCPU > 100.0 {
		estimatedCPU = 100.0
	}

	m.cpuStats.cpuPercent = estimatedCPU
	m.cpuStats.lastSampleTime = now
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.totalRequests, 0)
	atomic.StoreInt64(&m.cacheHits, 0)
	atomic.StoreInt64(&m.cacheMisses, 0)
	atomic.StoreInt64(&m.errors, 0)
	atomic.StoreInt64(&m.totalLatency, 0)
	atomic.StoreInt64(&m.latencyCount, 0)
	atomic.StoreInt64(&m.claudeInputTokens, 0)
	atomic.StoreInt64(&m.claudeOutputTokens, 0)
	atomic.StoreInt64(&m.claudeTotalCost, 0)
	atomic.StoreInt64(&m.claudeRequests, 0)
}

// MetricsStats holds snapshot of metrics
type MetricsStats struct {
	TotalRequests int64               `json:"total_requests"`
	CacheHits     int64               `json:"cache_hits"`
	CacheMisses   int64               `json:"cache_misses"`
	Errors        int64               `json:"errors"`
	CacheHitRatio float64             `json:"cache_hit_ratio"`
	AvgLatency    time.Duration       `json:"avg_latency"`
	MemoryUsageMB float64             `json:"memory_usage_mb"`
	CPUPercent    float64             `json:"cpu_percent"`
	ClaudeMetrics *ClaudeTokenMetrics `json:"claude_metrics,omitempty"`
}
