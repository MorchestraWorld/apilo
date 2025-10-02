// Package src provides integration and lifecycle management for all optimization components.
// This module orchestrates the HTTP/2 client, caching system, and monitoring framework.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// IntegratedOptimizer manages the complete optimization stack
type IntegratedOptimizer struct {
	// Core components
	client     *OptimizedClient
	benchmark  *BenchmarkEngine
	monitor    *Monitor

	// Configuration
	config     *IntegratedConfig

	// State management
	mu         sync.RWMutex
	running    bool
	initialized bool

	// Lifecycle management
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// Performance tracking
	startTime  time.Time
	stats      *IntegratedStats
}

// IntegratedConfig holds configuration for the complete optimization stack
type IntegratedConfig struct {
	// Client configuration
	ClientConfig *OptimizedClientConfig `yaml:"client"`

	// Benchmark configuration
	BenchmarkConfig *BenchmarkConfig `yaml:"benchmark"`

	// Monitoring configuration
	MonitoringConfig *MonitoringConfig `yaml:"monitoring"`

	// Integration settings
	WarmupEnabled    bool          `yaml:"warmup_enabled"`
	WarmupURLs       []string      `yaml:"warmup_urls"`
	WarmupTimeout    time.Duration `yaml:"warmup_timeout"`
	HealthCheckEnabled bool        `yaml:"health_check_enabled"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`

	// Performance targets
	TargetLatency    time.Duration `yaml:"target_latency"`
	MinCacheHitRatio float64       `yaml:"min_cache_hit_ratio"`
	MinConnectionReuse float64     `yaml:"min_connection_reuse"`
}

// DefaultIntegratedConfig returns an optimized configuration for API latency reduction
func DefaultIntegratedConfig() *IntegratedConfig {
	return &IntegratedConfig{
		ClientConfig:     DefaultOptimizedClientConfig(),
		BenchmarkConfig:  DefaultBenchmarkConfig(),
		MonitoringConfig: DefaultMonitoringConfig(),

		WarmupEnabled:       true,
		WarmupURLs:          []string{},
		WarmupTimeout:       30 * time.Second,
		HealthCheckEnabled:  true,
		HealthCheckInterval: 30 * time.Second,

		TargetLatency:       100 * time.Millisecond,
		MinCacheHitRatio:    0.6,
		MinConnectionReuse:  0.9,
	}
}

// IntegratedStats contains performance statistics for the complete optimization stack
type IntegratedStats struct {
	// Overall performance
	StartTime        time.Time     `json:"start_time"`
	Uptime          time.Duration `json:"uptime"`
	TotalRequests   int64         `json:"total_requests"`
	AverageLatency  time.Duration `json:"average_latency"`

	// Client statistics
	ClientStats     *OptimizedClientStats `json:"client_stats"`

	// Benchmark statistics
	BenchmarkStats  *BenchmarkStats       `json:"benchmark_stats"`

	// Target achievement
	LatencyTarget   bool    `json:"latency_target_met"`
	CacheTarget     bool    `json:"cache_target_met"`
	ConnectionTarget bool   `json:"connection_target_met"`
	OverallGrade    string  `json:"overall_grade"`

	// Health status
	HealthStatus    string  `json:"health_status"`
	LastHealthCheck time.Time `json:"last_health_check"`
}

// NewIntegratedOptimizer creates a new integrated optimization system
func NewIntegratedOptimizer(config *IntegratedConfig) (*IntegratedOptimizer, error) {
	if config == nil {
		config = DefaultIntegratedConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	optimizer := &IntegratedOptimizer{
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
		stats:     &IntegratedStats{
			StartTime: time.Now(),
			HealthStatus: "initializing",
		},
	}

	// Initialize optimized client
	client, err := NewOptimizedClient(config.ClientConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create optimized client: %w", err)
	}
	optimizer.client = client

	// Initialize benchmark engine with optimized client
	benchmarkConfig := config.BenchmarkConfig
	benchmarkConfig.OptimizedClient = client // Inject the optimized client

	benchmark, err := NewBenchmarkEngine(benchmarkConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create benchmark engine: %w", err)
	}
	optimizer.benchmark = benchmark

	// Initialize monitoring if enabled
	if config.MonitoringConfig.Enabled {
		monitor, err := NewMonitor(config.MonitoringConfig)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create monitor: %w", err)
		}
		optimizer.monitor = monitor
	}

	optimizer.initialized = true
	return optimizer, nil
}

// Start initializes and starts all optimization components
func (io *IntegratedOptimizer) Start() error {
	io.mu.Lock()
	defer io.mu.Unlock()

	if io.running {
		return fmt.Errorf("optimizer already running")
	}

	if !io.initialized {
		return fmt.Errorf("optimizer not initialized")
	}

	log.Println("Starting integrated API latency optimizer...")

	// Start monitoring first
	if io.monitor != nil {
		if err := io.monitor.Start(); err != nil {
			return fmt.Errorf("failed to start monitor: %w", err)
		}
		log.Println("✓ Monitoring system started")
	}

	// Perform cache warmup if enabled
	if io.config.WarmupEnabled {
		if err := io.performWarmup(); err != nil {
			log.Printf("⚠ Cache warmup failed: %v", err)
		} else {
			log.Println("✓ Cache warmup completed")
		}
	}

	// Start health checking if enabled
	if io.config.HealthCheckEnabled {
		io.wg.Add(1)
		go io.healthCheckLoop()
		log.Println("✓ Health checking started")
	}

	// Update statistics collection
	io.wg.Add(1)
	go io.statsCollectionLoop()

	io.running = true
	io.stats.HealthStatus = "running"

	log.Println("🚀 Integrated optimizer started successfully")
	return nil
}

// Stop gracefully shuts down all components
func (io *IntegratedOptimizer) Stop() error {
	io.mu.Lock()
	defer io.mu.Unlock()

	if !io.running {
		return nil
	}

	log.Println("Stopping integrated optimizer...")

	// Signal all goroutines to stop
	io.cancel()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		io.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✓ All background processes stopped")
	case <-time.After(10 * time.Second):
		log.Println("⚠ Timeout waiting for background processes")
	}

	// Stop client
	if io.client != nil {
		if err := io.client.Stop(); err != nil {
			log.Printf("⚠ Client stop error: %v", err)
		}
	}

	// Stop monitoring
	if io.monitor != nil {
		if err := io.monitor.Stop(); err != nil {
			log.Printf("⚠ Monitor stop error: %v", err)
		}
	}

	io.running = false
	io.stats.HealthStatus = "stopped"

	log.Println("✓ Integrated optimizer stopped")
	return nil
}

// performWarmup executes cache warmup with configured URLs
func (io *IntegratedOptimizer) performWarmup() error {
	if io.client == nil {
		return fmt.Errorf("client not initialized")
	}

	warmupURLs := io.config.WarmupURLs
	if len(warmupURLs) == 0 {
		// Use default warmup URLs if none configured
		warmupURLs = []string{
			"https://api.anthropic.com",
			"https://httpbin.org/get",
			"https://jsonplaceholder.typicode.com/posts/1",
		}
	}

	log.Printf("Starting cache warmup with %d URLs...", len(warmupURLs))

	// Set warmup timeout
	ctx, cancel := context.WithTimeout(io.ctx, io.config.WarmupTimeout)
	defer cancel()

	// Create a channel to track warmup completion
	done := make(chan error, 1)

	go func() {
		err := io.client.WarmupCache(warmupURLs)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("warmup failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("warmup timeout")
	}
}

// healthCheckLoop performs periodic health checks
func (io *IntegratedOptimizer) healthCheckLoop() {
	defer io.wg.Done()

	ticker := time.NewTicker(io.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-io.ctx.Done():
			return
		case <-ticker.C:
			io.performHealthCheck()
		}
	}
}

// performHealthCheck validates system health and performance
func (io *IntegratedOptimizer) performHealthCheck() {
	io.mu.Lock()
	defer io.mu.Unlock()

	io.stats.LastHealthCheck = time.Now()

	// Check client health
	if io.client == nil || !io.client.IsInitialized() {
		io.stats.HealthStatus = "unhealthy"
		return
	}

	// Get current statistics
	clientStats := io.client.GetStats()
	io.stats.ClientStats = clientStats

	// Check performance targets
	io.stats.LatencyTarget = clientStats.AverageLatency <= io.config.TargetLatency
	io.stats.CacheTarget = clientStats.CacheHitRatio >= io.config.MinCacheHitRatio
	io.stats.ConnectionTarget = clientStats.ConnectionReuseRatio >= io.config.MinConnectionReuse

	// Calculate overall grade
	targetsHit := 0
	if io.stats.LatencyTarget {
		targetsHit++
	}
	if io.stats.CacheTarget {
		targetsHit++
	}
	if io.stats.ConnectionTarget {
		targetsHit++
	}

	switch targetsHit {
	case 3:
		io.stats.OverallGrade = "A"
		io.stats.HealthStatus = "excellent"
	case 2:
		io.stats.OverallGrade = "B"
		io.stats.HealthStatus = "good"
	case 1:
		io.stats.OverallGrade = "C"
		io.stats.HealthStatus = "fair"
	default:
		io.stats.OverallGrade = "D"
		io.stats.HealthStatus = "poor"
	}

	// Log health status periodically
	if io.stats.LastHealthCheck.Sub(io.startTime) > time.Minute {
		log.Printf("Health Check - Grade: %s, Latency: %v, Cache: %.2f%%, Connections: %.2f%%",
			io.stats.OverallGrade,
			clientStats.AverageLatency,
			clientStats.CacheHitRatio*100,
			clientStats.ConnectionReuseRatio*100,
		)
	}
}

// statsCollectionLoop collects and updates performance statistics
func (io *IntegratedOptimizer) statsCollectionLoop() {
	defer io.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-io.ctx.Done():
			return
		case <-ticker.C:
			io.updateStats()
		}
	}
}

// updateStats updates performance statistics
func (io *IntegratedOptimizer) updateStats() {
	io.mu.Lock()
	defer io.mu.Unlock()

	io.stats.Uptime = time.Since(io.startTime)

	if io.client != nil {
		clientStats := io.client.GetStats()
		io.stats.ClientStats = clientStats
		io.stats.TotalRequests = clientStats.TotalRequests
		io.stats.AverageLatency = clientStats.AverageLatency
	}

	if io.benchmark != nil {
		benchmarkStats := io.benchmark.GetStats()
		io.stats.BenchmarkStats = benchmarkStats
	}
}

// RunBenchmark executes a performance benchmark using the optimized client
func (io *IntegratedOptimizer) RunBenchmark(config *BenchmarkRunConfig) (*BenchmarkResult, error) {
	if !io.running {
		return nil, fmt.Errorf("optimizer not running")
	}

	if io.benchmark == nil {
		return nil, fmt.Errorf("benchmark engine not initialized")
	}

	log.Printf("Running benchmark: %d requests, %d concurrent", config.TotalRequests, config.Concurrency)

	// Run the benchmark
	result, err := io.benchmark.Run(config)
	if err != nil {
		return nil, fmt.Errorf("benchmark failed: %w", err)
	}

	// Log results
	log.Printf("✓ Benchmark completed - P50: %v, P95: %v, Throughput: %.2f req/s",
		result.Latency.P50,
		result.Latency.P95,
		result.Throughput.RequestsPerSecond,
	)

	return result, nil
}

// RunComparisonBenchmark runs both optimized and baseline benchmarks for comparison
func (io *IntegratedOptimizer) RunComparisonBenchmark(config *BenchmarkRunConfig) (*ComparisonResult, error) {
	if !io.running {
		return nil, fmt.Errorf("optimizer not running")
	}

	log.Println("Running comparison benchmark (optimized vs baseline)...")

	// Run optimized benchmark
	optimizedResult, err := io.RunBenchmark(config)
	if err != nil {
		return nil, fmt.Errorf("optimized benchmark failed: %w", err)
	}

	// Create baseline client for comparison
	baselineConfig := &BenchmarkRunConfig{
		URL:              config.URL,
		TotalRequests:    config.TotalRequests,
		Concurrency:      config.Concurrency,
		Timeout:          config.Timeout,
		UseOptimizations: false, // Disable optimizations for baseline
	}

	baselineResult, err := io.benchmark.RunBaseline(baselineConfig)
	if err != nil {
		return nil, fmt.Errorf("baseline benchmark failed: %w", err)
	}

	// Calculate improvement metrics
	comparison := &ComparisonResult{
		Optimized: optimizedResult,
		Baseline:  baselineResult,
		Improvement: &ImprovementMetrics{
			LatencyImprovement: calculateLatencyImprovement(baselineResult, optimizedResult),
			ThroughputImprovement: calculateThroughputImprovement(baselineResult, optimizedResult),
			EfficiencyScore: calculateEfficiencyScore(baselineResult, optimizedResult),
		},
	}

	// Log comparison results
	log.Printf("✓ Comparison completed:")
	log.Printf("  Latency improvement: %.1f%% (P50: %v → %v)",
		comparison.Improvement.LatencyImprovement*100,
		baselineResult.Latency.P50,
		optimizedResult.Latency.P50,
	)
	log.Printf("  Throughput improvement: %.1f%% (%.2f → %.2f req/s)",
		comparison.Improvement.ThroughputImprovement*100,
		baselineResult.Throughput.RequestsPerSecond,
		optimizedResult.Throughput.RequestsPerSecond,
	)

	return comparison, nil
}

// GetCurrentStats returns current performance statistics
func (io *IntegratedOptimizer) GetCurrentStats() *IntegratedStats {
	io.mu.RLock()
	defer io.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *io.stats
	return &stats
}

// IsHealthy returns true if the system is performing within targets
func (io *IntegratedOptimizer) IsHealthy() bool {
	io.mu.RLock()
	defer io.mu.RUnlock()

	return io.stats.HealthStatus == "excellent" || io.stats.HealthStatus == "good"
}

// GetMonitoringURL returns the URL for the monitoring dashboard
func (io *IntegratedOptimizer) GetMonitoringURL() string {
	if io.monitor == nil {
		return ""
	}

	port := io.config.MonitoringConfig.DashboardPort
	return fmt.Sprintf("http://localhost:%d", port)
}

// ComparisonResult contains results from optimized vs baseline comparison
type ComparisonResult struct {
	Optimized   *BenchmarkResult     `json:"optimized"`
	Baseline    *BenchmarkResult     `json:"baseline"`
	Improvement *ImprovementMetrics  `json:"improvement"`
}

// ImprovementMetrics contains improvement calculations
type ImprovementMetrics struct {
	LatencyImprovement    float64 `json:"latency_improvement"`
	ThroughputImprovement float64 `json:"throughput_improvement"`
	EfficiencyScore       float64 `json:"efficiency_score"`
}

// calculateLatencyImprovement calculates the percentage improvement in latency
func calculateLatencyImprovement(baseline, optimized *BenchmarkResult) float64 {
	if baseline.Latency.P50 == 0 {
		return 0
	}

	return 1.0 - (float64(optimized.Latency.P50) / float64(baseline.Latency.P50))
}

// calculateThroughputImprovement calculates the percentage improvement in throughput
func calculateThroughputImprovement(baseline, optimized *BenchmarkResult) float64 {
	if baseline.Throughput.RequestsPerSecond == 0 {
		return 0
	}

	return (optimized.Throughput.RequestsPerSecond / baseline.Throughput.RequestsPerSecond) - 1.0
}

// calculateEfficiencyScore calculates an overall efficiency score
func calculateEfficiencyScore(baseline, optimized *BenchmarkResult) float64 {
	latencyImprovement := calculateLatencyImprovement(baseline, optimized)
	throughputImprovement := calculateThroughputImprovement(baseline, optimized)

	// Weighted average (latency more important than throughput)
	return (latencyImprovement * 0.7) + (throughputImprovement * 0.3)
}

// IsRunning returns true if the optimizer is currently running
func (io *IntegratedOptimizer) IsRunning() bool {
	io.mu.RLock()
	defer io.mu.RUnlock()
	return io.running
}

// IsInitialized returns true if the optimizer is fully initialized
func (io *IntegratedOptimizer) IsInitialized() bool {
	io.mu.RLock()
	defer io.mu.RUnlock()
	return io.initialized
}