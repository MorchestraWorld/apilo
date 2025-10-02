// Package src provides comprehensive tests for memory-bounded cache
package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestMemoryBoundedCacheBasicOperations tests basic cache operations
func TestMemoryBoundedCacheBasicOperations(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          1, // 1MB limit
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    10,
		MemoryCheckInterval:  time.Second,
		EnableGCOptimization: false, // Disable for predictable testing
		EnableMemoryTracker:  false,
		PressureThreshold:    0.85,
	}

	cache := NewMemoryBoundedCache(config)

	// Test Set and Get
	err := cache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Errorf("Failed to set key1: %v", err)
	}

	value, found := cache.Get("key1")
	if !found {
		t.Error("Key1 not found")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Test Get non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Non-existent key should not be found")
	}
}

// TestMemoryBoundedCacheMemoryLimit tests memory limit enforcement
func TestMemoryBoundedCacheMemoryLimit(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          1, // 1MB limit
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    50,
		MemoryCheckInterval:  time.Second,
		EnableGCOptimization: false,
		EnableMemoryTracker:  false,
		PressureThreshold:    0.85,
	}

	cache := NewMemoryBoundedCache(config)
	maxMemory := config.MaxMemoryMB * 1024 * 1024

	// Fill cache beyond memory limit
	largeValue := make([]byte, 10*1024) // 10KB per entry

	for i := 0; i < 200; i++ { // Attempt to store 2MB of data
		key := fmt.Sprintf("key%d", i)
		err := cache.Set(key, largeValue, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}

	// Verify memory limit is respected
	stats := cache.GetMemoryStats()
	if stats.CurrentMemoryBytes > maxMemory {
		t.Errorf("Memory usage %d exceeds limit %d", stats.CurrentMemoryBytes, maxMemory)
	}

	// Verify cache still functions after evictions
	testKey := "test_after_eviction"
	err := cache.Set(testKey, "test_value", 5*time.Minute)
	if err != nil {
		t.Errorf("Failed to set test key after evictions: %v", err)
	}

	value, found := cache.Get(testKey)
	if !found || value != "test_value" {
		t.Error("Cache not functioning correctly after evictions")
	}
}

// TestMemoryBoundedCacheExpiration tests TTL expiration
func TestMemoryBoundedCacheExpiration(t *testing.T) {
	config := DefaultMemoryBoundedConfig()
	config.EnableGCOptimization = false
	config.EnableMemoryTracker = false

	cache := NewMemoryBoundedCache(config)

	// Set item with short TTL
	err := cache.Set("expire_key", "expire_value", 100*time.Millisecond)
	if err != nil {
		t.Errorf("Failed to set expire_key: %v", err)
	}

	// Verify item exists before expiration
	value, found := cache.Get("expire_key")
	if !found || value != "expire_value" {
		t.Error("Item should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Verify item is expired
	_, found = cache.Get("expire_key")
	if found {
		t.Error("Item should be expired")
	}
}

// TestMemoryBoundedCacheConcurrency tests concurrent access
func TestMemoryBoundedCacheConcurrency(t *testing.T) {
	config := DefaultMemoryBoundedConfig()
	config.MaxMemoryMB = 10 // 10MB for concurrent testing
	config.EnableGCOptimization = false
	config.EnableMemoryTracker = false

	cache := NewMemoryBoundedCache(config)

	const numGoroutines = 100
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes and reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// Set
				err := cache.Set(key, value, 5*time.Minute)
				if err != nil {
					t.Errorf("Failed to set %s: %v", key, err)
					return
				}

				// Get
				retrievedValue, found := cache.Get(key)
				if found && retrievedValue != value {
					t.Errorf("Value mismatch for %s: expected %s, got %v", key, value, retrievedValue)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is still functional after concurrent operations
	stats := cache.GetMemoryStats()
	if stats.CurrentMemoryBytes > config.MaxMemoryMB*1024*1024 {
		t.Errorf("Memory limit exceeded after concurrent operations: %d", stats.CurrentMemoryBytes)
	}
}

// TestMemoryBoundedCacheGCOptimization tests GC optimization features
func TestMemoryBoundedCacheGCOptimization(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          5, // 5MB limit
		GCThresholdPercent:   0.7,
		GCInterval:           100 * time.Millisecond,
		EvictionBatchSize:    20,
		MemoryCheckInterval:  50 * time.Millisecond,
		EnableGCOptimization: true,
		EnableMemoryTracker:  true,
		PressureThreshold:    0.8,
	}

	cache := NewMemoryBoundedCache(config)

	// Fill cache to trigger GC
	largeValue := make([]byte, 50*1024) // 50KB per entry

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("gc_key%d", i)
		err := cache.Set(key, largeValue, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}

	// Wait for GC optimization to kick in
	time.Sleep(300 * time.Millisecond)

	stats := cache.GetMemoryStats()

	// Verify GC was triggered
	if stats.GCRunCount == 0 {
		t.Error("GC should have been triggered")
	}

	// Verify memory pressure is being tracked
	if stats.MemoryPressure < 0.0 || stats.MemoryPressure > 1.0 {
		t.Errorf("Invalid memory pressure value: %f", stats.MemoryPressure)
	}

	// Verify memory utilization is reasonable
	if stats.MemoryUtilization > 1.0 {
		t.Errorf("Memory utilization should not exceed 1.0: %f", stats.MemoryUtilization)
	}
}

// TestMemoryBoundedCacheMemoryPressure tests memory pressure handling
func TestMemoryBoundedCacheMemoryPressure(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          2, // 2MB limit for testing pressure
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    10,
		MemoryCheckInterval:  100 * time.Millisecond,
		EnableGCOptimization: false,
		EnableMemoryTracker:  true,
		PressureThreshold:    0.8,
	}

	cache := NewMemoryBoundedCache(config)

	// Gradually fill cache and monitor pressure
	largeValue := make([]byte, 100*1024) // 100KB per entry

	pressureReadings := make([]float64, 0)

	for i := 0; i < 25; i++ { // Up to 2.5MB of data
		key := fmt.Sprintf("pressure_key%d", i)
		err := cache.Set(key, largeValue, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}

		stats := cache.GetMemoryStats()
		pressureReadings = append(pressureReadings, stats.MemoryPressure)
	}

	// Verify pressure increases as memory fills
	if len(pressureReadings) < 2 {
		t.Error("Not enough pressure readings")
		return
	}

	// Check that pressure generally increases (allowing for some evictions)
	maxPressure := 0.0
	for _, pressure := range pressureReadings {
		if pressure > maxPressure {
			maxPressure = pressure
		}
	}

	if maxPressure < 0.5 {
		t.Errorf("Expected significant memory pressure, got max: %f", maxPressure)
	}
}

// TestMemoryBoundedCacheEvictionStrategy tests eviction strategies
func TestMemoryBoundedCacheEvictionStrategy(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          1, // 1MB limit
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    5,
		MemoryCheckInterval:  time.Second,
		EnableGCOptimization: false,
		EnableMemoryTracker:  false,
		PressureThreshold:    0.85,
	}

	cache := NewMemoryBoundedCache(config)

	// Add items with different access patterns
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("evict_key%d", i)
		value := make([]byte, 60*1024) // 60KB per entry
		err := cache.Set(key, value, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}

	// Access some keys to make them "hot"
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("evict_key%d", i)
		cache.Get(key)
		cache.Get(key) // Access twice to increase access count
		cache.Get(key)
	}

	// Add more items to trigger eviction
	for i := 20; i < 30; i++ {
		key := fmt.Sprintf("evict_key%d", i)
		value := make([]byte, 60*1024) // 60KB per entry
		err := cache.Set(key, value, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}

	// Verify that recently accessed keys are more likely to remain
	hotKeysRemaining := 0
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("evict_key%d", i)
		if _, found := cache.Get(key); found {
			hotKeysRemaining++
		}
	}

	coldKeysRemaining := 0
	for i := 10; i < 15; i++ { // Check middle range keys that weren't accessed
		key := fmt.Sprintf("evict_key%d", i)
		if _, found := cache.Get(key); found {
			coldKeysRemaining++
		}
	}

	// Hot keys should be more likely to remain than cold keys
	if hotKeysRemaining < coldKeysRemaining {
		t.Errorf("Eviction strategy not working correctly: hot=%d, cold=%d", hotKeysRemaining, coldKeysRemaining)
	}
}

// TestMemoryBoundedCacheStats tests statistics reporting
func TestMemoryBoundedCacheStats(t *testing.T) {
	config := DefaultMemoryBoundedConfig()
	config.EnableGCOptimization = false
	config.EnableMemoryTracker = true

	cache := NewMemoryBoundedCache(config)

	// Perform various operations
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("stats_key%d", i)
		value := fmt.Sprintf("stats_value%d", i)
		err := cache.Set(key, value, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}

	// Perform some gets (hits and misses)
	cache.Get("stats_key1")   // Hit
	cache.Get("stats_key2")   // Hit
	cache.Get("nonexistent1") // Miss
	cache.Get("nonexistent2") // Miss

	stats := cache.GetMemoryStats()

	// Verify stats are reasonable
	if stats.CurrentMemoryBytes <= 0 {
		t.Error("Current memory should be positive")
	}

	if stats.ItemCount != 10 {
		t.Errorf("Expected 10 items, got %d", stats.ItemCount)
	}

	if stats.MemoryUtilization < 0 || stats.MemoryUtilization > 1 {
		t.Errorf("Invalid memory utilization: %f", stats.MemoryUtilization)
	}

	if stats.HitRatio < 0 || stats.HitRatio > 1 {
		t.Errorf("Invalid hit ratio: %f", stats.HitRatio)
	}

	// Hit ratio should be 0.5 (2 hits out of 4 total requests)
	expectedHitRatio := 0.5
	if stats.HitRatio != expectedHitRatio {
		t.Errorf("Expected hit ratio %f, got %f", expectedHitRatio, stats.HitRatio)
	}
}

// TestMemoryBoundedCacheItemTooLarge tests handling of oversized items
func TestMemoryBoundedCacheItemTooLarge(t *testing.T) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          1, // 1MB limit
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    10,
		MemoryCheckInterval:  time.Second,
		EnableGCOptimization: false,
		EnableMemoryTracker:  false,
		PressureThreshold:    0.85,
	}

	cache := NewMemoryBoundedCache(config)

	// Try to set an item larger than the entire cache
	largeValue := make([]byte, 2*1024*1024) // 2MB value for 1MB cache
	err := cache.Set("large_key", largeValue, 5*time.Minute)

	if err != ErrItemTooLarge {
		t.Errorf("Expected ErrItemTooLarge, got %v", err)
	}

	// Verify cache is still functional
	err = cache.Set("normal_key", "normal_value", 5*time.Minute)
	if err != nil {
		t.Errorf("Failed to set normal key after large item rejection: %v", err)
	}

	value, found := cache.Get("normal_key")
	if !found || value != "normal_value" {
		t.Error("Cache should still function after rejecting large item")
	}
}

// BenchmarkMemoryBoundedCacheSet benchmarks cache set operations
func BenchmarkMemoryBoundedCacheSet(b *testing.B) {
	config := DefaultMemoryBoundedConfig()
	config.MaxMemoryMB = 100            // Large cache for benchmarking
	config.EnableGCOptimization = false // Disable for consistent benchmarking
	config.EnableMemoryTracker = false

	cache := NewMemoryBoundedCache(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			value := fmt.Sprintf("bench_value_%d", i)
			cache.Set(key, value, 5*time.Minute)
			i++
		}
	})
}

// BenchmarkMemoryBoundedCacheGet benchmarks cache get operations
func BenchmarkMemoryBoundedCacheGet(b *testing.B) {
	config := DefaultMemoryBoundedConfig()
	config.MaxMemoryMB = 100
	config.EnableGCOptimization = false
	config.EnableMemoryTracker = false

	cache := NewMemoryBoundedCache(config)

	// Pre-populate cache
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)
		cache.Set(key, value, 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%10000)
			cache.Get(key)
			i++
		}
	})
}

// BenchmarkMemoryBoundedCacheMemoryPressure benchmarks performance under memory pressure
func BenchmarkMemoryBoundedCacheMemoryPressure(b *testing.B) {
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          10, // Small cache to create pressure
		GCThresholdPercent:   0.8,
		GCInterval:           time.Minute,
		EvictionBatchSize:    50,
		MemoryCheckInterval:  time.Second,
		EnableGCOptimization: true, // Enable to test real-world scenario
		EnableMemoryTracker:  true,
		PressureThreshold:    0.85,
	}

	cache := NewMemoryBoundedCache(config)

	// Pre-populate to create memory pressure
	largeValue := make([]byte, 10*1024) // 10KB per entry
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("pressure_key_%d", i)
		cache.Set(key, largeValue, 5*time.Minute)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_pressure_key_%d", i)
			cache.Set(key, largeValue, 5*time.Minute)

			// Mix in some gets
			if i%4 == 0 {
				getKey := fmt.Sprintf("pressure_key_%d", i%500)
				cache.Get(getKey)
			}
			i++
		}
	})
}

// TestMemoryBoundedCacheMemoryTracker tests memory trend tracking
func TestMemoryBoundedCacheMemoryTracker(t *testing.T) {
	config := DefaultMemoryBoundedConfig()
	config.EnableMemoryTracker = true
	config.MemoryCheckInterval = 10 * time.Millisecond

	cache := NewMemoryBoundedCache(config)

	// Add items gradually to create trend
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("trend_key%d", i)
		value := make([]byte, 10*1024) // 10KB per entry
		err := cache.Set(key, value, 5*time.Minute)
		if err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
		time.Sleep(15 * time.Millisecond) // Allow trend tracking
	}

	stats := cache.GetMemoryStats()

	// Verify trend is being tracked
	if stats.Trend == TrendStable && stats.CurrentMemoryBytes > 100*1024 {
		// With significant memory usage, we should see some trend
		t.Error("Expected non-stable trend with significant memory usage")
	}
}

// Example usage function to demonstrate the API
func ExampleMemoryBoundedCache() {
	// Create configuration
	config := &MemoryBoundedConfig{
		MaxMemoryMB:          50,  // 50MB maximum
		GCThresholdPercent:   0.8, // Trigger GC at 80% memory usage
		GCInterval:           5 * time.Minute,
		EvictionBatchSize:    25,
		MemoryCheckInterval:  30 * time.Second,
		EnableGCOptimization: true,
		EnableMemoryTracker:  true,
		PressureThreshold:    0.85,
	}

	// Create cache
	cache := NewMemoryBoundedCache(config)

	// Store data
	err := cache.Set("user:123", "John Doe", 1*time.Hour)
	if err != nil {
		fmt.Printf("Error storing data: %v\n", err)
		return
	}

	// Retrieve data
	value, found := cache.Get("user:123")
	if found {
		fmt.Printf("Found user: %s\n", value)
	}

	// Check memory statistics
	stats := cache.GetMemoryStats()
	fmt.Printf("Memory usage: %.2f MB (%.1f%% of limit)\n",
		float64(stats.CurrentMemoryBytes)/(1024*1024),
		stats.MemoryUtilization*100)
	fmt.Printf("Cache hit ratio: %.2f%%\n", stats.HitRatio*100)
	fmt.Printf("Memory pressure: %.2f\n", stats.MemoryPressure)

	// Output:
	// Found user: John Doe
	// Memory usage: 0.00 MB (0.1% of limit)
	// Cache hit ratio: 100.00%
	// Memory pressure: 0.00
}
