package main

import (
	"fmt"
	"time"

	"question3/cache"
)

func main() {
	fmt.Println("=== Simple Cache Example ===")
	simpleCache := cache.NewSimpleCache()

	// Set and Get
	simpleCache.Set("user:1", "John Doe")
	simpleCache.Set("user:2", "Jane Smith")
	simpleCache.Set("count", 42)

	if value, exists := simpleCache.Get("user:1"); exists {
		fmt.Printf("Found: %v\n", value)
	}

	if value, exists := simpleCache.Get("count"); exists {
		fmt.Printf("Count: %v\n", value)
	}

	// Delete
	simpleCache.Delete("user:1")
	if _, exists := simpleCache.Get("user:1"); !exists {
		fmt.Println("user:1 has been deleted")
	}

	fmt.Println("\n=== TTL Cache Example ===")
	ttlCache := cache.NewTTLCache(2 * time.Second)
	defer ttlCache.Stop()

	// Set with default TTL
	ttlCache.Set("session:abc", "active")
	fmt.Println("Session set with 2-second TTL")

	// Immediate get
	if value, exists := ttlCache.Get("session:abc"); exists {
		fmt.Printf("Session status: %v\n", value)
	}

	// Wait 1 second
	fmt.Println("\nWaiting 1 second...")
	time.Sleep(1 * time.Second)
	if value, exists := ttlCache.Get("session:abc"); exists {
		fmt.Printf("Session still active: %v\n", value)
	}

	// Wait another 1.5 seconds (total 2.5 seconds)
	fmt.Println("\nWaiting another 1.5 seconds...")
	time.Sleep(1500 * time.Millisecond)
	if _, exists := ttlCache.Get("session:abc"); !exists {
		fmt.Println("Session has expired")
	}

	fmt.Println("\n=== Custom TTL Example ===")
	// Set with custom TTL
	ttlCache.SetWithTTL("temp:data", "short-lived", 500*time.Millisecond)
	fmt.Println("Set temp data with 500ms TTL")

	if value, exists := ttlCache.Get("temp:data"); exists {
		fmt.Printf("Temp data: %v\n", value)
	}

	time.Sleep(600 * time.Millisecond)
	if _, exists := ttlCache.Get("temp:data"); !exists {
		fmt.Println("Temp data has expired")
	}

	fmt.Println("\n=== Complex Data Types Example ===")
	type User struct {
		ID    int
		Name  string
		Email string
	}

	user := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	simpleCache.Set("user:object", user)

	if value, exists := simpleCache.Get("user:object"); exists {
		if u, ok := value.(User); ok {
			fmt.Printf("User: %s (%s)\n", u.Name, u.Email)
		}
	}

	fmt.Println("\n=== Cache Statistics ===")
	ttlCache.Clear()
	for i := 0; i < 10; i++ {
		ttlCache.Set(fmt.Sprintf("key:%d", i), i*10)
	}
	fmt.Printf("Cache size: %d items\n", ttlCache.Size())
}
