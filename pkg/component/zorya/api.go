package zorya

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/talav/talav/pkg/component/negotiation"
	"github.com/talav/talav/pkg/component/schema"
)

// Adapter is an interface that allows the API to be used with different HTTP
// routers and frameworks. It is designed to work with the standard library
// `http.Request` and `http.ResponseWriter` types as well as types like
// `gin.Context` or `fiber.Ctx` that provide both request and response
// functionality in one place, by using the `zorya.Context` interface which
// abstracts away those router-specific differences.
type Adapter interface {
	Handle(route *BaseRoute, handler func(ctx Context))
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Option configures an API.
type Option func(*api)

type API interface {
	// Adapter returns the router adapter for this API, providing a generic
	// interface to get request information and write responses.
	Adapter() Adapter

	// Middlewares returns a slice of middleware handler functions that will be
	// run for all operations. Middleware are run in the order they are added.
	// See also `huma.Operation{}.Middlewares` for adding operation-specific
	// middleware at operation registration time.
	Middlewares() Middlewares

	Codec() *schema.Codec

	// Negotiate returns the best content type for the response based on the
	// Accept header. If no match is found, returns the default format.
	Negotiate(accept string) (string, error)

	// Marshal writes the value to the writer using the format for the given
	// content type. Supports plus-segment matching (e.g., application/vnd.api+json).
	Marshal(w io.Writer, contentType string, v any) error

	// Validator returns the configured validator, or nil if validation is disabled.
	Validator() Validator

	// Transform runs all transformers on the response value.
	// Called automatically during response serialization.
	Transform(ctx Context, status string, v any) (any, error)

	// UseTransformer adds one or more transformer functions that will be
	// run on all responses.
	UseTransformer(transformers ...Transformer)
}

type api struct {
	adapter       Adapter
	middlewares   Middlewares
	codec         *schema.Codec
	formats       map[string]Format
	formatKeys    []string // Ordered keys for negotiation priority
	defaultFormat string
	negotiator    *negotiation.Negotiator
	validator     Validator
	transformers  []Transformer
}

func (a *api) Adapter() Adapter {
	return a.adapter
}

func (a *api) Middlewares() Middlewares {
	return a.middlewares
}

func (a *api) Codec() *schema.Codec {
	return a.codec
}

func (a *api) Validator() Validator {
	return a.validator
}

// Transform runs all transformers on the response value in the order they were added.
func (a *api) Transform(ctx Context, status string, v any) (any, error) {
	for _, t := range a.transformers {
		var err error
		v, err = t(ctx, status, v)
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

// UseTransformer adds one or more transformer functions that will be run on all responses.
func (a *api) UseTransformer(transformers ...Transformer) {
	a.transformers = append(a.transformers, transformers...)
}

// Negotiate returns the best content type based on the Accept header.
func (a *api) Negotiate(accept string) (string, error) {
	if accept == "" {
		return a.defaultFormat, nil
	}

	header, err := a.negotiator.Negotiate(accept, a.formatKeys, false)
	if errors.Is(err, negotiation.ErrNoMatch) {
		// Fallback to default format when no match
		return a.defaultFormat, nil
	}

	if err != nil {
		return "", fmt.Errorf("negotiation failed: %w", err)
	}

	return header.Type, nil
}

// Marshal writes the value using the format for the given content type.
func (a *api) Marshal(w io.Writer, ct string, v any) error {
	f, ok := a.formats[ct]
	if !ok {
		// Try extracting suffix from plus-segment (e.g., application/vnd.api+json -> json).
		if idx := strings.LastIndex(ct, "+"); idx != -1 {
			f, ok = a.formats[ct[idx+1:]]
		}
	}

	if !ok {
		return fmt.Errorf("unknown content type: %s", ct)
	}

	return f.Marshal(w, v)
}

// NewAPI creates a new API instance with the given adapter and options.
// The adapter is required; all other configuration is optional.
//
// Example:
//
//	api := zorya.NewAPI(adapter)
//	api := zorya.NewAPI(adapter, zorya.WithValidator(validator))
//	api := zorya.NewAPI(adapter, zorya.WithFormat("application/xml", xmlFormat))
//	api := zorya.NewAPI(adapter, zorya.WithFormats(customFormats))
//	api := zorya.NewAPI(adapter, zorya.WithFormatsReplace(formats)) // Replace all formats
func NewAPI(adapter Adapter, opts ...Option) API {
	a := &api{
		adapter:       adapter,
		middlewares:   Middlewares{},
		codec:         nil, // Set default below
		formats:       nil, // Set default below
		formatKeys:    nil,
		defaultFormat: "application/json",
		negotiator:    negotiation.NewMediaNegotiator(),
		validator:     nil,
		transformers:  []Transformer{},
	}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	// Set defaults for anything not configured
	if a.codec == nil {
		a.codec = schema.NewCodec()
	}

	if a.formats == nil {
		a.formats = DefaultFormats()
	}

	// Build format keys from formats
	if len(a.formatKeys) == 0 {
		a.formatKeys = make([]string, 0, len(a.formats))
		for k := range a.formats {
			// Only include full content types, not suffixes
			if strings.Contains(k, "/") {
				a.formatKeys = append(a.formatKeys, k)
			}
		}
	}

	return a
}

// WithValidator sets a validator for request validation.
func WithValidator(validator Validator) Option {
	return func(a *api) {
		a.validator = validator
	}
}

// WithFormat adds a single format for content negotiation.
// Multiple calls to WithFormat can be chained to add multiple formats.
// Formats are merged with default formats, with later formats taking precedence.
// Default formats are automatically included, so you don't need to add them manually.
func WithFormat(contentType string, format Format) Option {
	return func(a *api) {
		// Ensure defaults are loaded
		if a.formats == nil {
			a.formats = DefaultFormats()
		}
		a.formats[contentType] = format
	}
}

// WithFormats sets custom formats for content negotiation.
// Custom formats are merged with default formats, with custom formats taking precedence.
// Default formats are automatically included, so you don't need to add them manually.
func WithFormats(formats map[string]Format) Option {
	return func(a *api) {
		// Start with defaults and merge custom formats
		if a.formats == nil {
			a.formats = DefaultFormats()
		}
		for k, v := range formats {
			a.formats[k] = v
		}
	}
}

// WithFormatsReplace replaces all formats (does not merge with defaults).
// Use this when you want complete control over supported formats.
// Default formats are NOT included unless you explicitly add them.
func WithFormatsReplace(formats map[string]Format) Option {
	return func(a *api) {
		// Replace all formats - don't merge with defaults
		a.formats = make(map[string]Format, len(formats))
		for k, v := range formats {
			a.formats[k] = v
		}
	}
}

// WithCodec sets a custom codec for request/response encoding/decoding.
func WithCodec(codec *schema.Codec) Option {
	return func(a *api) {
		a.codec = codec
	}
}

// WithDefaultFormat sets the default content type when Accept header is missing or no match is found.
func WithDefaultFormat(format string) Option {
	return func(a *api) {
		a.defaultFormat = format
	}
}

// Register an operation handler for an API. The handler must be a function that
// takes a context and a pointer to the input struct and returns a pointer to the
// output struct and an error. The input struct must be a struct with fields
// for the request path/query/header/cookie parameters and/or body. The output
// struct must be a struct with fields for the output headers and body of the
// operation, if any.
//
//	huma.Register(api, huma.Operation{
//		OperationID: "get-greeting",
//		Method:      http.MethodGet,
//		Path:        "/greeting/{name}",
//		Summary:     "Get a greeting",
//	}, func(ctx context.Context, input *GreetingInput) (*GreetingOutput, error) {
//		if input.Name == "bob" {
//			return nil, huma.Error404NotFound("no greeting for bob")
//		}
//		resp := &GreetingOutput{}
//		resp.MyHeader = "MyValue"
//		resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
//		return resp, nil
//	})
func Register[I, O any](api API, route BaseRoute, handler func(context.Context, *I) (*O, error)) {
	inputType := reflect.TypeFor[I]()
	if inputType.Kind() != reflect.Struct {
		panic("input must be a struct")
	}
	outputType := reflect.TypeFor[O]()
	if outputType.Kind() != reflect.Struct {
		panic("output must be a struct")
	}

	responseMetadata := processOutputType(outputType)

	// validation: findResolvers in Huma
	// setting default parameters: findDefaults in Huma

	a := api.Adapter()
	a.Handle(&route, api.Middlewares().Handler(route.Middlewares.Handler(func(ctx Context) {
		setupRequestLimits(ctx, route)

		var input I
		if !decodeAndValidateRequest(api, ctx, &input) {
			return
		}

		// Execute handler.
		output, err := handler(ctx.Context(), &input)
		if err != nil {
			handleHandlerError(api, ctx, err)

			return
		}

		// Write response.
		defaultStatus := http.StatusOK
		if err := writeResponse(api, ctx, output, responseMetadata, defaultStatus); err != nil {
			_ = WriteErr(api, ctx, http.StatusInternalServerError, "failed to write response", err)

			return
		}
	})))
}

// Get registers a GET route handler.
func Get[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodGet, path, handler, options...)
}

// Post registers a POST route handler.
func Post[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPost, path, handler, options...)
}

// Put registers a PUT route handler.
func Put[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPut, path, handler, options...)
}

// Delete registers a DELETE route handler.
func Delete[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodDelete, path, handler, options...)
}

// Patch registers a PATCH route handler.
func Patch[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodPatch, path, handler, options...)
}

// Head registers a HEAD route handler.
func Head[I, O any](api API, path string, handler func(context.Context, *I) (*O, error), options ...func(*BaseRoute)) {
	convenience(api, http.MethodHead, path, handler, options...)
}

// convenience is a helper function used by Get, Post, Put, Delete, Patch, and Head.
func convenience[I, O any](api API, method, path string, handler func(context.Context, *I) (*O, error), options ...func(o *BaseRoute)) {
	// generate operation id, generate summary, generate base route, execute all options
	route := BaseRoute{
		Method: method,
		Path:   path,
	}
	for _, o := range options {
		o(&route)
	}
	Register(api, route, handler)
}

// setupRequestLimits configures body read timeout and size limits for the request.
func setupRequestLimits(ctx Context, route BaseRoute) {
	// Apply body read timeout.
	// This sets a deadline for reading the request body, helping prevent slow-loris attacks.
	// Default is 5 seconds if not explicitly configured.
	bodyTimeout := route.BodyReadTimeout
	if bodyTimeout == 0 {
		bodyTimeout = DefaultBodyReadTimeout
	}
	if bodyTimeout > 0 {
		_ = ctx.SetReadDeadline(time.Now().Add(bodyTimeout))
	} else {
		// Negative value disables any deadline.
		_ = ctx.SetReadDeadline(time.Time{})
	}

	// Apply body size limit if context supports it.
	// Default to 1MB if not explicitly configured.
	if limiter, ok := ctx.(BodyLimiter); ok {
		maxBytes := route.MaxBodyBytes
		if maxBytes == 0 {
			maxBytes = DefaultMaxBodyBytes
		}
		if maxBytes > 0 {
			limiter.SetBodyLimit(maxBytes)
		}
		// If maxBytes < 0, no limit is applied
	}
}

// decodeAndValidateRequest decodes and validates the request input.
// Returns false if decoding or validation failed (error already written).
func decodeAndValidateRequest[I any](api API, ctx Context, input *I) bool {
	if err := api.Codec().DecodeRequest(ctx.Request(), ctx.RouterParams(), input); err != nil {
		// Check if this is a body size limit error
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			_ = WriteErr(api, ctx, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("request body too large (limit: %d bytes)", maxBytesErr.Limit))

			return false
		}

		_ = WriteErr(api, ctx, http.StatusBadRequest, "failed to decode request", err)

		return false
	}

	// Validate input if validator configured
	if v := api.Validator(); v != nil {
		// Get struct metadata for validation location mapping
		var metadata *schema.StructMetadata
		codec := api.Codec()
		if codec != nil {
			typ := reflect.TypeOf(input)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if fieldCache := codec.FieldCache(); fieldCache != nil {
				if md, err := fieldCache.GetStructMetadata(typ); err == nil {
					metadata = md
				}
			}
		}

		if errs := v.Validate(ctx.Context(), input, metadata); len(errs) > 0 {
			_ = WriteErr(api, ctx, http.StatusUnprocessableEntity, "validation failed", errs...)

			return false
		}
	}

	return true
}

// handleHandlerError handles errors returned from the handler.
func handleHandlerError(api API, ctx Context, err error) {
	// Check if error implements HeadersError and set headers.
	var he HeadersError
	if errors.As(err, &he) {
		if w, ok := ctx.(ResponseWriter); ok {
			for k, values := range he.GetHeaders() {
				for _, v := range values {
					w.AppendHeader(k, v)
				}
			}
		}
	}

	status := http.StatusInternalServerError
	msg := err.Error()

	// Check if error implements StatusError.
	var statusErr StatusError
	if errors.As(err, &statusErr) {
		status = statusErr.GetStatus()
		msg = statusErr.Error()
	}

	_ = WriteErr(api, ctx, status, msg, err)
}

// writeErrorResponse writes an error response using content negotiation.
func writeErrorResponse(api API, ctx Context, status int, err StatusError) error {
	w, ok := ctx.(ResponseWriter)
	if !ok {
		return nil
	}

	// Check if error implements HeadersError and set headers.
	var he HeadersError
	if errors.As(err, &he) {
		for k, values := range he.GetHeaders() {
			for _, v := range values {
				w.AppendHeader(k, v)
			}
		}
	}

	// Negotiate content type for error response.
	ct, negErr := api.Negotiate(ctx.Header("Accept"))
	if negErr != nil {
		// Fallback to JSON if negotiation fails.
		ct = "application/json"
	}

	// Check if error implements ContentTypeFilter (e.g., ErrorModel).
	if ctf, ok := err.(ContentTypeFilter); ok {
		ct = ctf.ContentType(ct)
	}

	w.SetHeader("Content-Type", ct)
	w.SetStatus(status)

	// Marshal error using negotiated format.
	if err := api.Marshal(w, ct, err); err != nil {
		// Fallback to plain text if marshaling fails.
		w.SetHeader("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "Error %d: %s", status, err.Error())

		return err
	}

	return nil
}

// SetReadDeadline is a utility to set the read deadline on a response writer,
// if possible. It unwraps response writer wrappers until it finds one that
// supports SetReadDeadline, or returns an error if none is found.
// This approach avoids allocations (unlike the stdlib http.ResponseController).
// This is exported for use by adapters.
func SetReadDeadline(w http.ResponseWriter, deadline time.Time) error {
	for {
		switch t := w.(type) {
		case interface{ SetReadDeadline(time.Time) error }:
			return t.SetReadDeadline(deadline)
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
		default:
			// No response writer in the chain supports SetReadDeadline.
			// This is not necessarily an error - the server's connection-level
			// ReadTimeout will still apply.
			return nil
		}
	}
}
