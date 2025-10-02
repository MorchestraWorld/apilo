package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// CacheMetrics tracks performance metrics for cache operations
type CacheMetrics struct {
	// Request metrics (atomic counters for thread-safe increments)
	totalGets        int64
	totalHits        int64
	totalMisses      int64
	totalInserts     int64
	totalUpdates     int64
	totalEvictions   int64
	totalExpirations int64

	// Performance metrics
	avgAccessLatency time.Duration
	maxAccessLatency time.Duration
	minAccessLatency time.Duration

	// Memory metrics
	currentMemoryUsage int64
	peakMemoryUsage    int64

	// Time-based metrics
	startTime     time.Time
	lastResetTime time.Time

	// Detailed tracking
	hitsByHour     map[int]int64
	missByHour     map[int]int64
	latencyBuckets map[string]int64 // Histogram of latencies

	// Synchronization
	mu sync.RWMutex

	// Historical snapshots
	snapshots    []MetricsSnapshot
	maxSnapshots int
}

// MetricsSnapshot represents a point-in-time capture of metrics
type MetricsSnapshot struct {
	Timestamp         time.Time     `json:"timestamp"`
	TotalGets         int64         `json:"total_gets"`
	TotalHits         int64         `json:"total_hits"`
	TotalMisses       int64         `json:"total_misses"`
	HitRatio          float64       `json:"hit_ratio"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
	AvgAccessLatency  time.Duration `json:"avg_access_latency"`
	RequestsPerSecond float64       `json:"requests_per_second"`
}

// NewCacheMetrics creates a new metrics tracker
func NewCacheMetrics() *CacheMetrics {
	now := time.Now()
	return &CacheMetrics{
		startTime:        now,
		lastResetTime:    now,
		hitsByHour:       make(map[int]int64),
		missByHour:       make(map[int]int64),
		latencyBuckets:   make(map[string]int64),
		snapshots:        make([]MetricsSnapshot, 0, 1000),
		maxSnapshots:     1000,
		minAccessLatency: time.Hour, // Start with a high value
	}
}

// RecordGet records a cache get operation
func (m *CacheMetrics) RecordGet() {
	atomic.AddInt64(&m.totalGets, 1)
}

// RecordHit records a cache hit
func (m *CacheMetrics) RecordHit() {
	atomic.AddInt64(&m.totalHits, 1)

	m.mu.Lock()
	hour := time.Now().Hour()
	m.hitsByHour[hour]++
	m.mu.Unlock()
}

// RecordMiss records a cache miss
func (m *CacheMetrics) RecordMiss() {
	atomic.AddInt64(&m.totalMisses, 1)

	m.mu.Lock()
	hour := time.Now().Hour()
	m.missByHour[hour]++
	m.mu.Unlock()
}

// RecordInsert records a cache insertion
func (m *CacheMetrics) RecordInsert() {
	atomic.AddInt64(&m.totalInserts, 1)
}

// RecordUpdate records a cache update
func (m *CacheMetrics) RecordUpdate() {
	atomic.AddInt64(&m.totalUpdates, 1)
}

// RecordEviction records a cache eviction
func (m *CacheMetrics) RecordEviction() {
	atomic.AddInt64(&m.totalEvictions, 1)
}

// RecordExpiration records a cache expiration
func (m *CacheMetrics) RecordExpiration() {
	atomic.AddInt64(&m.totalExpirations, 1)
}

// RecordClear records a cache clear operation
func (m *CacheMetrics) RecordClear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all counters
	atomic.StoreInt64(&m.totalGets, 0)
	atomic.StoreInt64(&m.totalHits, 0)
	atomic.StoreInt64(&m.totalMisses, 0)
	atomic.StoreInt64(&m.totalInserts, 0)
	atomic.StoreInt64(&m.totalUpdates, 0)
	atomic.StoreInt64(&m.totalEvictions, 0)
	atomic.StoreInt64(&m.totalExpirations, 0)

	m.currentMemoryUsage = 0
	m.lastResetTime = time.Now()
}

// RecordAccessLatency records the latency of a cache access
func (m *CacheMetrics) RecordAccessLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update average (simple moving average)
	totalGets := atomic.LoadInt64(&m.totalGets)
	if totalGets > 0 {
		currentAvg := m.avgAccessLatency
		m.avgAccessLatency = (currentAvg*time.Duration(totalGets-1) + latency) / time.Duration(totalGets)
	} else {
		m.avgAccessLatency = latency
	}

	// Update min/max
	if latency < m.minAccessLatency {
		m.minAccessLatency = latency
	}
	if latency > m.maxAccessLatency {
		m.maxAccessLatency = latency
	}

	// Update latency histogram
	bucket := m.getLatencyBucket(latency)
	m.latencyBuckets[bucket]++
}

// RecordMemoryUsage records current memory usage
func (m *CacheMetrics) RecordMemoryUsage(bytes int64) {
	atomic.StoreInt64(&m.currentMemoryUsage, bytes)

	m.mu.Lock()
	if bytes > m.peakMemoryUsage {
		m.peakMemoryUsage = bytes
	}
	m.mu.Unlock()
}

// RecordCleanup records a cleanup operation
func (m *CacheMetrics) RecordCleanup(entriesRemoved int) {
	// Cleanup is essentially multiple expirations
	atomic.AddInt64(&m.totalExpirations, int64(entriesRemoved))
}

// HitRatio returns the cache hit ratio (0.0 to 1.0)
func (m *CacheMetrics) HitRatio() float64 {
	totalGets := atomic.LoadInt64(&m.totalGets)
	if totalGets == 0 {
		return 0.0
	}
	totalHits := atomic.LoadInt64(&m.totalHits)
	return float64(totalHits) / float64(totalGets)
}

// MissRatio returns the cache miss ratio (0.0 to 1.0)
func (m *CacheMetrics) MissRatio() float64 {
	return 1.0 - m.HitRatio()
}

// TotalGets returns total number of get operations
func (m *CacheMetrics) TotalGets() int64 {
	return atomic.LoadInt64(&m.totalGets)
}

// TotalHits returns total number of cache hits
func (m *CacheMetrics) TotalHits() int64 {
	return atomic.LoadInt64(&m.totalHits)
}

// TotalMisses returns total number of cache misses
func (m *CacheMetrics) TotalMisses() int64 {
	return atomic.LoadInt64(&m.totalMisses)
}

// TotalInserts returns total number of insertions
func (m *CacheMetrics) TotalInserts() int64 {
	return atomic.LoadInt64(&m.totalInserts)
}

// TotalUpdates returns total number of updates
func (m *CacheMetrics) TotalUpdates() int64 {
	return atomic.LoadInt64(&m.totalUpdates)
}

// TotalEvictions returns total number of evictions
func (m *CacheMetrics) TotalEvictions() int64 {
	return atomic.LoadInt64(&m.totalEvictions)
}

// TotalExpirations returns total number of expirations
func (m *CacheMetrics) TotalExpirations() int64 {
	return atomic.LoadInt64(&m.totalExpirations)
}

// AvgAccessLatency returns the average access latency
func (m *CacheMetrics) AvgAccessLatency() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.avgAccessLatency
}

// RequestsPerSecond calculates the current request rate
func (m *CacheMetrics) RequestsPerSecond() float64 {
	totalGets := atomic.LoadInt64(&m.totalGets)
	elapsed := time.Since(m.lastResetTime).Seconds()
	if elapsed > 0 {
		return float64(totalGets) / elapsed
	}
	return 0.0
}

// CurrentMemoryUsage returns current memory usage
func (m *CacheMetrics) CurrentMemoryUsage() int64 {
	return atomic.LoadInt64(&m.currentMemoryUsage)
}

// PeakMemoryUsage returns peak memory usage
func (m *CacheMetrics) PeakMemoryUsage() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peakMemoryUsage
}

// GetSnapshot creates a snapshot of current metrics
func (m *CacheMetrics) GetSnapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Timestamp:         time.Now(),
		TotalGets:         m.TotalGets(),
		TotalHits:         m.TotalHits(),
		TotalMisses:       m.TotalMisses(),
		HitRatio:          m.HitRatio(),
		MemoryUsage:       m.CurrentMemoryUsage(),
		AvgAccessLatency:  m.AvgAccessLatency(),
		RequestsPerSecond: m.RequestsPerSecond(),
	}
}

// CaptureSnapshot captures and stores a metrics snapshot
func (m *CacheMetrics) CaptureSnapshot() {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot := m.GetSnapshot()
	m.snapshots = append(m.snapshots, snapshot)

	// Limit snapshot history
	if len(m.snapshots) > m.maxSnapshots {
		m.snapshots = m.snapshots[1:]
	}
}

// GetSnapshots returns all captured snapshots
func (m *CacheMetrics) GetSnapshots() []MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	snapshots := make([]MetricsSnapshot, len(m.snapshots))
	copy(snapshots, m.snapshots)
	return snapshots
}

// GetDetailedStats returns comprehensive statistics
func (m *CacheMetrics) GetDetailedStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	return map[string]interface{}{
		// Basic counters
		"total_gets":        m.TotalGets(),
		"total_hits":        m.TotalHits(),
		"total_misses":      m.TotalMisses(),
		"total_inserts":     m.TotalInserts(),
		"total_updates":     m.TotalUpdates(),
		"total_evictions":   m.TotalEvictions(),
		"total_expirations": m.TotalExpirations(),

		// Ratios
		"hit_ratio":  m.HitRatio(),
		"miss_ratio": m.MissRatio(),

		// Performance
		"avg_access_latency_ms": m.avgAccessLatency.Milliseconds(),
		"min_access_latency_ms": m.minAccessLatency.Milliseconds(),
		"max_access_latency_ms": m.maxAccessLatency.Milliseconds(),
		"requests_per_second":   m.RequestsPerSecond(),

		// Memory
		"current_memory_bytes": m.CurrentMemoryUsage(),
		"peak_memory_bytes":    m.peakMemoryUsage,
		"current_memory_mb":    float64(m.CurrentMemoryUsage()) / (1024 * 1024),
		"peak_memory_mb":       float64(m.peakMemoryUsage) / (1024 * 1024),

		// Time
		"uptime_seconds":  uptime.Seconds(),
		"uptime_hours":    uptime.Hours(),
		"start_time":      m.startTime,
		"last_reset_time": m.lastResetTime,

		// Hourly distribution
		"hits_by_hour":   m.hitsByHour,
		"misses_by_hour": m.missByHour,

		// Latency distribution
		"latency_buckets": m.latencyBuckets,
	}
}

// GetSummary returns a human-readable summary
func (m *CacheMetrics) GetSummary() string {
	stats := m.GetDetailedStats()

	summary := "Cache Performance Metrics\n"
	summary += "=========================\n\n"
	summary += fmt.Sprintf("Hit Ratio: %.2f%%\n", m.HitRatio()*100)
	summary += fmt.Sprintf("Total Requests: %d\n", m.TotalGets())
	summary += fmt.Sprintf("Cache Hits: %d\n", m.TotalHits())
	summary += fmt.Sprintf("Cache Misses: %d\n", m.TotalMisses())
	summary += fmt.Sprintf("Average Access Latency: %v\n", m.AvgAccessLatency())
	summary += fmt.Sprintf("Requests/Second: %.2f\n", m.RequestsPerSecond())
	summary += fmt.Sprintf("Current Memory: %.2f MB\n", stats["current_memory_mb"])
	summary += fmt.Sprintf("Peak Memory: %.2f MB\n", stats["peak_memory_mb"])
	summary += fmt.Sprintf("Total Evictions: %d\n", m.TotalEvictions())
	summary += fmt.Sprintf("Total Expirations: %d\n", m.TotalExpirations())
	summary += fmt.Sprintf("Uptime: %.2f hours\n", stats["uptime_hours"])

	return summary
}

// ToJSON serializes metrics to JSON
func (m *CacheMetrics) ToJSON() ([]byte, error) {
	stats := m.GetDetailedStats()
	return json.MarshalIndent(stats, "", "  ")
}

// Reset resets all metrics
func (m *CacheMetrics) Reset() {
	m.RecordClear()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.avgAccessLatency = 0
	m.maxAccessLatency = 0
	m.minAccessLatency = time.Hour
	m.peakMemoryUsage = 0
	m.hitsByHour = make(map[int]int64)
	m.missByHour = make(map[int]int64)
	m.latencyBuckets = make(map[string]int64)
	m.snapshots = make([]MetricsSnapshot, 0, m.maxSnapshots)
	m.startTime = time.Now()
	m.lastResetTime = time.Now()
}

// StartPeriodicCapture starts capturing snapshots at regular intervals
func (m *CacheMetrics) StartPeriodicCapture(interval time.Duration) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CaptureSnapshot()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// getLatencyBucket categorizes latency into buckets for histogram
func (m *CacheMetrics) getLatencyBucket(latency time.Duration) string {
	microseconds := latency.Microseconds()

	switch {
	case microseconds < 100:
		return "<100us"
	case microseconds < 500:
		return "100-500us"
	case microseconds < 1000:
		return "500us-1ms"
	case microseconds < 5000:
		return "1-5ms"
	case microseconds < 10000:
		return "5-10ms"
	case microseconds < 50000:
		return "10-50ms"
	case microseconds < 100000:
		return "50-100ms"
	default:
		return ">100ms"
	}
}

// MeetsTarget checks if cache performance meets specified targets
func (m *CacheMetrics) MeetsTarget(hitRatioTarget float64, maxLatency time.Duration) bool {
	hitRatio := m.HitRatio()
	avgLatency := m.AvgAccessLatency()

	return hitRatio >= hitRatioTarget && avgLatency <= maxLatency
}

// PerformanceGrade returns a letter grade for cache performance
func (m *CacheMetrics) PerformanceGrade() string {
	hitRatio := m.HitRatio()
	latency := m.AvgAccessLatency()

	// Grading based on hit ratio and latency
	score := 0

	// Hit ratio scoring (60 points max)
	if hitRatio >= 0.75 {
		score += 60
	} else if hitRatio >= 0.60 {
		score += 50
	} else if hitRatio >= 0.50 {
		score += 40
	} else if hitRatio >= 0.40 {
		score += 30
	} else {
		score += int(hitRatio * 50)
	}

	// Latency scoring (40 points max)
	if latency < 500*time.Microsecond {
		score += 40
	} else if latency < 1*time.Millisecond {
		score += 35
	} else if latency < 5*time.Millisecond {
		score += 25
	} else if latency < 10*time.Millisecond {
		score += 15
	} else {
		score += 5
	}

	// Convert to letter grade
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
