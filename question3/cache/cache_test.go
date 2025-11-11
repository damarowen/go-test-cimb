package cache

import (
	"testing"
	"time"
)

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
