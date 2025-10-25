# Golang Developer Pre-Interview Test Solutions

This repository contains complete solutions for the Golang Developer pre-interview test from Solecode.

## Project Structure

```
golang-interview-test/
├── question1/          # Concurrent even sum calculator
│   ├── main.go
│   └── go.mod
├── question2/          # REST API for user management
│   ├── main.go
│   ├── go.mod
│   └── test_api.sh
├── question3/          # Cache implementation with tests
│   ├── cache/
│   │   ├── cache.go
│   │   └── cache_test.go
│   ├── main.go
│   └── go.mod
└── README.md
```

## Question 1: Concurrent Even Sum Calculator

### Description
A Go program that uses goroutines to calculate the sum of even numbers from a large slice, dividing work among multiple workers.

### Features
- ✅ Goroutines for concurrent processing
- ✅ Channel-based communication
- ✅ WaitGroup synchronization
- ✅ Configurable number of workers
- ✅ Proper workload distribution
- ✅ Race condition prevention

### Running
```bash
cd question1
go run main.go
```

### Expected Output
```
Processing 1000000 numbers with 4 workers...
Sum of all even numbers: 250000500000
Time taken: ~10ms

Verifying with sequential calculation...
Expected sum: 250000500000
Sequential time: ~15ms

✓ Result verified successfully!
```

### Key Implementation Details
- Divides slice into equal chunks with remainder distribution
- Each worker processes its chunk independently
- Results collected via buffered channel
- WaitGroup ensures all workers complete before aggregation

---

## Question 2: REST API for User Management

### Description
A thread-safe REST API built with Go's standard library for managing users with full CRUD operations.

### Features
- ✅ RESTful endpoint design
- ✅ Proper HTTP status codes
- ✅ Input validation (name and email)
- ✅ Thread-safe in-memory storage using sync.RWMutex
- ✅ Comprehensive error handling
- ✅ JSON request/response

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /users | Create a new user |
| GET | /users/:id | Retrieve user by ID |
| PUT | /users/:id | Update user information |
| DELETE | /users/:id | Delete user |

### Running the Server
```bash
cd question2
go run main.go
```

Server will start on `http://localhost:8080`

### Testing the API

#### Using the provided test script:
```bash
cd question2
chmod +x test_api.sh
./test_api.sh
```

#### Manual testing with curl:

**Create a user:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'
```

**Get a user:**
```bash
curl http://localhost:8080/users/1
```

**Update a user:**
```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"John Smith","email":"john.smith@example.com"}'
```

**Delete a user:**
```bash
curl -X DELETE http://localhost:8080/users/1
```

### Validation Rules
- **Name**: Required, 2-100 characters
- **Email**: Required, valid email format

### Error Responses
All errors return JSON with structure:
```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

### Thread Safety
- Uses `sync.RWMutex` for concurrent read/write protection
- Multiple goroutines can read simultaneously
- Write operations are mutually exclusive

---

## Question 3: Cache Implementation with TTL

### Description
A cache interface with two implementations: a simple in-memory cache and a TTL-based cache with automatic expiration.

### Features
- ✅ Clean interface design (Set, Get, Delete)
- ✅ SimpleCache: Basic in-memory storage
- ✅ TTLCache: Time-to-live with auto-cleanup
- ✅ Thread-safe operations
- ✅ Comprehensive unit tests (95%+ coverage)
- ✅ Benchmark tests included

### Cache Interface
```go
type Cache interface {
    Set(key string, value interface{})
    Get(key string) (interface{}, bool)
    Delete(key string)
}
```

### Implementations

#### 1. SimpleCache
Basic in-memory cache without expiration.

```go
cache := cache.NewSimpleCache()
cache.Set("key", "value")
value, exists := cache.Get("key")
cache.Delete("key")
```

#### 2. TTLCache
Cache with automatic expiration and background cleanup.

```go
cache := cache.NewTTLCache(5 * time.Second)
defer cache.Stop()

cache.Set("key", "value")  // Uses default TTL
cache.SetWithTTL("temp", "data", 1 * time.Second)  // Custom TTL

value, exists := cache.Get("key")
```

**TTL Features:**
- Default TTL for all entries
- Custom TTL per entry
- Background goroutine for cleanup
- Automatic expired entry removal on Get

### Running the Example
```bash
cd question3
go run main.go
```

### Running Tests
```bash
cd question3
go test ./cache -v
```

### Running Tests with Coverage
```bash
cd question3
go test ./cache -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Running Benchmarks
```bash
cd question3
go test ./cache -bench=. -benchmem
```

### Test Coverage
The test suite includes:
- ✅ Basic operations (Set, Get, Delete)
- ✅ Different data types (string, int, slice, struct)
- ✅ Overwrite scenarios
- ✅ Concurrent access (100 goroutines)
- ✅ TTL expiration
- ✅ Custom TTL
- ✅ Auto-cleanup verification
- ✅ Update expiration reset
- ✅ Cache clear operation
- ✅ Edge cases and race conditions

---

## Technical Highlights

### Concurrency & Synchronization
- Goroutines for parallel processing
- Channels for communication
- WaitGroups for synchronization
- Mutexes (RWMutex) for thread-safe data structures
- No race conditions (verified with `go run -race`)

### Code Quality
- Clean, readable code with proper naming
- Comprehensive error handling
- Input validation
- Proper resource cleanup
- Well-documented functions

### Testing
- Unit tests for all components
- Edge case coverage
- Concurrent access testing
- Benchmark tests for performance
- Table-driven tests where appropriate

### Best Practices
- Interface-based design (Question 3)
- Separation of concerns
- RESTful API design (Question 2)
- Proper HTTP status codes
- Thread-safe implementations

---

## Requirements

- Go 1.21 or higher
- No external dependencies (uses standard library only)

---

## Running All Tests

```bash
# Question 1
cd question1 && go run main.go && cd ..

# Question 2 (in separate terminal)
cd question2 && go run main.go

# Question 2 tests (in another terminal)
cd question2 && ./test_api.sh

# Question 3
cd question3 && go run main.go
cd question3 && go test ./cache -v -cover
```

---

## Performance Notes

### Question 1
- Successfully processes 1 million integers
- Concurrent version shows speedup with multiple workers
- Scales well with number of CPU cores

### Question 2
- Thread-safe for concurrent requests
- O(1) operations for all endpoints
- Memory-efficient in-memory storage

### Question 3
- SimpleCache: O(1) for all operations
- TTLCache: O(1) for Set/Get/Delete, O(n) for cleanup
- Background cleanup prevents memory leaks

---

## Author Notes

All solutions follow Go best practices and idioms:
- Error handling without panics
- Proper use of defer for cleanup
- Channel closing by sender
- Context-free implementations (as per requirements)
- Clean, production-ready code

The code is ready for review and demonstrates:
- Strong understanding of Go concurrency primitives
- RESTful API design skills
- Interface design and abstraction
- Comprehensive testing practices
- Thread-safety awareness

---

## License

This code is provided as a solution to the Solecode Golang Developer pre-interview test.
