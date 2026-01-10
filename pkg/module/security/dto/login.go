package dto

// LoginRequest represents the HTTP request to authenticate a user.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response with JWT token.
type LoginResponse struct {
	Token string `json:"token"`
}
