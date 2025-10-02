package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestCacheBasicOperations tests basic cache get/put/delete operations
func TestCacheBasicOperations(t *testing.T) {
	cache := NewLRUCache(100, 10)

	// Test Put and Get
	key := "test_key"
	entry := &CacheEntry{
		Key:          key,
		Value:        []byte("test_value"),
		StatusCode:   200,
		Size:         10,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		TTL:          5 * time.Minute,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}

	err := cache.Put(key, entry)
	if err != nil {
		t.Errorf("Put failed: %v", err)
	}

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Entry not found after Put")
	}

	if string(retrieved.Value) != "test_value" {
		t.Errorf("Retrieved value mismatch: got %s, want test_value", string(retrieved.Value))
	}

	// Test Delete
	deleted := cache.Delete(key)
	if !deleted {
		t.Error("Delete failed")
	}

	_, found = cache.Get(key)
	if found {
		t.Error("Entry still found after Delete")
	}
}

// TestCacheLRUEviction tests LRU eviction policy
func TestCacheLRUEviction(t *testing.T) {
	cache := NewLRUCache(3, 10) // Capacity of 3 entries

	// Add 3 entries
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte(fmt.Sprintf("value_%d", i)),
			StatusCode:   200,
			Size:         10,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}

	// Access key_0 to make it most recently used
	cache.Get("key_0")

	// Add a new entry, should evict key_1 (least recently used)
	newEntry := &CacheEntry{
		Key:          "key_3",
		Value:        []byte("value_3"),
		StatusCode:   200,
		Size:         10,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		TTL:          5 * time.Minute,
		ExpiresAt:    time.Now().Add(5 * time.Minute),
	}
	cache.Put("key_3", newEntry)

	// key_0 should still be present (accessed recently)
	if _, found := cache.Get("key_0"); !found {
		t.Error("key_0 should still be in cache")
	}

	// key_1 should be evicted
	if _, found := cache.Get("key_1"); found {
		t.Error("key_1 should have been evicted")
	}

	// key_3 should be present
	if _, found := cache.Get("key_3"); !found {
		t.Error("key_3 should be in cache")
	}
}

// TestCacheExpiration tests TTL-based expiration
func TestCacheExpiration(t *testing.T) {
	cache := NewLRUCache(100, 10)

	key := "expiring_key"
	entry := &CacheEntry{
		Key:          key,
		Value:        []byte("expiring_value"),
		StatusCode:   200,
		Size:         14,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		TTL:          100 * time.Millisecond,
		ExpiresAt:    time.Now().Add(100 * time.Millisecond),
	}

	cache.Put(key, entry)

	// Should be retrievable immediately
	if _, found := cache.Get(key); !found {
		t.Error("Entry should be found immediately after Put")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	if _, found := cache.Get(key); found {
		t.Error("Entry should have expired")
	}
}

// TestCacheConcurrency tests thread-safe concurrent operations
func TestCacheConcurrency(t *testing.T) {
	cache := NewLRUCache(1000, 100)
	numGoroutines := 50
	operationsPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				entry := &CacheEntry{
					Key:          key,
					Value:        []byte(fmt.Sprintf("value_%d_%d", id, j)),
					StatusCode:   200,
					Size:         20,
					CreatedAt:    time.Now(),
					LastAccessed: time.Now(),
					TTL:          5 * time.Minute,
					ExpiresAt:    time.Now().Add(5 * time.Minute),
				}
				cache.Put(key, entry)
			}
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and cache is in valid state
	size := cache.Size()
	if size <= 0 || size > cache.Capacity() {
		t.Errorf("Invalid cache size after concurrent operations: %d", size)
	}
}

// TestCacheMetrics tests metrics tracking
func TestCacheMetrics(t *testing.T) {
	cache := NewLRUCache(100, 10)

	// Add some entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte(fmt.Sprintf("value_%d", i)),
			StatusCode:   200,
			Size:         10,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	// Generate some hits
	for i := 0; i < 5; i++ {
		cache.Get(fmt.Sprintf("key_%d", i))
	}

	// Generate some misses
	for i := 10; i < 15; i++ {
		cache.Get(fmt.Sprintf("key_%d", i))
	}

	metrics := cache.GetMetrics()

	if metrics.TotalGets() != 10 {
		t.Errorf("Expected 10 total gets, got %d", metrics.TotalGets())
	}

	if metrics.TotalHits() != 5 {
		t.Errorf("Expected 5 hits, got %d", metrics.TotalHits())
	}

	if metrics.TotalMisses() != 5 {
		t.Errorf("Expected 5 misses, got %d", metrics.TotalMisses())
	}

	hitRatio := metrics.HitRatio()
	expectedRatio := 0.5
	if hitRatio != expectedRatio {
		t.Errorf("Expected hit ratio %.2f, got %.2f", expectedRatio, hitRatio)
	}
}

// TestCachePolicy tests different caching policies
func TestCachePolicy(t *testing.T) {
	// Test Default Policy
	defaultPolicy := NewDefaultPolicy()

	if !defaultPolicy.ShouldCache(200, 1024, nil) {
		t.Error("Should cache 200 OK responses")
	}

	if defaultPolicy.ShouldCache(500, 1024, nil) {
		t.Error("Should not cache 500 server errors")
	}

	if defaultPolicy.ShouldCache(200, 20*1024*1024, nil) {
		t.Error("Should not cache responses larger than max size")
	}

	// Test TTL computation
	entry := &CacheEntry{
		AccessCount: 50,
	}

	pattern := &AccessPattern{
		AccessCount: 50,
		Volatility:  0.2, // Stable pattern
	}

	ttl := defaultPolicy.ComputeTTL(entry, pattern)
	if ttl < defaultPolicy.MinTTL || ttl > defaultPolicy.MaxTTL {
		t.Errorf("TTL %v outside bounds [%v, %v]", ttl, defaultPolicy.MinTTL, defaultPolicy.MaxTTL)
	}
}

// TestAdaptivePolicy tests the adaptive caching policy
func TestAdaptivePolicy(t *testing.T) {
	policy := NewAdaptivePolicy()

	// Record some access patterns
	for i := 0; i < 10; i++ {
		policy.RecordAccess("resource_1")
		time.Sleep(10 * time.Millisecond)
	}

	pattern := policy.getPattern("resource_1")
	if pattern.AccessCount != 10 {
		t.Errorf("Expected access count 10, got %d", pattern.AccessCount)
	}

	// Test TTL computation with learned pattern
	entry := &CacheEntry{
		Key:         "resource_1",
		AccessCount: 10,
	}

	ttl := policy.ComputeTTL(entry, pattern)
	if ttl < policy.MinTTL || ttl > policy.MaxTTL {
		t.Errorf("Adaptive TTL %v outside bounds", ttl)
	}
}

// TestCacheWarmup tests cache warming strategies
func TestCacheWarmup(t *testing.T) {
	cache := NewLRUCache(100, 10)

	// Test static warmup
	urls := []string{"http://example.com/1", "http://example.com/2", "http://example.com/3"}
	staticWarmup := NewStaticWarmup(urls)

	ctx := context.Background()
	err := staticWarmup.Warmup(ctx, cache)
	if err != nil {
		t.Errorf("Static warmup failed: %v", err)
	}

	if cache.Size() != 3 {
		t.Errorf("Expected 3 entries after warmup, got %d", cache.Size())
	}
}

// TestPredictiveWarmup tests predictive cache warming
func TestPredictiveWarmup(t *testing.T) {
	warmup := NewPredictiveWarmup(30*time.Minute, 5)

	// Create some access patterns
	for i := 0; i < 5; i++ {
		pattern := &AccessPattern{
			Key:              fmt.Sprintf("resource_%d", i),
			AccessCount:      int64(10 * (i + 1)),
			PredictedNextUse: time.Now().Add(15 * time.Minute),
			Volatility:       0.3,
		}
		warmup.LearnPattern(fmt.Sprintf("resource_%d", i), pattern)
	}

	predictions := warmup.Predict()
	if len(predictions) == 0 {
		t.Error("Expected some predictions, got none")
	}

	if len(predictions) > warmup.topN {
		t.Errorf("Expected at most %d predictions, got %d", warmup.topN, len(predictions))
	}

	// Verify predictions are sorted by priority
	for i := 1; i < len(predictions); i++ {
		if predictions[i].Priority > predictions[i-1].Priority {
			t.Error("Predictions should be sorted by priority (descending)")
		}
	}
}

// TestCacheWarmer tests the cache warmer orchestration
func TestCacheWarmer(t *testing.T) {
	cache := NewLRUCache(100, 10)
	urls := []string{"http://example.com/test"}
	strategy := NewStaticWarmup(urls)
	warmer := NewCacheWarmer(cache, strategy)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test immediate warmup
	err := warmer.WarmupNow(ctx)
	if err != nil {
		t.Errorf("Immediate warmup failed: %v", err)
	}

	if cache.Size() == 0 {
		t.Error("Cache should have entries after warmup")
	}

	// Test prefetch queue
	warmer.Prefetch("http://example.com/prefetch", 10)
	if warmer.queue.Size() != 1 {
		t.Errorf("Expected 1 item in prefetch queue, got %d", warmer.queue.Size())
	}

	err = warmer.ProcessPrefetchQueue(ctx, 10)
	if err != nil {
		t.Errorf("Prefetch processing failed: %v", err)
	}

	if warmer.queue.Size() != 0 {
		t.Error("Prefetch queue should be empty after processing")
	}
}

// TestCacheCleanup tests automatic cleanup of expired entries
func TestCacheCleanup(t *testing.T) {
	cache := NewLRUCache(100, 10)

	// Add entries with short TTL
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte(fmt.Sprintf("value_%d", i)),
			StatusCode:   200,
			Size:         10,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          50 * time.Millisecond,
			ExpiresAt:    time.Now().Add(50 * time.Millisecond),
		}
		cache.Put(key, entry)
	}

	if cache.Size() != 5 {
		t.Errorf("Expected 5 entries, got %d", cache.Size())
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Run cleanup
	evicted := cache.EvictExpired()
	if evicted != 5 {
		t.Errorf("Expected to evict 5 entries, evicted %d", evicted)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", cache.Size())
	}
}

// TestCacheSnapshot tests cache persistence via snapshots
func TestCacheSnapshot(t *testing.T) {
	cache := NewLRUCache(100, 10)

	// Add some entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte(fmt.Sprintf("value_%d", i)),
			StatusCode:   200,
			Size:         10,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	// Create snapshot
	snapshot := cache.Snapshot()
	if len(snapshot) != 10 {
		t.Errorf("Expected 10 entries in snapshot, got %d", len(snapshot))
	}

	// Clear cache
	cache.Clear()
	if cache.Size() != 0 {
		t.Error("Cache should be empty after Clear")
	}

	// Restore from snapshot
	err := cache.LoadSnapshot(snapshot)
	if err != nil {
		t.Errorf("Failed to load snapshot: %v", err)
	}

	if cache.Size() != 10 {
		t.Errorf("Expected 10 entries after loading snapshot, got %d", cache.Size())
	}
}

// TestCacheMemoryLimit tests memory-based eviction
func TestCacheMemoryLimit(t *testing.T) {
	maxMemoryMB := int64(1) // 1MB limit
	cache := NewLRUCache(1000, maxMemoryMB)

	// Add entries until we hit memory limit
	entrySize := int64(100 * 1024) // 100KB per entry
	maxEntries := (maxMemoryMB * 1024 * 1024) / entrySize

	for i := 0; i < int(maxEntries)+5; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        make([]byte, entrySize),
			StatusCode:   200,
			Size:         entrySize,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	// Memory usage should not exceed limit
	memoryUsage := cache.MemoryUsage()
	maxMemory := maxMemoryMB * 1024 * 1024
	if memoryUsage > maxMemory {
		t.Errorf("Memory usage %d exceeds limit %d", memoryUsage, maxMemory)
	}
}

// BenchmarkCacheGet benchmarks cache get operations
func BenchmarkCacheGet(b *testing.B) {
	cache := NewLRUCache(10000, 100)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte("benchmark_value"),
			StatusCode:   200,
			Size:         15,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%1000)
		cache.Get(key)
	}
}

// BenchmarkCachePut benchmarks cache put operations
func BenchmarkCachePut(b *testing.B) {
	cache := NewLRUCache(10000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte("benchmark_value"),
			StatusCode:   200,
			Size:         15,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}
}

// BenchmarkCacheConcurrentAccess benchmarks concurrent cache access
func BenchmarkCacheConcurrentAccess(b *testing.B) {
	cache := NewLRUCache(10000, 100)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		entry := &CacheEntry{
			Key:          key,
			Value:        []byte("benchmark_value"),
			StatusCode:   200,
			Size:         15,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			TTL:          5 * time.Minute,
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		cache.Put(key, entry)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			cache.Get(key)
			i++
		}
	})
}
