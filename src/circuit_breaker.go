// Package src provides circuit breaker and failover mechanisms for fault tolerance
// This addresses the critical single point of failure risk
package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// CircuitState represents the current state of the circuit breaker
type CircuitState int32

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (cs CircuitState) String() string {
	switch cs {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker provides fault tolerance through circuit breaking pattern
type CircuitBreaker struct {
	// Configuration
	config *CircuitBreakerConfig

	// State management
	state          int32 // CircuitState stored as int32 for atomic operations
	failureCount   int64
	successCount   int64
	requestCount   int64
	lastFailTime   int64 // Unix timestamp in nanoseconds
	lastStateChange int64 // Unix timestamp in nanoseconds

	// Metrics
	metrics *CircuitBreakerMetrics

	// Synchronization
	mutex sync.RWMutex

	// Half-open state management
	halfOpenRequests   int64
	halfOpenSuccesses  int64
	halfOpenStart      int64

	// Generation counter for state changes
	generation int64
}

// CircuitBreakerConfig configures the circuit breaker behavior
type CircuitBreakerConfig struct {
	// Failure threshold
	FailureThreshold     int           `yaml:"failure_threshold"`
	FailureRate          float64       `yaml:"failure_rate"`
	MinimumRequests      int           `yaml:"minimum_requests"`

	// Timing
	OpenTimeout          time.Duration `yaml:"open_timeout"`
	HalfOpenTimeout      time.Duration `yaml:"half_open_timeout"`
	ResetTimeout         time.Duration `yaml:"reset_timeout"`

	// Half-open state
	HalfOpenMaxRequests  int           `yaml:"half_open_max_requests"`
	HalfOpenSuccessThreshold int       `yaml:"half_open_success_threshold"`

	// Advanced settings
	ExponentialBackoff   bool          `yaml:"exponential_backoff"`
	MaxBackoffTime      time.Duration `yaml:"max_backoff_time"`
	BackoffMultiplier   float64       `yaml:"backoff_multiplier"`

	// Monitoring
	EnableMetrics       bool          `yaml:"enable_metrics"`
	MetricsWindowSize   int           `yaml:"metrics_window_size"`

	// Custom failure detection
	IsFailure           func(error) bool
	ShouldTrip          func(*CircuitBreakerMetrics) bool
}

// CircuitBreakerMetrics tracks circuit breaker performance
type CircuitBreakerMetrics struct {
	// Counters
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	RejectedRequests    int64

	// State changes
	StateChanges        int64
	OpenCount           int64
	HalfOpenCount       int64
	ClosedCount         int64

	// Timing
	SuccessRate         float64
	FailureRate         float64
	AverageLatency      time.Duration
	LastFailure         time.Time
	LastSuccess         time.Time

	// Windows for rolling metrics
	requestWindow       *RollingWindow
	latencyWindow       *RollingWindow

	mutex               sync.RWMutex
}

// RollingWindow maintains rolling statistics
type RollingWindow struct {
	values   []float64
	index    int
	size     int
	full     bool
	mutex    sync.RWMutex
}

// FailoverManager manages multiple circuit breakers and failover logic
type FailoverManager struct {
	// Primary and backup services
	primary   *CircuitBreaker
	backups   []*CircuitBreaker

	// Configuration
	config    *FailoverConfig

	// State
	currentService int32 // Index of currently active service
	fallbackMode   int32 // 0 = normal, 1 = fallback active

	// Metrics
	metrics   *FailoverMetrics

	// Health checking
	healthChecker *HealthChecker

	mutex     sync.RWMutex
}

// FailoverConfig configures failover behavior
type FailoverConfig struct {
	// Failover strategy
	Strategy             FailoverStrategy  `yaml:"strategy"`
	MaxRetries           int               `yaml:"max_retries"`
	RetryDelay           time.Duration     `yaml:"retry_delay"`

	// Health checking
	HealthCheckInterval  time.Duration     `yaml:"health_check_interval"`
	HealthCheckTimeout   time.Duration     `yaml:"health_check_timeout"`

	// Automatic recovery
	AutoRecovery         bool              `yaml:"auto_recovery"`
	RecoveryCheckInterval time.Duration    `yaml:"recovery_check_interval"`

	// Fallback behavior
	EnableFallback       bool              `yaml:"enable_fallback"`
	FallbackTimeout      time.Duration     `yaml:"fallback_timeout"`
}

// FailoverStrategy defines how failover should behave
type FailoverStrategy int

const (
	FailoverImmediate FailoverStrategy = iota
	FailoverGradual
	FailoverRoundRobin
	FailoverWeighted
)

// FailoverMetrics tracks failover performance
type FailoverMetrics struct {
	FailoverCount         int64
	RecoveryCount         int64
	FallbackActivations   int64
	HealthCheckFailures   int64
	TotalSwitches         int64

	CurrentServiceIndex   int32
	ServiceHealthStatus   map[int]bool

	mutex                 sync.RWMutex
}

// HealthChecker performs health checks on services
type HealthChecker struct {
	config    *HealthCheckConfig
	checkers  map[int]func() error
	results   map[int]*HealthCheckResult
	mutex     sync.RWMutex
}

// HealthCheckConfig configures health checking
type HealthCheckConfig struct {
	Interval           time.Duration `yaml:"interval"`
	Timeout            time.Duration `yaml:"timeout"`
	HealthyThreshold   int           `yaml:"healthy_threshold"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"`
}

// HealthCheckResult stores health check results
type HealthCheckResult struct {
	Healthy          bool
	LastCheck        time.Time
	ConsecutiveFails int
	ConsecutiveSuccess int
	LastError        error
}

// Common errors
var (
	ErrCircuitOpen         = errors.New("circuit breaker is open")
	ErrHalfOpenLimitExceeded = errors.New("half-open request limit exceeded")
	ErrAllServicesDown     = errors.New("all services are down")
	ErrNoHealthyService    = errors.New("no healthy service available")
	ErrFailoverInProgress  = errors.New("failover is in progress")
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	cb := &CircuitBreaker{
		config:  config,
		state:   int32(CircuitClosed),
		metrics: NewCircuitBreakerMetrics(config.MetricsWindowSize),
	}

	return cb
}

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return cb.ExecuteWithContext(context.Background(), fn)
}

// ExecuteWithContext runs a function through the circuit breaker with context
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Check if we can execute
	if err := cb.canExecute(); err != nil {
		atomic.AddInt64(&cb.metrics.RejectedRequests, 1)
		return nil, err
	}

	// Record request
	atomic.AddInt64(&cb.requestCount, 1)
	atomic.AddInt64(&cb.metrics.TotalRequests, 1)

	start := time.Now()
	result, err := fn()
	latency := time.Since(start)

	// Record latency
	cb.metrics.recordLatency(latency)

	// Handle result
	if err != nil {
		cb.onFailure(err)
		return nil, err
	}

	cb.onSuccess()
	return result, nil
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() error {
	state := CircuitState(atomic.LoadInt32(&cb.state))

	switch state {
	case CircuitClosed:
		return nil
	case CircuitOpen:
		if cb.shouldAttemptReset() {
			cb.transitionToHalfOpen()
			return nil
		}
		return ErrCircuitOpen
	case CircuitHalfOpen:
		if cb.canExecuteInHalfOpen() {
			return nil
		}
		return ErrHalfOpenLimitExceeded
	default:
		return ErrCircuitOpen
	}
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	atomic.AddInt64(&cb.successCount, 1)
	atomic.AddInt64(&cb.metrics.SuccessfulRequests, 1)

	cb.metrics.mutex.Lock()
	cb.metrics.LastSuccess = time.Now()
	cb.metrics.mutex.Unlock()

	state := CircuitState(atomic.LoadInt32(&cb.state))

	if state == CircuitHalfOpen {
		atomic.AddInt64(&cb.halfOpenSuccesses, 1)
		if cb.shouldTransitionToClosed() {
			cb.transitionToClosed()
		}
	}

	// Update success rate
	cb.updateMetrics()
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure(err error) {
	// Check if this error should be considered a failure
	if cb.config.IsFailure != nil && !cb.config.IsFailure(err) {
		return
	}

	atomic.AddInt64(&cb.failureCount, 1)
	atomic.AddInt64(&cb.metrics.FailedRequests, 1)
	atomic.StoreInt64(&cb.lastFailTime, time.Now().UnixNano())

	cb.metrics.mutex.Lock()
	cb.metrics.LastFailure = time.Now()
	cb.metrics.mutex.Unlock()

	state := CircuitState(atomic.LoadInt32(&cb.state))

	if state == CircuitHalfOpen {
		// Any failure in half-open state should open the circuit
		cb.transitionToOpen()
	} else if state == CircuitClosed {
		if cb.shouldTrip() {
			cb.transitionToOpen()
		}
	}

	// Update failure rate
	cb.updateMetrics()
}

// shouldTrip determines if the circuit should trip to open state
func (cb *CircuitBreaker) shouldTrip() bool {
	// Custom trip condition
	if cb.config.ShouldTrip != nil {
		return cb.config.ShouldTrip(cb.metrics)
	}

	// Default trip conditions
	requests := atomic.LoadInt64(&cb.requestCount)
	failures := atomic.LoadInt64(&cb.failureCount)

	// Must have minimum number of requests
	if requests < int64(cb.config.MinimumRequests) {
		return false
	}

	// Check failure threshold
	if failures >= int64(cb.config.FailureThreshold) {
		return true
	}

	// Check failure rate
	if cb.config.FailureRate > 0 {
		failureRate := float64(failures) / float64(requests)
		return failureRate >= cb.config.FailureRate
	}

	return false
}

// shouldAttemptReset determines if we should attempt to reset from open state
func (cb *CircuitBreaker) shouldAttemptReset() bool {
	lastFailTime := atomic.LoadInt64(&cb.lastFailTime)
	if lastFailTime == 0 {
		return true
	}

	timeout := cb.config.OpenTimeout
	if cb.config.ExponentialBackoff {
		// Calculate exponential backoff
		failures := atomic.LoadInt64(&cb.failureCount)
		backoffTime := time.Duration(math.Pow(cb.config.BackoffMultiplier, float64(failures))) * cb.config.OpenTimeout
		if backoffTime > cb.config.MaxBackoffTime {
			backoffTime = cb.config.MaxBackoffTime
		}
		timeout = backoffTime
	}

	elapsed := time.Duration(time.Now().UnixNano() - lastFailTime)
	return elapsed >= timeout
}

// canExecuteInHalfOpen checks if execution is allowed in half-open state
func (cb *CircuitBreaker) canExecuteInHalfOpen() bool {
	halfOpenRequests := atomic.LoadInt64(&cb.halfOpenRequests)
	return halfOpenRequests < int64(cb.config.HalfOpenMaxRequests)
}

// shouldTransitionToClosed determines if half-open circuit should close
func (cb *CircuitBreaker) shouldTransitionToClosed() bool {
	successes := atomic.LoadInt64(&cb.halfOpenSuccesses)
	return successes >= int64(cb.config.HalfOpenSuccessThreshold)
}

// State transitions
func (cb *CircuitBreaker) transitionToOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if CircuitState(atomic.LoadInt32(&cb.state)) != CircuitOpen {
		atomic.StoreInt32(&cb.state, int32(CircuitOpen))
		atomic.AddInt64(&cb.generation, 1)
		atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())

		atomic.AddInt64(&cb.metrics.StateChanges, 1)
		atomic.AddInt64(&cb.metrics.OpenCount, 1)

		cb.resetCounts()
	}
}

func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if CircuitState(atomic.LoadInt32(&cb.state)) != CircuitHalfOpen {
		atomic.StoreInt32(&cb.state, int32(CircuitHalfOpen))
		atomic.AddInt64(&cb.generation, 1)
		atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())
		atomic.StoreInt64(&cb.halfOpenStart, time.Now().UnixNano())

		atomic.AddInt64(&cb.metrics.StateChanges, 1)
		atomic.AddInt64(&cb.metrics.HalfOpenCount, 1)

		atomic.StoreInt64(&cb.halfOpenRequests, 0)
		atomic.StoreInt64(&cb.halfOpenSuccesses, 0)
	}
}

func (cb *CircuitBreaker) transitionToClosed() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if CircuitState(atomic.LoadInt32(&cb.state)) != CircuitClosed {
		atomic.StoreInt32(&cb.state, int32(CircuitClosed))
		atomic.AddInt64(&cb.generation, 1)
		atomic.StoreInt64(&cb.lastStateChange, time.Now().UnixNano())

		atomic.AddInt64(&cb.metrics.StateChanges, 1)
		atomic.AddInt64(&cb.metrics.ClosedCount, 1)

		cb.resetCounts()
	}
}

// resetCounts resets the failure and success counters
func (cb *CircuitBreaker) resetCounts() {
	atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt64(&cb.successCount, 0)
	atomic.StoreInt64(&cb.requestCount, 0)
}

// updateMetrics updates the rolling metrics
func (cb *CircuitBreaker) updateMetrics() {
	if !cb.config.EnableMetrics {
		return
	}

	cb.metrics.mutex.Lock()
	defer cb.metrics.mutex.Unlock()

	requests := atomic.LoadInt64(&cb.requestCount)
	successes := atomic.LoadInt64(&cb.successCount)
	failures := atomic.LoadInt64(&cb.failureCount)

	if requests > 0 {
		cb.metrics.SuccessRate = float64(successes) / float64(requests)
		cb.metrics.FailureRate = float64(failures) / float64(requests)
	}

	// Update rolling windows
	if cb.metrics.requestWindow != nil {
		cb.metrics.requestWindow.Add(float64(requests))
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.metrics.mutex.RLock()
	defer cb.metrics.mutex.RUnlock()

	return CircuitBreakerMetrics{
		TotalRequests:      atomic.LoadInt64(&cb.metrics.TotalRequests),
		SuccessfulRequests: atomic.LoadInt64(&cb.metrics.SuccessfulRequests),
		FailedRequests:     atomic.LoadInt64(&cb.metrics.FailedRequests),
		RejectedRequests:   atomic.LoadInt64(&cb.metrics.RejectedRequests),
		StateChanges:       atomic.LoadInt64(&cb.metrics.StateChanges),
		OpenCount:          atomic.LoadInt64(&cb.metrics.OpenCount),
		HalfOpenCount:      atomic.LoadInt64(&cb.metrics.HalfOpenCount),
		ClosedCount:        atomic.LoadInt64(&cb.metrics.ClosedCount),
		SuccessRate:        cb.metrics.SuccessRate,
		FailureRate:        cb.metrics.FailureRate,
		AverageLatency:     cb.metrics.AverageLatency,
		LastFailure:        cb.metrics.LastFailure,
		LastSuccess:        cb.metrics.LastSuccess,
	}
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(primary *CircuitBreaker, backups []*CircuitBreaker, config *FailoverConfig) *FailoverManager {
	if config == nil {
		config = DefaultFailoverConfig()
	}

	fm := &FailoverManager{
		primary:        primary,
		backups:        backups,
		config:         config,
		currentService: 0, // Start with primary
		metrics:        NewFailoverMetrics(),
		healthChecker:  NewHealthChecker(&HealthCheckConfig{
			Interval:           config.HealthCheckInterval,
			Timeout:            config.HealthCheckTimeout,
			HealthyThreshold:   2,
			UnhealthyThreshold: 3,
		}),
	}

	// Start health checking if auto-recovery is enabled
	if config.AutoRecovery {
		go fm.healthCheckLoop()
	}

	return fm
}

// Execute attempts to execute function with failover
func (fm *FailoverManager) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return fm.ExecuteWithContext(context.Background(), fn)
}

// ExecuteWithContext executes function with context and failover
func (fm *FailoverManager) ExecuteWithContext(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	maxAttempts := 1 + len(fm.backups)
	if fm.config.MaxRetries > 0 && fm.config.MaxRetries < maxAttempts {
		maxAttempts = fm.config.MaxRetries
	}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Get current service
		serviceIndex := atomic.LoadInt32(&fm.currentService)
		var cb *CircuitBreaker

		if serviceIndex == 0 {
			cb = fm.primary
		} else if int(serviceIndex-1) < len(fm.backups) {
			cb = fm.backups[serviceIndex-1]
		} else {
			// Reset to primary if index is out of bounds
			atomic.StoreInt32(&fm.currentService, 0)
			cb = fm.primary
		}

		// Try to execute
		result, err := cb.ExecuteWithContext(ctx, fn)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// If circuit is open, try failover
		if err == ErrCircuitOpen || err == ErrHalfOpenLimitExceeded {
			if attempt < maxAttempts-1 {
				fm.attemptFailover()
				if fm.config.RetryDelay > 0 {
					time.Sleep(fm.config.RetryDelay)
				}
				continue
			}
		}

		break
	}

	// If all services failed, check if we should activate fallback
	if fm.config.EnableFallback && atomic.LoadInt32(&fm.fallbackMode) == 0 {
		atomic.StoreInt32(&fm.fallbackMode, 1)
		atomic.AddInt64(&fm.metrics.FallbackActivations, 1)

		// Implement fallback logic here
		return fm.executeFallback(ctx, fn)
	}

	return nil, fmt.Errorf("all services failed, last error: %w", lastErr)
}

// attemptFailover tries to switch to the next available service
func (fm *FailoverManager) attemptFailover() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	currentIndex := atomic.LoadInt32(&fm.currentService)

	// Try next service
	nextIndex := (currentIndex + 1) % int32(1+len(fm.backups))

	atomic.StoreInt32(&fm.currentService, nextIndex)
	atomic.AddInt64(&fm.metrics.FailoverCount, 1)
	atomic.AddInt64(&fm.metrics.TotalSwitches, 1)

	fm.metrics.mutex.Lock()
	fm.metrics.CurrentServiceIndex = nextIndex
	fm.metrics.mutex.Unlock()
}

// executeFallback executes fallback logic
func (fm *FailoverManager) executeFallback(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Implement fallback logic - this could be:
	// - Return cached data
	// - Return default values
	// - Execute simplified logic
	// For now, return an error indicating fallback mode
	return nil, fmt.Errorf("service in fallback mode")
}

// healthCheckLoop runs periodic health checks for auto-recovery
func (fm *FailoverManager) healthCheckLoop() {
	ticker := time.NewTicker(fm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		fm.performHealthChecks()
		fm.attemptRecovery()
	}
}

// performHealthChecks checks the health of all services
func (fm *FailoverManager) performHealthChecks() {
	// Check primary
	primaryHealthy := fm.isServiceHealthy(fm.primary)

	// Check backups
	backupHealth := make([]bool, len(fm.backups))
	for i, backup := range fm.backups {
		backupHealth[i] = fm.isServiceHealthy(backup)
	}

	// Update metrics
	fm.metrics.mutex.Lock()
	if fm.metrics.ServiceHealthStatus == nil {
		fm.metrics.ServiceHealthStatus = make(map[int]bool)
	}
	fm.metrics.ServiceHealthStatus[0] = primaryHealthy
	for i, healthy := range backupHealth {
		fm.metrics.ServiceHealthStatus[i+1] = healthy
	}
	fm.metrics.mutex.Unlock()
}

// isServiceHealthy checks if a service is healthy
func (fm *FailoverManager) isServiceHealthy(cb *CircuitBreaker) bool {
	state := cb.GetState()
	return state == CircuitClosed || state == CircuitHalfOpen
}

// attemptRecovery tries to recover to a healthier service
func (fm *FailoverManager) attemptRecovery() {
	if !fm.config.AutoRecovery {
		return
	}

	currentIndex := atomic.LoadInt32(&fm.currentService)

	// If not on primary, check if primary is healthy
	if currentIndex != 0 && fm.isServiceHealthy(fm.primary) {
		atomic.StoreInt32(&fm.currentService, 0)
		atomic.StoreInt32(&fm.fallbackMode, 0)
		atomic.AddInt64(&fm.metrics.RecoveryCount, 1)

		fm.metrics.mutex.Lock()
		fm.metrics.CurrentServiceIndex = 0
		fm.metrics.mutex.Unlock()
	}
}

// Helper functions and default configurations

func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:         5,
		FailureRate:              0.5,
		MinimumRequests:          10,
		OpenTimeout:              30 * time.Second,
		HalfOpenTimeout:          5 * time.Second,
		ResetTimeout:             60 * time.Second,
		HalfOpenMaxRequests:      3,
		HalfOpenSuccessThreshold: 2,
		ExponentialBackoff:       true,
		MaxBackoffTime:          5 * time.Minute,
		BackoffMultiplier:       2.0,
		EnableMetrics:           true,
		MetricsWindowSize:       100,
	}
}

func DefaultFailoverConfig() *FailoverConfig {
	return &FailoverConfig{
		Strategy:              FailoverImmediate,
		MaxRetries:            3,
		RetryDelay:            100 * time.Millisecond,
		HealthCheckInterval:   30 * time.Second,
		HealthCheckTimeout:    5 * time.Second,
		AutoRecovery:          true,
		RecoveryCheckInterval: 60 * time.Second,
		EnableFallback:        true,
		FallbackTimeout:       10 * time.Second,
	}
}

func NewCircuitBreakerMetrics(windowSize int) *CircuitBreakerMetrics {
	return &CircuitBreakerMetrics{
		requestWindow: NewRollingWindow(windowSize),
		latencyWindow: NewRollingWindow(windowSize),
	}
}

func NewFailoverMetrics() *FailoverMetrics {
	return &FailoverMetrics{
		ServiceHealthStatus: make(map[int]bool),
	}
}

func NewHealthChecker(config *HealthCheckConfig) *HealthChecker {
	return &HealthChecker{
		config:   config,
		checkers: make(map[int]func() error),
		results:  make(map[int]*HealthCheckResult),
	}
}

func NewRollingWindow(size int) *RollingWindow {
	return &RollingWindow{
		values: make([]float64, size),
		size:   size,
	}
}

func (rw *RollingWindow) Add(value float64) {
	rw.mutex.Lock()
	defer rw.mutex.Unlock()

	rw.values[rw.index] = value
	rw.index = (rw.index + 1) % rw.size
	if !rw.full && rw.index == 0 {
		rw.full = true
	}
}

func (rw *RollingWindow) Average() float64 {
	rw.mutex.RLock()
	defer rw.mutex.RUnlock()

	var sum float64
	count := rw.size
	if !rw.full {
		count = rw.index
	}

	if count == 0 {
		return 0
	}

	for i := 0; i < count; i++ {
		sum += rw.values[i]
	}

	return sum / float64(count)
}

// recordLatency records latency in the metrics
func (cbm *CircuitBreakerMetrics) recordLatency(latency time.Duration) {
	if cbm.latencyWindow != nil {
		cbm.latencyWindow.Add(float64(latency.Nanoseconds()))

		cbm.mutex.Lock()
		cbm.AverageLatency = time.Duration(cbm.latencyWindow.Average())
		cbm.mutex.Unlock()
	}
}