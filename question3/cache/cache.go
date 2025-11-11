package cache

import (
	"log"
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

// TTLCache is a cache implementation with time-to-live functionality
type TTLCache struct {
	data          map[string]*cacheItem // ← Shared data!
	mu            sync.RWMutex
	defaultTTL    time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
	wg            sync.WaitGroup
}

// NewTTLCache creates a new TTLCache instance with specified default TTL
func NewTTLCache(defaultTTL time.Duration) *TTLCache {
	cache := &TTLCache{
		data:        make(map[string]*cacheItem),
		defaultTTL:  defaultTTL,
		stopCleanup: make(chan bool),
	}

	// Start a background cleanup goroutine
	cache.startCleanup()

	return cache
}

// startCleanup starts a background goroutine to periodically clean expired entries
func (c *TTLCache) startCleanup() {
	// Run cleanup every minute or every TTL/2, whichever is shorter
	//this is primary logic to determine cleanup interval, dibagi 2 adalah agar memiliki interval yang lebih ideal
	cleanupInterval := c.defaultTTL / 2
	if cleanupInterval > time.Minute {
		cleanupInterval = time.Minute // Max: 1 minute
	}
	if cleanupInterval < time.Second {
		cleanupInterval = time.Second // Min: 1 second
	}

	//start ticker, check for expired items every cleanupInterval, seperti setInterval() di js
	c.cleanupTicker = time.NewTicker(cleanupInterval)
	c.wg.Add(1)

	go func() {
		defer c.wg.Done() //use defer so it will panic-safe if something goes wrong
		for {
			select {
			case <-c.cleanupTicker.C:
				log.Printf("cleanup called, checking expired items every %v", cleanupInterval)
				c.deleteExpired()
			case <-c.stopCleanup: //stop the loop
				log.Println("cleanup stopped")
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
	//check if any item has expired > now
	for key, item := range c.data {
		if now.After(item.expiration) {
			log.Printf("delete expired item %s", key)
			delete(c.data, key)
		}
	}
}

// Set stores a value in the cache with default TTL
func (c *TTLCache) SetWithDefaultTTL(key string, value interface{}) {
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
	log.Printf("Set %s to %v with TTL %v", key, value, ttl)
}

// Get retrieves a value from the cache if it exists and hasn't expired
func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if an item has expired, prevent returning expired items
	// For memory-critical applications, consider (delete-on-get).
	if time.Now().After(item.expiration) {
		return nil, false
	}

	log.Printf("Get %s from cache success", key)
	return item.value, true
}

// Delete removes a value from the cache
func (c *TTLCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("Delete %s from cache", key)
	delete(c.data, key)
}

// Stop stops the background cleanup goroutine
func (c *TTLCache) Stop() {
	c.cleanupTicker.Stop() // Stop ticker first
	close(c.stopCleanup)   // Close instead of send
	c.wg.Wait()            // ← Wait for goroutine to finish
}

// Clear removes all entries from the cache
func (c *TTLCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*cacheItem)
}
