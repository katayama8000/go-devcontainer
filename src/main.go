package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
)

// Define the Go struct for our Todo data
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// In-memory data store (populated with some sample data)
var todos = []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a GraphQL Server", Completed: false},
	{ID: 3, Title: "Buy milk", Completed: true},
}

func main() {
	// === 1. Define GraphQL Types ===

	// Define the 'Todo' object type for GraphQL
	todoType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Todo",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.Int},
			"title":     &graphql.Field{Type: graphql.String},
			"completed": &graphql.Field{Type: graphql.Boolean},
		},
	})

	// Define the Root Query object
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// Field to get a single todo by ID
			"todo": &graphql.Field{
				Type:        todoType,
				Description: "Get a single todo by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if ok {
						for _, todo := range todos {
							if todo.ID == id {
								return todo, nil
							}
						}
					}
					return nil, nil // Not found
				},
			},
			// Field to get all todos
			"todos": &graphql.Field{
				Type:        graphql.NewList(todoType),
				Description: "Get all todos",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return todos, nil
				},
			},
		},
	})

	// Define the Root Mutation object
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			// Field to update a todo
			"updateTodo": &graphql.Field{
				Type:        todoType, // Returns the updated Todo
				Description: "Update a todo's completed status by its ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"completed": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Boolean),
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id, _ := params.Args["id"].(int)
					completed, _ := params.Args["completed"].(bool)
					var updatedTodo Todo

					// Find the todo and update it
					for i, t := range todos {
						if t.ID == id {
							todos[i].Completed = completed
							updatedTodo = todos[i]
							break
						}
					}
					return updatedTodo, nil
				},
			},
		},
	})

	// === 2. Create the Schema ===
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType, // Add Mutation to the schema
	})
	if err != nil {
		log.Fatalf("Failed to create new schema, error: %v", err)
	}

	// === 3. Create an HTTP handler ===
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Execute the GraphQL query or mutation
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: r.URL.Query().Get("query"), // Support GET for simple queries
		})

		// For POST requests, read the JSON body
		if r.Method == http.MethodPost {
			var params struct {
				Query string `json:"query"`
			}
			if err := json.NewDecoder(r.Body).Decode(&params); err == nil {
				result = graphql.Do(graphql.Params{
					Schema:        schema,
					RequestString: params.Query,
				})
			}
		}

		// Handle any errors from execution
		if len(result.Errors) > 0 {
			log.Printf("GraphQL errors: %v", result.Errors)
		}

		// Write the JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	port := ":8080"
	log.Printf("GraphQL server is running on http://localhost%s/graphql", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}