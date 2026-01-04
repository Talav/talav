package fxhttpserver

import (
	"net/http"
	"sort"
)

// MiddlewareEntry represents a registered middleware with its priority.
type MiddlewareEntry struct {
	// Middleware is the middleware function to apply.
	Middleware func(http.Handler) http.Handler

	// Priority determines execution order. Lower values execute first.
	// Middlewares with the same priority execute in registration order.
	Priority int

	// Name is an optional name for debugging and logging.
	Name string
}

const (
	// PriorityRequestID is the priority for RequestID middleware.
	PriorityRequestID = 100

	// PriorityHTTPLog is the priority for HTTPLog middleware.
	PriorityHTTPLog = 200

	// PriorityBeforeZorya is the default priority for user middlewares.
	// Use this or higher for middlewares that should run after infrastructure.
	PriorityBeforeZorya = 250
)

// sortMiddlewares sorts middleware entries by priority (lower first).
// Entries with the same priority maintain their relative order (stable sort).
func sortMiddlewares(entries []MiddlewareEntry) []MiddlewareEntry {
	sorted := make([]MiddlewareEntry, len(entries))
	copy(sorted, entries)

	// Sort by priority, maintaining stable order for equal priorities
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})

	return sorted
}
