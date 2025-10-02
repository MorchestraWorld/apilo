// Package src provides advanced cache invalidation strategies
// This addresses the critical cache invalidation complexity issue
package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// InvalidationStrategy defines the interface for cache invalidation strategies
type InvalidationStrategy interface {
	ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool
	GetName() string
	GetPriority() int
}

// InvalidationMetadata contains contextual information for invalidation decisions
type InvalidationMetadata struct {
	CurrentTime     time.Time
	RequestContext  context.Context
	DependencyGraph *DependencyGraph
	DataVersion     string
	SourceUpdate    time.Time
	UpdateType      UpdateType
	AffectedTags    []string
	UserContext     *UserContext
}

// UpdateType defines the type of data update
type UpdateType int

const (
	UpdateTypeCreate UpdateType = iota
	UpdateTypeModify
	UpdateTypeDelete
	UpdateTypeInvalidate
)

// UserContext provides user-specific invalidation context
type UserContext struct {
	UserID      string
	Permissions []string
	Region      string
	TenantID    string
}

// AdvancedInvalidationManager manages multiple invalidation strategies
type AdvancedInvalidationManager struct {
	strategies       []InvalidationStrategy
	dependencyGraph  *DependencyGraph
	taggedCache      *TaggedCacheIndex
	versionManager   *VersionManager
	config          *InvalidationConfig
	metrics         *InvalidationMetrics
	mu              sync.RWMutex
}

// InvalidationConfig configures the invalidation manager
type InvalidationConfig struct {
	EnableTagBasedInvalidation    bool          `yaml:"enable_tag_based_invalidation"`
	EnableDependencyTracking      bool          `yaml:"enable_dependency_tracking"`
	EnableVersionBasedInvalidation bool          `yaml:"enable_version_based_invalidation"`
	EnablePatternMatching         bool          `yaml:"enable_pattern_matching"`
	MaxDependencyDepth            int           `yaml:"max_dependency_depth"`
	InvalidationBatchSize         int           `yaml:"invalidation_batch_size"`
	AsyncInvalidation            bool          `yaml:"async_invalidation"`
	InvalidationTimeout          time.Duration `yaml:"invalidation_timeout"`
}

// InvalidationMetrics tracks invalidation performance
type InvalidationMetrics struct {
	StrategyExecutions   map[string]int64
	TotalInvalidations   int64
	BatchInvalidations   int64
	TagInvalidations     int64
	DependencyInvalidations int64
	VersionInvalidations int64
	PatternInvalidations int64
	InvalidationLatency  time.Duration
	mu                  sync.RWMutex
}

// TaggedCacheIndex maintains tag-to-key mappings for efficient invalidation
type TaggedCacheIndex struct {
	tagToKeys   map[string]map[string]bool // tag -> set of keys
	keyToTags   map[string]map[string]bool // key -> set of tags
	mu          sync.RWMutex
}

// DependencyGraph tracks cache entry dependencies
type DependencyGraph struct {
	dependencies map[string]map[string]bool // key -> set of dependent keys
	dependents   map[string]map[string]bool // key -> set of keys it depends on
	mu           sync.RWMutex
}

// VersionManager tracks data versions for invalidation
type VersionManager struct {
	keyVersions    map[string]string // key -> version
	globalVersion  string
	versionHistory map[string]time.Time // version -> timestamp
	mu             sync.RWMutex
}

// TTLInvalidationStrategy implements time-based invalidation
type TTLInvalidationStrategy struct {
	name     string
	priority int
}

// TagBasedInvalidationStrategy implements tag-based invalidation
type TagBasedInvalidationStrategy struct {
	name     string
	priority int
	tagIndex *TaggedCacheIndex
}

// DependencyInvalidationStrategy implements dependency-based invalidation
type DependencyInvalidationStrategy struct {
	name            string
	priority        int
	dependencyGraph *DependencyGraph
	maxDepth        int
}

// VersionBasedInvalidationStrategy implements version-based invalidation
type VersionBasedInvalidationStrategy struct {
	name           string
	priority       int
	versionManager *VersionManager
}

// PatternBasedInvalidationStrategy implements pattern-based invalidation
type PatternBasedInvalidationStrategy struct {
	name     string
	priority int
	patterns []*regexp.Regexp
}

// ConditionalInvalidationStrategy implements conditional invalidation
type ConditionalInvalidationStrategy struct {
	name      string
	priority  int
	condition func(*CacheElement, InvalidationMetadata) bool
}

// NewAdvancedInvalidationManager creates a new invalidation manager
func NewAdvancedInvalidationManager(config *InvalidationConfig) *AdvancedInvalidationManager {
	if config == nil {
		config = DefaultInvalidationConfig()
	}

	manager := &AdvancedInvalidationManager{
		strategies:      make([]InvalidationStrategy, 0),
		dependencyGraph: NewDependencyGraph(),
		taggedCache:     NewTaggedCacheIndex(),
		versionManager:  NewVersionManager(),
		config:          config,
		metrics:         NewInvalidationMetrics(),
	}

	// Register default strategies
	manager.registerDefaultStrategies()

	return manager
}

// registerDefaultStrategies registers the default invalidation strategies
func (aim *AdvancedInvalidationManager) registerDefaultStrategies() {
	// TTL strategy (highest priority)
	aim.AddStrategy(&TTLInvalidationStrategy{
		name:     "TTL",
		priority: 100,
	})

	if aim.config.EnableTagBasedInvalidation {
		aim.AddStrategy(&TagBasedInvalidationStrategy{
			name:     "TagBased",
			priority: 90,
			tagIndex: aim.taggedCache,
		})
	}

	if aim.config.EnableDependencyTracking {
		aim.AddStrategy(&DependencyInvalidationStrategy{
			name:            "Dependency",
			priority:        80,
			dependencyGraph: aim.dependencyGraph,
			maxDepth:        aim.config.MaxDependencyDepth,
		})
	}

	if aim.config.EnableVersionBasedInvalidation {
		aim.AddStrategy(&VersionBasedInvalidationStrategy{
			name:           "Version",
			priority:       70,
			versionManager: aim.versionManager,
		})
	}

	if aim.config.EnablePatternMatching {
		patterns := []*regexp.Regexp{
			regexp.MustCompile(`^user:\d+:`),     // User-specific keys
			regexp.MustCompile(`^api:v\d+:`),    // API version keys
			regexp.MustCompile(`^temp:`),         // Temporary keys
		}

		aim.AddStrategy(&PatternBasedInvalidationStrategy{
			name:     "Pattern",
			priority: 60,
			patterns: patterns,
		})
	}
}

// AddStrategy adds an invalidation strategy
func (aim *AdvancedInvalidationManager) AddStrategy(strategy InvalidationStrategy) {
	aim.mu.Lock()
	defer aim.mu.Unlock()

	aim.strategies = append(aim.strategies, strategy)

	// Sort strategies by priority (highest first)
	for i := len(aim.strategies) - 1; i > 0; i-- {
		if aim.strategies[i].GetPriority() > aim.strategies[i-1].GetPriority() {
			aim.strategies[i], aim.strategies[i-1] = aim.strategies[i-1], aim.strategies[i]
		} else {
			break
		}
	}
}

// ShouldInvalidate determines if a cache entry should be invalidated
func (aim *AdvancedInvalidationManager) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	aim.mu.RLock()
	defer aim.mu.RUnlock()

	start := time.Now()
	defer func() {
		aim.metrics.mu.Lock()
		aim.metrics.InvalidationLatency = time.Since(start)
		aim.metrics.mu.Unlock()
	}()

	// Execute strategies in priority order
	for _, strategy := range aim.strategies {
		if strategy.ShouldInvalidate(entry, metadata) {
			aim.metrics.mu.Lock()
			aim.metrics.StrategyExecutions[strategy.GetName()]++
			aim.metrics.TotalInvalidations++
			aim.metrics.mu.Unlock()
			return true
		}
	}

	return false
}

// InvalidateByTag invalidates all cache entries with specific tags
func (aim *AdvancedInvalidationManager) InvalidateByTag(tags []string, cache InvalidatableCache) error {
	if !aim.config.EnableTagBasedInvalidation {
		return fmt.Errorf("tag-based invalidation is disabled")
	}

	keys := aim.taggedCache.GetKeysByTags(tags)

	if aim.config.AsyncInvalidation {
		go aim.batchInvalidate(keys, cache)
	} else {
		return aim.batchInvalidate(keys, cache)
	}

	aim.metrics.mu.Lock()
	aim.metrics.TagInvalidations++
	aim.metrics.mu.Unlock()

	return nil
}

// InvalidateByPattern invalidates cache entries matching patterns
func (aim *AdvancedInvalidationManager) InvalidateByPattern(pattern string, cache InvalidatableCache) error {
	if !aim.config.EnablePatternMatching {
		return fmt.Errorf("pattern-based invalidation is disabled")
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}

	// This would need to be implemented by the cache to provide all keys
	// For now, we'll assume a method to get matching keys
	keys := cache.GetKeysMatchingPattern(regex)

	if aim.config.AsyncInvalidation {
		go aim.batchInvalidate(keys, cache)
	} else {
		return aim.batchInvalidate(keys, cache)
	}

	aim.metrics.mu.Lock()
	aim.metrics.PatternInvalidations++
	aim.metrics.mu.Unlock()

	return nil
}

// InvalidateByDependency invalidates cache entries based on dependencies
func (aim *AdvancedInvalidationManager) InvalidateByDependency(sourceKey string, cache InvalidatableCache) error {
	if !aim.config.EnableDependencyTracking {
		return fmt.Errorf("dependency tracking is disabled")
	}

	dependentKeys := aim.dependencyGraph.GetDependentKeys(sourceKey, aim.config.MaxDependencyDepth)

	if aim.config.AsyncInvalidation {
		go aim.batchInvalidate(dependentKeys, cache)
	} else {
		return aim.batchInvalidate(dependentKeys, cache)
	}

	aim.metrics.mu.Lock()
	aim.metrics.DependencyInvalidations++
	aim.metrics.mu.Unlock()

	return nil
}

// InvalidateByVersion invalidates cache entries based on version changes
func (aim *AdvancedInvalidationManager) InvalidateByVersion(newVersion string, cache InvalidatableCache) error {
	if !aim.config.EnableVersionBasedInvalidation {
		return fmt.Errorf("version-based invalidation is disabled")
	}

	outdatedKeys := aim.versionManager.GetOutdatedKeys(newVersion)

	if aim.config.AsyncInvalidation {
		go aim.batchInvalidate(outdatedKeys, cache)
	} else {
		return aim.batchInvalidate(outdatedKeys, cache)
	}

	aim.metrics.mu.Lock()
	aim.metrics.VersionInvalidations++
	aim.metrics.mu.Unlock()

	return nil
}

// batchInvalidate performs batch invalidation of keys
func (aim *AdvancedInvalidationManager) batchInvalidate(keys []string, cache InvalidatableCache) error {
	batchSize := aim.config.InvalidationBatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		for _, key := range batch {
			cache.Delete(key)
		}

		aim.metrics.mu.Lock()
		aim.metrics.BatchInvalidations++
		aim.metrics.mu.Unlock()

		// Small delay between batches to avoid overwhelming the system
		if i+batchSize < len(keys) {
			time.Sleep(time.Millisecond)
		}
	}

	return nil
}

// TTL Strategy Implementation
func (tis *TTLInvalidationStrategy) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	return metadata.CurrentTime.After(entry.expiresAt)
}

func (tis *TTLInvalidationStrategy) GetName() string {
	return tis.name
}

func (tis *TTLInvalidationStrategy) GetPriority() int {
	return tis.priority
}

// Tag-Based Strategy Implementation
func (tbis *TagBasedInvalidationStrategy) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	if len(metadata.AffectedTags) == 0 {
		return false
	}

	entryTags := tbis.tagIndex.GetTagsForKey(entry.key)
	for _, affectedTag := range metadata.AffectedTags {
		for entryTag := range entryTags {
			if entryTag == affectedTag {
				return true
			}
		}
	}

	return false
}

func (tbis *TagBasedInvalidationStrategy) GetName() string {
	return tbis.name
}

func (tbis *TagBasedInvalidationStrategy) GetPriority() int {
	return tbis.priority
}

// Dependency Strategy Implementation
func (dis *DependencyInvalidationStrategy) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	// Check if this entry depends on any updated data
	dependencies := dis.dependencyGraph.GetDependencies(entry.key)

	// Simple implementation: check if source update affects any dependencies
	for dependency := range dependencies {
		if strings.Contains(dependency, metadata.DataVersion) {
			return true
		}
	}

	return false
}

func (dis *DependencyInvalidationStrategy) GetName() string {
	return dis.name
}

func (dis *DependencyInvalidationStrategy) GetPriority() int {
	return dis.priority
}

// Version Strategy Implementation
func (vbis *VersionBasedInvalidationStrategy) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	if metadata.DataVersion == "" {
		return false
	}

	entryVersion := vbis.versionManager.GetKeyVersion(entry.key)
	return entryVersion != metadata.DataVersion
}

func (vbis *VersionBasedInvalidationStrategy) GetName() string {
	return vbis.name
}

func (vbis *VersionBasedInvalidationStrategy) GetPriority() int {
	return vbis.priority
}

// Pattern Strategy Implementation
func (pbis *PatternBasedInvalidationStrategy) ShouldInvalidate(entry *CacheElement, metadata InvalidationMetadata) bool {
	for _, pattern := range pbis.patterns {
		if pattern.MatchString(entry.key) {
			// Additional logic could be added here based on metadata
			return metadata.UpdateType == UpdateTypeInvalidate
		}
	}
	return false
}

func (pbis *PatternBasedInvalidationStrategy) GetName() string {
	return pbis.name
}

func (pbis *PatternBasedInvalidationStrategy) GetPriority() int {
	return pbis.priority
}

// Tagged Cache Index Implementation
func NewTaggedCacheIndex() *TaggedCacheIndex {
	return &TaggedCacheIndex{
		tagToKeys: make(map[string]map[string]bool),
		keyToTags: make(map[string]map[string]bool),
	}
}

func (tci *TaggedCacheIndex) AddKeyWithTags(key string, tags []string) {
	tci.mu.Lock()
	defer tci.mu.Unlock()

	// Initialize key entry if not exists
	if tci.keyToTags[key] == nil {
		tci.keyToTags[key] = make(map[string]bool)
	}

	for _, tag := range tags {
		// Add key to tag mapping
		if tci.tagToKeys[tag] == nil {
			tci.tagToKeys[tag] = make(map[string]bool)
		}
		tci.tagToKeys[tag][key] = true

		// Add tag to key mapping
		tci.keyToTags[key][tag] = true
	}
}

func (tci *TaggedCacheIndex) RemoveKey(key string) {
	tci.mu.Lock()
	defer tci.mu.Unlock()

	// Remove key from all tag mappings
	if tags, exists := tci.keyToTags[key]; exists {
		for tag := range tags {
			if keys, exists := tci.tagToKeys[tag]; exists {
				delete(keys, key)
				if len(keys) == 0 {
					delete(tci.tagToKeys, tag)
				}
			}
		}
		delete(tci.keyToTags, key)
	}
}

func (tci *TaggedCacheIndex) GetKeysByTags(tags []string) []string {
	tci.mu.RLock()
	defer tci.mu.RUnlock()

	keySet := make(map[string]bool)

	for _, tag := range tags {
		if keys, exists := tci.tagToKeys[tag]; exists {
			for key := range keys {
				keySet[key] = true
			}
		}
	}

	result := make([]string, 0, len(keySet))
	for key := range keySet {
		result = append(result, key)
	}

	return result
}

func (tci *TaggedCacheIndex) GetTagsForKey(key string) map[string]bool {
	tci.mu.RLock()
	defer tci.mu.RUnlock()

	if tags, exists := tci.keyToTags[key]; exists {
		// Return copy to avoid race conditions
		result := make(map[string]bool, len(tags))
		for tag, val := range tags {
			result[tag] = val
		}
		return result
	}

	return make(map[string]bool)
}

// Dependency Graph Implementation
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(map[string]map[string]bool),
		dependents:   make(map[string]map[string]bool),
	}
}

func (dg *DependencyGraph) AddDependency(key, dependsOn string) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// Add to dependencies (key depends on dependsOn)
	if dg.dependencies[key] == nil {
		dg.dependencies[key] = make(map[string]bool)
	}
	dg.dependencies[key][dependsOn] = true

	// Add to dependents (dependsOn is depended on by key)
	if dg.dependents[dependsOn] == nil {
		dg.dependents[dependsOn] = make(map[string]bool)
	}
	dg.dependents[dependsOn][key] = true
}

func (dg *DependencyGraph) GetDependentKeys(key string, maxDepth int) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	visited := make(map[string]bool)
	result := make([]string, 0)

	dg.getDependentKeysRecursive(key, maxDepth, 0, visited, &result)

	return result
}

func (dg *DependencyGraph) getDependentKeysRecursive(key string, maxDepth, currentDepth int, visited map[string]bool, result *[]string) {
	if currentDepth >= maxDepth || visited[key] {
		return
	}

	visited[key] = true

	if dependents, exists := dg.dependents[key]; exists {
		for dependent := range dependents {
			*result = append(*result, dependent)
			dg.getDependentKeysRecursive(dependent, maxDepth, currentDepth+1, visited, result)
		}
	}
}

func (dg *DependencyGraph) GetDependencies(key string) map[string]bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	if deps, exists := dg.dependencies[key]; exists {
		// Return copy
		result := make(map[string]bool, len(deps))
		for dep, val := range deps {
			result[dep] = val
		}
		return result
	}

	return make(map[string]bool)
}

// Version Manager Implementation
func NewVersionManager() *VersionManager {
	return &VersionManager{
		keyVersions:    make(map[string]string),
		versionHistory: make(map[string]time.Time),
		globalVersion:  "1.0.0",
	}
}

func (vm *VersionManager) SetKeyVersion(key, version string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.keyVersions[key] = version
	vm.versionHistory[version] = time.Now()
}

func (vm *VersionManager) GetKeyVersion(key string) string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if version, exists := vm.keyVersions[key]; exists {
		return version
	}
	return vm.globalVersion
}

func (vm *VersionManager) GetOutdatedKeys(newVersion string) []string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	result := make([]string, 0)
	for key, version := range vm.keyVersions {
		if version != newVersion {
			result = append(result, key)
		}
	}

	return result
}

func (vm *VersionManager) UpdateGlobalVersion(version string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.globalVersion = version
	vm.versionHistory[version] = time.Now()
}

// InvalidatableCache interface that caches must implement
type InvalidatableCache interface {
	Delete(key string)
	GetKeysMatchingPattern(pattern *regexp.Regexp) []string
}

// Helper functions and default configurations
func DefaultInvalidationConfig() *InvalidationConfig {
	return &InvalidationConfig{
		EnableTagBasedInvalidation:     true,
		EnableDependencyTracking:       true,
		EnableVersionBasedInvalidation: true,
		EnablePatternMatching:          true,
		MaxDependencyDepth:            3,
		InvalidationBatchSize:         100,
		AsyncInvalidation:            false,
		InvalidationTimeout:          30 * time.Second,
	}
}

func NewInvalidationMetrics() *InvalidationMetrics {
	return &InvalidationMetrics{
		StrategyExecutions: make(map[string]int64),
	}
}

// GetMetrics returns current invalidation metrics
func (aim *AdvancedInvalidationManager) GetMetrics() InvalidationMetrics {
	aim.metrics.mu.RLock()
	defer aim.metrics.mu.RUnlock()

	// Return copy of metrics
	metrics := InvalidationMetrics{
		StrategyExecutions:      make(map[string]int64),
		TotalInvalidations:      aim.metrics.TotalInvalidations,
		BatchInvalidations:      aim.metrics.BatchInvalidations,
		TagInvalidations:        aim.metrics.TagInvalidations,
		DependencyInvalidations: aim.metrics.DependencyInvalidations,
		VersionInvalidations:    aim.metrics.VersionInvalidations,
		PatternInvalidations:    aim.metrics.PatternInvalidations,
		InvalidationLatency:     aim.metrics.InvalidationLatency,
	}

	for strategy, count := range aim.metrics.StrategyExecutions {
		metrics.StrategyExecutions[strategy] = count
	}

	return metrics
}