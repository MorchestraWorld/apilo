// Package src provides an optimized HTTP client that combines HTTP/2, caching, and monitoring
// for maximum API latency reduction. This unified client integrates all Phase 1 optimizations.
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// OptimizedClient combines HTTP/2 client, caching, and monitoring for maximum performance
type OptimizedClient struct {
	// Core components
	http2Client    *HTTP2Client
	cache          *Cache
	monitor        *Monitor
	metricsCollector *MetricsCollector

	// Configuration
	config         *OptimizedClientConfig

	// State management
	mu             sync.RWMutex
	initialized    bool
	warmedUp       bool
	requestCount   int64

	// Performance tracking
	totalLatency   time.Duration
	cacheHits      int64
	cacheMisses    int64
	connectionReuse int64
	errors         int64
}

// OptimizedClientConfig holds configuration for the unified client
type OptimizedClientConfig struct {
	// HTTP/2 Configuration
	HTTP2Config struct {
		MaxConnectionsPerHost int           `yaml:"max_connections_per_host"`
		IdleConnTimeout      time.Duration `yaml:"idle_conn_timeout"`
		TLSHandshakeTimeout  time.Duration `yaml:"tls_handshake_timeout"`
		DisableCompression   bool          `yaml:"disable_compression"`
		EnablePush           bool          `yaml:"enable_push"`
	} `yaml:"http2"`

	// Cache Configuration
	CacheConfig struct {
		Enabled      bool          `yaml:"enabled"`
		Capacity     int           `yaml:"capacity"`
		DefaultTTL   time.Duration `yaml:"default_ttl"`
		PolicyType   string        `yaml:"policy_type"`
		WarmupEnabled bool         `yaml:"warmup_enabled"`
	} `yaml:"cache"`

	// Monitoring Configuration
	MonitoringConfig struct {
		Enabled      bool `yaml:"enabled"`
		DashboardPort int `yaml:"dashboard_port"`
		AlertsEnabled bool `yaml:"alerts_enabled"`
		PrometheusEnabled bool `yaml:"prometheus_enabled"`
	} `yaml:"monitoring"`

	// Integration Configuration
	MaxRetries       int           `yaml:"max_retries"`
	RetryBackoff     time.Duration `yaml:"retry_backoff"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	EnableMetrics    bool          `yaml:"enable_metrics"`
}

// DefaultOptimizedClientConfig returns a configuration optimized for API latency reduction
func DefaultOptimizedClientConfig() *OptimizedClientConfig {
	return &OptimizedClientConfig{
		HTTP2Config: struct {
			MaxConnectionsPerHost int           `yaml:"max_connections_per_host"`
			IdleConnTimeout      time.Duration `yaml:"idle_conn_timeout"`
			TLSHandshakeTimeout  time.Duration `yaml:"tls_handshake_timeout"`
			DisableCompression   bool          `yaml:"disable_compression"`
			EnablePush           bool          `yaml:"enable_push"`
		}{
			MaxConnectionsPerHost: 10,
			IdleConnTimeout:      90 * time.Second,
			TLSHandshakeTimeout:  10 * time.Second,
			DisableCompression:   false,
			EnablePush:           true,
		},
		CacheConfig: struct {
			Enabled      bool          `yaml:"enabled"`
			Capacity     int           `yaml:"capacity"`
			DefaultTTL   time.Duration `yaml:"default_ttl"`
			PolicyType   string        `yaml:"policy_type"`
			WarmupEnabled bool         `yaml:"warmup_enabled"`
		}{
			Enabled:      true,
			Capacity:     10000,
			DefaultTTL:   5 * time.Minute,
			PolicyType:   "adaptive",
			WarmupEnabled: true,
		},
		MonitoringConfig: struct {
			Enabled      bool `yaml:"enabled"`
			DashboardPort int `yaml:"dashboard_port"`
			AlertsEnabled bool `yaml:"alerts_enabled"`
			PrometheusEnabled bool `yaml:"prometheus_enabled"`
		}{
			Enabled:      true,
			DashboardPort: 8080,
			AlertsEnabled: false,
			PrometheusEnabled: false,
		},
		MaxRetries:       3,
		RetryBackoff:     100 * time.Millisecond,
		RequestTimeout:   30 * time.Second,
		EnableMetrics:    true,
	}
}

// NewOptimizedClient creates a new unified client with all optimizations enabled
func NewOptimizedClient(config *OptimizedClientConfig) (*OptimizedClient, error) {
	if config == nil {
		config = DefaultOptimizedClientConfig()
	}

	client := &OptimizedClient{
		config: config,
	}

	// Initialize HTTP/2 client
	http2Config := &HTTP2ClientConfig{
		MaxConnectionsPerHost: config.HTTP2Config.MaxConnectionsPerHost,
		IdleConnTimeout:      config.HTTP2Config.IdleConnTimeout,
		TLSHandshakeTimeout:  config.HTTP2Config.TLSHandshakeTimeout,
		DisableCompression:   config.HTTP2Config.DisableCompression,
		EnableHTTP2Push:      config.HTTP2Config.EnablePush,
	}

	var err error
	client.http2Client, err = NewHTTP2Client(http2Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP/2 client: %w", err)
	}

	// Initialize cache if enabled
	if config.CacheConfig.Enabled {
		cacheConfig := &CacheConfig{
			Capacity:   config.CacheConfig.Capacity,
			DefaultTTL: config.CacheConfig.DefaultTTL,
			Policy:     config.CacheConfig.PolicyType,
		}

		client.cache = NewCache(cacheConfig)

		// Initialize cache warmup if enabled
		if config.CacheConfig.WarmupEnabled {
			warmupConfig := &WarmupConfig{
				Strategy: "adaptive",
			}
			client.cache.InitializeWarmup(warmupConfig)
		}
	}

	// Initialize monitoring if enabled
	if config.MonitoringConfig.Enabled {
		monitorConfig := &MonitoringConfig{
			DashboardEnabled:    true,
			DashboardPort:      config.MonitoringConfig.DashboardPort,
			AlertsEnabled:      config.MonitoringConfig.AlertsEnabled,
			PrometheusEnabled:  config.MonitoringConfig.PrometheusEnabled,
		}

		client.monitor, err = NewMonitor(monitorConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create monitor: %w", err)
		}

		client.metricsCollector = NewMetricsCollector(1000) // Default max snapshots
	}

	client.initialized = true
	return client, nil
}

// OptimizedRequest represents a request with optimization context
type OptimizedRequest struct {
	*http.Request
	CacheKey     string
	CacheTTL     time.Duration
	UseCache     bool
	EnableMetrics bool
	Metadata     map[string]interface{}
}

// OptimizedResponse represents a response with optimization metadata
type OptimizedResponse struct {
	*http.Response
	CacheHit      bool
	CacheAge      time.Duration
	ConnectionReused bool
	TotalLatency  time.Duration
	DNSLatency    time.Duration
	ConnectLatency time.Duration
	TLSLatency    time.Duration
	TTFBLatency   time.Duration
	ProcessingLatency time.Duration
	Metadata      map[string]interface{}
}

// Do executes an HTTP request using all available optimizations
func (c *OptimizedClient) Do(req *OptimizedRequest) (*OptimizedResponse, error) {
	start := time.Now()

	// Increment request counter
	c.mu.Lock()
	c.requestCount++
	reqID := c.requestCount
	c.mu.Unlock()

	// Create optimization context
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Add request timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	req.Request = req.Request.WithContext(ctx)

	// Initialize response
	response := &OptimizedResponse{
		Metadata: make(map[string]interface{}),
	}
	response.Metadata["request_id"] = reqID
	response.Metadata["start_time"] = start

	// Try cache first if enabled
	if c.cache != nil && req.UseCache {
		if cached := c.tryCache(req, response); cached != nil {
			response.CacheHit = true
			response.TotalLatency = time.Since(start)

			c.mu.Lock()
			c.cacheHits++
			c.totalLatency += response.TotalLatency
			c.mu.Unlock()

			// Record metrics
			if c.metricsCollector != nil && req.EnableMetrics {
				c.recordCacheHit(req, response)
			}

			return cached, nil
		}

		c.mu.Lock()
		c.cacheMisses++
		c.mu.Unlock()
	}

	// Execute HTTP/2 request with detailed timing
	httpResponse, timing, err := c.executeHTTP2Request(req)
	if err != nil {
		c.mu.Lock()
		c.errors++
		c.mu.Unlock()

		// Retry logic
		if c.shouldRetry(err, 0) {
			return c.retryRequest(req, response, 1)
		}

		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Populate response timing
	response.Response = httpResponse
	response.DNSLatency = timing.DNSLatency
	response.ConnectLatency = timing.ConnectLatency
	response.TLSLatency = timing.TLSLatency
	response.TTFBLatency = timing.TTFBLatency
	response.ProcessingLatency = timing.ProcessingLatency
	response.TotalLatency = time.Since(start)
	response.ConnectionReused = timing.ConnectionReused

	if timing.ConnectionReused {
		c.mu.Lock()
		c.connectionReuse++
		c.mu.Unlock()
	}

	// Cache response if enabled and cacheable
	if c.cache != nil && req.UseCache && c.isCacheable(httpResponse) {
		c.cacheResponse(req, httpResponse)
	}

	// Record metrics
	if c.metricsCollector != nil && req.EnableMetrics {
		c.recordRequest(req, response)
	}

	// Update running statistics
	c.mu.Lock()
	c.totalLatency += response.TotalLatency
	c.mu.Unlock()

	return response, nil
}

// HTTP2RequestTiming contains detailed timing information for HTTP/2 requests
// MOVED TO types.go
// type HTTP2RequestTiming struct {
// 	DNSLatency        time.Duration
// 	ConnectLatency    time.Duration
// 	TLSLatency        time.Duration
// 	TTFBLatency       time.Duration
// 	ProcessingLatency time.Duration
// 	ConnectionReused  bool
// }

// executeHTTP2Request performs the actual HTTP/2 request with detailed timing
func (c *OptimizedClient) executeHTTP2Request(req *OptimizedRequest) (*http.Response, *HTTP2RequestTiming, error) {
	timing := &HTTP2RequestTiming{}

	// Use the HTTP/2 client's Do method which provides detailed timing
	response, err := c.http2Client.Do(req.Request)
	if err != nil {
		return nil, timing, err
	}

	// Get timing information from HTTP/2 client
	if clientTiming := c.http2Client.GetLastRequestTiming(); clientTiming != nil {
		timing.DNSLatency = clientTiming.DNSLatency
		timing.ConnectLatency = clientTiming.ConnectLatency
		timing.TLSLatency = clientTiming.TLSLatency
		timing.TTFBLatency = clientTiming.TTFBLatency
		timing.ProcessingLatency = clientTiming.ProcessingLatency
		timing.ConnectionReused = clientTiming.ConnectionReused
	}

	return response, timing, nil
}

// tryCache attempts to retrieve a cached response
func (c *OptimizedClient) tryCache(req *OptimizedRequest, response *OptimizedResponse) *OptimizedResponse {
	key := req.CacheKey
	if key == "" {
		key = generateCacheKey(req.Request)
	}

	cached, age, found := c.cache.GetWithAge(key)
	if !found {
		return nil
	}

	// Parse cached response
	cachedResp, ok := cached.(*http.Response)
	if !ok {
		// Cache corruption, remove entry
		c.cache.Delete(key)
		return nil
	}

	// Create optimized response from cache
	optimized := &OptimizedResponse{
		Response:    cachedResp,
		CacheHit:    true,
		CacheAge:    age,
		Metadata:    response.Metadata,
	}

	return optimized
}

// cacheResponse stores a response in the cache
func (c *OptimizedClient) cacheResponse(req *OptimizedRequest, resp *http.Response) {
	key := req.CacheKey
	if key == "" {
		key = generateCacheKey(req.Request)
	}

	ttl := req.CacheTTL
	if ttl == 0 {
		ttl = c.config.CacheConfig.DefaultTTL
	}

	// Clone response for caching (to avoid body consumption issues)
	clonedResp := c.cloneResponse(resp)
	c.cache.SetWithTTL(key, clonedResp, ttl)
}

// cloneResponse creates a copy of an HTTP response for caching
func (c *OptimizedClient) cloneResponse(resp *http.Response) *http.Response {
	cloned := &http.Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Header:           resp.Header.Clone(),
		ContentLength:    resp.ContentLength,
		TransferEncoding: resp.TransferEncoding,
		Close:            resp.Close,
		Uncompressed:     resp.Uncompressed,
		Trailer:          resp.Trailer,
		Request:          resp.Request,
		TLS:              resp.TLS,
	}

	// Clone body if present
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			cloned.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	return cloned
}

// isCacheable determines if a response should be cached
func (c *OptimizedClient) isCacheable(resp *http.Response) bool {
	// Only cache successful responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	// Check cache-control headers
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "no-cache" || cacheControl == "no-store" {
		return false
	}

	// Don't cache very large responses (>1MB)
	if resp.ContentLength > 1024*1024 {
		return false
	}

	return true
}

// shouldRetry determines if a request should be retried
func (c *OptimizedClient) shouldRetry(err error, attempt int) bool {
	if attempt >= c.config.MaxRetries {
		return false
	}

	// Retry on timeout and connection errors
	switch err.(type) {
	case *tls.CertificateVerificationError:
		return false // Don't retry certificate errors
	default:
		return true // Retry other errors
	}
}

// retryRequest performs request retry with exponential backoff
func (c *OptimizedClient) retryRequest(req *OptimizedRequest, response *OptimizedResponse, attempt int) (*OptimizedResponse, error) {
	// Exponential backoff
	backoff := time.Duration(attempt) * c.config.RetryBackoff
	time.Sleep(backoff)

	// Retry the request
	return c.Do(req)
}

// recordRequest records metrics for a completed request
func (c *OptimizedClient) recordRequest(req *OptimizedRequest, resp *OptimizedResponse) {
	if c.metricsCollector == nil {
		return
	}

	// Record latency metrics
	c.metricsCollector.RecordLatency("total", resp.TotalLatency)
	c.metricsCollector.RecordLatency("dns", resp.DNSLatency)
	c.metricsCollector.RecordLatency("connect", resp.ConnectLatency)
	c.metricsCollector.RecordLatency("tls", resp.TLSLatency)
	c.metricsCollector.RecordLatency("ttfb", resp.TTFBLatency)

	// Record cache metrics
	if resp.CacheHit {
		c.metricsCollector.RecordCacheHit()
	} else {
		c.metricsCollector.RecordCacheMiss()
	}

	// Record connection reuse
	if resp.ConnectionReused {
		c.metricsCollector.RecordConnectionReuse()
	}

	// Record response size
	if resp.Response != nil {
		c.metricsCollector.RecordResponseSize(resp.Response.ContentLength)
	}
}

// recordCacheHit records metrics for a cache hit
func (c *OptimizedClient) recordCacheHit(req *OptimizedRequest, resp *OptimizedResponse) {
	if c.metricsCollector == nil {
		return
	}

	c.metricsCollector.RecordCacheHit()
	c.metricsCollector.RecordLatency("total", resp.TotalLatency)
}

// generateCacheKey creates a cache key from an HTTP request
func generateCacheKey(req *http.Request) string {
	return fmt.Sprintf("%s:%s", req.Method, req.URL.String())
}

// WarmupCache performs cache warming with common URLs
func (c *OptimizedClient) WarmupCache(urls []string) error {
	if c.cache == nil {
		return fmt.Errorf("caching not enabled")
	}

	c.mu.Lock()
	if c.warmedUp {
		c.mu.Unlock()
		return nil
	}
	c.warmedUp = true
	c.mu.Unlock()

	// Use cache warmup system if available
	if warmup := c.cache.GetWarmup(); warmup != nil {
		return warmup.WarmupURLs(urls)
	}

	// Fallback: simple URL warming
	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		optimizedReq := &OptimizedRequest{
			Request:  req,
			UseCache: true,
			EnableMetrics: false, // Don't count warmup in metrics
		}

		// Fire and forget warmup requests
		go func() {
			c.Do(optimizedReq)
		}()
	}

	return nil
}

// GetStats returns current client performance statistics
func (c *OptimizedClient) GetStats() *OptimizedClientStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &OptimizedClientStats{
		TotalRequests:    c.requestCount,
		CacheHits:        c.cacheHits,
		CacheMisses:      c.cacheMisses,
		ConnectionReuse:  c.connectionReuse,
		Errors:           c.errors,
		Initialized:      c.initialized,
		WarmedUp:         c.warmedUp,
	}

	if c.requestCount > 0 {
		stats.AverageLatency = c.totalLatency / time.Duration(c.requestCount)
		stats.CacheHitRatio = float64(c.cacheHits) / float64(c.cacheHits+c.cacheMisses)
		stats.ConnectionReuseRatio = float64(c.connectionReuse) / float64(c.requestCount)
		stats.ErrorRate = float64(c.errors) / float64(c.requestCount)
	}

	return stats
}

// OptimizedClientStats contains performance statistics
type OptimizedClientStats struct {
	TotalRequests        int64         `json:"total_requests"`
	CacheHits            int64         `json:"cache_hits"`
	CacheMisses          int64         `json:"cache_misses"`
	ConnectionReuse      int64         `json:"connection_reuse"`
	Errors               int64         `json:"errors"`
	AverageLatency       time.Duration `json:"average_latency"`
	CacheHitRatio        float64       `json:"cache_hit_ratio"`
	ConnectionReuseRatio float64       `json:"connection_reuse_ratio"`
	ErrorRate            float64       `json:"error_rate"`
	Initialized          bool          `json:"initialized"`
	WarmedUp             bool          `json:"warmed_up"`
}

// Stop gracefully shuts down the optimized client
func (c *OptimizedClient) Stop() error {
	var errs []error

	// Stop monitoring
	if c.monitor != nil {
		if err := c.monitor.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("monitor stop error: %w", err))
		}
	}

	// Stop HTTP/2 client
	if c.http2Client != nil {
		if err := c.http2Client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("HTTP/2 client close error: %w", err))
		}
	}

	// Stop cache cleanup
	if c.cache != nil {
		c.cache.Stop()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}

// IsInitialized returns true if the client is fully initialized
func (c *OptimizedClient) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// IsWarmedUp returns true if cache warmup has been performed
func (c *OptimizedClient) IsWarmedUp() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.warmedUp
}