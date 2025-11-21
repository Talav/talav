package schema

import (
	"fmt"
	"slices"
)

// ParameterLocation represents the location of a parameter in an OpenAPI spec.
type ParameterLocation string

const (
	// LocationQuery represents query parameters.
	LocationQuery ParameterLocation = "query"
	// LocationPath represents path parameters.
	LocationPath ParameterLocation = "path"
	// LocationHeader represents header parameters.
	LocationHeader ParameterLocation = "header"
	// LocationCookie represents cookie parameters.
	LocationCookie ParameterLocation = "cookie"
)

// Style represents the serialization style for a parameter.
type Style string

const (
	// StyleMatrix is used for path parameters.
	// Values are prefixed with a semicolon (;) and key-value pairs are separated by an equals sign (=).
	StyleMatrix Style = "matrix"
	// StyleLabel is used for path parameters.
	// Values are prefixed with a period (.).
	StyleLabel Style = "label"
	// StyleForm is commonly used for query and cookie parameters.
	// Values are serialized as form data.
	StyleForm Style = "form"
	// StyleSimple is applicable to path and header parameters.
	// Values are serialized without any additional formatting.
	StyleSimple Style = "simple"
	// StyleSpaceDelimited is used for query parameters.
	// Array values are separated by spaces.
	StyleSpaceDelimited Style = "spaceDelimited"
	// StylePipeDelimited is used for query parameters.
	// Array values are separated by pipes (|).
	StylePipeDelimited Style = "pipeDelimited"
	// StyleDeepObject is used for query parameters.
	// Allows for complex objects to be represented in a deep object style.
	StyleDeepObject Style = "deepObject"
)

// DefaultStyle returns the default style for the given parameter location.
// Default values (based on value of in):
//   - query - form
//   - path - simple
//   - header - simple
//   - cookie - form
func DefaultStyle(location ParameterLocation) Style {
	switch location {
	case LocationQuery:
		return StyleForm
	case LocationPath:
		return StyleSimple
	case LocationHeader:
		return StyleSimple
	case LocationCookie:
		return StyleForm
	default:
		return StyleForm
	}
}

// AllowedStyles returns all allowed style values for the given parameter location.
func AllowedStyles(location ParameterLocation) []Style {
	switch location {
	case LocationQuery:
		return []Style{
			StyleForm,
			StyleSpaceDelimited,
			StylePipeDelimited,
			StyleDeepObject,
		}
	case LocationPath:
		return []Style{
			StyleSimple,
			StyleLabel,
			StyleMatrix,
		}
	case LocationHeader:
		return []Style{
			StyleSimple,
		}
	case LocationCookie:
		return []Style{
			StyleForm,
		}
	default:
		return nil
	}
}

// IsStyleAllowed checks if the given style is allowed for the specified parameter location.
func IsStyleAllowed(location ParameterLocation, style Style) bool {
	allowed := AllowedStyles(location)

	return slices.Contains(allowed, style)
}

// ValidateStyle validates that the given style is allowed for the specified parameter location.
// If style is empty, it returns the default style for the location and nil error.
// Returns an error if the style is not allowed for the location.
func ValidateStyle(location ParameterLocation, style Style) (Style, error) {
	// Empty style means use default
	if style == "" {
		return DefaultStyle(location), nil
	}

	if !IsStyleAllowed(location, style) {
		allowed := AllowedStyles(location)

		return "", fmt.Errorf(
			"style %q is not allowed for parameter location %q. Allowed styles: %v",
			style,
			location,
			allowed,
		)
	}

	return style, nil
}
