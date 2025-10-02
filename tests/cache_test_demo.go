// Cache Test Demo - Tests functional cache implementation
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

// SimpleOptimizedClient demonstrates functional caching
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

// Do executes an HTTP request with caching
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

// TestCacheEffectiveness demonstrates cache functionality
func TestCacheEffectiveness(url string) {
	fmt.Printf(`
🧪 Cache Effectiveness Test
===========================
Target URL: %s
Test: 10 repeated requests to same URL

`, url)

	client := NewSimpleOptimizedClient()

	var latencies []time.Duration

	for i := 0; i < 10; i++ {
		fmt.Printf("Request %d: ", i+1)

		resp, latency, cacheHit, err := client.Do(url, true)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		latencies = append(latencies, latency)

		if cacheHit {
			fmt.Printf("✅ CACHE HIT (%v)\n", latency)
		} else {
			fmt.Printf("○ Cache miss (%v)\n", latency)
		}

		// Small delay to show timing
		time.Sleep(100 * time.Millisecond)
	}

	// Print results
	stats := client.GetStats()
	hitRatio := float64(stats.CacheHits) / float64(stats.CacheHits + stats.CacheMisses) * 100

	fmt.Printf(`
📊 Cache Test Results
=====================
Total Requests:    %d
Cache Hits:        %d
Cache Misses:      %d
Cache Hit Ratio:   %.1f%%

Performance Impact:
- First request:   %v (cache miss)
- Cached requests: %v average
- Speedup:         %.1fx faster

`,
		stats.TotalRequests,
		stats.CacheHits,
		stats.CacheMisses,
		hitRatio,
		latencies[0],
		calculateAverageCachedLatency(latencies[1:]),
		float64(latencies[0])/float64(calculateAverageCachedLatency(latencies[1:])),
	)

	if hitRatio >= 80 {
		fmt.Println("✅ CACHE WORKING PERFECTLY - 80%+ hit ratio achieved!")
	} else if hitRatio >= 60 {
		fmt.Println("✅ CACHE WORKING WELL - 60%+ hit ratio achieved!")
	} else if hitRatio > 0 {
		fmt.Println("⚠️ CACHE WORKING PARTIALLY - Some hits but could be better")
	} else {
		fmt.Println("❌ CACHE NOT WORKING - 0% hit ratio")
	}
}

func calculateAverageCachedLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total / time.Duration(len(latencies))
}

func main() {
	// Test cache with a simple endpoint
	TestCacheEffectiveness("https://httpbin.org/get")
}