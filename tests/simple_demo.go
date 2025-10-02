// Simple Phase 1 demonstration of API latency optimization
package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// SimpleOptimizedClient demonstrates Phase 1 optimizations
type SimpleOptimizedClient struct {
	client      *http.Client
	cache       map[string]*CachedResponse
	cacheMutex  sync.RWMutex
	stats       *SimpleStats
	statsMutex  sync.Mutex
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
	Timestamp  time.Time
	TTL        time.Duration
}

// SimpleStats tracks client performance
type SimpleStats struct {
	TotalRequests   int64
	CacheHits       int64
	CacheMisses     int64
	TotalLatency    time.Duration
	ConnectionReuse int64
}

// NewSimpleOptimizedClient creates a new optimized client
func NewSimpleOptimizedClient() *SimpleOptimizedClient {
	// Create HTTP/2 enabled client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:       100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:    90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		ForceAttemptHTTP2: true, // Enable HTTP/2
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &SimpleOptimizedClient{
		client: client,
		cache:  make(map[string]*CachedResponse),
		stats:  &SimpleStats{},
	}
}

// Do executes an HTTP request with optimizations
func (c *SimpleOptimizedClient) Do(url string, useCache bool) (*http.Response, time.Duration, bool, error) {
	start := time.Now()

	// Try cache first if enabled
	if useCache {
		if cached := c.getFromCache(url); cached != nil {
			c.recordCacheHit()
			latency := time.Since(start)

			// Create response from cache
			resp := &http.Response{
				Status:        fmt.Sprintf("%d OK", cached.StatusCode),
				StatusCode:    cached.StatusCode,
				Header:        cached.Headers,
				Body:          io.NopCloser(bytes.NewReader(cached.Body)),
				ContentLength: int64(len(cached.Body)),
			}

			return resp, latency, true, nil
		}
		c.recordCacheMiss()
	}

	// Execute HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, false, err
	}

	latency := time.Since(start)

	// Cache response if successful and caching is enabled
	if useCache && resp.StatusCode == 200 {
		c.cacheResponse(url, resp)
	}

	c.recordRequest(latency)
	return resp, latency, false, nil
}

// getFromCache retrieves a response from cache
func (c *SimpleOptimizedClient) getFromCache(url string) *CachedResponse {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	cached, exists := c.cache[url]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(cached.Timestamp) > cached.TTL {
		// Cache expired, remove it
		go func() {
			c.cacheMutex.Lock()
			delete(c.cache, url)
			c.cacheMutex.Unlock()
		}()
		return nil
	}

	return cached
}

// cacheResponse stores a response in cache
func (c *SimpleOptimizedClient) cacheResponse(url string, resp *http.Response) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Replace response body so it can still be read by caller
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// Store in cache
	cached := &CachedResponse{
		Body:       body,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Timestamp:  time.Now(),
		TTL:        5 * time.Minute, // 5 minute TTL
	}

	c.cacheMutex.Lock()
	c.cache[url] = cached
	c.cacheMutex.Unlock()
}

// recordRequest records request statistics
func (c *SimpleOptimizedClient) recordRequest(latency time.Duration) {
	c.statsMutex.Lock()
	c.stats.TotalRequests++
	c.stats.TotalLatency += latency
	c.statsMutex.Unlock()
}

// recordCacheHit records a cache hit
func (c *SimpleOptimizedClient) recordCacheHit() {
	c.statsMutex.Lock()
	c.stats.CacheHits++
	c.statsMutex.Unlock()
}

// recordCacheMiss records a cache miss
func (c *SimpleOptimizedClient) recordCacheMiss() {
	c.statsMutex.Lock()
	c.stats.CacheMisses++
	c.statsMutex.Unlock()
}

// GetStats returns current performance statistics
func (c *SimpleOptimizedClient) GetStats() SimpleStats {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()
	return *c.stats
}

// SimpleBenchmark runs a benchmark comparison
func SimpleBenchmark(url string, requests int, concurrency int) {
	fmt.Printf(`
🚀 Phase 1 API Latency Optimization Demo
=========================================
Target URL: %s
Requests: %d
Concurrency: %d

`, url, requests, concurrency)

	// Run baseline benchmark
	fmt.Println("📊 Running baseline benchmark (no optimizations)...")
	baselineStats := runBenchmark(url, requests, concurrency, false, false)

	// Run optimized benchmark
	fmt.Println("⚡ Running optimized benchmark (HTTP/2 + caching)...")
	optimizedStats := runBenchmark(url, requests, concurrency, true, true)

	// Calculate improvements
	latencyImprovement := 1.0 - (float64(optimizedStats.AverageLatency) / float64(baselineStats.AverageLatency))
	throughputImprovement := (optimizedStats.RequestsPerSecond / baselineStats.RequestsPerSecond) - 1.0

	// Print results
	fmt.Printf(`
📈 Results Summary
==================

Baseline Performance:
  Average Latency:  %v
  Throughput:       %.2f req/s

Optimized Performance:
  Average Latency:  %v
  Throughput:       %.2f req/s
  Cache Hit Ratio:  %.1f%%

Improvements:
  Latency:          %.1f%% reduction
  Throughput:       %.1f%% increase

Phase 1 Status:
%s

`,
		baselineStats.AverageLatency,
		baselineStats.RequestsPerSecond,
		optimizedStats.AverageLatency,
		optimizedStats.RequestsPerSecond,
		optimizedStats.CacheHitRatio,
		latencyImprovement*100,
		throughputImprovement*100,
		getPhase1Status(latencyImprovement, optimizedStats.CacheHitRatio),
	)
}

// BenchmarkStats holds benchmark results
type BenchmarkStats struct {
	AverageLatency     time.Duration
	RequestsPerSecond  float64
	CacheHitRatio      float64
	TotalRequests      int64
}

// runBenchmark executes a benchmark
func runBenchmark(url string, requests int, concurrency int, useHTTP2 bool, useCache bool) BenchmarkStats {
	var client *SimpleOptimizedClient
	var standardClient *http.Client

	if useHTTP2 {
		client = NewSimpleOptimizedClient()
	} else {
		// Standard HTTP/1.1 client
		standardClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				ForceAttemptHTTP2: false, // Disable HTTP/2
			},
		}
	}

	startTime := time.Now()
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var totalLatency time.Duration
	var latencyMutex sync.Mutex
	var successCount int64
	var successMutex sync.Mutex

	// Execute requests
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			reqStart := time.Now()
			var err error

			if useHTTP2 {
				_, _, _, err = client.Do(url, useCache)
			} else {
				req, reqErr := http.NewRequest("GET", url, nil)
				if reqErr != nil {
					return
				}
				resp, respErr := standardClient.Do(req)
				if respErr != nil {
					err = respErr
				} else {
					resp.Body.Close()
				}
			}

			reqLatency := time.Since(reqStart)

			if err == nil {
				latencyMutex.Lock()
				totalLatency += reqLatency
				latencyMutex.Unlock()

				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Calculate statistics
	avgLatency := time.Duration(0)
	if successCount > 0 {
		avgLatency = totalLatency / time.Duration(successCount)
	}

	requestsPerSecond := float64(successCount) / duration.Seconds()

	var cacheHitRatio float64
	if useHTTP2 && client != nil {
		stats := client.GetStats()
		if stats.CacheHits+stats.CacheMisses > 0 {
			cacheHitRatio = float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		}
	}

	return BenchmarkStats{
		AverageLatency:    avgLatency,
		RequestsPerSecond: requestsPerSecond,
		CacheHitRatio:     cacheHitRatio,
		TotalRequests:     successCount,
	}
}

// getPhase1Status determines Phase 1 completion status
func getPhase1Status(latencyImprovement float64, cacheHitRatio float64) string {
	score := 0
	total := 3

	// Check targets
	if latencyImprovement >= 0.1 { // 10% improvement
		score++
	}
	if cacheHitRatio >= 60 { // 60% cache hit ratio
		score++
	}
	if latencyImprovement >= 0.2 { // 20% improvement (stretch goal)
		score++
	}

	percentage := float64(score) / float64(total) * 100

	switch {
	case percentage >= 80:
		return "🎉 PHASE 1 COMPLETE - Exceeds targets! Ready for Phase 2."
	case percentage >= 60:
		return "✅ PHASE 1 SUCCESS - Targets achieved. Minor optimizations possible."
	case percentage >= 40:
		return "⚠️ PHASE 1 PARTIAL - Some targets met. Additional work needed."
	default:
		return "❌ PHASE 1 INCOMPLETE - Targets not met. Requires investigation."
	}
}

func main() {
	// Run demo with Anthropic API
	SimpleBenchmark("https://api.anthropic.com", 20, 5)
}