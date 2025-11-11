package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//prevent race condition
//sync.RWMutex - allows multiple concurrent reads (RLock) but exclusive writes (Lock)
//mu.Lock() in Create, Update, Delete - ensures only one goroutine modifies the map at a time
//mu.RLock() in Get - allows concurrent reads without blocking other reads

// User represents a user in the system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserStore manages user data with thread-safe operations
// ini akan di akses bareng2 , maka perlu di lindungi dengan mutex
// Goroutine 1 (POST /users)
//s.users[s.nextID] = user  // ✍️ WRITE
// Goroutine 2 (GET /users/1)
//user, exists := s.users[id]  // 👁️ READ
// Goroutine 3 (PUT /users/2)
//s.users[id].Name = name  // ✍️ WRITE
// Goroutine 4 (DELETE /users/3)

type UserStore struct {
	users  map[int]*User
	nextID int
	mu     sync.RWMutex
}

// NewUserStore creates a new UserStore instance
func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// Create adds a new user to the store
func (s *UserStore) Create(name, email string) (*User, error) {
	s.mu.Lock()
	//s.mu.Lock() memastikan hanya 1 goroutine yang bisa menjalankan kode ini pada satu waktu
	//Jadi tidak akan ada 2 user dengan ID yang sama
	defer s.mu.Unlock()

	//check if email already exists
	for _, user := range s.users {
		log.Printf("Created user: %v\n", *user)
		if user.Email == email {
			return nil, fmt.Errorf("email already exists")
		}
	}

	user := &User{
		ID:    s.nextID,
		Name:  name,
		Email: email,
	}
	s.users[s.nextID] = user //← Multiple goroutines writing here
	s.nextID++
	log.Printf("Created user: %v\n", user)

	return user, nil
}

// Get retrieves a user by ID
func (s *UserStore) Get(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	return user, exists
}

// Update modifies an existing user
func (s *UserStore) Update(id int, name, email string) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, false
	}

	user.Name = name
	user.Email = email
	return user, true
}

// Delete removes a user from the store
func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.users[id]
	if !exists {
		return false
	}

	delete(s.users, id)
	return true
}

// APIError represents an error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	// Email validation regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// validateName validates user name
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	if len(name) > 100 {
		return fmt.Errorf("name must not exceed 100 characters")
	}
	return nil
}

// validateEmail validates user email
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, error, message string) {
	respondWithJSON(w, status, APIError{
		Error:   error,
		Message: message,
	})
}

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	store *UserStore
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(store *UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload")
		return
	}

	// Validate name
	if err := validateName(req.Name); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Validate email
	if err := validateEmail(req.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Create user
	user, err := h.store.Create(strings.TrimSpace(req.Name), strings.TrimSpace(req.Email))
	if err != nil {
		if err.Error() == "email already exists" {
			respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "internal_error", "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only GET method is allowed")
		return
	}

	// Extract ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid URL format")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid user ID")
		return
	}

	user, exists := h.store.Get(id)
	if !exists {
		respondWithError(w, http.StatusNotFound, "not_found", "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only PUT method is allowed")
		return
	}

	// Extract ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid URL format")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload")
		return
	}

	// Validate name
	if err := validateName(req.Name); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Validate email
	if err := validateEmail(req.Email); err != nil {
		respondWithError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Update user
	user, exists := h.store.Update(id, strings.TrimSpace(req.Name), strings.TrimSpace(req.Email))
	if !exists {
		respondWithError(w, http.StatusNotFound, "not_found", "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only DELETE method is allowed")
		return
	}

	// Extract ID from URL
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid URL format")
		return
	}

	id, err := strconv.Atoi(pathParts[1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid user ID")
		return
	}

	if !h.store.Delete(id) {
		respondWithError(w, http.StatusNotFound, "not_found", "User not found")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Router handles routing logic
func (h *UserHandler) Router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// POST /users
	if path == "/users" && r.Method == http.MethodPost {
		h.CreateUser(w, r)
		return
	}

	// GET, PUT, DELETE /users/:id
	if strings.HasPrefix(path, "/users/") {
		switch r.Method {
		case http.MethodGet:
			h.GetUser(w, r)
		case http.MethodPut:
			h.UpdateUser(w, r)
		case http.MethodDelete:
			h.DeleteUser(w, r)
		default:
			respondWithError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		}
		return
	}

	respondWithError(w, http.StatusNotFound, "not_found", "Endpoint not found")
}

func main() {
	store := NewUserStore()
	handler := NewUserHandler(store)

	http.HandleFunc("/", handler.Router)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST   /users       - Create a new user")
	fmt.Println("  GET    /users/:id   - Get user by ID")
	fmt.Println("  PUT    /users/:id   - Update user by ID")
	fmt.Println("  DELETE /users/:id   - Delete user by ID")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
