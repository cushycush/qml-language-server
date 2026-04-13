package handler

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]cacheItem[V]
	maxAge   time.Duration
	maxItems int
}

type cacheItem[V any] struct {
	value     V
	timestamp time.Time
}

func NewCache[K comparable, V any](maxAge time.Duration, maxItems int) *Cache[K, V] {
	return &Cache[K, V]{
		items:    make(map[K]cacheItem[V]),
		maxAge:   maxAge,
		maxItems: maxItems,
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	if time.Since(item.timestamp) > c.maxAge {
		return *new(V), false
	}

	return item.value, true
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.maxItems {
		c.evictOldest()
	}

	c.items[key] = cacheItem[V]{
		value:     value,
		timestamp: time.Now(),
	}
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]cacheItem[V])
}

func (c *Cache[K, V]) evictOldest() {
	var oldestKey K
	var oldestTime time.Time
	first := true

	for k, item := range c.items {
		if first || item.timestamp.Before(oldestTime) {
			oldestKey = k
			oldestTime = item.timestamp
			first = false
		}
	}

	delete(c.items, oldestKey)
}

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

type DiagnosticsCache struct {
	cache *Cache[string, []CachedDiagnostic]
}

type CachedDiagnostic struct {
	Range    [4]int
	Severity int
	Message  string
	Source   string
}

func NewDiagnosticsCache() *DiagnosticsCache {
	return &DiagnosticsCache{
		cache: NewCache[string, []CachedDiagnostic](5*time.Minute, 100),
	}
}

func (dc *DiagnosticsCache) Get(uri string) ([]CachedDiagnostic, bool) {
	return dc.cache.Get(uri)
}

func (dc *DiagnosticsCache) Set(uri string, diagnostics []CachedDiagnostic) {
	dc.cache.Set(uri, diagnostics)
}

func (dc *DiagnosticsCache) Invalidate(uri string) {
	dc.cache.Delete(uri)
}
