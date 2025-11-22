package negotiation

import "fmt"

// InvalidArgumentError is returned when an invalid argument is provided.
type InvalidArgumentError struct {
	Message string
}

func (e *InvalidArgumentError) Error() string {
	return e.Message
}

// InvalidHeaderError is returned when a header cannot be parsed.
type InvalidHeaderError struct {
	Header string
}

func (e *InvalidHeaderError) Error() string {
	return fmt.Sprintf("failed to parse accept header: %q", e.Header)
}

// InvalidMediaTypeError is returned when a media type is invalid.
type InvalidMediaTypeError struct{}

func (e *InvalidMediaTypeError) Error() string {
	return "invalid media type"
}

// InvalidLanguageError is returned when a language tag is invalid.
type InvalidLanguageError struct{}

func (e *InvalidLanguageError) Error() string {
	return "invalid language"
}

// ErrNoMatch is returned when no matching header is found.
var ErrNoMatch = &InvalidArgumentError{Message: "no matching header found"}
