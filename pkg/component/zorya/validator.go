package zorya

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/schema"
)

// Validator validates input structs after request decoding.
// Each returned error should implement ErrorDetailer for RFC 9457 compliant responses.
type Validator interface {
	// Validate validates the input struct.
	// Returns nil if validation succeeds, or a slice of errors if validation fails.
	Validate(ctx context.Context, input any, metadata *schema.StructMetadata) []error
}

// PlaygroundValidator adapts go-playground/validator to Zorya's Validator interface.
type PlaygroundValidator struct {
	validate *validator.Validate
}

// NewPlaygroundValidator creates a new validator adapter for go-playground/validator.
func NewPlaygroundValidator(v *validator.Validate) *PlaygroundValidator {
	return &PlaygroundValidator{validate: v}
}

// Validate validates the input struct using go-playground/validator.
// Returns validation errors as ErrorDetail objects with code, message, and location.
func (v *PlaygroundValidator) Validate(ctx context.Context, input any, metadata *schema.StructMetadata) []error {
	err := v.validate.StructCtx(ctx, input)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// Non-validation error (e.g., invalid input type)
		return []error{&ErrorDetail{
			Code:    "validation_error",
			Message: err.Error(),
		}}
	}

	errs := make([]error, len(validationErrors))
	for i, e := range validationErrors {
		// Determine correct location prefix
		var location string
		if metadata != nil {
			if loc, ok := metadata.LocationForNamespace(e.Namespace()); ok {
				// Remove struct name prefix from namespace if present
				namespace := e.Namespace()
				if parts := strings.Split(namespace, "."); len(parts) > 1 {
					namespace = strings.Join(parts[1:], ".")
				}
				location = loc + "." + namespace
			} else {
				location = "body." + e.Namespace()
			}
		} else {
			location = "body." + e.Namespace()
		}

		errs[i] = &ErrorDetail{
			Code:     e.Tag(),   // "required", "email", "min", or custom tag
			Message:  e.Error(), // Human-readable message from validator
			Location: location,  // "query.email", "path.id", "header.auth", "body.User.email"
		}
	}

	return errs
}
