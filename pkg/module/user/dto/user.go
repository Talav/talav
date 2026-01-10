package dto

// CreateUserRequest represents the HTTP request to create a user.
type CreateUserRequest struct {
	Name     string   `json:"name" validate:"required,min=2"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,password"`
	Roles    []string `json:"roles,omitempty"`
}

// UserResponse represents the HTTP response for user data.
type UserResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Roles   []string `json:"roles"`
	IsAdmin bool     `json:"is_admin"`
}

// CreateUserOutput represents the output for user creation.
type CreateUserOutput struct {
	Status int          `status:"201"`
	Body   UserResponse `body:"structured"`
}

// GetUserInput represents the input for getting a user.
type GetUserInput struct {
	ID string `schema:"id,location=path"`
}

// GetUserOutput represents the output for getting a user.
type GetUserOutput struct {
	Body UserResponse `body:"structured"`
}

// ListUsersInput represents the input for listing users.
type ListUsersInput struct {
	Cursor string `schema:"cursor,location=query"`
	Limit  int    `schema:"limit,location=query" default:"10"`
	Email  string `schema:"email,location=query"`
	Name   string `schema:"name,location=query"`
}

// ListUsersOutput represents the output for listing users.
type ListUsersOutput struct {
	Body ListUsersResponse `body:"structured"`
}

// ListUsersResponse represents the response body for listing users.
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}
