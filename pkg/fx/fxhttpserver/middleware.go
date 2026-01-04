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

	// order is the registration order for stable sorting when priorities are equal.
	// Lower values were registered first.
	order int
}

const (
	// PriorityRequestID is the priority for RequestID middleware.
	PriorityRequestID = 100

	// PriorityCORS is the priority for CORS middleware.
	PriorityCORS = 150

	// PriorityHTTPLog is the priority for HTTPLog middleware.
	PriorityHTTPLog = 200

	// PriorityBeforeZorya is the default priority for user middlewares.
	// Use this or higher for middlewares that should run after infrastructure.
	PriorityBeforeZorya = 250
)

// sortMiddlewares sorts middleware entries by priority (lower first).
// Entries with the same priority are sorted by registration order (order field).
func sortMiddlewares(entries []MiddlewareEntry) []MiddlewareEntry {
	sorted := make([]MiddlewareEntry, len(entries))
	copy(sorted, entries)

	// Sort by priority first, then by registration order for equal priorities
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}

		return sorted[i].order < sorted[j].order
	})

	return sorted
}
