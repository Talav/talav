package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/component/zorya/adapters"
)

// CreateUserInput represents the request body for creating a user.
type CreateUserInput struct {
	Body struct {
		Name  string `schema:"name" json:"name"`
		Email string `schema:"email" json:"email"`
	}
}

// CreateUserOutput represents the response for creating a user.
type CreateUserOutput struct {
	Status int `json:"-"`
	Body   struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"body"`
}

func main() {
	router := chi.NewMux()
	adapter := adapters.NewChi(router)
	api := zorya.NewAPI(adapter)

	// Register POST /users endpoint
	zorya.Post(api, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
		output := &CreateUserOutput{
			Status: http.StatusCreated,
		}
		output.Body.ID = 1
		output.Body.Name = input.Body.Name
		output.Body.Email = input.Body.Email

		return output, nil
	})

	log.Println("=== Zorya API Test ===")
	log.Println("\n1. Registered endpoints:")
	log.Println("   POST /users - Create a new user")

	// Test the endpoint using httptest
	log.Println("\n2. Testing POST /users endpoint...")

	requestBody := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	log.Printf("   Request: POST /users\n")
	log.Printf("   Request Body: %s\n", string(jsonData))
	log.Printf("   Status: %d %s\n", resp.StatusCode, resp.Status)
	log.Printf("   Response Body:\n")

	// Pretty print JSON response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "     "); err != nil {
		log.Printf("   %s\n", string(body))
	} else {
		log.Printf("   %s\n", prettyJSON.String())
	}

	log.Println("\n3. Test completed successfully!")
	log.Println("\nTo use in a real server, start the router with:")
	log.Println("   http.ListenAndServe(\":8080\", router)")
}
