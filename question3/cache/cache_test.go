package cache

import (
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

// TestSimpleCache_DifferentTypes tests storing different data types
func TestSimpleCache_DifferentTypes(t *testing.T) {
	cache := NewSimpleCache()

	// Test string
	cache.Set("string", "test")
	value, _ := cache.Get("string")
	if value.(string) != "test" {
		t.Error("String value mismatch")
	}

	// Test int
	cache.Set("int", 42)
	value, _ = cache.Get("int")
	if value.(int) != 42 {
		t.Error("Int value mismatch")
	}

	// Test slice
	testSlice := []int{1, 2, 3}
	cache.Set("slice", testSlice)
	value, _ = cache.Get("slice")
	retrievedSlice := value.([]int)
	if len(retrievedSlice) != 3 || retrievedSlice[0] != 1 {
		t.Error("Slice value mismatch")
	}

	// Test struct
	type TestStruct struct {
		Name string
		Age  int
	}
	testStruct := TestStruct{Name: "John", Age: 30}
	cache.Set("struct", testStruct)
	value, _ = cache.Get("struct")
	retrievedStruct := value.(TestStruct)
	if retrievedStruct.Name != "John" || retrievedStruct.Age != 30 {
		t.Error("Struct value mismatch")
	}
}

// TestSimpleCache_Overwrite tests overwriting existing keys
func TestSimpleCache_Overwrite(t *testing.T) {
	cache := NewSimpleCache()

	cache.Set("key", "value1")
	cache.Set("key", "value2")

	value, exists := cache.Get("key")
	if !exists {
		t.Error("Expected key to exist")
	}
	if value != "value2" {
		t.Errorf("Expected value2, got %v", value)
	}
}

// TestSimpleCache_ConcurrentAccess tests thread safety
func TestSimpleCache_ConcurrentAccess(t *testing.T) {
	cache := NewSimpleCache()
	var wg sync.WaitGroup
	numGoroutines := 100
	numOperations := 100

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
				cache.Set(key, id*numOperations+j)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
				cache.Get(key)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent mixed operations
	wg.Add(numGoroutines * 3)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + (id % 26)))
			cache.Set(key, id)
		}(i)
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + (id % 26)))
			cache.Get(key)
		}(i)
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + (id % 26)))
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
	cache.Set("key1", "value1")
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

// TestTTLCache_Expiration tests that entries expire after TTL
func TestTTLCache_Expiration(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	cache.Set("key1", "value1")

	// Should exist immediately
	_, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(ttl + 50*time.Millisecond)

	// Should not exist after expiration
	_, exists = cache.Get("key1")
	if exists {
		t.Error("Expected key1 to be expired")
	}
}

// TestTTLCache_CustomTTL tests setting custom TTL for entries
func TestTTLCache_CustomTTL(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()

	shortTTL := 50 * time.Millisecond
	cache.SetWithTTL("short", "value", shortTTL)

	// Should exist immediately
	_, exists := cache.Get("short")
	if !exists {
		t.Error("Expected short to exist immediately")
	}

	// Wait for short TTL to expire
	time.Sleep(shortTTL + 20*time.Millisecond)

	// Should be expired
	_, exists = cache.Get("short")
	if exists {
		t.Error("Expected short to be expired")
	}
}

// TestTTLCache_NoExpiration tests that items don't expire prematurely
func TestTTLCache_NoExpiration(t *testing.T) {
	ttl := 200 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	cache.Set("key", "value")

	// Check multiple times before expiration
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)
		_, exists := cache.Get("key")
		if !exists {
			t.Errorf("Expected key to exist at check %d", i+1)
		}
	}
}

// TestTTLCache_AutoCleanup tests automatic cleanup of expired entries
func TestTTLCache_AutoCleanup(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewTTLCache(ttl)
	defer cache.Stop()

	// Add multiple entries
	for i := 0; i < 10; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	initialSize := cache.Size()
	if initialSize != 10 {
		t.Errorf("Expected 10 entries, got %d", initialSize)
	}

	// Wait for cleanup to run (TTL + cleanup interval)
	time.Sleep(ttl + 200*time.Millisecond)

	finalSize := cache.Size()
	if finalSize >= initialSize {
		t.Errorf("Expected cleanup to reduce size, initial: %d, final: %d", initialSize, finalSize)
	}
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
				key := string(rune('a' + (id % 26)))
				cache.Set(key, id*numOperations+j)
				time.Sleep(time.Millisecond)
			}
		}(i)

		// Readers
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
				cache.Get(key)
				time.Sleep(time.Millisecond)
			}
		}(i)

		// Deleters
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := string(rune('a' + (id % 26)))
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

	cache.Set("key", "value1")

	// Wait half the TTL
	time.Sleep(ttl / 2)

	// Update the key
	cache.Set("key", "value2")

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

// TestTTLCache_Clear tests clearing the cache
func TestTTLCache_Clear(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()

	// Add entries
	for i := 0; i < 10; i++ {
		cache.Set(string(rune('a'+i)), i)
	}

	if cache.Size() != 10 {
		t.Errorf("Expected 10 entries, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", cache.Size())
	}

	// Verify items don't exist
	for i := 0; i < 10; i++ {
		_, exists := cache.Get(string(rune('a' + i)))
		if exists {
			t.Errorf("Expected key %c to not exist after clear", 'a'+i)
		}
	}
}

// TestCache_Interface tests that both implementations satisfy the Cache interface
func TestCache_Interface(t *testing.T) {
	var _ Cache = (*SimpleCache)(nil)
	var _ Cache = (*TTLCache)(nil)

	testCases := []struct {
		name  string
		cache Cache
	}{
		{"SimpleCache", NewSimpleCache()},
		{"TTLCache", NewTTLCache(5 * time.Second)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache

			// Test Set and Get
			cache.Set("test", "value")
			value, exists := cache.Get("test")
			if !exists {
				t.Error("Expected test key to exist")
			}
			if value != "value" {
				t.Errorf("Expected 'value', got %v", value)
			}

			// Test Delete
			cache.Delete("test")
			_, exists = cache.Get("test")
			if exists {
				t.Error("Expected test key to be deleted")
			}
		})
	}
}

// BenchmarkSimpleCache_Set benchmarks Set operation for SimpleCache
func BenchmarkSimpleCache_Set(b *testing.B) {
	cache := NewSimpleCache()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

// BenchmarkSimpleCache_Get benchmarks Get operation for SimpleCache
func BenchmarkSimpleCache_Get(b *testing.B) {
	cache := NewSimpleCache()
	cache.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

// BenchmarkTTLCache_Set benchmarks Set operation for TTLCache
func BenchmarkTTLCache_Set(b *testing.B) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

// BenchmarkTTLCache_Get benchmarks Get operation for TTLCache
func BenchmarkTTLCache_Get(b *testing.B) {
	cache := NewTTLCache(5 * time.Second)
	defer cache.Stop()
	cache.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}
