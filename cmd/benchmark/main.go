// Comprehensive Performance Baseline Establishment
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"
)

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	Timestamp  time.Time
	TTL        time.Duration
}

// PerformanceBaseline comprehensive system performance metrics
type PerformanceBaseline struct {
	SystemMetrics       *SystemMetrics
	NetworkMetrics      *NetworkMetrics
	ApplicationMetrics  *ApplicationMetrics
	ResourceMetrics     *ResourceMetrics
	Timestamp          time.Time
}

// SystemMetrics tracks system-level performance
type SystemMetrics struct {
	CPUCores           int
	MemoryTotal        uint64
	GoroutineCount     int
	GCStats            runtime.MemStats
	OSThreads          int
}

// NetworkMetrics tracks network performance
type NetworkMetrics struct {
	DNSLatency         time.Duration
	ConnectLatency     time.Duration
	TLSLatency         time.Duration
	TTFBLatency        time.Duration
	TotalLatency       time.Duration
	ConnectionReused   bool
	Protocol           string
	StatusCode         int
}

// ApplicationMetrics tracks application performance
type ApplicationMetrics struct {
	CacheHitRatio      float64
	CacheSize          int
	RequestsPerSecond  float64
	AverageLatency     time.Duration
	P50Latency         time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	ErrorRate          float64
}

// ResourceMetrics tracks resource utilization
type ResourceMetrics struct {
	MemoryUsage        uint64
	MemoryAllocated    uint64
	GCPauseTime        time.Duration
	GCFrequency        uint32
	HeapSize           uint64
	StackSize          uint64
}

// ComprehensiveBenchmark performs full system performance analysis
type ComprehensiveBenchmark struct {
	client        *http.Client
	cache         map[string]*CachedResponse
	cacheMutex    sync.RWMutex
	metrics       []NetworkMetrics
	metricsMutex  sync.Mutex
}

// NewComprehensiveBenchmark creates a new comprehensive benchmark
func NewComprehensiveBenchmark() *ComprehensiveBenchmark {
	// Create optimized HTTP/2 client
	transport := &http.Transport{
		MaxIdleConns:       100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:    90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		ForceAttemptHTTP2: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &ComprehensiveBenchmark{
		client: client,
		cache:  make(map[string]*CachedResponse),
		metrics: make([]NetworkMetrics, 0),
	}
}

// EstablishBaseline creates comprehensive performance baseline
func (cb *ComprehensiveBenchmark) EstablishBaseline(url string, iterations int) (*PerformanceBaseline, error) {
	fmt.Printf(`
⚡ Phase 1: Performance Baseline Establishment
=============================================
Target URL: %s
Iterations: %d
Analysis: Comprehensive system performance profiling

`, url, iterations)

	// Collect system metrics
	systemMetrics := cb.collectSystemMetrics()

	fmt.Println("📊 System Configuration:")
	fmt.Printf("  CPU Cores:     %d\n", systemMetrics.CPUCores)
	fmt.Printf("  Memory Total:  %d MB\n", systemMetrics.MemoryTotal/(1024*1024))
	fmt.Printf("  Goroutines:    %d\n", systemMetrics.GoroutineCount)
	fmt.Printf("  OS Threads:    %d\n", systemMetrics.OSThreads)

	// Perform baseline measurements
	fmt.Println("\n🔍 Collecting baseline performance data...")

	var networkMetrics []NetworkMetrics
	start := time.Now()

	for i := 0; i < iterations; i++ {
		metric, err := cb.measureSingleRequest(url, false)
		if err != nil {
			fmt.Printf("❌ Request %d failed: %v\n", i+1, err)
			continue
		}
		networkMetrics = append(networkMetrics, *metric)

		if (i+1) % 25 == 0 {
			fmt.Printf("  Progress: %d/%d requests completed\n", i+1, iterations)
		}
	}

	duration := time.Since(start)

	// Calculate application metrics
	appMetrics := cb.calculateApplicationMetrics(networkMetrics, duration)

	// Collect resource metrics
	resourceMetrics := cb.collectResourceMetrics()

	baseline := &PerformanceBaseline{
		SystemMetrics:      systemMetrics,
		NetworkMetrics:     &networkMetrics[0], // Representative sample
		ApplicationMetrics: appMetrics,
		ResourceMetrics:    resourceMetrics,
		Timestamp:         time.Now(),
	}

	cb.printBaselineReport(baseline, networkMetrics)

	return baseline, nil
}

// collectSystemMetrics gathers system-level performance data
func (cb *ComprehensiveBenchmark) collectSystemMetrics() *SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &SystemMetrics{
		CPUCores:       runtime.NumCPU(),
		MemoryTotal:    memStats.Sys,
		GoroutineCount: runtime.NumGoroutine(),
		GCStats:        memStats,
		OSThreads:      runtime.GOMAXPROCS(0),
	}
}

// measureSingleRequest performs detailed timing measurement
func (cb *ComprehensiveBenchmark) measureSingleRequest(url string, useCache bool) (*NetworkMetrics, error) {
	start := time.Now()

	// Check cache first if enabled
	if useCache {
		if cached := cb.getFromCache(url); cached != nil {
			return &NetworkMetrics{
				TotalLatency:     time.Since(start),
				ConnectionReused: true,
				Protocol:        "cached",
				StatusCode:      200,
			}, nil
		}
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Measure request timing
	reqStart := time.Now()
	resp, err := cb.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	totalLatency := time.Since(reqStart)

	// Cache response if enabled
	if useCache && resp.StatusCode == 200 {
		cb.cacheResponse(url, resp)
	}

	return &NetworkMetrics{
		TotalLatency:     totalLatency,
		ConnectionReused: false, // Simplified for baseline
		Protocol:        resp.Proto,
		StatusCode:      resp.StatusCode,
	}, nil
}

// getFromCache retrieves cached response
func (cb *ComprehensiveBenchmark) getFromCache(url string) *CachedResponse {
	cb.cacheMutex.RLock()
	defer cb.cacheMutex.RUnlock()

	cached, exists := cb.cache[url]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(cached.Timestamp) > cached.TTL {
		go func() {
			cb.cacheMutex.Lock()
			delete(cb.cache, url)
			cb.cacheMutex.Unlock()
		}()
		return nil
	}

	return cached
}

// cacheResponse stores response in cache
func (cb *ComprehensiveBenchmark) cacheResponse(url string, resp *http.Response) {
	cached := &CachedResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Timestamp:  time.Now(),
		TTL:        5 * time.Minute,
	}

	cb.cacheMutex.Lock()
	cb.cache[url] = cached
	cb.cacheMutex.Unlock()
}

// calculateApplicationMetrics computes application performance metrics
func (cb *ComprehensiveBenchmark) calculateApplicationMetrics(metrics []NetworkMetrics, duration time.Duration) *ApplicationMetrics {
	if len(metrics) == 0 {
		return &ApplicationMetrics{}
	}

	// Sort latencies for percentile calculation
	latencies := make([]time.Duration, len(metrics))
	var totalLatency time.Duration
	var errorCount int

	for i, m := range metrics {
		latencies[i] = m.TotalLatency
		totalLatency += m.TotalLatency
		if m.StatusCode >= 400 {
			errorCount++
		}
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate percentiles
	p50Index := len(latencies) * 50 / 100
	p95Index := len(latencies) * 95 / 100
	p99Index := len(latencies) * 99 / 100

	// Calculate cache metrics
	cacheHits := 0
	for _, m := range metrics {
		if m.Protocol == "cached" {
			cacheHits++
		}
	}

	return &ApplicationMetrics{
		CacheHitRatio:     float64(cacheHits) / float64(len(metrics)) * 100,
		CacheSize:         len(cb.cache),
		RequestsPerSecond: float64(len(metrics)) / duration.Seconds(),
		AverageLatency:    totalLatency / time.Duration(len(metrics)),
		P50Latency:        latencies[p50Index],
		P95Latency:        latencies[p95Index],
		P99Latency:        latencies[p99Index],
		ErrorRate:         float64(errorCount) / float64(len(metrics)) * 100,
	}
}

// collectResourceMetrics gathers resource utilization data
func (cb *ComprehensiveBenchmark) collectResourceMetrics() *ResourceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &ResourceMetrics{
		MemoryUsage:     memStats.Alloc,
		MemoryAllocated: memStats.TotalAlloc,
		GCPauseTime:     time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		GCFrequency:     memStats.NumGC,
		HeapSize:        memStats.HeapAlloc,
		StackSize:       memStats.StackInuse,
	}
}

// printBaselineReport displays comprehensive baseline analysis
func (cb *ComprehensiveBenchmark) printBaselineReport(baseline *PerformanceBaseline, metrics []NetworkMetrics) {
	fmt.Printf(`
📊 Performance Baseline Report
==============================
Timestamp: %s

🖥️  System Performance Characteristics
======================================
CPU Cores:              %d
Memory Total:           %d MB
Active Goroutines:      %d
OS Thread Limit:        %d

📡 Network Performance Baseline
===============================
Average Latency:        %v
P50 Latency:           %v
P95 Latency:           %v
P99 Latency:           %v
Requests/Second:       %.2f
Error Rate:            %.2f%%

💾 Resource Utilization Baseline
=================================
Memory Usage:          %d MB
Heap Size:             %d MB
Stack Usage:           %d KB
GC Frequency:          %d cycles
Last GC Pause:         %v

🔧 Application Performance Baseline
====================================
Cache Hit Ratio:       %.1f%%
Cache Size:            %d entries
Protocol Usage:        %s
Success Rate:          %.2f%%

📈 Performance Classification
=============================
`,
		baseline.Timestamp.Format("2006-01-02 15:04:05"),
		baseline.SystemMetrics.CPUCores,
		baseline.SystemMetrics.MemoryTotal/(1024*1024),
		baseline.SystemMetrics.GoroutineCount,
		baseline.SystemMetrics.OSThreads,
		baseline.ApplicationMetrics.AverageLatency,
		baseline.ApplicationMetrics.P50Latency,
		baseline.ApplicationMetrics.P95Latency,
		baseline.ApplicationMetrics.P99Latency,
		baseline.ApplicationMetrics.RequestsPerSecond,
		baseline.ApplicationMetrics.ErrorRate,
		baseline.ResourceMetrics.MemoryUsage/(1024*1024),
		baseline.ResourceMetrics.HeapSize/(1024*1024),
		baseline.ResourceMetrics.StackSize/1024,
		baseline.ResourceMetrics.GCFrequency,
		baseline.ResourceMetrics.GCPauseTime,
		baseline.ApplicationMetrics.CacheHitRatio,
		baseline.ApplicationMetrics.CacheSize,
		baseline.NetworkMetrics.Protocol,
		100.0-baseline.ApplicationMetrics.ErrorRate,
	)

	// Performance classification
	avgLatency := baseline.ApplicationMetrics.AverageLatency
	rps := baseline.ApplicationMetrics.RequestsPerSecond

	fmt.Printf("Latency Classification: %s\n", classifyLatency(avgLatency))
	fmt.Printf("Throughput Classification: %s\n", classifyThroughput(rps))
	fmt.Printf("Cache Effectiveness: %s\n", classifyCacheEffectiveness(baseline.ApplicationMetrics.CacheHitRatio))
	fmt.Printf("Resource Efficiency: %s\n", classifyResourceEfficiency(baseline.ResourceMetrics))

	fmt.Printf(`
✅ Baseline Establishment Complete
==================================
Performance baseline successfully established with %d samples.
System ready for bottleneck analysis and optimization.

Next Phase: Bottleneck Identification & Analysis
`, len(metrics))
}

// Classification helper functions
func classifyLatency(latency time.Duration) string {
	if latency < 50*time.Millisecond {
		return "⚡ EXCELLENT (< 50ms)"
	} else if latency < 100*time.Millisecond {
		return "✅ GOOD (50-100ms)"
	} else if latency < 200*time.Millisecond {
		return "⚠️ ACCEPTABLE (100-200ms)"
	} else {
		return "❌ NEEDS OPTIMIZATION (> 200ms)"
	}
}

func classifyThroughput(rps float64) string {
	if rps > 100 {
		return "⚡ HIGH (> 100 RPS)"
	} else if rps > 50 {
		return "✅ MODERATE (50-100 RPS)"
	} else if rps > 20 {
		return "⚠️ LOW (20-50 RPS)"
	} else {
		return "❌ VERY LOW (< 20 RPS)"
	}
}

func classifyCacheEffectiveness(hitRatio float64) string {
	if hitRatio > 90 {
		return "⚡ EXCELLENT (> 90%)"
	} else if hitRatio > 70 {
		return "✅ GOOD (70-90%)"
	} else if hitRatio > 50 {
		return "⚠️ MODERATE (50-70%)"
	} else {
		return "❌ POOR (< 50%)"
	}
}

func classifyResourceEfficiency(metrics *ResourceMetrics) string {
	memUsageMB := metrics.MemoryUsage / (1024 * 1024)
	if memUsageMB < 50 {
		return "⚡ EXCELLENT (< 50MB)"
	} else if memUsageMB < 100 {
		return "✅ GOOD (50-100MB)"
	} else if memUsageMB < 200 {
		return "⚠️ MODERATE (100-200MB)"
	} else {
		return "❌ HIGH (> 200MB)"
	}
}

func main() {
	benchmark := NewComprehensiveBenchmark()

	// Establish comprehensive baseline with 100 samples for statistical significance
	baseline, err := benchmark.EstablishBaseline("https://httpbin.org/get", 100)
	if err != nil {
		fmt.Printf("❌ Baseline establishment failed: %v\n", err)
		return
	}

	fmt.Printf("\n🎯 Baseline establishment completed successfully!\n")
	fmt.Printf("Performance baseline established with timestamp: %s\n", baseline.Timestamp.Format("2006-01-02 15:04:05"))
}