package adapters

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/talav/talav/pkg/component/zorya"
)

// ChiAdapter implements zorya.Adapter for Chi router.
type ChiAdapter struct {
	router chi.Router
}

// NewChi creates a new adapter for the given chi router.
func NewChi(r chi.Router) *ChiAdapter {
	return &ChiAdapter{router: r}
}

func (a *ChiAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *ChiAdapter) Handle(route *zorya.BaseRoute, handler func(zorya.Context)) {
	a.router.MethodFunc(route.Method, route.Path, func(w http.ResponseWriter, r *http.Request) {
		// Extract path parameters from chi context
		routerParams := make(map[string]string)
		chiCtx := chi.RouteContext(r.Context())
		if chiCtx != nil {
			for i, key := range chiCtx.URLParams.Keys {
				if i < len(chiCtx.URLParams.Values) {
					routerParams[key] = chiCtx.URLParams.Values[i]
				}
			}
		}

		handler(&chiContext{
			route:        route,
			r:            r,
			w:            w,
			routerParams: routerParams,
		})
	})
}

type chiContext struct {
	route        *zorya.BaseRoute
	r            *http.Request
	w            http.ResponseWriter
	status       int
	routerParams map[string]string
}

func (c *chiContext) Context() context.Context {
	return c.r.Context()
}

func (c *chiContext) Request() *http.Request {
	return c.r
}

func (c *chiContext) RouterParams() map[string]string {
	return c.routerParams
}

func (c *chiContext) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *chiContext) SetReadDeadline(t time.Time) error {
	return zorya.SetReadDeadline(c.w, t)
}

func (c *chiContext) SetBodyLimit(limit int64) {
	if limit > 0 {
		c.r.Body = http.MaxBytesReader(c.w, c.r.Body, limit)
	}
}

// ResponseWriter implementation

func (c *chiContext) SetStatus(status int) {
	c.status = status
	c.w.WriteHeader(status)
}

func (c *chiContext) SetHeader(name, value string) {
	c.w.Header().Set(name, value)
}

func (c *chiContext) AppendHeader(name, value string) {
	c.w.Header().Add(name, value)
}

func (c *chiContext) Write(data []byte) (int, error) {
	return c.w.Write(data)
}

func (c *chiContext) BodyWriter() http.ResponseWriter {
	return c.w
}
