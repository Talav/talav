package fxhttpserver

import (
	"net/http"
	"sort"
)

// middlewareEntry represents a registered middleware with its priority.
type middlewareEntry struct {
	// middleware is the middleware function to apply.
	middleware func(http.Handler) http.Handler

	// priority determines execution order. Lower values execute first.
	// middlewares with the same priority execute in registration order.
	priority int

	// name is an optional name for debugging and logging.
	name string

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
func sortMiddlewares(entries []middlewareEntry) []middlewareEntry {
	sorted := make([]middlewareEntry, len(entries))
	copy(sorted, entries)

	// Sort by priority first, then by registration order for equal priorities
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].priority != sorted[j].priority {
			return sorted[i].priority < sorted[j].priority
		}

		return sorted[i].order < sorted[j].order
	})

	return sorted
}
