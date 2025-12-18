package zorya

import (
	"fmt"
	"net/http"
	"reflect"
)

// ResponseMetadata contains metadata about the output struct fields.
type ResponseMetadata struct {
	StatusIndex int  // Index of Status field (-1 if none)
	BodyIndex   int  // Index of Body field (-1 if none)
	BodyFunc    bool // Is Body a callback function?
	Headers     []HeaderField
}

// HeaderField contains information about a header field in the output struct.
type HeaderField struct {
	FieldIndex int    // Index of the field in the struct
	Name       string // Header name (from tag or field name)
	IsSlice    bool   // Whether the field is a slice (multiple header values)
}

// processOutputType analyzes the output struct type and extracts metadata
// about Status, Body, and Header fields. Similar to Huma's processOutputType.
func processOutputType(outputType reflect.Type) *ResponseMetadata {
	meta := &ResponseMetadata{
		StatusIndex: -1,
		BodyIndex:   -1,
		BodyFunc:    false,
		Headers:     []HeaderField{},
	}

	// Find Status field.
	if f, ok := outputType.FieldByName("Status"); ok {
		meta.StatusIndex = f.Index[0]
		if f.Type.Kind() != reflect.Int {
			panic("Status field must be an int")
		}
	}

	// Find Body field.
	if f, ok := outputType.FieldByName("Body"); ok {
		meta.BodyIndex = f.Index[0]
		meta.BodyFunc = isBodyFunc(f.Type)
	}

	// Find header fields (fields with "header" tag).
	meta.Headers = findHeaderFields(outputType)

	return meta
}

// isBodyFunc checks if the type is a valid body callback function and validates its signature.
func isBodyFunc(t reflect.Type) bool {
	if t.Kind() != reflect.Func {
		return false
	}

	// Validate function signature: func(Context).
	if t.NumIn() != 1 || t.NumOut() != 0 {
		panic("Body function must have signature func(Context)")
	}

	// Check if first parameter is Context.
	if !t.In(0).Implements(reflect.TypeOf((*Context)(nil)).Elem()) {
		panic("Body function parameter must implement Context interface")
	}

	return true
}

// findHeaderFields finds all fields in the struct that have a "header" tag.
func findHeaderFields(typ reflect.Type) []HeaderField {
	var headers []HeaderField

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		headerTag := field.Tag.Get("header")
		if headerTag == "" || headerTag == "-" {
			continue
		}

		headerName := headerTag
		if headerName == "" {
			// Use field name if tag is empty
			headerName = field.Name
		}

		isSlice := field.Type.Kind() == reflect.Slice

		headers = append(headers, HeaderField{
			FieldIndex: i,
			Name:       headerName,
			IsSlice:    isSlice,
		})
	}

	return headers
}

// writeResponse transforms the output struct into an HTTP response.
func writeResponse(api API, ctx Context, output any, meta *ResponseMetadata, defaultStatus int) error {
	if output == nil {
		// Special case: No output, just set default status
		if w, ok := ctx.(ResponseWriter); ok {
			w.SetStatus(defaultStatus)
		}

		return nil
	}

	vo := reflect.ValueOf(output)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}

	// Extract and write headers
	if w, ok := ctx.(ResponseWriter); ok {
		writeHeaders(w, vo, meta.Headers)
	}

	// Determine status code
	status := defaultStatus
	if meta.StatusIndex != -1 {
		statusField := vo.Field(meta.StatusIndex)
		if statusField.IsValid() && statusField.CanInt() {
			status = int(statusField.Int())
		}
	}

	// No body field, just set status.
	if meta.BodyIndex == -1 {
		if w, ok := ctx.(ResponseWriter); ok {
			w.SetStatus(status)
		}

		return nil
	}

	// Extract and write body.
	return writeBody(api, ctx, vo, meta, status)
}

// writeHeaders writes header fields from the output struct to the HTTP response.
func writeHeaders(w ResponseWriter, vo reflect.Value, headers []HeaderField) {
	for _, hf := range headers {
		field := vo.Field(hf.FieldIndex)
		if !field.IsValid() {
			continue
		}

		field = reflect.Indirect(field)
		if field.Kind() == reflect.Invalid {
			continue
		}

		if hf.IsSlice {
			// Multiple header values - append each
			for i := 0; i < field.Len(); i++ {
				value := field.Index(i)
				headerValue := formatHeaderValue(value)
				w.AppendHeader(hf.Name, headerValue)
			}
		} else {
			// Single header value
			headerValue := formatHeaderValue(field)
			w.SetHeader(hf.Name, headerValue)
		}
	}
}

// formatHeaderValue converts a reflect.Value to a string for use as a header value.
func formatHeaderValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}

	// Handle fmt.Stringer interface.
	if v.CanInterface() {
		if str, ok := v.Interface().(interface{ String() string }); ok {
			return str.String()
		}
	}

	// Convert based on kind.
	//nolint:exhaustive // Only handle primitive types, default handles the rest.
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", v.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// writeBody handles body extraction and writing.
func writeBody(api API, ctx Context, vo reflect.Value, meta *ResponseMetadata, status int) error {
	bodyField := vo.Field(meta.BodyIndex)
	if !bodyField.IsValid() {
		return writeStatusOnly(ctx, status)
	}

	if meta.BodyFunc {
		return writeBodyFunc(ctx, bodyField, status)
	}

	body := bodyField.Interface()

	// Handle []byte (raw bytes) - no content negotiation.
	if b, ok := body.([]byte); ok {
		return writeRawBody(ctx, status, b)
	}

	return writeNegotiatedBody(api, ctx, status, body)
}

// writeStatusOnly writes only the status code when body is invalid.
func writeStatusOnly(ctx Context, status int) error {
	if w, ok := ctx.(ResponseWriter); ok {
		w.SetStatus(status)
	}

	return nil
}

// writeBodyFunc executes a body callback function for streaming responses.
// Status and headers from struct fields should already be set before this is called.
func writeBodyFunc(ctx Context, bodyField reflect.Value, status int) error {
	// Set status before streaming starts
	if w, ok := ctx.(ResponseWriter); ok {
		w.SetStatus(status)
	}

	if fn, ok := bodyField.Interface().(func(Context)); ok {
		fn(ctx)
	}

	return nil
}

// writeRawBody writes raw bytes without content negotiation.
func writeRawBody(ctx Context, status int, data []byte) error {
	if w, ok := ctx.(ResponseWriter); ok {
		w.SetStatus(status)
		_, _ = w.Write(data)
	}

	return nil
}

// writeNegotiatedBody negotiates content type and marshals the body.
func writeNegotiatedBody(api API, ctx Context, status int, body any) error {
	w, ok := ctx.(ResponseWriter)
	if !ok {
		return nil
	}

	ct, err := api.Negotiate(ctx.Header("Accept"))
	if err != nil {
		return writeNegotiationError(w, err)
	}

	// Run transformers on the body before serialization.
	statusStr := fmt.Sprintf("%d", status)
	body, err = api.Transform(ctx, statusStr, body)
	if err != nil {
		return fmt.Errorf("transformer error: %w", err)
	}

	// Check if body implements ContentTypeFilter (e.g., ErrorModel).
	if ctf, ok := body.(ContentTypeFilter); ok {
		ct = ctf.ContentType(ct)
	}

	w.SetHeader("Content-Type", ct)
	w.SetStatus(status)

	if err := api.Marshal(w, ct, body); err != nil {
		return fmt.Errorf("failed to marshal response body: %w", err)
	}

	return nil
}

// writeNegotiationError writes a 406 Not Acceptable error response.
func writeNegotiationError(w ResponseWriter, err error) error {
	w.SetStatus(http.StatusNotAcceptable)
	w.SetHeader("Content-Type", "application/json")

	// Use JSON as fallback for error responses
	jsonFmt := JSONFormat()

	return jsonFmt.Marshal(w, map[string]any{
		"status": http.StatusNotAcceptable,
		"title":  "Not Acceptable",
		"detail": err.Error(),
	})
}

// ResponseWriter extends Context with methods to write HTTP responses.
// This interface should be implemented by adapter contexts (like chiContext).
type ResponseWriter interface {
	Context
	// SetStatus sets the HTTP status code.
	SetStatus(status int)
	// SetHeader sets a single header value.
	SetHeader(name, value string)
	// AppendHeader appends a header value (for multiple values).
	AppendHeader(name, value string)
	// Write writes raw bytes to the response body.
	Write(data []byte) (int, error)
}
