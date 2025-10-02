// Package src provides a functional LRU cache implementation
// This replaces the stub cache implementations to enable actual caching
package main

import (
	"container/list"
	"sync"
	"time"
)

// CacheEntry represents a cache entry with TTL
type CacheEntry struct {
	key        string
	value      interface{}
	expiration time.Time
	createdAt  time.Time
}

// FunctionalCache provides a working LRU cache with TTL
type FunctionalCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	lru      *list.List
}

// NewFunctionalCache creates a new functional LRU cache
func NewFunctionalCache(capacity int) *FunctionalCache {
	return &FunctionalCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

// GetWithAge retrieves an item from cache with age information
func (fc *FunctionalCache) GetWithAge(key string) (interface{}, time.Duration, bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	element, exists := fc.items[key]
	if !exists {
		return nil, 0, false
	}

	entry := element.Value.(*CacheEntry)

	// Check if expired
	if time.Now().After(entry.expiration) {
		// Remove expired entry
		fc.lru.Remove(element)
		delete(fc.items, key)
		return nil, 0, false
	}

	// Move to front (most recently used)
	fc.lru.MoveToFront(element)

	// Calculate age
	age := time.Since(entry.createdAt)

	return entry.value, age, true
}

// SetWithTTL sets an item in cache with TTL
func (fc *FunctionalCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	now := time.Now()
	entry := &CacheEntry{
		key:        key,
		value:      value,
		expiration: now.Add(ttl),
		createdAt:  now,
	}

	// Check if key already exists
	if element, exists := fc.items[key]; exists {
		// Update existing entry
		element.Value = entry
		fc.lru.MoveToFront(element)
		return
	}

	// Add new entry
	element := fc.lru.PushFront(entry)
	fc.items[key] = element

	// Check capacity and evict if necessary
	if fc.lru.Len() > fc.capacity {
		fc.evictOldest()
	}
}

// Delete removes an item from cache
func (fc *FunctionalCache) Delete(key string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if element, exists := fc.items[key]; exists {
		fc.lru.Remove(element)
		delete(fc.items, key)
	}
}

// evictOldest removes the oldest entry from cache
func (fc *FunctionalCache) evictOldest() {
	oldest := fc.lru.Back()
	if oldest != nil {
		entry := oldest.Value.(*CacheEntry)
		fc.lru.Remove(oldest)
		delete(fc.items, entry.key)
	}
}

// Size returns the current number of items in cache
func (fc *FunctionalCache) Size() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.lru.Len()
}

// Clear removes all items from cache
func (fc *FunctionalCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.lru.Init()
	fc.items = make(map[string]*list.Element)
}

// Stats returns cache statistics
func (fc *FunctionalCache) Stats() map[string]interface{} {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	expired := 0
	now := time.Now()

	// Count expired entries
	for element := fc.lru.Front(); element != nil; element = element.Next() {
		entry := element.Value.(*CacheEntry)
		if now.After(entry.expiration) {
			expired++
		}
	}

	return map[string]interface{}{
		"size":     fc.lru.Len(),
		"capacity": fc.capacity,
		"expired":  expired,
	}
}