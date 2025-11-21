package schema

import "errors"

// Package-level error types for consistent error handling.
var (
	// ErrUnsupportedLocation indicates an unsupported parameter location.
	ErrUnsupportedLocation = errors.New("unsupported location")

	// ErrUnsupportedStyle indicates an unsupported style for a given location.
	ErrUnsupportedStyle = errors.New("unsupported style")

	// ErrInvalidStyle indicates an invalid style value.
	ErrInvalidStyle = errors.New("invalid style")

	// ErrInvalidFormat indicates an invalid format in the input data.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrUnsupportedType indicates an unsupported type for marshaling.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrInvalidElement indicates an invalid element value.
	ErrInvalidElement = errors.New("invalid element")

	// ErrUnsupportedSliceElementType indicates an unsupported slice element type.
	ErrUnsupportedSliceElementType = errors.New("unsupported slice element type")

	// ErrInvalidOptions indicates invalid options configuration.
	ErrInvalidOptions = errors.New("invalid options")
)
