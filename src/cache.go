package main

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a single cached item with metadata
type CacheEntry struct {
	Key          string        // Cache key (typically URL + params hash)
	Value        []byte        // Cached response body
	StatusCode   int           // HTTP status code
	Headers      map[string]string // Important response headers
	Size         int64         // Entry size in bytes
	CreatedAt    time.Time     // When entry was created
	LastAccessed time.Time     // Last access timestamp
	AccessCount  int64         // Number of times accessed
	TTL          time.Duration // Time-to-live duration
	ExpiresAt    time.Time     // Absolute expiration time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Age returns how long the entry has been in cache
func (e *CacheEntry) Age() time.Duration {
	return time.Since(e.CreatedAt)
}

// CacheKey represents a request identifier for caching
type CacheKey struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    string
}

// Hash generates a unique hash for the cache key
func (k *CacheKey) Hash() string {
	hasher := sha256.New()
	hasher.Write([]byte(k.Method))
	hasher.Write([]byte(k.URL))

	// Include relevant headers in hash
	for key, value := range k.Headers {
		hasher.Write([]byte(key))
		hasher.Write([]byte(value))
	}

	if k.Body != "" {
		hasher.Write([]byte(k.Body))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// LRUCache implements a thread-safe Least Recently Used cache
type LRUCache struct {
	capacity     int                        // Maximum number of entries
	maxMemory    int64                      // Maximum memory usage in bytes
	currentSize  int64                      // Current memory usage
	entries      map[string]*list.Element   // Hash map for O(1) lookups
	evictionList *list.List                 // Doubly linked list for LRU ordering
	mu           sync.RWMutex               // Read-write mutex for thread safety
	metrics      *CacheMetrics              // Performance metrics
	policy       CachePolicy                // Eviction and TTL policy
	onEvict      func(key string, entry *CacheEntry) // Eviction callback
}

// cacheItem wraps a cache entry with its key for the LRU list
type cacheItem struct {
	key   string
	entry *CacheEntry
}

// NewLRUCache creates a new LRU cache with specified capacity
func NewLRUCache(capacity int, maxMemoryMB int64) *LRUCache {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}

	cache := &LRUCache{
		capacity:     capacity,
		maxMemory:    maxMemoryMB * 1024 * 1024, // Convert MB to bytes
		currentSize:  0,
		entries:      make(map[string]*list.Element),
		evictionList: list.New(),
		metrics:      NewCacheMetrics(),
		policy:       NewDefaultPolicy(),
	}

	return cache
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Record the get operation
	c.metrics.RecordGet()

	element, exists := c.entries[key]
	if !exists {
		c.metrics.RecordMiss()
		return nil, false
	}

	item := element.Value.(*cacheItem)
	entry := item.entry

	// Check if expired
	if entry.IsExpired() {
		c.removeElement(element)
		c.metrics.RecordExpiration()
		c.metrics.RecordMiss()
		return nil, false
	}

	// Update access metadata
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	// Move to front (most recently used)
	c.evictionList.MoveToFront(element)

	c.metrics.RecordHit()
	c.metrics.RecordAccessLatency(time.Since(entry.LastAccessed))

	return entry, true
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key string, entry *CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if entry already exists
	if element, exists := c.entries[key]; exists {
		// Update existing entry
		oldItem := element.Value.(*cacheItem)
		c.currentSize -= oldItem.entry.Size

		item := &cacheItem{key: key, entry: entry}
		element.Value = item
		c.evictionList.MoveToFront(element)
		c.currentSize += entry.Size

		c.metrics.RecordUpdate()
		return nil
	}

	// Check capacity and evict if necessary
	if c.evictionList.Len() >= c.capacity || c.currentSize+entry.Size > c.maxMemory {
		c.evictOldest()
	}

	// Add new entry
	item := &cacheItem{key: key, entry: entry}
	element := c.evictionList.PushFront(item)
	c.entries[key] = element
	c.currentSize += entry.Size

	c.metrics.RecordInsert()
	c.metrics.RecordMemoryUsage(c.currentSize)

	return nil
}

// Delete removes an entry from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.entries[key]
	if !exists {
		return false
	}

	c.removeElement(element)
	c.metrics.RecordEviction()

	return true
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.evictionList = list.New()
	c.currentSize = 0
	c.metrics.RecordClear()
}

// Size returns the current number of entries in the cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// MemoryUsage returns current memory usage in bytes
func (c *LRUCache) MemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// Capacity returns the maximum capacity
func (c *LRUCache) Capacity() int {
	return c.capacity
}

// HitRatio returns the current cache hit ratio
func (c *LRUCache) HitRatio() float64 {
	return c.metrics.HitRatio()
}

// GetMetrics returns the cache metrics
func (c *LRUCache) GetMetrics() *CacheMetrics {
	return c.metrics
}

// SetPolicy sets the cache policy
func (c *LRUCache) SetPolicy(policy CachePolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.policy = policy
}

// SetEvictionCallback sets a callback function to be called when entries are evicted
func (c *LRUCache) SetEvictionCallback(callback func(key string, entry *CacheEntry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = callback
}

// EvictExpired removes all expired entries
func (c *LRUCache) EvictExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	evicted := 0
	now := time.Now()

	// Iterate through all entries
	for key, element := range c.entries {
		item := element.Value.(*cacheItem)
		if now.After(item.entry.ExpiresAt) {
			c.removeElement(element)
			c.metrics.RecordExpiration()
			evicted++
			delete(c.entries, key)
		}
	}

	return evicted
}

// StartCleanupRoutine starts a background goroutine to periodically clean up expired entries
func (c *LRUCache) StartCleanupRoutine(interval time.Duration) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				evicted := c.EvictExpired()
				if evicted > 0 {
					c.metrics.RecordCleanup(evicted)
				}
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// GetStats returns current cache statistics
func (c *LRUCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"size":          len(c.entries),
		"capacity":      c.capacity,
		"memory_usage":  c.currentSize,
		"max_memory":    c.maxMemory,
		"memory_utilization": float64(c.currentSize) / float64(c.maxMemory) * 100,
		"hit_ratio":     c.metrics.HitRatio(),
		"total_gets":    c.metrics.TotalGets(),
		"total_hits":    c.metrics.TotalHits(),
		"total_misses":  c.metrics.TotalMisses(),
	}
}

// evictOldest removes the least recently used entry
func (c *LRUCache) evictOldest() {
	element := c.evictionList.Back()
	if element != nil {
		c.removeElement(element)
		c.metrics.RecordEviction()
	}
}

// removeElement removes an element from the cache
func (c *LRUCache) removeElement(element *list.Element) {
	c.evictionList.Remove(element)
	item := element.Value.(*cacheItem)
	delete(c.entries, item.key)
	c.currentSize -= item.entry.Size

	// Call eviction callback if set
	if c.onEvict != nil {
		c.onEvict(item.key, item.entry)
	}
}

// Keys returns all cache keys (primarily for debugging/testing)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	return keys
}

// Snapshot creates a snapshot of cache entries for persistence
func (c *LRUCache) Snapshot() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CacheEntry, 0, len(c.entries))
	for _, element := range c.entries {
		item := element.Value.(*cacheItem)
		if !item.entry.IsExpired() {
			entries = append(entries, item.entry)
		}
	}
	return entries
}

// LoadSnapshot loads entries from a snapshot (for cache persistence)
func (c *LRUCache) LoadSnapshot(entries []*CacheEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	loaded := 0
	now := time.Now()

	for _, entry := range entries {
		// Skip expired entries
		if now.After(entry.ExpiresAt) {
			continue
		}

		// Check capacity
		if c.evictionList.Len() >= c.capacity {
			break
		}

		key := fmt.Sprintf("%s_%d", entry.Key, entry.CreatedAt.Unix())
		item := &cacheItem{key: key, entry: entry}
		element := c.evictionList.PushFront(item)
		c.entries[key] = element
		c.currentSize += entry.Size
		loaded++
	}

	return nil
}
