package main

import (
	"fmt"
	"time"

	"question3/cache"
)

func main() {
	fmt.Println("=== Simple Cache Example start ===")
	simpleCache := cache.NewSimpleCache()

	// Set and Get
	simpleCache.Set("user:1", "John Doe")

	if value, exists := simpleCache.Get("user:1"); exists {
		fmt.Printf("Found: %v,%v\n", value, "user:1")
	}

	// Delete
	simpleCache.Delete("user:1")
	if _, exists := simpleCache.Get("user:1"); !exists {
		fmt.Println("user:1 has been deleted")
	}

	fmt.Println("\n=== TTL Cache Example start ===")
	ttlCache := cache.NewTTLCache(5 * time.Second) //default TTL is 5 seconds
	defer ttlCache.Stop()

	// Set with default TTL
	ttlCache.SetWithDefaultTTL("session:abc", "active")

	// Immediate get
	if value, exists := ttlCache.Get("session:abc"); exists {
		fmt.Printf("session:abc status: %v\n", value)
	}

	// Wait 1 second
	fmt.Println("\nWaiting 1 second...")
	time.Sleep(1 * time.Second)

	if value, exists := ttlCache.Get("session:abc"); exists {
		fmt.Printf("session:abc status: %v\n", value)
	}
	// Wait another 1.5 seconds (total 2.5 seconds)
	fmt.Println("\nWaiting another 4 seconds...")
	time.Sleep(4000 * time.Millisecond)

	fmt.Println("\n=== Custom TTL Example ===")
	// Set with custom TTL
	ttlCache.SetWithTTL("temp:data", "one-minute", 1*time.Minute)
	ttlCache.SetWithTTL("user_id_1", "Alice", 2*time.Second) // Expires in 2s
	ttlCache.SetWithTTL("user_id_2", "Bob", 30*time.Second)  // Expires in 30s
	time.Sleep(65 * time.Second)
	if _, exists := ttlCache.Get("temp:data"); !exists {
		fmt.Println("Temp data has expired")
	}
	if _, exists := ttlCache.Get("user_id_2"); !exists {
		fmt.Println("Temp user_id_2 has expired")
	}
	ttlCache.Clear()
}
