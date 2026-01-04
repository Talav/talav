package main

import (
	"context"
	"net/http"
	"time"

	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxhttpserver"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
)

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Body struct {
		Name  string `json:"name" validate:"required,min=2" doc:"User's full name"`
		Email string `json:"email" validate:"required,email" doc:"User's email address"`
	}
}

// CreateUserResponse represents the response for creating a user.
type CreateUserResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID        string `json:"id" doc:"Generated user ID"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}
}

// GetUserRequest represents the request for getting a user by ID.
type GetUserRequest struct {
	ID string `schema:"id,location=path" doc:"User ID"`
}

// GetUserResponse represents the response for getting a user.
type GetUserResponse struct {
	Status int `json:"-"`
	Body   struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}
}

// HealthRequest represents the health check request.
type HealthRequest struct{}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status int `json:"-"`
	Body   struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}
}

func main() {
	fx.New(
		fxconfig.FxConfigModule,
		fxlogger.FxLoggerModule,
		fxhttpserver.FxHTTPServerModule,
		fx.Invoke(RegisterRoutes),
	).Run()
}

// RegisterRoutes registers all API routes with Zorya.
func RegisterRoutes(api zorya.API) {
	// Health check endpoint
	zorya.Register(api, zorya.BaseRoute{
		Method: http.MethodGet,
		Path:   "/health",
		Operation: &zorya.Operation{
			Summary:     "Health check",
			Description: "Returns the health status of the API",
			Tags:        []string{"system"},
		},
	}, healthHandler)

	// Create user endpoint
	zorya.Register(api, zorya.BaseRoute{
		Method: http.MethodPost,
		Path:   "/users",
		Operation: &zorya.Operation{
			Summary:     "Create a new user",
			Description: "Creates a new user with the provided information",
			Tags:        []string{"users"},
		},
	}, createUserHandler)

	// Get user endpoint
	zorya.Register(api, zorya.BaseRoute{
		Method: http.MethodGet,
		Path:   "/users/{id}",
		Operation: &zorya.Operation{
			Summary:     "Get user by ID",
			Description: "Retrieves a user by their unique identifier",
			Tags:        []string{"users"},
		},
	}, getUserHandler)
}

// healthHandler handles health check requests.
func healthHandler(ctx context.Context, input *HealthRequest) (*HealthResponse, error) {
	return &HealthResponse{
		Status: http.StatusOK,
		Body: struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

// createUserHandler handles user creation requests.
func createUserHandler(ctx context.Context, input *CreateUserRequest) (*CreateUserResponse, error) {
	// In a real application, you would save to a database here
	userID := "user-" + time.Now().Format("20060102150405")

	return &CreateUserResponse{
		Status: http.StatusCreated,
		Body: struct {
			ID        string `json:"id" doc:"Generated user ID"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        userID,
			Name:      input.Body.Name,
			Email:     input.Body.Email,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

// getUserHandler handles user retrieval requests.
func getUserHandler(ctx context.Context, input *GetUserRequest) (*GetUserResponse, error) {
	// In a real application, you would fetch from a database here
	// For this example, we'll return a mock user
	if input.ID == "" {
		return &GetUserResponse{
			Status: http.StatusNotFound,
			Body: struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				CreatedAt string `json:"created_at"`
			}{},
		}, nil
	}

	return &GetUserResponse{
		Status: http.StatusOK,
		Body: struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        input.ID,
			Name:      "Example User",
			Email:     "user@example.com",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}
