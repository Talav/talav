package adapters

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/talav/talav/pkg/component/zorya"
)

// FiberAdapter implements zorya.Adapter for Fiber router.
type FiberAdapter struct {
	app *fiber.App
}

// NewFiber creates a new adapter for the given Fiber app.
func NewFiber(app *fiber.App) *FiberAdapter {
	return &FiberAdapter{app: app}
}

func (a *FiberAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use Fiber's Test method to handle http.Request
	resp, err := a.app.Test(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Copy headers
	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (a *FiberAdapter) Handle(route *zorya.BaseRoute, handler func(zorya.Context)) {
	// Convert {param} to :param for Fiber
	path := route.Path
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")

	a.app.Add(route.Method, path, func(c *fiber.Ctx) error {
		// Extract path parameters
		routerParams := make(map[string]string)
		if c.Route() != nil {
			for _, param := range c.Route().Params {
				routerParams[param] = c.Params(param)
			}
		}

		handler(&fiberContext{
			route:        route,
			fiberCtx:     c,
			routerParams: routerParams,
		})

		return nil
	})
}

type fiberContext struct {
	route        *zorya.BaseRoute
	fiberCtx     *fiber.Ctx
	status       int
	routerParams map[string]string
	bodyLimit    int64
	bodyLimitSet bool
}

func (c *fiberContext) Context() context.Context {
	return c.fiberCtx.UserContext()
}

func (c *fiberContext) Request() *http.Request {
	// Convert Fiber request to http.Request for compatibility
	req := c.fiberCtx.Request()
	bodyBytes := c.fiberCtx.BodyRaw()

	// Check body limit
	var bodyReader io.Reader = bytes.NewReader(bodyBytes)
	if c.bodyLimitSet && c.bodyLimit > 0 && int64(len(bodyBytes)) > c.bodyLimit {
		// Body exceeds limit - provide a reader that will error
		bodyReader = &limitExceededReader{limit: c.bodyLimit}
	}

	r, _ := http.NewRequestWithContext(
		c.fiberCtx.UserContext(),
		string(req.Header.Method()),
		c.fiberCtx.OriginalURL(),
		io.NopCloser(bodyReader),
	)

	// Copy headers
	req.Header.VisitAll(func(key, value []byte) {
		r.Header.Set(string(key), string(value))
	})

	return r
}

func (c *fiberContext) RouterParams() map[string]string {
	return c.routerParams
}

func (c *fiberContext) Header(name string) string {
	return c.fiberCtx.Get(name)
}

func (c *fiberContext) SetReadDeadline(t time.Time) error {
	return c.fiberCtx.Context().Conn().SetReadDeadline(t)
}

func (c *fiberContext) SetBodyLimit(limit int64) {
	c.bodyLimit = limit
	c.bodyLimitSet = true
}

// ResponseWriter implementation

func (c *fiberContext) SetStatus(status int) {
	c.status = status
	c.fiberCtx.Status(status)
}

func (c *fiberContext) SetHeader(name, value string) {
	c.fiberCtx.Set(name, value)
}

func (c *fiberContext) AppendHeader(name, value string) {
	c.fiberCtx.Append(name, value)
}

func (c *fiberContext) Write(data []byte) (int, error) {
	return c.fiberCtx.Write(data)
}

func (c *fiberContext) BodyWriter() http.ResponseWriter {
	// Fiber doesn't expose http.ResponseWriter directly.
	// Return a wrapper that implements http.ResponseWriter using Fiber's API.
	return &fiberResponseWriter{ctx: c.fiberCtx}
}

type fiberResponseWriter struct {
	ctx *fiber.Ctx
}

func (w *fiberResponseWriter) Header() http.Header {
	h := make(http.Header)
	w.ctx.Response().Header.VisitAll(func(key, value []byte) {
		h.Add(string(key), string(value))
	})

	return h
}

func (w *fiberResponseWriter) Write(data []byte) (int, error) {
	return w.ctx.Write(data)
}

func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	w.ctx.Status(statusCode)
}

// limitExceededReader returns an error when read, indicating body too large.
type limitExceededReader struct {
	limit int64
}

func (r *limitExceededReader) Read(p []byte) (int, error) {
	return 0, &http.MaxBytesError{Limit: r.limit}
}
