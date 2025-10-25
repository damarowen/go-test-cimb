package cache

import (
	"sync"
	"time"
)

// Cache defines the interface for cache operations
type Cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
}

// SimpleCache is a basic in-memory cache implementation
type SimpleCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSimpleCache creates a new SimpleCache instance
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the cache
func (c *SimpleCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get retrieves a value from the cache
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// Delete removes a value from the cache
func (c *SimpleCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// cacheItem represents an item in the TTL cache with expiration time
type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// isExpired checks if the cache item has expired
func (item *cacheItem) isExpired() bool {
	return time.Now().After(item.expiration)
}

// TTLCache is a cache implementation with time-to-live functionality
type TTLCache struct {
	data          map[string]*cacheItem
	mu            sync.RWMutex
	defaultTTL    time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// NewTTLCache creates a new TTLCache instance with specified default TTL
func NewTTLCache(defaultTTL time.Duration) *TTLCache {
	cache := &TTLCache{
		data:        make(map[string]*cacheItem),
		defaultTTL:  defaultTTL,
		stopCleanup: make(chan bool),
	}

	// Start background cleanup goroutine
	cache.startCleanup()

	return cache
}

// startCleanup starts a background goroutine to periodically clean expired entries
func (c *TTLCache) startCleanup() {
	// Run cleanup every minute or every TTL/2, whichever is shorter
	cleanupInterval := c.defaultTTL / 2
	if cleanupInterval > time.Minute {
		cleanupInterval = time.Minute
	}
	if cleanupInterval < time.Second {
		cleanupInterval = time.Second
	}

	c.cleanupTicker = time.NewTicker(cleanupInterval)

	go func() {
		for {
			select {
			case <-c.cleanupTicker.C:
				c.deleteExpired()
			case <-c.stopCleanup:
				c.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// deleteExpired removes all expired entries from the cache
func (c *TTLCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiration) {
			delete(c.data, key)
		}
	}
}

// Set stores a value in the cache with default TTL
func (c *TTLCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (c *TTLCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a value from the cache if it exists and hasn't expired
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if item.isExpired() {
		return nil, false
	}

	return item.value, true
}

// Delete removes a value from the cache
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Stop stops the background cleanup goroutine
func (c *TTLCache) Stop() {
	c.stopCleanup <- true
}

// Clear removes all entries from the cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*cacheItem)
}

// Size returns the number of items in the cache (including expired ones)
func (c *TTLCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
