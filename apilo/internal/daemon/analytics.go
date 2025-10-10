package daemon

import (
	"sort"
	"sync"
	"time"
)

// RequestRecord represents a single request for analytics
type RequestRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	URL          string    `json:"url"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	Latency      int64     `json:"latency"` // nanoseconds
	CacheHit     bool      `json:"cache_hit"`
	Error        string    `json:"error,omitempty"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	TotalTokens  int64     `json:"total_tokens"`
	IsEstimated  bool      `json:"is_estimated"`
}

// Analytics provides enhanced metrics tracking and analysis
type Analytics struct {
	requestHistory []RequestRecord
	latencyHistory []int64 // For percentile calculations
	maxHistory     int
	errorBreakdown map[string]int64
	urlStats       map[string]*URLStats
	mu             sync.RWMutex
}

// URLStats tracks per-URL statistics
type URLStats struct {
	TotalRequests int64
	CacheHits     int64
	CacheMisses   int64
	TotalLatency  int64
	MinLatency    int64
	MaxLatency    int64
}

// AnalyticsSnapshot represents current analytics state
type AnalyticsSnapshot struct {
	RecentRequests     []RequestRecord        `json:"recent_requests"`
	LatencyPercentiles map[string]float64     `json:"latency_percentiles"`
	ErrorBreakdown     map[string]int64       `json:"error_breakdown"`
	TopURLs            []URLAnalytics         `json:"top_urls"`
	RequestRate        float64                `json:"request_rate"` // requests per second
	CacheEfficiency    CacheEfficiencyMetrics `json:"cache_efficiency"`
	TimeSeriesData     TimeSeriesData         `json:"time_series"`
	TokenSavings       TokenSavingsMetrics    `json:"token_savings"`
	TokenUsageMetrics  TokenUsageMetrics      `json:"token_usage_metrics"`
}

// URLAnalytics provides per-URL analytics
type URLAnalytics struct {
	URL           string  `json:"url"`
	TotalRequests int64   `json:"total_requests"`
	CacheHitRatio float64 `json:"cache_hit_ratio"`
	AvgLatency    int64   `json:"avg_latency"`
	MinLatency    int64   `json:"min_latency"`
	MaxLatency    int64   `json:"max_latency"`
}

// CacheEfficiencyMetrics provides cache performance details
type CacheEfficiencyMetrics struct {
	HitRate        float64 `json:"hit_rate"`
	MissRate       float64 `json:"miss_rate"`
	TotalHits      int64   `json:"total_hits"`
	TotalMisses    int64   `json:"total_misses"`
	AvgHitLatency  int64   `json:"avg_hit_latency"`
	AvgMissLatency int64   `json:"avg_miss_latency"`
}

// TimeSeriesData provides time-bucketed metrics
type TimeSeriesData struct {
	LastMinute   TimeBucket `json:"last_minute"`
	Last5Minutes TimeBucket `json:"last_5_minutes"`
	LastHour     TimeBucket `json:"last_hour"`
}

// TimeBucket represents metrics for a time period
type TimeBucket struct {
	Requests    int64   `json:"requests"`
	CacheHits   int64   `json:"cache_hits"`
	CacheMisses int64   `json:"cache_misses"`
	Errors      int64   `json:"errors"`
	AvgLatency  int64   `json:"avg_latency"`
	RequestRate float64 `json:"request_rate"`
}

// TokenSavingsMetrics represents estimated token and cost savings
type TokenSavingsMetrics struct {
	TotalCacheHits       int64   `json:"total_cache_hits"`
	EstimatedTokensSaved int64   `json:"estimated_tokens_saved"`
	EstimatedCostSavings float64 `json:"estimated_cost_savings"` // in dollars
	AvgTokensPerRequest  int64   `json:"avg_tokens_per_request"`
	CostPerMillionTokens float64 `json:"cost_per_million_tokens"`
	LatencySavingsMs     int64   `json:"latency_savings_ms"`
	ApiCallsSaved        int64   `json:"api_calls_saved"`
}

// TokenUsageMetrics represents overall token consumption tracking
type TokenUsageMetrics struct {
	TotalInputTokens    int64   `json:"total_input_tokens"`
	TotalOutputTokens   int64   `json:"total_output_tokens"`
	TotalTokens         int64   `json:"total_tokens"`
	EstimatedRequests   int64   `json:"estimated_requests"`
	ActualRequests      int64   `json:"actual_requests"`
	TotalRequests       int64   `json:"total_requests"`
	AvgInputPerRequest  float64 `json:"avg_input_per_request"`
	AvgOutputPerRequest float64 `json:"avg_output_per_request"`
	EstimatedCost       float64 `json:"estimated_cost"` // in dollars
}

// NewAnalytics creates a new analytics tracker
func NewAnalytics(maxHistory int) *Analytics {
	return &Analytics{
		requestHistory: make([]RequestRecord, 0, maxHistory),
		latencyHistory: make([]int64, 0, maxHistory),
		maxHistory:     maxHistory,
		errorBreakdown: make(map[string]int64),
		urlStats:       make(map[string]*URLStats),
	}
}

// RecordRequest records a request for analytics
func (a *Analytics) RecordRequest(record RequestRecord) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to request history (circular buffer)
	if len(a.requestHistory) >= a.maxHistory {
		a.requestHistory = a.requestHistory[1:]
	}
	a.requestHistory = append(a.requestHistory, record)

	// Add to latency history
	if len(a.latencyHistory) >= a.maxHistory {
		a.latencyHistory = a.latencyHistory[1:]
	}
	a.latencyHistory = append(a.latencyHistory, record.Latency)

	// Track errors
	if record.Error != "" {
		a.errorBreakdown[record.Error]++
	}

	// Track per-URL stats
	if stats, exists := a.urlStats[record.URL]; exists {
		stats.TotalRequests++
		stats.TotalLatency += record.Latency
		if record.CacheHit {
			stats.CacheHits++
		} else {
			stats.CacheMisses++
		}
		if record.Latency < stats.MinLatency || stats.MinLatency == 0 {
			stats.MinLatency = record.Latency
		}
		if record.Latency > stats.MaxLatency {
			stats.MaxLatency = record.Latency
		}
	} else {
		hits := int64(0)
		misses := int64(0)
		if record.CacheHit {
			hits = 1
		} else {
			misses = 1
		}
		a.urlStats[record.URL] = &URLStats{
			TotalRequests: 1,
			CacheHits:     hits,
			CacheMisses:   misses,
			TotalLatency:  record.Latency,
			MinLatency:    record.Latency,
			MaxLatency:    record.Latency,
		}
	}
}

// GetSnapshot returns current analytics snapshot with default limits
func (a *Analytics) GetSnapshot() *AnalyticsSnapshot {
	return a.GetSnapshotWithLimit(20)
}

// GetSnapshotWithLimit returns current analytics snapshot with configurable request limit
func (a *Analytics) GetSnapshotWithLimit(requestLimit int) *AnalyticsSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Enforce reasonable limits
	if requestLimit < 1 {
		requestLimit = 20
	}
	if requestLimit > a.maxHistory {
		requestLimit = a.maxHistory
	}

	snapshot := &AnalyticsSnapshot{
		RecentRequests:     a.getRecentRequests(requestLimit),
		LatencyPercentiles: a.calculatePercentiles(),
		ErrorBreakdown:     a.copyErrorBreakdown(),
		TopURLs:            a.getTopURLs(10),
		RequestRate:        a.calculateRequestRate(),
		CacheEfficiency:    a.calculateCacheEfficiency(),
		TimeSeriesData:     a.calculateTimeSeries(),
		TokenSavings:       a.calculateTokenSavings(),
		TokenUsageMetrics:  a.calculateTokenUsage(),
	}

	return snapshot
}

// getRecentRequests returns the N most recent requests
func (a *Analytics) getRecentRequests(n int) []RequestRecord {
	if len(a.requestHistory) == 0 {
		return []RequestRecord{}
	}

	start := 0
	if len(a.requestHistory) > n {
		start = len(a.requestHistory) - n
	}

	// Return copy
	records := make([]RequestRecord, len(a.requestHistory)-start)
	copy(records, a.requestHistory[start:])

	// Reverse to show newest first
	for i := 0; i < len(records)/2; i++ {
		j := len(records) - 1 - i
		records[i], records[j] = records[j], records[i]
	}

	return records
}

// calculatePercentiles calculates latency percentiles
func (a *Analytics) calculatePercentiles() map[string]float64 {
	if len(a.latencyHistory) == 0 {
		return map[string]float64{
			"p50": 0,
			"p95": 0,
			"p99": 0,
		}
	}

	// Create sorted copy
	sorted := make([]int64, len(a.latencyHistory))
	copy(sorted, a.latencyHistory)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	return map[string]float64{
		"p50": float64(sorted[len(sorted)*50/100]),
		"p95": float64(sorted[len(sorted)*95/100]),
		"p99": float64(sorted[len(sorted)*99/100]),
	}
}

// copyErrorBreakdown returns copy of error breakdown
func (a *Analytics) copyErrorBreakdown() map[string]int64 {
	breakdown := make(map[string]int64)
	for k, v := range a.errorBreakdown {
		breakdown[k] = v
	}
	return breakdown
}

// getTopURLs returns top N URLs by request count
func (a *Analytics) getTopURLs(n int) []URLAnalytics {
	urls := make([]URLAnalytics, 0, len(a.urlStats))

	for url, stats := range a.urlStats {
		hitRatio := 0.0
		if stats.TotalRequests > 0 {
			hitRatio = float64(stats.CacheHits) / float64(stats.TotalRequests)
		}
		avgLatency := int64(0)
		if stats.TotalRequests > 0 {
			avgLatency = stats.TotalLatency / stats.TotalRequests
		}

		urls = append(urls, URLAnalytics{
			URL:           url,
			TotalRequests: stats.TotalRequests,
			CacheHitRatio: hitRatio,
			AvgLatency:    avgLatency,
			MinLatency:    stats.MinLatency,
			MaxLatency:    stats.MaxLatency,
		})
	}

	// Sort by request count
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].TotalRequests > urls[j].TotalRequests
	})

	if len(urls) > n {
		urls = urls[:n]
	}

	return urls
}

// calculateRequestRate calculates requests per second based on recent history
func (a *Analytics) calculateRequestRate() float64 {
	if len(a.requestHistory) < 2 {
		return 0
	}

	first := a.requestHistory[0].Timestamp
	last := a.requestHistory[len(a.requestHistory)-1].Timestamp
	duration := last.Sub(first).Seconds()

	if duration == 0 {
		return 0
	}

	return float64(len(a.requestHistory)) / duration
}

// calculateCacheEfficiency calculates cache performance metrics
func (a *Analytics) calculateCacheEfficiency() CacheEfficiencyMetrics {
	var totalHits, totalMisses int64
	var hitLatency, missLatency int64
	var hitCount, missCount int64

	for _, record := range a.requestHistory {
		if record.CacheHit {
			totalHits++
			hitLatency += record.Latency
			hitCount++
		} else {
			totalMisses++
			missLatency += record.Latency
			missCount++
		}
	}

	total := totalHits + totalMisses
	hitRate := 0.0
	missRate := 0.0
	if total > 0 {
		hitRate = float64(totalHits) / float64(total)
		missRate = float64(totalMisses) / float64(total)
	}

	avgHitLatency := int64(0)
	if hitCount > 0 {
		avgHitLatency = hitLatency / hitCount
	}

	avgMissLatency := int64(0)
	if missCount > 0 {
		avgMissLatency = missLatency / missCount
	}

	return CacheEfficiencyMetrics{
		HitRate:        hitRate,
		MissRate:       missRate,
		TotalHits:      totalHits,
		TotalMisses:    totalMisses,
		AvgHitLatency:  avgHitLatency,
		AvgMissLatency: avgMissLatency,
	}
}

// calculateTimeSeries calculates time-bucketed metrics
func (a *Analytics) calculateTimeSeries() TimeSeriesData {
	now := time.Now()

	return TimeSeriesData{
		LastMinute:   a.calculateBucket(now.Add(-1 * time.Minute)),
		Last5Minutes: a.calculateBucket(now.Add(-5 * time.Minute)),
		LastHour:     a.calculateBucket(now.Add(-1 * time.Hour)),
	}
}

// calculateBucket calculates metrics for requests since cutoff time
func (a *Analytics) calculateBucket(cutoff time.Time) TimeBucket {
	var requests, cacheHits, cacheMisses, errors int64
	var totalLatency int64

	for _, record := range a.requestHistory {
		if record.Timestamp.Before(cutoff) {
			continue
		}

		requests++
		totalLatency += record.Latency

		if record.CacheHit {
			cacheHits++
		} else {
			cacheMisses++
		}

		if record.Error != "" {
			errors++
		}
	}

	avgLatency := int64(0)
	if requests > 0 {
		avgLatency = totalLatency / requests
	}

	duration := time.Since(cutoff).Seconds()
	requestRate := 0.0
	if duration > 0 {
		requestRate = float64(requests) / duration
	}

	return TimeBucket{
		Requests:    requests,
		CacheHits:   cacheHits,
		CacheMisses: cacheMisses,
		Errors:      errors,
		AvgLatency:  avgLatency,
		RequestRate: requestRate,
	}
}

// calculateTokenSavings estimates token and cost savings from cache hits
func (a *Analytics) calculateTokenSavings() TokenSavingsMetrics {
	// Count cache hits and track actual token usage
	var totalCacheHits int64
	var totalHitLatency int64
	var totalMissLatency int64
	var hitCount, missCount int64
	var totalInputFromCacheHits, totalOutputFromCacheHits int64

	for _, record := range a.requestHistory {
		if record.CacheHit {
			totalCacheHits++
			totalHitLatency += record.Latency
			hitCount++
			totalInputFromCacheHits += record.InputTokens
			totalOutputFromCacheHits += record.OutputTokens
		} else {
			totalMissLatency += record.Latency
			missCount++
		}
	}

	// Calculate average tokens from ACTUAL measured data
	avgTokensPerRequest := int64(0)
	if totalCacheHits > 0 {
		avgTokensPerRequest = (totalInputFromCacheHits + totalOutputFromCacheHits) / totalCacheHits
	} else {
		// Fallback: calculate average from all requests if no cache hits yet
		var totalTokensAll int64
		for _, record := range a.requestHistory {
			totalTokensAll += record.TotalTokens
		}
		if len(a.requestHistory) > 0 {
			avgTokensPerRequest = totalTokensAll / int64(len(a.requestHistory))
		} else {
			// Ultimate fallback for empty history
			avgTokensPerRequest = 1000
		}
	}

	// Total tokens saved (actual measurement, not estimate)
	estimatedTokensSaved := totalInputFromCacheHits + totalOutputFromCacheHits

	// Calculate cost savings using ACTUAL input/output pricing
	// Input: $3/MTok, Output: $15/MTok (Claude Sonnet-4)
	inputCostSavings := (float64(totalInputFromCacheHits) / 1000000.0) * 3.0
	outputCostSavings := (float64(totalOutputFromCacheHits) / 1000000.0) * 15.0
	estimatedCostSavings := inputCostSavings + outputCostSavings

	// Calculate blended cost per million tokens for display
	costPerMillionTokens := 0.0
	if estimatedTokensSaved > 0 {
		costPerMillionTokens = (estimatedCostSavings * 1000000.0) / float64(estimatedTokensSaved)
	}

	// Calculate latency savings (time saved by not making external API calls)
	avgMissLatency := int64(0)
	if missCount > 0 {
		avgMissLatency = totalMissLatency / missCount
	}

	latencySavingsNs := totalCacheHits * avgMissLatency
	latencySavingsMs := latencySavingsNs / 1000000 // Convert to milliseconds

	return TokenSavingsMetrics{
		TotalCacheHits:       totalCacheHits,
		EstimatedTokensSaved: estimatedTokensSaved,
		EstimatedCostSavings: estimatedCostSavings,
		AvgTokensPerRequest:  avgTokensPerRequest,
		CostPerMillionTokens: costPerMillionTokens,
		LatencySavingsMs:     latencySavingsMs,
		ApiCallsSaved:        totalCacheHits,
	}
}

// calculateTokenUsage calculates overall token consumption metrics
func (a *Analytics) calculateTokenUsage() TokenUsageMetrics {
	var totalInputTokens, totalOutputTokens int64
	var estimatedRequests, actualRequests int64

	for _, record := range a.requestHistory {
		totalInputTokens += record.InputTokens
		totalOutputTokens += record.OutputTokens

		if record.IsEstimated {
			estimatedRequests++
		} else {
			actualRequests++
		}
	}

	totalTokens := totalInputTokens + totalOutputTokens
	totalRequests := int64(len(a.requestHistory))

	avgInputPerRequest := 0.0
	if totalRequests > 0 {
		avgInputPerRequest = float64(totalInputTokens) / float64(totalRequests)
	}

	avgOutputPerRequest := 0.0
	if totalRequests > 0 {
		avgOutputPerRequest = float64(totalOutputTokens) / float64(totalRequests)
	}

	// Estimate cost based on Claude Sonnet-4 pricing
	// Input: $3/MTok, Output: $15/MTok
	inputCost := (float64(totalInputTokens) / 1000000.0) * 3.0
	outputCost := (float64(totalOutputTokens) / 1000000.0) * 15.0
	estimatedCost := inputCost + outputCost

	return TokenUsageMetrics{
		TotalInputTokens:    totalInputTokens,
		TotalOutputTokens:   totalOutputTokens,
		TotalTokens:         totalTokens,
		EstimatedRequests:   estimatedRequests,
		ActualRequests:      actualRequests,
		TotalRequests:       totalRequests,
		AvgInputPerRequest:  avgInputPerRequest,
		AvgOutputPerRequest: avgOutputPerRequest,
		EstimatedCost:       estimatedCost,
	}
}
