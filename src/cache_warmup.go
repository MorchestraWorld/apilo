package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// WarmupStrategy defines the interface for cache warming strategies
type WarmupStrategy interface {
	// Warmup performs the cache warming operation
	Warmup(ctx context.Context, cache *LRUCache) error

	// Name returns the strategy name
	Name() string
}

// PrefetchRequest represents a request to prefetch into cache
type PrefetchRequest struct {
	URL        string
	Priority   int       // Higher priority = fetch first
	ExpectedAt time.Time // When we expect this to be needed
}

// PrefetchQueue manages prioritized prefetch requests
type PrefetchQueue struct {
	requests []*PrefetchRequest
	mu       sync.Mutex
}

// NewPrefetchQueue creates a new prefetch queue
func NewPrefetchQueue() *PrefetchQueue {
	return &PrefetchQueue{
		requests: make([]*PrefetchRequest, 0),
	}
}

// Add adds a prefetch request to the queue
func (q *PrefetchQueue) Add(req *PrefetchRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.requests = append(q.requests, req)

	// Sort by priority (descending)
	sort.Slice(q.requests, func(i, j int) bool {
		return q.requests[i].Priority > q.requests[j].Priority
	})
}

// Pop removes and returns the highest priority request
func (q *PrefetchQueue) Pop() *PrefetchRequest {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.requests) == 0 {
		return nil
	}

	req := q.requests[0]
	q.requests = q.requests[1:]
	return req
}

// Size returns the number of pending requests
func (q *PrefetchQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.requests)
}

// Clear removes all pending requests
func (q *PrefetchQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.requests = make([]*PrefetchRequest, 0)
}

// StaticWarmup implements a static list-based warmup strategy
type StaticWarmup struct {
	urls []string // URLs to warm up
}

// NewStaticWarmup creates a static warmup strategy
func NewStaticWarmup(urls []string) *StaticWarmup {
	return &StaticWarmup{urls: urls}
}

func (s *StaticWarmup) Name() string {
	return "static"
}

func (s *StaticWarmup) Warmup(ctx context.Context, cache *LRUCache) error {
	// This would typically fetch the URLs and populate the cache
	// For now, we just create placeholder entries
	for _, url := range s.urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create a dummy cache entry
			key := (&CacheKey{URL: url, Method: "GET"}).Hash()
			entry := &CacheEntry{
				Key:          key,
				Value:        []byte("warmup_placeholder"),
				StatusCode:   200,
				Size:         20,
				CreatedAt:    time.Now(),
				LastAccessed: time.Now(),
				TTL:          5 * time.Minute,
				ExpiresAt:    time.Now().Add(5 * time.Minute),
			}
			cache.Put(key, entry)
		}
	}

	return nil
}

// PredictiveWarmup implements a pattern-based predictive warmup strategy
type PredictiveWarmup struct {
	patterns         map[string]*AccessPattern // Historical access patterns
	predictionWindow time.Duration             // How far ahead to predict
	topN             int                       // Number of top predictions to warm
	mu               sync.RWMutex
}

// NewPredictiveWarmup creates a predictive warmup strategy
func NewPredictiveWarmup(predictionWindow time.Duration, topN int) *PredictiveWarmup {
	return &PredictiveWarmup{
		patterns:         make(map[string]*AccessPattern),
		predictionWindow: predictionWindow,
		topN:             topN,
	}
}

func (p *PredictiveWarmup) Name() string {
	return "predictive"
}

// LearnPattern learns an access pattern for a resource
func (p *PredictiveWarmup) LearnPattern(key string, pattern *AccessPattern) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.patterns[key] = pattern
}

// Predict predicts which resources will be accessed soon
func (p *PredictiveWarmup) Predict() []*PrefetchRequest {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now()
	predictions := make([]*PrefetchRequest, 0)

	for key, pattern := range p.patterns {
		// Check if predicted next use is within our prediction window
		if pattern.PredictedNextUse.After(now) &&
			pattern.PredictedNextUse.Before(now.Add(p.predictionWindow)) {

			// Calculate priority based on access frequency and prediction confidence
			priority := int(pattern.AccessCount)

			// Higher priority for more stable patterns (lower volatility)
			if pattern.Volatility < 0.3 {
				priority *= 2
			}

			predictions = append(predictions, &PrefetchRequest{
				URL:        key,
				Priority:   priority,
				ExpectedAt: pattern.PredictedNextUse,
			})
		}
	}

	// Sort by priority and take top N
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Priority > predictions[j].Priority
	})

	if len(predictions) > p.topN {
		predictions = predictions[:p.topN]
	}

	return predictions
}

func (p *PredictiveWarmup) Warmup(ctx context.Context, cache *LRUCache) error {
	predictions := p.Predict()

	for _, pred := range predictions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create cache entry based on prediction
			entry := &CacheEntry{
				Key:          pred.URL,
				Value:        []byte("predictive_warmup"),
				StatusCode:   200,
				Size:         18,
				CreatedAt:    time.Now(),
				LastAccessed: time.Now(),
				TTL:          5 * time.Minute,
				ExpiresAt:    time.Now().Add(5 * time.Minute),
			}
			cache.Put(pred.URL, entry)
		}
	}

	return nil
}

// TimeBasedWarmup implements warmup based on time-of-day patterns
type TimeBasedWarmup struct {
	schedules map[int][]string // Hour -> URLs to warm
}

// NewTimeBasedWarmup creates a time-based warmup strategy
func NewTimeBasedWarmup() *TimeBasedWarmup {
	return &TimeBasedWarmup{
		schedules: make(map[int][]string),
	}
}

func (t *TimeBasedWarmup) Name() string {
	return "time_based"
}

// AddSchedule adds URLs to warm at a specific hour
func (t *TimeBasedWarmup) AddSchedule(hour int, urls []string) {
	if hour < 0 || hour > 23 {
		return
	}
	t.schedules[hour] = urls
}

func (t *TimeBasedWarmup) Warmup(ctx context.Context, cache *LRUCache) error {
	currentHour := time.Now().Hour()
	urls, exists := t.schedules[currentHour]

	if !exists {
		return nil // No warmup scheduled for this hour
	}

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			key := (&CacheKey{URL: url, Method: "GET"}).Hash()
			entry := &CacheEntry{
				Key:          key,
				Value:        []byte("time_based_warmup"),
				StatusCode:   200,
				Size:         17,
				CreatedAt:    time.Now(),
				LastAccessed: time.Now(),
				TTL:          1 * time.Hour,
				ExpiresAt:    time.Now().Add(1 * time.Hour),
			}
			cache.Put(key, entry)
		}
	}

	return nil
}

// AdaptiveWarmup combines multiple strategies
type AdaptiveWarmup struct {
	strategies []WarmupStrategy
	weights    map[string]float64 // Strategy performance weights
	mu         sync.RWMutex
}

// NewAdaptiveWarmup creates an adaptive warmup combining multiple strategies
func NewAdaptiveWarmup(strategies ...WarmupStrategy) *AdaptiveWarmup {
	weights := make(map[string]float64)
	for _, strategy := range strategies {
		weights[strategy.Name()] = 1.0 // Equal weight initially
	}

	return &AdaptiveWarmup{
		strategies: strategies,
		weights:    weights,
	}
}

func (a *AdaptiveWarmup) Name() string {
	return "adaptive"
}

func (a *AdaptiveWarmup) Warmup(ctx context.Context, cache *LRUCache) error {
	// Execute all strategies in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, len(a.strategies))

	for _, strategy := range a.strategies {
		wg.Add(1)
		go func(s WarmupStrategy) {
			defer wg.Done()
			if err := s.Warmup(ctx, cache); err != nil {
				errChan <- fmt.Errorf("%s warmup failed: %w", s.Name(), err)
			}
		}(strategy)
	}

	wg.Wait()
	close(errChan)

	// Collect any errors
	for err := range errChan {
		return err
	}

	return nil
}

// UpdateWeights adjusts strategy weights based on performance
func (a *AdaptiveWarmup) UpdateWeights(strategyName string, performance float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Simple exponential moving average
	currentWeight := a.weights[strategyName]
	a.weights[strategyName] = 0.7*currentWeight + 0.3*performance
}

// CacheWarmer manages cache warming operations
type CacheWarmer struct {
	cache          *LRUCache
	strategy       WarmupStrategy
	queue          *PrefetchQueue
	running        bool
	stopChan       chan struct{}
	warmupInterval time.Duration
	mu             sync.Mutex
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(cache *LRUCache, strategy WarmupStrategy) *CacheWarmer {
	return &CacheWarmer{
		cache:          cache,
		strategy:       strategy,
		queue:          NewPrefetchQueue(),
		warmupInterval: 15 * time.Minute, // Default: warm every 15 minutes
	}
}

// SetWarmupInterval sets how frequently to perform warmup
func (w *CacheWarmer) SetWarmupInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.warmupInterval = interval
}

// SetStrategy changes the warmup strategy
func (w *CacheWarmer) SetStrategy(strategy WarmupStrategy) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.strategy = strategy
}

// Start begins periodic cache warming
func (w *CacheWarmer) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("cache warmer already running")
	}
	w.running = true
	w.stopChan = make(chan struct{})
	interval := w.warmupInterval
	w.mu.Unlock()

	// Perform initial warmup
	if err := w.WarmupNow(ctx); err != nil {
		return fmt.Errorf("initial warmup failed: %w", err)
	}

	// Start periodic warmup
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				warmupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				if err := w.WarmupNow(warmupCtx); err != nil {
					// Log error but continue
					fmt.Printf("Warmup failed: %v\n", err)
				}
				cancel()

			case <-w.stopChan:
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop halts cache warming
func (w *CacheWarmer) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		close(w.stopChan)
		w.running = false
	}
}

// WarmupNow performs immediate cache warmup
func (w *CacheWarmer) WarmupNow(ctx context.Context) error {
	w.mu.Lock()
	strategy := w.strategy
	w.mu.Unlock()

	if strategy == nil {
		return fmt.Errorf("no warmup strategy configured")
	}

	return strategy.Warmup(ctx, w.cache)
}

// Prefetch adds a URL to the prefetch queue
func (w *CacheWarmer) Prefetch(url string, priority int) {
	w.queue.Add(&PrefetchRequest{
		URL:        url,
		Priority:   priority,
		ExpectedAt: time.Now().Add(5 * time.Minute),
	})
}

// ProcessPrefetchQueue processes pending prefetch requests
func (w *CacheWarmer) ProcessPrefetchQueue(ctx context.Context, maxItems int) error {
	processed := 0

	for processed < maxItems {
		req := w.queue.Pop()
		if req == nil {
			break // Queue empty
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create cache entry for prefetched item
			key := (&CacheKey{URL: req.URL, Method: "GET"}).Hash()
			entry := &CacheEntry{
				Key:          key,
				Value:        []byte("prefetched"),
				StatusCode:   200,
				Size:         10,
				CreatedAt:    time.Now(),
				LastAccessed: time.Now(),
				TTL:          5 * time.Minute,
				ExpiresAt:    time.Now().Add(5 * time.Minute),
			}
			w.cache.Put(key, entry)
			processed++
		}
	}

	return nil
}

// WarmupConfig represents warmup configuration
// MOVED TO types.go
// type WarmupConfig struct {
// 	Enabled        bool          `yaml:"enabled" json:"enabled"`
// 	Strategy       string        `yaml:"strategy" json:"strategy"` // "static", "predictive", "time_based", "adaptive"
// 	Interval       string        `yaml:"interval" json:"interval"`
// 	StaticURLs     []string      `yaml:"static_urls" json:"static_urls"`
// 	PredictionWindow string      `yaml:"prediction_window" json:"prediction_window"`
// 	TopN           int           `yaml:"top_n" json:"top_n"`
// }

// CreateWarmupStrategy creates a warmup strategy from configuration
func CreateWarmupStrategy(config WarmupConfig) (WarmupStrategy, error) {
	switch config.Strategy {
	case "static":
		return NewStaticWarmup(config.StaticURLs), nil

	case "predictive":
		window := 30 * time.Minute
		if config.PredictionWindow != "" {
			if parsed, err := time.ParseDuration(config.PredictionWindow); err == nil {
				window = parsed
			}
		}
		topN := config.TopN
		if topN == 0 {
			topN = 10
		}
		return NewPredictiveWarmup(window, topN), nil

	case "time_based":
		return NewTimeBasedWarmup(), nil

	case "adaptive":
		// Create adaptive strategy with static and predictive
		static := NewStaticWarmup(config.StaticURLs)
		predictive := NewPredictiveWarmup(30*time.Minute, 10)
		return NewAdaptiveWarmup(static, predictive), nil

	default:
		return nil, fmt.Errorf("unknown warmup strategy: %s", config.Strategy)
	}
}
