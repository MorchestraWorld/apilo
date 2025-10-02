// Package src provides a memory-bounded cache implementation with GC pressure mitigation
// This addresses the critical memory growth issue (10K entries = 500MB RAM)
package main

import (
	"container/list"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryBoundedCache provides a cache with strict memory limits and GC optimization
type MemoryBoundedCache struct {
	// Core cache components
	mu           sync.RWMutex
	items        map[string]*CacheElement
	lru          *list.List

	// Memory management
	maxMemoryBytes   int64
	currentMemory    int64
	itemCount        int64

	// GC optimization
	gcThreshold      int64    // Memory threshold to trigger GC
	gcRunning        int32    // Atomic flag for GC in progress
	lastGCRun        time.Time
	gcInterval       time.Duration

	// Memory pressure management
	memoryPressure   float64  // 0.0 to 1.0, indicates memory pressure
	evictionRate     float64  // Dynamic eviction rate based on pressure

	// Monitoring and metrics
	metrics          *EnhancedCacheMetrics
	memoryTracker    *MemoryTracker

	// Configuration
	config           *MemoryBoundedConfig
}

// CacheElement represents an enhanced cache element with memory tracking
type CacheElement struct {
	listElement  *list.Element
	key          string
	value        interface{}
	memorySize   int64
	createdAt    time.Time
	lastAccessed time.Time
	accessCount  int64
	ttl          time.Duration
	expiresAt    time.Time
}

// MemoryBoundedConfig configures the memory-bounded cache
type MemoryBoundedConfig struct {
	MaxMemoryMB          int64         `yaml:"max_memory_mb"`
	GCThresholdPercent   float64       `yaml:"gc_threshold_percent"`
	GCInterval           time.Duration `yaml:"gc_interval"`
	EvictionBatchSize    int           `yaml:"eviction_batch_size"`
	MemoryCheckInterval  time.Duration `yaml:"memory_check_interval"`
	EnableGCOptimization bool          `yaml:"enable_gc_optimization"`
	EnableMemoryTracker  bool          `yaml:"enable_memory_tracker"`
	PressureThreshold    float64       `yaml:"pressure_threshold"`
}

// EnhancedCacheMetrics tracks comprehensive cache metrics
type EnhancedCacheMetrics struct {
	// Memory metrics
	currentMemoryBytes   int64
	peakMemoryBytes      int64
	memoryPressureValue  float64
	gcRunCount           int64
	evictionCount        int64

	// Performance metrics
	hitCount             int64
	missCount            int64
	setCount             int64
	deleteCount          int64

	// GC metrics
	gcDuration           time.Duration
	lastGCTime           time.Time
	memoryFreedBytes     int64

	// Access patterns
	avgAccessCount       float64
	hotKeyCount          int64
	coldKeyEvictions     int64

	mutex                sync.RWMutex
}

// MemoryTracker provides advanced memory usage tracking
type MemoryTracker struct {
	samples            []MemorySample
	maxSamples         int
	currentIndex       int
	mutex              sync.RWMutex

	// Memory trend analysis
	trendDirection     MemoryTrend
	averageUsage       int64
	peakUsage          int64
	growthRate         float64
}

// MemorySample represents a memory usage sample
type MemorySample struct {
	timestamp     time.Time
	memoryBytes   int64
	itemCount     int64
	gcRunning     bool
}

// MemoryTrend indicates memory usage trend
type MemoryTrend int

const (
	TrendStable MemoryTrend = iota
	TrendIncreasing
	TrendDecreasing
	TrendVolatile
)

// NewMemoryBoundedCache creates a new memory-bounded cache
func NewMemoryBoundedCache(config *MemoryBoundedConfig) *MemoryBoundedCache {
	if config == nil {
		config = DefaultMemoryBoundedConfig()
	}

	cache := &MemoryBoundedCache{
		items:            make(map[string]*CacheElement),
		lru:              list.New(),
		maxMemoryBytes:   config.MaxMemoryMB * 1024 * 1024, // Convert MB to bytes
		gcThreshold:      int64(float64(config.MaxMemoryMB*1024*1024) * config.GCThresholdPercent),
		gcInterval:       config.GCInterval,
		config:           config,
		metrics:          NewEnhancedCacheMetrics(),
		memoryTracker:    NewMemoryTracker(1000), // Keep 1000 samples
	}

	// Start background memory management
	if config.EnableMemoryTracker {
		go cache.memoryManagementLoop()
	}

	if config.EnableGCOptimization {
		go cache.gcOptimizationLoop()
	}

	return cache
}

// Get retrieves an item from cache with memory pressure awareness
func (mbc *MemoryBoundedCache) Get(key string) (interface{}, bool) {
	mbc.mu.Lock()
	defer mbc.mu.Unlock()

	element, exists := mbc.items[key]
	if !exists {
		atomic.AddInt64(&mbc.metrics.missCount, 1)
		return nil, false
	}

	// Check expiration
	if time.Now().After(element.expiresAt) {
		mbc.removeElementUnsafe(element)
		atomic.AddInt64(&mbc.metrics.missCount, 1)
		return nil, false
	}

	// Update access patterns
	element.lastAccessed = time.Now()
	atomic.AddInt64(&element.accessCount, 1)

	// Move to front (LRU)
	mbc.lru.MoveToFront(element.listElement)

	atomic.AddInt64(&mbc.metrics.hitCount, 1)
	return element.value, true
}

// Set stores an item in cache with memory management
func (mbc *MemoryBoundedCache) Set(key string, value interface{}, ttl time.Duration) error {
	memorySize := mbc.calculateMemorySize(key, value)

	// Check if this single item would exceed memory limit
	if memorySize > mbc.maxMemoryBytes {
		return ErrItemTooLarge
	}

	mbc.mu.Lock()
	defer mbc.mu.Unlock()

	// Remove existing item if present
	if existing, exists := mbc.items[key]; exists {
		mbc.removeElementUnsafe(existing)
	}

	// Ensure we have space for the new item
	mbc.ensureMemorySpaceUnsafe(memorySize)

	// Create new cache element
	now := time.Now()
	element := &CacheElement{
		key:          key,
		value:        value,
		memorySize:   memorySize,
		createdAt:    now,
		lastAccessed: now,
		accessCount:  0,
		ttl:          ttl,
		expiresAt:    now.Add(ttl),
	}

	// Add to cache
	listElement := mbc.lru.PushFront(element)
	element.listElement = listElement
	mbc.items[key] = element

	// Update memory tracking
	atomic.AddInt64(&mbc.currentMemory, memorySize)
	atomic.AddInt64(&mbc.itemCount, 1)
	atomic.AddInt64(&mbc.metrics.setCount, 1)

	// Update memory pressure
	mbc.updateMemoryPressure()

	return nil
}

// ensureMemorySpaceUnsafe ensures sufficient memory space by evicting items
func (mbc *MemoryBoundedCache) ensureMemorySpaceUnsafe(requiredMemory int64) {
	neededMemory := mbc.currentMemory + requiredMemory - mbc.maxMemoryBytes

	if neededMemory <= 0 {
		return
	}

	// Increase eviction aggressiveness based on memory pressure
	batchSize := mbc.config.EvictionBatchSize
	if mbc.memoryPressure > mbc.config.PressureThreshold {
		batchSize = int(float64(batchSize) * (1.0 + mbc.memoryPressure))
	}

	evicted := 0
	freedMemory := int64(0)

	// Evict items starting from the least recently used
	for mbc.lru.Len() > 0 && freedMemory < neededMemory && evicted < batchSize*2 {
		oldest := mbc.lru.Back()
		if oldest == nil {
			break
		}

		element := oldest.Value.(*CacheElement)
		freedMemory += element.memorySize
		mbc.removeElementUnsafe(element)
		evicted++
	}

	atomic.AddInt64(&mbc.metrics.evictionCount, int64(evicted))
	atomic.AddInt64(&mbc.metrics.memoryFreedBytes, freedMemory)
}

// removeElementUnsafe removes an element from cache (must hold lock)
func (mbc *MemoryBoundedCache) removeElementUnsafe(element *CacheElement) {
	if element.listElement != nil {
		mbc.lru.Remove(element.listElement)
	}
	delete(mbc.items, element.key)

	atomic.AddInt64(&mbc.currentMemory, -element.memorySize)
	atomic.AddInt64(&mbc.itemCount, -1)
}

// updateMemoryPressure calculates current memory pressure (0.0 to 1.0)
func (mbc *MemoryBoundedCache) updateMemoryPressure() {
	if mbc.maxMemoryBytes == 0 {
		mbc.memoryPressure = 0.0
		return
	}

	pressure := float64(mbc.currentMemory) / float64(mbc.maxMemoryBytes)

	// Apply exponential curve for pressure sensitivity
	if pressure > 0.8 {
		pressure = 0.8 + (pressure-0.8)*2.0 // Accelerate pressure above 80%
	}

	mbc.memoryPressure = pressure
	mbc.metrics.mutex.Lock()
	mbc.metrics.memoryPressureValue = pressure
	mbc.metrics.mutex.Unlock()
}

// memoryManagementLoop runs background memory management tasks
func (mbc *MemoryBoundedCache) memoryManagementLoop() {
	ticker := time.NewTicker(mbc.config.MemoryCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		mbc.performMemoryCheck()
		mbc.recordMemorySample()
	}
}

// performMemoryCheck performs periodic memory health checks
func (mbc *MemoryBoundedCache) performMemoryCheck() {
	mbc.mu.RLock()
	currentMemory := mbc.currentMemory
	pressure := mbc.memoryPressure
	mbc.mu.RUnlock()

	// Check for memory pressure
	if pressure > mbc.config.PressureThreshold {
		// Trigger aggressive cleanup
		mbc.performEmergencyCleanup()
	}

	// Update peak memory tracking
	if currentMemory > mbc.metrics.peakMemoryBytes {
		atomic.StoreInt64(&mbc.metrics.peakMemoryBytes, currentMemory)
	}

	// Check for potential memory leaks
	if pressure > 0.95 && time.Since(mbc.lastGCRun) > mbc.gcInterval {
		mbc.triggerGC()
	}
}

// performEmergencyCleanup performs emergency memory cleanup
func (mbc *MemoryBoundedCache) performEmergencyCleanup() {
	mbc.mu.Lock()
	defer mbc.mu.Unlock()

	now := time.Now()
	expired := 0
	freedMemory := int64(0)

	// First pass: Remove all expired items
	for _, element := range mbc.items {
		if now.After(element.expiresAt) {
			freedMemory += element.memorySize
			mbc.removeElementUnsafe(element)
			expired++
		}
	}

	// Second pass: Remove cold items (low access count)
	if mbc.memoryPressure > 0.9 {
		coldThreshold := mbc.calculateColdThreshold()
		for _, element := range mbc.items {
			if element.accessCount < coldThreshold {
				freedMemory += element.memorySize
				mbc.removeElementUnsafe(element)
				atomic.AddInt64(&mbc.metrics.coldKeyEvictions, 1)
			}
		}
	}

	atomic.AddInt64(&mbc.metrics.memoryFreedBytes, freedMemory)
}

// calculateColdThreshold calculates threshold for identifying cold keys
func (mbc *MemoryBoundedCache) calculateColdThreshold() int64 {
	totalAccess := int64(0)
	count := int64(0)

	for _, element := range mbc.items {
		totalAccess += element.accessCount
		count++
	}

	if count == 0 {
		return 0
	}

	avgAccess := totalAccess / count
	return avgAccess / 4 // Consider items with <25% of average access as cold
}

// gcOptimizationLoop runs GC optimization in background
func (mbc *MemoryBoundedCache) gcOptimizationLoop() {
	ticker := time.NewTicker(mbc.gcInterval)
	defer ticker.Stop()

	for range ticker.C {
		if mbc.shouldRunGC() {
			mbc.triggerGC()
		}
	}
}

// shouldRunGC determines if GC should be triggered
func (mbc *MemoryBoundedCache) shouldRunGC() bool {
	// Don't run GC if already running
	if atomic.LoadInt32(&mbc.gcRunning) == 1 {
		return false
	}

	// Run GC if memory usage exceeds threshold
	return mbc.currentMemory > mbc.gcThreshold
}

// triggerGC triggers garbage collection with optimization
func (mbc *MemoryBoundedCache) triggerGC() {
	if !atomic.CompareAndSwapInt32(&mbc.gcRunning, 0, 1) {
		return // GC already running
	}

	start := time.Now()

	// Get memory stats before GC
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Force garbage collection
	runtime.GC()

	// Get memory stats after GC
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	duration := time.Since(start)
	freedBytes := int64(memBefore.Alloc - memAfter.Alloc)

	// Update metrics
	atomic.AddInt64(&mbc.metrics.gcRunCount, 1)
	mbc.metrics.mutex.Lock()
	mbc.metrics.gcDuration = duration
	mbc.metrics.lastGCTime = time.Now()
	mbc.metrics.mutex.Unlock()
	atomic.AddInt64(&mbc.metrics.memoryFreedBytes, freedBytes)

	mbc.lastGCRun = time.Now()
	atomic.StoreInt32(&mbc.gcRunning, 0)
}

// calculateMemorySize estimates memory usage of a cache entry
func (mbc *MemoryBoundedCache) calculateMemorySize(key string, value interface{}) int64 {
	// Base size for cache element structure
	size := int64(200) // Rough estimate for CacheElement struct

	// Add key size
	size += int64(len(key))

	// Estimate value size based on type
	switch v := value.(type) {
	case string:
		size += int64(len(v))
	case []byte:
		size += int64(len(v))
	case int, int32, int64:
		size += 8
	case float32, float64:
		size += 8
	case bool:
		size += 1
	default:
		// For complex types, use a rough estimate
		size += 1024 // 1KB default estimate
	}

	return size
}

// recordMemorySample records a memory usage sample for trend analysis
func (mbc *MemoryBoundedCache) recordMemorySample() {
	sample := MemorySample{
		timestamp:   time.Now(),
		memoryBytes: mbc.currentMemory,
		itemCount:   mbc.itemCount,
		gcRunning:   atomic.LoadInt32(&mbc.gcRunning) == 1,
	}

	mbc.memoryTracker.AddSample(sample)
}

// GetMemoryStats returns comprehensive memory statistics
func (mbc *MemoryBoundedCache) GetMemoryStats() MemoryStats {
	mbc.mu.RLock()
	defer mbc.mu.RUnlock()

	return MemoryStats{
		CurrentMemoryBytes:  mbc.currentMemory,
		MaxMemoryBytes:      mbc.maxMemoryBytes,
		MemoryPressure:      mbc.memoryPressure,
		ItemCount:           mbc.itemCount,
		MemoryUtilization:   float64(mbc.currentMemory) / float64(mbc.maxMemoryBytes),
		GCRunCount:          atomic.LoadInt64(&mbc.metrics.gcRunCount),
		EvictionCount:       atomic.LoadInt64(&mbc.metrics.evictionCount),
		MemoryFreedBytes:    atomic.LoadInt64(&mbc.metrics.memoryFreedBytes),
		HitRatio:            mbc.calculateHitRatio(),
		Trend:               mbc.memoryTracker.GetTrend(),
	}
}

// MemoryStats represents comprehensive memory statistics
type MemoryStats struct {
	CurrentMemoryBytes  int64       `json:"current_memory_bytes"`
	MaxMemoryBytes      int64       `json:"max_memory_bytes"`
	MemoryPressure      float64     `json:"memory_pressure"`
	ItemCount           int64       `json:"item_count"`
	MemoryUtilization   float64     `json:"memory_utilization"`
	GCRunCount          int64       `json:"gc_run_count"`
	EvictionCount       int64       `json:"eviction_count"`
	MemoryFreedBytes    int64       `json:"memory_freed_bytes"`
	HitRatio            float64     `json:"hit_ratio"`
	Trend               MemoryTrend `json:"trend"`
}

// calculateHitRatio calculates cache hit ratio
func (mbc *MemoryBoundedCache) calculateHitRatio() float64 {
	hits := atomic.LoadInt64(&mbc.metrics.hitCount)
	misses := atomic.LoadInt64(&mbc.metrics.missCount)
	total := hits + misses

	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total)
}

// Helper functions and error definitions

var (
	ErrItemTooLarge = fmt.Errorf("item size exceeds maximum memory limit")
)

// DefaultMemoryBoundedConfig returns default configuration
func DefaultMemoryBoundedConfig() *MemoryBoundedConfig {
	return &MemoryBoundedConfig{
		MaxMemoryMB:          100,
		GCThresholdPercent:   0.8,
		GCInterval:           5 * time.Minute,
		EvictionBatchSize:    50,
		MemoryCheckInterval:  30 * time.Second,
		EnableGCOptimization: true,
		EnableMemoryTracker:  true,
		PressureThreshold:    0.85,
	}
}

// NewEnhancedCacheMetrics creates new enhanced cache metrics
func NewEnhancedCacheMetrics() *EnhancedCacheMetrics {
	return &EnhancedCacheMetrics{}
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker(maxSamples int) *MemoryTracker {
	return &MemoryTracker{
		samples:    make([]MemorySample, maxSamples),
		maxSamples: maxSamples,
	}
}

// AddSample adds a memory sample to the tracker
func (mt *MemoryTracker) AddSample(sample MemorySample) {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	mt.samples[mt.currentIndex] = sample
	mt.currentIndex = (mt.currentIndex + 1) % mt.maxSamples

	// Update trend analysis
	mt.updateTrendAnalysis()
}

// GetTrend returns the current memory trend
func (mt *MemoryTracker) GetTrend() MemoryTrend {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()
	return mt.trendDirection
}

// updateTrendAnalysis analyzes memory usage trends
func (mt *MemoryTracker) updateTrendAnalysis() {
	// Simple trend analysis based on recent samples
	if mt.currentIndex < 10 {
		mt.trendDirection = TrendStable
		return
	}

	recentSamples := 10
	increasing := 0
	decreasing := 0

	for i := 1; i < recentSamples; i++ {
		prevIdx := (mt.currentIndex - i - 1 + mt.maxSamples) % mt.maxSamples
		currIdx := (mt.currentIndex - i + mt.maxSamples) % mt.maxSamples

		if mt.samples[currIdx].memoryBytes > mt.samples[prevIdx].memoryBytes {
			increasing++
		} else if mt.samples[currIdx].memoryBytes < mt.samples[prevIdx].memoryBytes {
			decreasing++
		}
	}

	if increasing > decreasing*2 {
		mt.trendDirection = TrendIncreasing
	} else if decreasing > increasing*2 {
		mt.trendDirection = TrendDecreasing
	} else if increasing > 3 && decreasing > 3 {
		mt.trendDirection = TrendVolatile
	} else {
		mt.trendDirection = TrendStable
	}
}