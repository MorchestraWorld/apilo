package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// MonitoringSnapshot represents a complete point-in-time capture of all metrics
type MonitoringSnapshot struct {
	Timestamp time.Time `json:"timestamp"`

	// Cache metrics
	CacheHitRatio      float64 `json:"cache_hit_ratio"`
	CacheMissRatio     float64 `json:"cache_miss_ratio"`
	CacheSize          int     `json:"cache_size"`
	CacheCapacity      int     `json:"cache_capacity"`
	CacheMemoryUsageMB float64 `json:"cache_memory_usage_mb"`
	CachePeakMemoryMB  float64 `json:"cache_peak_memory_mb"`
	CacheTotalGets     int64   `json:"cache_total_gets"`
	CacheTotalHits     int64   `json:"cache_total_hits"`
	CacheTotalMisses   int64   `json:"cache_total_misses"`
	CacheEvictions     int64   `json:"cache_evictions"`
	CacheExpirations   int64   `json:"cache_expirations"`

	// Benchmark metrics (from last run)
	LatencyP50  float64 `json:"latency_p50_ms"`
	LatencyP95  float64 `json:"latency_p95_ms"`
	LatencyP99  float64 `json:"latency_p99_ms"`
	LatencyMean float64 `json:"latency_mean_ms"`
	LatencyMin  float64 `json:"latency_min_ms"`
	LatencyMax  float64 `json:"latency_max_ms"`

	TTFBP50 float64 `json:"ttfb_p50_ms"`
	TTFBP95 float64 `json:"ttfb_p95_ms"`
	TTFBP99 float64 `json:"ttfb_p99_ms"`

	// Throughput metrics
	RequestsPerSecond float64 `json:"requests_per_second"`
	BytesPerSecond    float64 `json:"bytes_per_second"`

	// Error metrics
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	ErrorRate          float64 `json:"error_rate"`

	// Connection pool metrics
	ConnectionReuseRate float64 `json:"connection_reuse_rate"`

	// System metrics
	UptimeSeconds float64 `json:"uptime_seconds"`

	// Performance grade
	PerformanceGrade string `json:"performance_grade"`
	PerformanceScore int    `json:"performance_score"`
}

// MetricsCollector aggregates metrics from all system components
type MetricsCollector struct {
	// Component references
	benchmarker *Benchmarker
	cache       *LRUCache

	// Historical data
	snapshots    []MonitoringSnapshot
	maxSnapshots int

	// Current metrics
	currentSnapshot     *MonitoringSnapshot
	lastBenchmarkResult *BenchmarkResult

	// Synchronization
	mu sync.RWMutex

	// Timing
	collectionStart time.Time
	lastCollection  time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(maxSnapshots int) *MetricsCollector {
	if maxSnapshots <= 0 {
		maxSnapshots = 1000
	}

	return &MetricsCollector{
		snapshots:       make([]MonitoringSnapshot, 0, maxSnapshots),
		maxSnapshots:    maxSnapshots,
		collectionStart: time.Now(),
		lastCollection:  time.Now(),
	}
}

// AttachBenchmarker attaches a benchmarker for monitoring
func (mc *MetricsCollector) AttachBenchmarker(b *Benchmarker) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.benchmarker = b
}

// AttachCache attaches a cache for monitoring
func (mc *MetricsCollector) AttachCache(c *LRUCache) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cache = c
}

// UpdateBenchmarkResult updates the last benchmark result
func (mc *MetricsCollector) UpdateBenchmarkResult(result *BenchmarkResult) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.lastBenchmarkResult = result
}

// Collect gathers current metrics from all components
func (mc *MetricsCollector) Collect() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	snapshot := MonitoringSnapshot{
		Timestamp:     time.Now(),
		UptimeSeconds: time.Since(mc.collectionStart).Seconds(),
	}

	// Collect cache metrics
	if mc.cache != nil {
		cacheMetrics := mc.cache.GetMetrics()
		snapshot.CacheHitRatio = cacheMetrics.HitRatio()
		snapshot.CacheMissRatio = cacheMetrics.MissRatio()
		snapshot.CacheSize = mc.cache.Size()
		snapshot.CacheCapacity = mc.cache.Capacity()
		snapshot.CacheMemoryUsageMB = float64(mc.cache.MemoryUsage()) / (1024 * 1024)
		snapshot.CachePeakMemoryMB = float64(cacheMetrics.PeakMemoryUsage()) / (1024 * 1024)
		snapshot.CacheTotalGets = cacheMetrics.TotalGets()
		snapshot.CacheTotalHits = cacheMetrics.TotalHits()
		snapshot.CacheTotalMisses = cacheMetrics.TotalMisses()
		snapshot.CacheEvictions = cacheMetrics.TotalEvictions()
		snapshot.CacheExpirations = cacheMetrics.TotalExpirations()
		snapshot.PerformanceGrade = cacheMetrics.PerformanceGrade()
	}

	// Collect benchmark metrics (from last run)
	if mc.lastBenchmarkResult != nil {
		result := mc.lastBenchmarkResult
		snapshot.LatencyP50 = result.LatencyStats.P50
		snapshot.LatencyP95 = result.LatencyStats.P95
		snapshot.LatencyP99 = result.LatencyStats.P99
		snapshot.LatencyMean = result.LatencyStats.Mean
		snapshot.LatencyMin = result.LatencyStats.Min
		snapshot.LatencyMax = result.LatencyStats.Max

		snapshot.TTFBP50 = result.TTFBStats.P50
		snapshot.TTFBP95 = result.TTFBStats.P95
		snapshot.TTFBP99 = result.TTFBStats.P99

		snapshot.RequestsPerSecond = result.RequestsPerSecond
		snapshot.BytesPerSecond = result.BytesPerSecond

		snapshot.TotalRequests = result.TotalRequests
		snapshot.SuccessfulRequests = result.SuccessfulReqs
		snapshot.FailedRequests = result.FailedReqs

		if result.TotalRequests > 0 {
			snapshot.ErrorRate = float64(result.FailedReqs) / float64(result.TotalRequests)
		}

		// Calculate connection reuse rate (approximation based on TCP connection time)
		if result.ConnectionStats.Samples > 0 && result.TotalRequests > 0 {
			// If TCP connection time is minimal, connections are being reused
			newConnections := result.ConnectionStats.Samples
			snapshot.ConnectionReuseRate = 1.0 - (float64(newConnections) / float64(result.TotalRequests))
		}
	}

	// Calculate performance score
	snapshot.PerformanceScore = mc.calculatePerformanceScore(snapshot)

	mc.currentSnapshot = &snapshot
	mc.lastCollection = time.Now()
}

// CaptureSnapshot captures and stores the current snapshot
func (mc *MetricsCollector) CaptureSnapshot() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.currentSnapshot == nil {
		// If no current snapshot, collect one first
		mc.mu.Unlock()
		mc.Collect()
		mc.mu.Lock()
	}

	// Store snapshot
	mc.snapshots = append(mc.snapshots, *mc.currentSnapshot)

	// Limit snapshot history
	if len(mc.snapshots) > mc.maxSnapshots {
		mc.snapshots = mc.snapshots[1:]
	}
}

// GetSnapshot returns the current monitoring snapshot
func (mc *MetricsCollector) GetSnapshot() *MonitoringSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.currentSnapshot == nil {
		return nil
	}

	// Return a copy
	snapshot := *mc.currentSnapshot
	return &snapshot
}

// GetSnapshots returns all historical snapshots
func (mc *MetricsCollector) GetSnapshots() []MonitoringSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy
	snapshots := make([]MonitoringSnapshot, len(mc.snapshots))
	copy(snapshots, mc.snapshots)
	return snapshots
}

// GetSnapshotsSince returns snapshots since the given time
func (mc *MetricsCollector) GetSnapshotsSince(since time.Time) []MonitoringSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	filtered := make([]MonitoringSnapshot, 0)
	for _, snapshot := range mc.snapshots {
		if snapshot.Timestamp.After(since) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

// GetSnapshotsInRange returns snapshots within a time range
func (mc *MetricsCollector) GetSnapshotsInRange(start, end time.Time) []MonitoringSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	filtered := make([]MonitoringSnapshot, 0)
	for _, snapshot := range mc.snapshots {
		if snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

// CleanupOldSnapshots removes snapshots older than the retention period
func (mc *MetricsCollector) CleanupOldSnapshots(retentionPeriod time.Duration) int {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	cutoff := time.Now().Add(-retentionPeriod)
	newSnapshots := make([]MonitoringSnapshot, 0, len(mc.snapshots))

	for _, snapshot := range mc.snapshots {
		if snapshot.Timestamp.After(cutoff) {
			newSnapshots = append(newSnapshots, snapshot)
		}
	}

	removed := len(mc.snapshots) - len(newSnapshots)
	mc.snapshots = newSnapshots

	return removed
}

// GetMetricsSummary returns a summary of key metrics
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.currentSnapshot == nil {
		return map[string]interface{}{
			"status": "no data available",
		}
	}

	snapshot := mc.currentSnapshot

	return map[string]interface{}{
		"cache": map[string]interface{}{
			"hit_ratio":       snapshot.CacheHitRatio,
			"size":            snapshot.CacheSize,
			"capacity":        snapshot.CacheCapacity,
			"memory_usage_mb": snapshot.CacheMemoryUsageMB,
		},
		"latency": map[string]interface{}{
			"p50_ms":  snapshot.LatencyP50,
			"p95_ms":  snapshot.LatencyP95,
			"p99_ms":  snapshot.LatencyP99,
			"mean_ms": snapshot.LatencyMean,
		},
		"throughput": map[string]interface{}{
			"requests_per_second": snapshot.RequestsPerSecond,
			"bytes_per_second":    snapshot.BytesPerSecond,
		},
		"errors": map[string]interface{}{
			"error_rate":      snapshot.ErrorRate,
			"total_requests":  snapshot.TotalRequests,
			"failed_requests": snapshot.FailedRequests,
		},
		"performance": map[string]interface{}{
			"grade": snapshot.PerformanceGrade,
			"score": snapshot.PerformanceScore,
		},
	}
}

// GetTrendAnalysis analyzes trends in the collected metrics
func (mc *MetricsCollector) GetTrendAnalysis(duration time.Duration) map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	since := time.Now().Add(-duration)
	relevantSnapshots := make([]MonitoringSnapshot, 0)

	for _, snapshot := range mc.snapshots {
		if snapshot.Timestamp.After(since) {
			relevantSnapshots = append(relevantSnapshots, snapshot)
		}
	}

	if len(relevantSnapshots) < 2 {
		return map[string]interface{}{
			"status": "insufficient data for trend analysis",
		}
	}

	// Calculate trends
	first := relevantSnapshots[0]
	last := relevantSnapshots[len(relevantSnapshots)-1]

	return map[string]interface{}{
		"period":  duration.String(),
		"samples": len(relevantSnapshots),
		"cache_hit_ratio": map[string]interface{}{
			"start":  first.CacheHitRatio,
			"end":    last.CacheHitRatio,
			"change": last.CacheHitRatio - first.CacheHitRatio,
		},
		"latency_p95": map[string]interface{}{
			"start":  first.LatencyP95,
			"end":    last.LatencyP95,
			"change": last.LatencyP95 - first.LatencyP95,
		},
		"latency_p99": map[string]interface{}{
			"start":  first.LatencyP99,
			"end":    last.LatencyP99,
			"change": last.LatencyP99 - first.LatencyP99,
		},
		"throughput": map[string]interface{}{
			"start":  first.RequestsPerSecond,
			"end":    last.RequestsPerSecond,
			"change": last.RequestsPerSecond - first.RequestsPerSecond,
		},
		"error_rate": map[string]interface{}{
			"start":  first.ErrorRate,
			"end":    last.ErrorRate,
			"change": last.ErrorRate - first.ErrorRate,
		},
	}
}

// SaveReport saves a comprehensive metrics report to a file
func (mc *MetricsCollector) SaveReport(filepath string) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	report := map[string]interface{}{
		"generated_at":       time.Now(),
		"uptime_seconds":     time.Since(mc.collectionStart).Seconds(),
		"current_snapshot":   mc.currentSnapshot,
		"total_snapshots":    len(mc.snapshots),
		"metrics_summary":    mc.GetMetricsSummary(),
		"trend_analysis_1h":  mc.GetTrendAnalysis(1 * time.Hour),
		"trend_analysis_24h": mc.GetTrendAnalysis(24 * time.Hour),
		"snapshots":          mc.snapshots,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// calculatePerformanceScore calculates an overall performance score
func (mc *MetricsCollector) calculatePerformanceScore(snapshot MonitoringSnapshot) int {
	score := 0

	// Cache performance (30 points)
	if snapshot.CacheHitRatio >= 0.80 {
		score += 30
	} else if snapshot.CacheHitRatio >= 0.70 {
		score += 25
	} else if snapshot.CacheHitRatio >= 0.60 {
		score += 20
	} else if snapshot.CacheHitRatio >= 0.50 {
		score += 15
	} else {
		score += int(snapshot.CacheHitRatio * 20)
	}

	// Latency performance (40 points)
	if snapshot.LatencyP95 < 100 {
		score += 40
	} else if snapshot.LatencyP95 < 200 {
		score += 35
	} else if snapshot.LatencyP95 < 300 {
		score += 30
	} else if snapshot.LatencyP95 < 500 {
		score += 20
	} else if snapshot.LatencyP95 < 1000 {
		score += 10
	} else {
		score += 5
	}

	// Error rate (20 points)
	if snapshot.ErrorRate == 0 {
		score += 20
	} else if snapshot.ErrorRate < 0.01 {
		score += 15
	} else if snapshot.ErrorRate < 0.05 {
		score += 10
	} else if snapshot.ErrorRate < 0.10 {
		score += 5
	}

	// Throughput (10 points)
	if snapshot.RequestsPerSecond > 1000 {
		score += 10
	} else if snapshot.RequestsPerSecond > 500 {
		score += 8
	} else if snapshot.RequestsPerSecond > 100 {
		score += 6
	} else if snapshot.RequestsPerSecond > 50 {
		score += 4
	} else {
		score += 2
	}

	return score
}

// PrintSummary displays a summary of current metrics
func (s *MonitoringSnapshot) PrintSummary() {
	fmt.Printf("=== Performance Metrics Summary ===\n\n")

	// Cache metrics
	fmt.Printf("--- Cache Performance ---\n")
	fmt.Printf("Hit Ratio: %.2f%%\n", s.CacheHitRatio*100)
	fmt.Printf("Size: %d / %d (%.1f%% full)\n", s.CacheSize, s.CacheCapacity,
		float64(s.CacheSize)/float64(s.CacheCapacity)*100)
	fmt.Printf("Memory Usage: %.2f MB (Peak: %.2f MB)\n",
		s.CacheMemoryUsageMB, s.CachePeakMemoryMB)
	fmt.Printf("Total Gets: %d (Hits: %d, Misses: %d)\n",
		s.CacheTotalGets, s.CacheTotalHits, s.CacheTotalMisses)
	fmt.Printf("Evictions: %d, Expirations: %d\n",
		s.CacheEvictions, s.CacheExpirations)

	// Latency metrics
	fmt.Printf("\n--- Latency Statistics ---\n")
	fmt.Printf("P50: %.2f ms\n", s.LatencyP50)
	fmt.Printf("P95: %.2f ms\n", s.LatencyP95)
	fmt.Printf("P99: %.2f ms\n", s.LatencyP99)
	fmt.Printf("Mean: %.2f ms (Min: %.2f, Max: %.2f)\n",
		s.LatencyMean, s.LatencyMin, s.LatencyMax)

	// TTFB metrics
	fmt.Printf("\n--- Time to First Byte ---\n")
	fmt.Printf("P50: %.2f ms\n", s.TTFBP50)
	fmt.Printf("P95: %.2f ms\n", s.TTFBP95)
	fmt.Printf("P99: %.2f ms\n", s.TTFBP99)

	// Throughput metrics
	fmt.Printf("\n--- Throughput ---\n")
	fmt.Printf("Requests/sec: %.2f\n", s.RequestsPerSecond)
	fmt.Printf("Bytes/sec: %.2f (%.2f KB/s)\n",
		s.BytesPerSecond, s.BytesPerSecond/1024)

	// Error metrics
	fmt.Printf("\n--- Reliability ---\n")
	fmt.Printf("Total Requests: %d\n", s.TotalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n",
		s.SuccessfulRequests,
		float64(s.SuccessfulRequests)/float64(s.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.2f%% error rate)\n",
		s.FailedRequests, s.ErrorRate*100)

	// Connection metrics
	fmt.Printf("\n--- Connection Pool ---\n")
	fmt.Printf("Connection Reuse Rate: %.2f%%\n", s.ConnectionReuseRate*100)

	// Overall performance
	fmt.Printf("\n--- Overall Performance ---\n")
	fmt.Printf("Grade: %s\n", s.PerformanceGrade)
	fmt.Printf("Score: %d/100\n", s.PerformanceScore)
	fmt.Printf("Uptime: %.2f seconds\n", s.UptimeSeconds)
}
