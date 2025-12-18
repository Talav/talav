package zorya

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorDetailer returns error details for responses & debugging. This enables
// the use of custom error types. See `NewError` for more details.
type ErrorDetailer interface {
	ErrorDetail() *ErrorDetail
}

// ErrorDetail provides details about a specific error.
//
//nolint:errname // This is a data structure, not an error type.
type ErrorDetail struct {
	// Code is a machine-readable error code (e.g., "required", "email", "min").
	// This enables frontend translation and automated error handling.
	Code string `json:"code,omitempty"`

	// Message is a human-readable explanation of the error (optional).
	// Useful for developers, logs, and debugging.
	Message string `json:"message,omitempty"`

	// Location is a path-like string indicating where the error occurred.
	// It typically begins with `path`, `query`, `header`, or `body`. Example:
	// `body.items[3].tags` or `path.thing-id`.
	Location string `json:"location,omitempty"`
}

// Error returns the error message / satisfies the `error` interface.
func (e *ErrorDetail) Error() string {
	if e.Message != "" {
		if e.Location != "" {
			return fmt.Sprintf("%s (%s)", e.Message, e.Location)
		}

		return e.Message
	}

	if e.Code != "" && e.Location != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Location)
	}

	if e.Code != "" {
		return e.Code
	}

	if e.Location != "" {
		return e.Location
	}

	return "validation error"
}

// ErrorDetail satisfies the `ErrorDetailer` interface.
func (e *ErrorDetail) ErrorDetail() *ErrorDetail {
	return e
}

// ErrorModel defines a basic error message model based on RFC 9457 Problem
// Details for HTTP APIs (https://datatracker.ietf.org/doc/html/rfc9457). It
// is augmented with an `errors` field of `ErrorDetail` objects that
// can help provide exhaustive & descriptive errors.
//
//	err := &ErrorModel{
//		Title:  http.StatusText(http.StatusBadRequest),
//		Status: http.StatusBadRequest,
//		Detail: "Validation failed",
//		Errors: []*ErrorDetail{
//			{
//				Code:     "required",
//				Message:  "expected required property id to be present",
//				Location: "body.friends[0]",
//			},
//			{
//				Code:     "type",
//				Message:  "expected boolean",
//				Location: "body.friends[1].active",
//			},
//		},
//	}
//
//nolint:errname // This is a data structure implementing error interface, not a pure error type.
type ErrorModel struct {
	// Type is a URI to get more information about the error type.
	Type string `json:"type,omitempty"`

	// Title provides a short static summary of the problem. Zorya will default this
	// to the HTTP response status code text if not present.
	Title string `json:"title,omitempty"`

	// Status provides the HTTP status code for client convenience. Zorya will
	// default this to the response status code if unset. This SHOULD match the
	// response status code (though proxies may modify the actual status code).
	Status int `json:"status,omitempty"`

	// Detail is an explanation specific to this error occurrence.
	Detail string `json:"detail,omitempty"`

	// Instance is a URI to get more info about this error occurrence.
	Instance string `json:"instance,omitempty"`

	// Errors provides an optional mechanism of passing additional error details
	// as a list.
	Errors []*ErrorDetail `json:"errors,omitempty"`
}

// Error satisfies the `error` interface. It returns the error's detail field.
func (e *ErrorModel) Error() string {
	return e.Detail
}

// Add an error to the `Errors` slice. If passed a struct that satisfies the
// `ErrorDetailer` interface, then it is used, otherwise the error
// string is used as the error detail message.
//
//	err := &ErrorModel{ /* ... */ }
//	err.Add(&ErrorDetail{
//		Code:     "type",
//		Message:  "expected boolean",
//		Location: "body.friends[1].active",
//	})
func (e *ErrorModel) Add(err error) {
	if converted, ok := err.(ErrorDetailer); ok {
		e.Errors = append(e.Errors, converted.ErrorDetail())

		return
	}

	if err != nil {
		e.Errors = append(e.Errors, &ErrorDetail{Message: err.Error()})
	}
}

// GetStatus returns the HTTP status that should be returned to the client
// for this error.
func (e *ErrorModel) GetStatus() int {
	return e.Status
}

// ContentType provides a filter to adjust response content types. This is
// used to ensure e.g. `application/problem+json` content types defined in
// RFC 9457 Problem Details for HTTP APIs are used in responses to clients.
func (e *ErrorModel) ContentType(ct string) string {
	if ct == "application/json" {
		return "application/problem+json"
	}
	if ct == "application/cbor" {
		return "application/problem+cbor"
	}

	return ct
}

// ContentTypeFilter allows you to override the content type for responses,
// allowing you to return a different content type like
// `application/problem+json` after using the `application/json` marshaller.
// This should be implemented by the response body struct.
type ContentTypeFilter interface {
	ContentType(string) string
}

// StatusError is an error that has an HTTP status code. When returned from
// an operation handler, this sets the response status code before sending it
// to the client.
type StatusError interface {
	GetStatus() int
	Error() string
}

// HeadersError is an error that has HTTP headers. When returned from an
// operation handler, these headers are set on the response before sending it
// to the client. Use `ErrorWithHeaders` to wrap an error like
// `Error400BadRequest` with additional headers.
type HeadersError interface {
	GetHeaders() http.Header
	Error() string
}

//nolint:errname // This is an internal wrapper type, not a public error type.
type errWithHeaders struct {
	err     error
	headers http.Header
}

func (e *errWithHeaders) Error() string {
	return e.err.Error()
}

func (e *errWithHeaders) Unwrap() error {
	return e.err
}

func (e *errWithHeaders) GetHeaders() http.Header {
	return e.headers
}

// ErrorWithHeaders wraps an error with additional headers to be sent to the
// client. This is useful for e.g. caching, rate limiting, or other metadata.
func ErrorWithHeaders(err error, headers http.Header) error {
	var he HeadersError
	if errors.As(err, &he) {
		// There is already a headers error, so we need to merge the headers. This
		// lets you chain multiple calls together and have all the headers set.
		orig := he.GetHeaders()
		for k, values := range headers {
			for _, v := range values {
				orig.Add(k, v)
			}
		}

		return err
	}

	return &errWithHeaders{err: err, headers: headers}
}

// NewError creates a new instance of an error model with the given status code,
// message, and optional error details. If the error details implement the
// `ErrorDetailer` interface, the error details will be used. Otherwise, the
// error string will be used as the message.
//
// Replace this function to use your own error type. Example:
//
//	type MyDetail struct {
//		Message string `json:"message"`
//		Location string `json:"location"`
//	}
//
//	type MyError struct {
//		status  int
//		Message string `json:"message"`
//		Errors  []error `json:"errors"`
//	}
//
//	func (e *MyError) Error() string {
//		return e.Message
//	}
//
//	func (e *MyError) GetStatus() int {
//		return e.status
//	}
//
//	zorya.NewError = func(status int, msg string, errs ...error) StatusError {
//		return &MyError{
//			status:  status,
//			Message: msg,
//			Errors:  errs,
//		}
//	}
var NewError = func(status int, msg string, errs ...error) StatusError {
	details := make([]*ErrorDetail, 0, len(errs))
	for i := 0; i < len(errs); i++ {
		if errs[i] == nil {
			continue
		}
		if converted, ok := errs[i].(ErrorDetailer); ok {
			details = append(details, converted.ErrorDetail())
		} else {
			details = append(details, &ErrorDetail{Message: errs[i].Error()})
		}
	}

	title := http.StatusText(status)
	if title == "" {
		title = "Error"
	}

	return &ErrorModel{
		Status: status,
		Title:  title,
		Detail: msg,
		Errors: details,
	}
}

// NewErrorWithContext creates a new error with context. By default, it delegates
// to NewError, but can be replaced to provide context-aware error creation.
var NewErrorWithContext = func(_ Context, status int, msg string, errs ...error) StatusError {
	return NewError(status, msg, errs...)
}

// WriteErr writes an error response with the given context, using the
// configured error type and with the given status code and message. It is
// marshaled using the API's content negotiation methods.
func WriteErr(api API, ctx Context, status int, msg string, errs ...error) error {
	err := NewErrorWithContext(ctx, status, msg, errs...)

	// NewError may have modified the status code, so update it here if needed.
	// If it was not modified then this is a no-op.
	status = err.GetStatus()

	writeErr := writeErrorResponse(api, ctx, status, err)
	if writeErr != nil {
		// If we can't write the error, log it so we know what happened.
		// Note: In production, you might want to use a proper logger here.
		_ = writeErr
	}

	return writeErr
}

// Status304NotModified returns a 304. This is not really an error, but
// provides a way to send non-default responses.
func Status304NotModified() StatusError {
	return NewError(http.StatusNotModified, "")
}

// Error400BadRequest returns a 400.
func Error400BadRequest(msg string, errs ...error) StatusError {
	return NewError(http.StatusBadRequest, msg, errs...)
}

// Error401Unauthorized returns a 401.
func Error401Unauthorized(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnauthorized, msg, errs...)
}

// Error403Forbidden returns a 403.
func Error403Forbidden(msg string, errs ...error) StatusError {
	return NewError(http.StatusForbidden, msg, errs...)
}

// Error404NotFound returns a 404.
func Error404NotFound(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotFound, msg, errs...)
}

// Error405MethodNotAllowed returns a 405.
func Error405MethodNotAllowed(msg string, errs ...error) StatusError {
	return NewError(http.StatusMethodNotAllowed, msg, errs...)
}

// Error406NotAcceptable returns a 406.
func Error406NotAcceptable(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotAcceptable, msg, errs...)
}

// Error409Conflict returns a 409.
func Error409Conflict(msg string, errs ...error) StatusError {
	return NewError(http.StatusConflict, msg, errs...)
}

// Error410Gone returns a 410.
func Error410Gone(msg string, errs ...error) StatusError {
	return NewError(http.StatusGone, msg, errs...)
}

// Error412PreconditionFailed returns a 412.
func Error412PreconditionFailed(msg string, errs ...error) StatusError {
	return NewError(http.StatusPreconditionFailed, msg, errs...)
}

// Error415UnsupportedMediaType returns a 415.
func Error415UnsupportedMediaType(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnsupportedMediaType, msg, errs...)
}

// Error422UnprocessableEntity returns a 422.
func Error422UnprocessableEntity(msg string, errs ...error) StatusError {
	return NewError(http.StatusUnprocessableEntity, msg, errs...)
}

// Error429TooManyRequests returns a 429.
func Error429TooManyRequests(msg string, errs ...error) StatusError {
	return NewError(http.StatusTooManyRequests, msg, errs...)
}

// Error500InternalServerError returns a 500.
func Error500InternalServerError(msg string, errs ...error) StatusError {
	return NewError(http.StatusInternalServerError, msg, errs...)
}

// Error501NotImplemented returns a 501.
func Error501NotImplemented(msg string, errs ...error) StatusError {
	return NewError(http.StatusNotImplemented, msg, errs...)
}

// Error502BadGateway returns a 502.
func Error502BadGateway(msg string, errs ...error) StatusError {
	return NewError(http.StatusBadGateway, msg, errs...)
}

// Error503ServiceUnavailable returns a 503.
func Error503ServiceUnavailable(msg string, errs ...error) StatusError {
	return NewError(http.StatusServiceUnavailable, msg, errs...)
}

// Error504GatewayTimeout returns a 504.
func Error504GatewayTimeout(msg string, errs ...error) StatusError {
	return NewError(http.StatusGatewayTimeout, msg, errs...)
}
