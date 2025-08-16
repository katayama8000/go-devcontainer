package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Todo defines the structure for a todo item.
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// ---- In-memory data store ----
var (
	// todos is a map to store todo items.
	todos = make(map[int]Todo)
	// nextID is the counter for the next available ID.
	nextID = 1
	// mu is a Mutex to safely access data from multiple requests.
	mu sync.Mutex
)

// todosHandler handles requests to the /todos endpoint.
func todosHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		// Get all items.
		list := make([]Todo, 0, len(todos))
		for _, todo := range todos {
			list = append(list, todo)
		}
		json.NewEncoder(w).Encode(list)

	case http.MethodPost:
		// Create a new item.
		var newTodo Todo
		if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newTodo.ID = nextID
		nextID++
		todos[newTodo.ID] = newTodo
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTodo)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// todoByIDHandler handles requests to the /todos/{id} endpoint.
func todoByIDHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Get ID from URL path.
	id, err := strconv.Atoi(r.URL.Path[len("/todos/"):])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if _, ok := todos[id]; !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get a single item.
		json.NewEncoder(w).Encode(todos[id])

	case http.MethodPut:
		// Update an item (toggles completed status).
		todo := todos[id]
		todo.Completed = !todo.Completed
		todos[id] = todo
		json.NewEncoder(w).Encode(todo)

	case http.MethodDelete:
		// Delete an item.
		delete(todos, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// Middleware to set Content-Type and route requests.
	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		todosHandler(w, r)
	})
	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		todoByIDHandler(w, r)
	})

	port := ":8080"
	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}