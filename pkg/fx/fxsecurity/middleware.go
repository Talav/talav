package fxsecurity

import (
	"net/http"
	"strings"

	"github.com/talav/talav/pkg/component/security"
)

// JWTAuthMiddleware creates a middleware that extracts and validates JWT tokens.
func JWTAuthMiddleware(jwtService security.JWTService, cfg security.SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r, cfg.TokenSource)
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user := &security.AuthUser{
				ID:    claims.Subject,
				Roles: claims.Roles,
			}

			r = security.SetAuthUser(r, user)
			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts the JWT token from the request based on configured sources.
func extractToken(r *http.Request, cfg security.TokenSourceConfig) string {
	sources := cfg.Sources
	if len(sources) == 0 {
		sources = []string{"header", "cookie"}
	}

	for _, source := range sources {
		switch source {
		case "header":
			if token := extractTokenFromHeader(r, cfg.HeaderName); token != "" {
				return token
			}
		case "cookie":
			if token := extractTokenFromCookie(r, cfg.CookieName); token != "" {
				return token
			}
		}
	}

	return ""
}

// extractTokenFromHeader extracts token from Authorization header.
func extractTokenFromHeader(r *http.Request, headerName string) string {
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

// extractTokenFromCookie extracts token from cookie.
func extractTokenFromCookie(r *http.Request, cookieName string) string {
	if cookieName == "" {
		cookieName = "access_token"
	}

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
}

