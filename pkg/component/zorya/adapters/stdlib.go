//go:build go1.22

package adapters

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/talav/talav/pkg/component/zorya"
)

// Mux is an interface for HTTP muxes that support Go 1.22+ routing.
// This includes http.ServeMux and any compatible mux implementation.
type Mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// StdlibAdapter implements zorya.Adapter for net/http (Go 1.22+) router.
type StdlibAdapter struct {
	mux    Mux
	prefix string
}

// NewStdlib creates a new adapter for the given HTTP mux (Go 1.22+).
//
//	mux := http.NewServeMux()
//	adapter := adapters.NewStdlib(mux)
//	api := zorya.NewAPI(adapter)
func NewStdlib(mux Mux) *StdlibAdapter {
	return &StdlibAdapter{mux: mux, prefix: ""}
}

// NewStdlibWithPrefix creates a new adapter with a URL prefix.
// This behaves similar to router groups, adding the prefix before each route path.
//
//	mux := http.NewServeMux()
//	adapter := adapters.NewStdlibWithPrefix(mux, "/api")
//	api := zorya.NewAPI(adapter)
func NewStdlibWithPrefix(mux Mux, prefix string) *StdlibAdapter {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}

	return &StdlibAdapter{mux: mux, prefix: prefix}
}

func (a *StdlibAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *StdlibAdapter) Handle(route *zorya.BaseRoute, handler func(zorya.Context)) {
	// Go 1.22+ ServeMux uses "METHOD PATH" pattern format
	pattern := strings.ToUpper(route.Method) + " " + a.prefix + route.Path

	a.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Extract path parameters using Go 1.22+ PathValue method
		routerParams := make(map[string]string)

		// Extract parameter names from route pattern
		pathSegments := strings.Split(route.Path, "/")
		for _, segment := range pathSegments {
			if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
				paramName := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
				// Go 1.22+ PathValue extracts path parameters automatically
				if val := r.PathValue(paramName); val != "" {
					routerParams[paramName] = val
				}
			}
		}

		handler(&httpContext{
			route:        route,
			r:            r,
			w:            w,
			routerParams: routerParams,
		})
	})
}

type httpContext struct {
	route        *zorya.BaseRoute
	r            *http.Request
	w            http.ResponseWriter
	status       int
	routerParams map[string]string
}

func (c *httpContext) Context() context.Context {
	return c.r.Context()
}

func (c *httpContext) Request() *http.Request {
	return c.r
}

func (c *httpContext) RouterParams() map[string]string {
	return c.routerParams
}

func (c *httpContext) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *httpContext) SetReadDeadline(t time.Time) error {
	return zorya.SetReadDeadline(c.w, t)
}

func (c *httpContext) SetBodyLimit(limit int64) {
	if limit > 0 {
		c.r.Body = http.MaxBytesReader(c.w, c.r.Body, limit)
	}
}

// ResponseWriter implementation

func (c *httpContext) SetStatus(status int) {
	c.status = status
	c.w.WriteHeader(status)
}

func (c *httpContext) SetHeader(name, value string) {
	c.w.Header().Set(name, value)
}

func (c *httpContext) AppendHeader(name, value string) {
	c.w.Header().Add(name, value)
}

func (c *httpContext) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *httpContext) BodyWriter() http.ResponseWriter {
	return c.w
}
