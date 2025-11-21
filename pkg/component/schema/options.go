package schema

import "fmt"

// Options configures how parameters are decoded/encoded.
// Options is a validated value object - it must be created using NewOptions or DefaultOptions.
type Options struct {
	location ParameterLocation
	style    Style
	explode  *bool // nil = use default for style
}

// NewOptions creates validated options for the given location and style.
// If style is empty, uses the default style for the location.
// explode is variadic: no argument means use default, one bool argument means explicit value.
// Returns an error if the style is not allowed for the location.
func NewOptions(location ParameterLocation, style Style, explode ...bool) (Options, error) {
	validatedStyle, err := ValidateStyle(location, style)
	if err != nil {
		return Options{}, fmt.Errorf("%w: %w", ErrInvalidOptions, err)
	}

	var explodePtr *bool
	switch len(explode) {
	case 0:
		explodePtr = nil // use default
	case 1:
		explodePtr = &explode[0] // explicit value
	default:
		return Options{}, fmt.Errorf("explode: expected 0 or 1 arguments, got %d", len(explode))
	}

	return Options{
		location: location,
		style:    validatedStyle,
		explode:  explodePtr,
	}, nil
}

// DefaultOptions returns default options for the given parameter location.
// Default style is based on location (see DefaultStyle).
// Default explode is true for form style, false for all other styles.
func DefaultOptions(location ParameterLocation) Options {
	style := DefaultStyle(location)

	// Use default explode (nil) - NewOptions will handle it
	opts, _ := NewOptions(location, style)

	return opts
}

// Style returns the validated style.
func (o Options) Style() Style {
	return o.style
}

// Location returns the parameter location.
func (o Options) Location() ParameterLocation {
	return o.location
}

// getExplode returns the explode value, using default if nil.
func (o Options) getExplode() bool {
	if o.explode != nil {
		return *o.explode
	}

	// Default: true for form, false for others
	return o.style == StyleForm
}
