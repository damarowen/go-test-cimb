package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestSimpleCache_BasicOperations tests basic cache operations
func TestSimpleCache_BasicOperations(t *testing.T) {
	cache := NewSimpleCache()

	// Test Set and Get
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Get non-existent key
	_, exists = cache.Get("nonexistent")
	if exists {
		t.Error("Expected nonexistent key to not exist")
	}

	// Test Delete
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	if exists {
		t.Error("Expected key1 to be deleted")
	}
}

// TestSimpleCache_ConcurrentAccess tests thread safety
func TestSimpleCache_ConcurrentAccess(t *testing.T) {
	cache := NewSimpleCache()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent mixed operations
	wg.Add(numGoroutines * 3)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id%26)
			cache.Set(key, id)
		}(i)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id%26)
			cache.Get(key)
		}(i)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id%26)
			cache.Delete(key)
		}(i)
	}
	wg.Wait()
}

// TestTTLCache_BasicOperations tests basic TTL cache operations
func TestTTLCache_BasicOperations(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()

	// Test Set and Get
	cache.SetWithDefaultTTL("key1", "value1")
	value, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Delete
	cache.Delete("key1")
	_, exists = cache.Get("key1")
	if exists {
		t.Error("Expected key1 to be deleted")
	}
}

// TestTTLCache_NoExpiration tests that items don't expire prematurely
func TestTTLCache_NoExpiration(t *testing.T) {
	ttl := 200 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	cache.SetWithDefaultTTL("key", "value")

	// Check 30ms times before expiration
	time.Sleep(30 * time.Millisecond)
	_, exists := cache.Get("key")
	if !exists {
		t.Error("Expected key to still exist")
	}
}

// TestTTLCache_AutoCleanup tests automatic cleanup of expired entries
func TestTTLCache_AutoCleanup(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	// Add multiple entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", 1)
		cache.SetWithDefaultTTL(key, i)
	}

	// Wait for cleanup to run (TTL + cleanup interval)
	time.Sleep(ttl + 200*time.Millisecond)
}

// TestTTLCache_ConcurrentAccess tests thread safety with TTL
func TestTTLCache_ConcurrentAccess(t *testing.T) {
	cache := NewTTLCache(500 * time.Millisecond)
	defer cache.Stop()

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 50

	// Concurrent operations
	wg.Add(numGoroutines * 3)
	for i := 0; i < numGoroutines; i++ {
		// Writers
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key%d", id%26)
				cache.SetWithDefaultTTL(key, id*numOperations+j)
				time.Sleep(time.Millisecond)
			}
		}(i)

		// Readers
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key%d", id%26)
				cache.Get(key)
				time.Sleep(time.Millisecond)
			}
		}(i)

		// Deleters
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key%d", id%26)
				cache.Delete(key)
				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}

// TestTTLCache_UpdateExpiration tests that updating a key resets expiration
func TestTTLCache_UpdateExpiration(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	cache.SetWithDefaultTTL("key", "value1")

	// Wait half the TTL
	time.Sleep(ttl / 2)

	// Update the key
	cache.SetWithDefaultTTL("key", "value2")

	// Wait another half TTL (should still exist)
	time.Sleep(ttl / 2)

	value, exists := cache.Get("key")
	if !exists {
		t.Error("Expected key to still exist after update")
	}
	if value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}

	// Wait for full TTL from last update
	time.Sleep(ttl)

	_, exists = cache.Get("key")
	if exists {
		t.Error("Expected key to be expired")
	}
}

// TestTTLCache_DeleteExpired tests the cleanup of expired entries
func TestTTLCache_DeleteExpired(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	// Add 3 items
	cache.SetWithDefaultTTL("item1", "value1")
	cache.SetWithDefaultTTL("item2", "value2")
	cache.SetWithDefaultTTL("item3", "value3")

	// Wait for items to expire
	time.Sleep(150 * time.Millisecond)

	// Trigger cleanup
	cache.deleteExpired()

	// Check all items are gone
	if _, exists := cache.Get("item1"); exists {
		t.Error("item1 should be deleted")
	}
	if _, exists := cache.Get("item2"); exists {
		t.Error("item2 should be deleted")
	}
	if _, exists := cache.Get("item3"); exists {
		t.Error("item3 should be deleted")
	}
}

// TestTTLCache_Clear tests clearing the cache
func TestTTLCache_Clear(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()

	// Add some entries
	cache.SetWithDefaultTTL("key1", "value1")
	cache.SetWithDefaultTTL("key2", "value2")

	// Clear the cache
	cache.Clear()

	// Verify all items are gone
	if _, exists := cache.Get("key1"); exists {
		t.Error("Expected key1 to not exist after clear")
	}
	if _, exists := cache.Get("key2"); exists {
		t.Error("Expected key2 to not exist after clear")
	}
}

// TestGapBehavior validates that expired items are inaccessible during the gap
func TestGapBehavior(t *testing.T) {
	ttlCache := NewTTLCache(5 * time.Second)
	defer ttlCache.Stop()

	// Add item with 2-second TTL
	// Cleanup interval will be 5s/2 = 2.5s
	ttlCache.SetWithTTL("user_id_1", "Alice", 2*time.Second)

	// Should exist immediately
	if val, exists := ttlCache.Get("user_id_1"); !exists {
		t.Error("Expected item to exist immediately after setting")
	} else {
		t.Logf("T+0s: Item exists, value=%v", val)
	}

	// Wait until item expires
	time.Sleep(2100 * time.Millisecond) // T+2.1s (item expired at T+2s)

	// Item should be expired but still in memory (gap period)
	t.Log("T+2.1s: Item has expired (in gap period)")

	// Get() should return nil/false even though item is still in map
	if val, exists := ttlCache.Get("user_id_1"); exists {
		t.Errorf("Expected Get() to return false for expired item during gap, got value=%v", val)
	} else {
		t.Log("✅ Get() correctly returns nil/false for expired item")
	}

	// Check if item is actually still in memory
	ttlCache.mu.RLock()
	_, stillInMap := ttlCache.data["user_id_1"]
	ttlCache.mu.RUnlock()

	if stillInMap {
		t.Log("⚠️ Item is still in memory (gap confirmed)")
	} else {
		t.Log("Item already deleted (no gap detected)")
	}

	// Wait for cleanup to run (cleanup interval is 2.5s)
	time.Sleep(500 * time.Millisecond) // T+2.6s (cleanup should have run)

	// Verify item is now deleted from memory
	ttlCache.mu.RLock()
	_, stillExists := ttlCache.data["user_id_1"]
	ttlCache.mu.RUnlock()

	if stillExists {
		t.Error("Expected item to be deleted from memory after cleanup")
	} else {
		t.Log("✅ Item deleted from memory after cleanup")
	}
}
