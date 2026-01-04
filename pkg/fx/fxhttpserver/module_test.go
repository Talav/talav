package fxhttpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/httpserver"
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxlogger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// executionOrder tracks the order middleware executes.
var executionOrder []string

func resetExecutionOrder() {
	executionOrder = nil
}

func createTestMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, name)
			next.ServeHTTP(w, r)
		})
	}
}

// setupTestAPI creates a test API with the given middleware options.
func setupTestAPI(t *testing.T, cfg httpserver.Config, opts ...fx.Option) zorya.API {
	t.Helper()
	resetExecutionOrder()

	var api zorya.API

	allOpts := []fx.Option{
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxlogger.FxLoggerModule,
		FxHTTPServerModule,
		fx.Replace(cfg),
	}
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, fx.Populate(&api))

	fxtest.New(t, allOpts...).RequireStart().RequireStop()

	require.NotNil(t, api)

	return api
}

// findPosition returns the index of name in executionOrder, or -1 if not found.
func findPosition(name string) int {
	for i, n := range executionOrder {
		if n == name {
			return i
		}
	}

	return -1
}

// makeRequest makes a test HTTP request to the API.
func makeRequest(t *testing.T, api zorya.API) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	api.Adapter().ServeHTTP(rec, req)

	return rec
}

func TestModule_MiddlewareOrdering(t *testing.T) {
	t.Run("lower priority executes first", func(t *testing.T) {
		api := setupTestAPI(t, httpserver.DefaultConfig(),
			AsMiddleware(createTestMiddleware("mw-230"), 230, "mw-230"),
			AsMiddleware(createTestMiddleware("mw-240"), 240, "mw-240"),
			AsMiddleware(createTestMiddleware("mw-250"), 250, "mw-250"),
		)

		makeRequest(t, api)

		pos230 := findPosition("mw-230")
		pos240 := findPosition("mw-240")
		pos250 := findPosition("mw-250")

		assert.NotEqual(t, -1, pos230)
		assert.NotEqual(t, -1, pos240)
		assert.NotEqual(t, -1, pos250)
		assert.Less(t, pos230, pos240)
		assert.Less(t, pos240, pos250)
	})

	t.Run("same priority maintains registration order", func(t *testing.T) {
		api := setupTestAPI(t, httpserver.DefaultConfig(),
			AsMiddleware(createTestMiddleware("first"), 250, "first"),
			AsMiddleware(createTestMiddleware("second"), 250, "second"),
			AsMiddleware(createTestMiddleware("third"), 250, "third"),
		)

		makeRequest(t, api)

		posFirst := findPosition("first")
		posSecond := findPosition("second")
		posThird := findPosition("third")

		assert.NotEqual(t, -1, posFirst)
		assert.NotEqual(t, -1, posSecond)
		assert.NotEqual(t, -1, posThird)
		assert.Less(t, posFirst, posSecond)
		assert.Less(t, posSecond, posThird)
	})

	t.Run("built-in middlewares at correct positions", func(t *testing.T) {
		cfg := httpserver.DefaultConfig()
		cfg.Logging.Enabled = true

		api := setupTestAPI(t, cfg,
			AsMiddleware(createTestMiddleware("before"), 50, "before"),
			AsMiddleware(createTestMiddleware("between"), 150, "between"),
			AsMiddleware(createTestMiddleware("after"), 250, "after"),
		)

		makeRequest(t, api)

		posBefore := findPosition("before")
		posBetween := findPosition("between")
		posAfter := findPosition("after")

		assert.NotEqual(t, -1, posBefore)
		assert.NotEqual(t, -1, posBetween)
		assert.NotEqual(t, -1, posAfter)
		assert.Less(t, posBefore, posBetween)
		assert.Less(t, posBetween, posAfter)
	})

	t.Run("all priority ranges work", func(t *testing.T) {
		api := setupTestAPI(t, httpserver.DefaultConfig(),
			AsMiddleware(createTestMiddleware("p50"), 50, "p50"),
			AsMiddleware(createTestMiddleware("p150"), 150, "p150"),
			AsMiddleware(createTestMiddleware("p250"), 250, "p250"),
			AsMiddleware(createTestMiddleware("p350"), 350, "p350"),
		)

		makeRequest(t, api)

		assert.NotEqual(t, -1, findPosition("p50"))
		assert.NotEqual(t, -1, findPosition("p150"))
		assert.NotEqual(t, -1, findPosition("p250"))
		assert.NotEqual(t, -1, findPosition("p350"))
		assert.Less(t, findPosition("p50"), findPosition("p150"))
		assert.Less(t, findPosition("p150"), findPosition("p250"))
		assert.Less(t, findPosition("p250"), findPosition("p350"))
	})
}

func TestModule_MiddlewareRegistration(t *testing.T) {
	t.Run("works without middlewares", func(t *testing.T) {
		api := setupTestAPI(t, httpserver.DefaultConfig())
		rec := makeRequest(t, api)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("registers and executes middlewares", func(t *testing.T) {
		api := setupTestAPI(t, httpserver.DefaultConfig(),
			AsMiddleware(createTestMiddleware("test"), 250, "test"),
		)

		makeRequest(t, api)
		assert.Greater(t, len(executionOrder), 0)
	})
}

func TestModule_MiddlewarePriorityConstants(t *testing.T) {
	// Verify priority constants are set correctly
	assert.Equal(t, 100, PriorityRequestID, "PriorityRequestID should be 100")
	assert.Equal(t, 200, PriorityHTTPLog, "PriorityHTTPLog should be 200")
	assert.Equal(t, 250, PriorityBeforeZorya, "PriorityBeforeZorya should be 250")
}
