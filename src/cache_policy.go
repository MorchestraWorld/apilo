package main

import (
	"time"
)

// CachePolicy defines the interface for cache policies
type CachePolicy interface {
	// ComputeTTL calculates the TTL for a cache entry based on various factors
	ComputeTTL(entry *CacheEntry, pattern *AccessPattern) time.Duration

	// ShouldCache determines if a response should be cached
	ShouldCache(statusCode int, size int64, headers map[string]string) bool

	// CanEvict determines if an entry can be evicted based on policy rules
	CanEvict(entry *CacheEntry) bool

	// Name returns the policy name
	Name() string
}

// AccessPattern represents observed access patterns for a resource
type AccessPattern struct {
	Key               string        // Resource identifier
	AccessCount       int64         // Total number of accesses
	LastAccess        time.Time     // Last access timestamp
	FirstAccess       time.Time     // First access timestamp
	AverageInterval   time.Duration // Average time between accesses
	RecentAccesses    []time.Time   // Recent access timestamps (sliding window)
	PredictedNextUse  time.Time     // Predicted next access time
	Volatility        float64       // Measure of access pattern stability (0-1)
}

// DefaultPolicy implements a sensible default caching policy
type DefaultPolicy struct {
	BaseTTL        time.Duration // Default TTL
	MinTTL         time.Duration // Minimum allowed TTL
	MaxTTL         time.Duration // Maximum allowed TTL
	MaxCacheSize   int64         // Maximum cacheable response size
	CacheableStatusCodes map[int]bool // HTTP status codes to cache
}

// NewDefaultPolicy creates a new default policy
func NewDefaultPolicy() *DefaultPolicy {
	return &DefaultPolicy{
		BaseTTL:      5 * time.Minute,
		MinTTL:       30 * time.Second,
		MaxTTL:       30 * time.Minute,
		MaxCacheSize: 10 * 1024 * 1024, // 10MB
		CacheableStatusCodes: map[int]bool{
			200: true, // OK
			203: true, // Non-Authoritative Information
			204: true, // No Content
			206: true, // Partial Content
			300: true, // Multiple Choices
			301: true, // Moved Permanently
			404: true, // Not Found (cache negative results)
			410: true, // Gone
		},
	}
}

func (p *DefaultPolicy) Name() string {
	return "default"
}

func (p *DefaultPolicy) ComputeTTL(entry *CacheEntry, pattern *AccessPattern) time.Duration {
	ttl := p.BaseTTL

	// If we have access pattern data, adjust TTL based on frequency
	if pattern != nil && pattern.AccessCount > 0 {
		// High frequency access = longer TTL
		if pattern.AccessCount > 100 {
			ttl = p.MaxTTL
		} else if pattern.AccessCount > 50 {
			ttl = 15 * time.Minute
		} else if pattern.AccessCount > 10 {
			ttl = 10 * time.Minute
		}

		// Adjust based on access pattern stability
		if pattern.Volatility < 0.3 { // Stable pattern
			ttl = time.Duration(float64(ttl) * 1.5)
		} else if pattern.Volatility > 0.7 { // Unstable pattern
			ttl = time.Duration(float64(ttl) * 0.7)
		}
	}

	// Clamp to min/max bounds
	if ttl < p.MinTTL {
		ttl = p.MinTTL
	}
	if ttl > p.MaxTTL {
		ttl = p.MaxTTL
	}

	return ttl
}

func (p *DefaultPolicy) ShouldCache(statusCode int, size int64, headers map[string]string) bool {
	// Check status code
	if !p.CacheableStatusCodes[statusCode] {
		return false
	}

	// Check size limit
	if size > p.MaxCacheSize {
		return false
	}

	// Check Cache-Control headers if present
	if cacheControl, exists := headers["Cache-Control"]; exists {
		if cacheControl == "no-store" || cacheControl == "no-cache" {
			return false
		}
	}

	return true
}

func (p *DefaultPolicy) CanEvict(entry *CacheEntry) bool {
	// Don't evict recently created entries
	if time.Since(entry.CreatedAt) < 1*time.Minute {
		return false
	}

	// Don't evict frequently accessed entries
	if entry.AccessCount > 100 {
		return false
	}

	return true
}

// AdaptivePolicy implements an intelligent adaptive caching policy
type AdaptivePolicy struct {
	*DefaultPolicy
	patterns map[string]*AccessPattern // Access patterns per resource
}

// NewAdaptivePolicy creates a new adaptive policy
func NewAdaptivePolicy() *AdaptivePolicy {
	return &AdaptivePolicy{
		DefaultPolicy: NewDefaultPolicy(),
		patterns:      make(map[string]*AccessPattern),
	}
}

func (p *AdaptivePolicy) Name() string {
	return "adaptive"
}

func (p *AdaptivePolicy) ComputeTTL(entry *CacheEntry, pattern *AccessPattern) time.Duration {
	if pattern == nil {
		pattern = p.getPattern(entry.Key)
	}

	// Base computation from default policy
	ttl := p.DefaultPolicy.ComputeTTL(entry, pattern)

	// Adaptive adjustments based on observed patterns
	if pattern != nil {
		// Calculate access frequency (accesses per hour)
		age := time.Since(pattern.FirstAccess)
		if age > 0 {
			frequency := float64(pattern.AccessCount) / age.Hours()

			// Very high frequency (>10/hour) = max TTL
			if frequency > 10 {
				ttl = p.MaxTTL
			} else if frequency > 5 {
				// High frequency (5-10/hour) = extended TTL
				ttl = time.Duration(float64(ttl) * 1.5)
			} else if frequency < 1 {
				// Low frequency (<1/hour) = reduced TTL
				ttl = time.Duration(float64(ttl) * 0.7)
			}
		}

		// Predict next use based on average interval
		if pattern.AverageInterval > 0 {
			predictedTTL := pattern.AverageInterval * 2
			if predictedTTL > p.MinTTL && predictedTTL < p.MaxTTL {
				// Use predicted TTL if it's reasonable
				ttl = predictedTTL
			}
		}
	}

	// Ensure within bounds
	if ttl < p.MinTTL {
		ttl = p.MinTTL
	}
	if ttl > p.MaxTTL {
		ttl = p.MaxTTL
	}

	return ttl
}

func (p *AdaptivePolicy) getPattern(key string) *AccessPattern {
	pattern, exists := p.patterns[key]
	if !exists {
		pattern = &AccessPattern{
			Key:            key,
			FirstAccess:    time.Now(),
			RecentAccesses: make([]time.Time, 0, 100),
		}
		p.patterns[key] = pattern
	}
	return pattern
}

// RecordAccess records an access event for pattern learning
func (p *AdaptivePolicy) RecordAccess(key string) {
	pattern := p.getPattern(key)
	now := time.Now()

	pattern.AccessCount++
	pattern.LastAccess = now
	pattern.RecentAccesses = append(pattern.RecentAccesses, now)

	// Keep only last 100 accesses
	if len(pattern.RecentAccesses) > 100 {
		pattern.RecentAccesses = pattern.RecentAccesses[1:]
	}

	// Calculate average interval
	if len(pattern.RecentAccesses) > 1 {
		totalInterval := time.Duration(0)
		for i := 1; i < len(pattern.RecentAccesses); i++ {
			interval := pattern.RecentAccesses[i].Sub(pattern.RecentAccesses[i-1])
			totalInterval += interval
		}
		pattern.AverageInterval = totalInterval / time.Duration(len(pattern.RecentAccesses)-1)

		// Predict next use
		pattern.PredictedNextUse = now.Add(pattern.AverageInterval)

		// Calculate volatility (standard deviation of intervals)
		intervals := make([]time.Duration, 0, len(pattern.RecentAccesses)-1)
		for i := 1; i < len(pattern.RecentAccesses); i++ {
			intervals = append(intervals, pattern.RecentAccesses[i].Sub(pattern.RecentAccesses[i-1]))
		}
		pattern.Volatility = calculateVolatility(intervals, pattern.AverageInterval)
	}
}

// TTLPolicy implements a simple time-based TTL policy
type TTLPolicy struct {
	TTL time.Duration // Fixed TTL for all entries
}

// NewTTLPolicy creates a new fixed TTL policy
func NewTTLPolicy(ttl time.Duration) *TTLPolicy {
	return &TTLPolicy{TTL: ttl}
}

func (p *TTLPolicy) Name() string {
	return "ttl"
}

func (p *TTLPolicy) ComputeTTL(entry *CacheEntry, pattern *AccessPattern) time.Duration {
	return p.TTL
}

func (p *TTLPolicy) ShouldCache(statusCode int, size int64, headers map[string]string) bool {
	// Cache successful responses only
	return statusCode >= 200 && statusCode < 300
}

func (p *TTLPolicy) CanEvict(entry *CacheEntry) bool {
	return true // Allow any eviction
}

// LFUPolicy implements a Least Frequently Used eviction policy
type LFUPolicy struct {
	*DefaultPolicy
	minAccessCount int64 // Minimum access count to prevent eviction
}

// NewLFUPolicy creates a new LFU policy
func NewLFUPolicy(minAccessCount int64) *LFUPolicy {
	return &LFUPolicy{
		DefaultPolicy:  NewDefaultPolicy(),
		minAccessCount: minAccessCount,
	}
}

func (p *LFUPolicy) Name() string {
	return "lfu"
}

func (p *LFUPolicy) CanEvict(entry *CacheEntry) bool {
	// Protect frequently accessed entries
	if entry.AccessCount >= p.minAccessCount {
		return false
	}
	return p.DefaultPolicy.CanEvict(entry)
}

// calculateVolatility computes the coefficient of variation for intervals
func calculateVolatility(intervals []time.Duration, mean time.Duration) float64 {
	if len(intervals) == 0 || mean == 0 {
		return 0
	}

	// Calculate variance
	var variance float64
	meanFloat := float64(mean)
	for _, interval := range intervals {
		diff := float64(interval) - meanFloat
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	// Standard deviation
	stdDev := variance
	if variance > 0 {
		// Simple approximation of sqrt
		stdDev = meanFloat
		for i := 0; i < 10; i++ {
			stdDev = (stdDev + variance/stdDev) / 2
		}
	}

	// Coefficient of variation (normalized volatility)
	return stdDev / meanFloat
}

// PolicyConfig represents serializable policy configuration
type PolicyConfig struct {
	Type          string        `yaml:"type" json:"type"` // "default", "adaptive", "ttl", "lfu"
	BaseTTL       string        `yaml:"base_ttl" json:"base_ttl"`
	MinTTL        string        `yaml:"min_ttl" json:"min_ttl"`
	MaxTTL        string        `yaml:"max_ttl" json:"max_ttl"`
	MaxCacheSize  int64         `yaml:"max_cache_size_mb" json:"max_cache_size_mb"`
	MinAccessCount int64        `yaml:"min_access_count" json:"min_access_count"`
}

// CreatePolicy creates a policy from configuration
func CreatePolicy(config PolicyConfig) (CachePolicy, error) {
	switch config.Type {
	case "adaptive":
		policy := NewAdaptivePolicy()
		if config.BaseTTL != "" {
			if ttl, err := time.ParseDuration(config.BaseTTL); err == nil {
				policy.BaseTTL = ttl
			}
		}
		if config.MinTTL != "" {
			if ttl, err := time.ParseDuration(config.MinTTL); err == nil {
				policy.MinTTL = ttl
			}
		}
		if config.MaxTTL != "" {
			if ttl, err := time.ParseDuration(config.MaxTTL); err == nil {
				policy.MaxTTL = ttl
			}
		}
		if config.MaxCacheSize > 0 {
			policy.MaxCacheSize = config.MaxCacheSize * 1024 * 1024
		}
		return policy, nil

	case "ttl":
		ttl := 5 * time.Minute
		if config.BaseTTL != "" {
			if parsedTTL, err := time.ParseDuration(config.BaseTTL); err == nil {
				ttl = parsedTTL
			}
		}
		return NewTTLPolicy(ttl), nil

	case "lfu":
		minAccess := config.MinAccessCount
		if minAccess == 0 {
			minAccess = 10
		}
		return NewLFUPolicy(minAccess), nil

	default: // "default"
		policy := NewDefaultPolicy()
		if config.BaseTTL != "" {
			if ttl, err := time.ParseDuration(config.BaseTTL); err == nil {
				policy.BaseTTL = ttl
			}
		}
		if config.MinTTL != "" {
			if ttl, err := time.ParseDuration(config.MinTTL); err == nil {
				policy.MinTTL = ttl
			}
		}
		if config.MaxTTL != "" {
			if ttl, err := time.ParseDuration(config.MaxTTL); err == nil {
				policy.MaxTTL = ttl
			}
		}
		if config.MaxCacheSize > 0 {
			policy.MaxCacheSize = config.MaxCacheSize * 1024 * 1024
		}
		return policy, nil
	}
}
