package security

import (
	"net/http"
	"strings"
)

// NewJWTAuthMiddleware creates a middleware that extracts and validates JWT tokens.
func NewJWTAuthMiddleware(jwtService JWTService, cfg SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ExtractToken(r, cfg.TokenSource)
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)

				return
			}

			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)

				return
			}

			user := &AuthUser{
				ID:    claims.Subject,
				Roles: claims.Roles,
			}

			r = SetAuthUser(r, user)
			next.ServeHTTP(w, r)
		})
	}
}

// ExtractToken extracts the JWT token from the request based on configured sources.
// It tries each configured source in order and returns the first token found.
func ExtractToken(r *http.Request, cfg TokenSourceConfig) string {
	sources := cfg.Sources
	if len(sources) == 0 {
		sources = []string{"header", "cookie"}
	}

	for _, source := range sources {
		switch source {
		case "header":
			if token := ExtractTokenFromHeader(r, cfg.HeaderName); token != "" {
				return token
			}
		case "cookie":
			if token := ExtractTokenFromCookie(r, cfg.CookieName); token != "" {
				return token
			}
		}
	}

	return ""
}

// ExtractTokenFromHeader extracts token from Authorization header.
// Supports "Bearer <token>" format.
// If headerName is empty, defaults to "Authorization".
func ExtractTokenFromHeader(r *http.Request, headerName string) string {
	if headerName == "" {
		headerName = "Authorization"
	}

	authHeader := r.Header.Get(headerName)
	if authHeader == "" {
		return ""
	}

	// Support "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// ExtractTokenFromCookie extracts token from cookie.
// If cookieName is empty, defaults to "access_token".
func ExtractTokenFromCookie(r *http.Request, cookieName string) string {
	if cookieName == "" {
		cookieName = "access_token"
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
}
