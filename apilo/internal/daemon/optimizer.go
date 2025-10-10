package daemon

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Optimizer handles request optimization logic
type Optimizer struct {
	config     *DaemonConfig
	cache      *Cache
	httpClient *http.Client
	logger     *Logger
	mu         sync.RWMutex
}

// NewOptimizer creates a new optimizer
func NewOptimizer(config *DaemonConfig, logger *Logger) (*Optimizer, error) {
	opt := &Optimizer{
		config: config,
		cache:  NewCache(config.CacheMaxMemoryMB, config.CacheDefaultTTL, logger),
		logger: logger,
	}

	// Configure HTTP client with HTTP/2 support
	transport := &http.Transport{
		MaxIdleConns:        config.MaxConnections,
		MaxIdleConnsPerHost: config.MaxConnections,
		IdleConnTimeout:     config.IdleTimeout,
		DisableKeepAlives:   false,
	}

	if config.EnableHTTP2 {
		// Enable HTTP/2
		transport.ForceAttemptHTTP2 = true
	}

	opt.httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return opt, nil
}

// Optimize optimizes an API request
func (opt *Optimizer) Optimize(req *OptimizationRequest) (*OptimizationResponse, error) {
	// Generate cache key
	cacheKey := opt.generateCacheKey(req)

	// Check cache
	if cached, found := opt.cache.Get(cacheKey); found {
		opt.logger.LogCacheOperation("GET", cacheKey, true)
		return &OptimizationResponse{
			StatusCode: cached.StatusCode,
			Headers:    cached.Headers,
			Body:       cached.Body,
			CacheHit:   true,
			Optimized:  true,
			Metadata: ResponseMetadata{
				CacheStatus:      "hit",
				OptimizationType: "cached",
				ConnectionReused: true,
				HTTP2Used:        opt.config.EnableHTTP2,
				TokenUsage:       cached.TokenUsage,
			},
		}, nil
	}
	opt.logger.LogCacheOperation("GET", cacheKey, false)

	// Make HTTP request
	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	httpResp, err := opt.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range httpResp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Estimate token usage for this request/response
	tokenUsage := estimateTokens(req.Body, body)

	// Cache the response with token data
	opt.cache.Set(cacheKey, &CacheEntry{
		StatusCode: httpResp.StatusCode,
		Headers:    headers,
		Body:       body,
		CachedAt:   time.Now(),
		TokenUsage: tokenUsage,
	})
	opt.logger.LogCacheOperation("SET", cacheKey, true)

	return &OptimizationResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    headers,
		Body:       body,
		CacheHit:   false,
		Optimized:  true,
		Metadata: ResponseMetadata{
			CacheStatus:      "miss",
			OptimizationType: "http2+cache",
			ConnectionReused: httpResp.Request.Response != nil,
			HTTP2Used:        httpResp.ProtoMajor == 2,
			LatencySaved:     0,
			TokenUsage:       tokenUsage,
		},
	}, nil
}

// generateCacheKey generates a unique cache key for a request
func (opt *Optimizer) generateCacheKey(req *OptimizationRequest) string {
	hasher := sha256.New()
	hasher.Write([]byte(req.Method))
	hasher.Write([]byte(req.URL))

	// Include relevant headers
	for key, value := range req.Headers {
		hasher.Write([]byte(key))
		hasher.Write([]byte(value))
	}

	if len(req.Body) > 0 {
		hasher.Write(req.Body)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// InvalidateCache clears the entire cache
func (opt *Optimizer) InvalidateCache() {
	opt.cache.Clear()
}

// estimateTokens estimates token usage based on content size
// Uses approximation: ~4 characters per token for English text
func estimateTokens(requestBody, responseBody []byte) *TokenUsage {
	// Estimate input tokens from request body
	inputChars := len(requestBody)
	inputTokens := int64(inputChars / 4)
	if inputTokens == 0 && inputChars > 0 {
		inputTokens = 1 // At least 1 token for non-empty content
	}

	// Estimate output tokens from response body
	outputChars := len(responseBody)
	outputTokens := int64(outputChars / 4)
	if outputTokens == 0 && outputChars > 0 {
		outputTokens = 1 // At least 1 token for non-empty content
	}

	return &TokenUsage{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		IsEstimated:  true,
	}
}

// Cache implements a simple thread-safe cache
type Cache struct {
	data          map[string]*CacheEntry
	maxMemory     int64
	currentMemory int64
	defaultTTL    time.Duration
	logger        *Logger
	mu            sync.RWMutex
}

// CacheEntry represents a cached response
type CacheEntry struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	CachedAt   time.Time
	TokenUsage *TokenUsage
}

// NewCache creates a new cache
func NewCache(maxMemoryMB int64, defaultTTL time.Duration, logger *Logger) *Cache {
	return &Cache{
		data:          make(map[string]*CacheEntry),
		maxMemory:     maxMemoryMB * 1024 * 1024,
		currentMemory: 0,
		defaultTTL:    defaultTTL,
		logger:        logger,
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.data[key]
	if !found {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.CachedAt) > c.defaultTTL {
		return nil, false
	}

	return entry, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entrySize := int64(len(entry.Body))

	// Check if we need to evict entries
	if c.currentMemory+entrySize > c.maxMemory {
		c.evictLRU(entrySize)
	}

	c.data[key] = entry
	c.currentMemory += entrySize
}

// evictLRU evicts least recently used entries to make room
func (c *Cache) evictLRU(neededSpace int64) {
	// Find oldest entries
	type entryAge struct {
		key string
		age time.Time
	}

	var entries []entryAge
	for k, v := range c.data {
		entries = append(entries, entryAge{key: k, age: v.CachedAt})
	}

	// Sort by age (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].age.After(entries[j].age) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict until we have enough space
	freedSpace := int64(0)
	for _, e := range entries {
		if freedSpace >= neededSpace {
			break
		}

		if entry, exists := c.data[e.key]; exists {
			entrySize := int64(len(entry.Body))
			delete(c.data, e.key)
			c.currentMemory -= entrySize
			freedSpace += entrySize
			c.logger.Debug("Cache eviction - Key: %s, Size: %d bytes", e.key[:min(16, len(e.key))], entrySize)
		}
	}
}

// Clear clears all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
	c.currentMemory = 0
	c.logger.Info("Cache cleared")
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.data)
}

// GetStats returns detailed cache statistics
func (c *Cache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]CacheEntryInfo, 0, len(c.data))

	for key, entry := range c.data {
		age := time.Since(entry.CachedAt)
		ttlRemaining := c.defaultTTL - age

		entries = append(entries, CacheEntryInfo{
			Key:          key[:min(16, len(key))], // Truncate for display
			SizeBytes:    int64(len(entry.Body)),
			Age:          age,
			TTLRemaining: ttlRemaining,
			Expired:      age > c.defaultTTL,
		})
	}

	return &CacheStats{
		Entries:       len(c.data),
		MemoryUsedMB:  float64(c.currentMemory) / 1024 / 1024,
		MemoryLimitMB: float64(c.maxMemory) / 1024 / 1024,
		MemoryPercent: float64(c.currentMemory) / float64(c.maxMemory) * 100,
		DefaultTTL:    c.defaultTTL,
		EntryDetails:  entries,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	Entries       int              `json:"entries"`
	MemoryUsedMB  float64          `json:"memory_used_mb"`
	MemoryLimitMB float64          `json:"memory_limit_mb"`
	MemoryPercent float64          `json:"memory_percent"`
	DefaultTTL    time.Duration    `json:"default_ttl"`
	EntryDetails  []CacheEntryInfo `json:"entry_details"`
}

// CacheEntryInfo holds information about a cache entry
type CacheEntryInfo struct {
	Key          string        `json:"key"`
	SizeBytes    int64         `json:"size_bytes"`
	Age          time.Duration `json:"age"`
	TTLRemaining time.Duration `json:"ttl_remaining"`
	Expired      bool          `json:"expired"`
}
