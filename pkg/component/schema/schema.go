package schema

import (
	"errors"
	"fmt"
	"reflect"
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

var validLocations = []ParameterLocation{
	LocationQuery,
	LocationPath,
	LocationHeader,
	LocationCookie,
}

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

// styleGroup represents a key for grouping fields with the same style and explode value.
type styleGroup struct {
	Style   Style
	Explode bool
}

// BodyType represents the type of request body.
type BodyType string

const (
	BodyTypeStructured BodyType = "structured" // JSON, XML
	BodyTypeFile       BodyType = "file"       // File upload
	BodyTypeMultipart  BodyType = "multipart"  // Multipart form
)

var validBodyTypes = []BodyType{
	BodyTypeStructured,
	BodyTypeFile,
	BodyTypeMultipart,
}

// FieldMetadata represents a cached struct field metadata.
// It can represent both parameter fields (schema tag) and body fields (body tag).
type FieldMetadata struct {
	// StructFieldName is the name of the struct field in Go source code.
	StructFieldName string
	// ParamName is the parameter name as specified in the schema tag (for parameters) or empty (for body fields).
	ParamName string
	// MapKey is the key used to extract the value from the decoded map during unmarshaling.
	MapKey string
	// Index is the field index in the struct (used for reflection-based field access).
	Index int
	// Embedded indicates whether this field is an embedded/anonymous struct field.
	Embedded bool
	// Type is the reflect.Type of the field.
	Type reflect.Type

	// Parameter metadata (for schema tags)
	// IsParameter indicates whether this field represents a parameter (query, path, header, cookie).
	IsParameter bool
	// Location specifies where the parameter is located (query, path, header, cookie).
	Location ParameterLocation
	// Style specifies the serialization style for the parameter (form, simple, deepObject, etc.).
	Style Style
	// Explode indicates whether arrays and objects should be exploded (OpenAPI v3 parameter serialization).
	Explode bool
	// Required indicates whether the parameter is required.
	Required bool

	// Body metadata (for body tags)
	// IsBody indicates whether this field represents a request body.
	IsBody bool
	// BodyType specifies the type of request body (structured, file, multipart).
	BodyType BodyType
}

type StructMetadata struct {
	Fields       []FieldMetadata
	fieldsByName map[string]*FieldMetadata
}

// NewStructMetadata creates a new struct metadata.
func NewStructMetadata(fields []FieldMetadata) (*StructMetadata, error) {
	// Validate all fields and collect errors
	var errs []error
	for i, field := range fields {
		if err := validateField(field); err != nil {
			errs = append(errs, fmt.Errorf("field[%d] %q: %w", i, field.StructFieldName, err))
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %w", errors.Join(errs...))
	}

	// Build map for O(1) lookup by StructFieldName
	fieldsByName := make(map[string]*FieldMetadata, len(fields))
	for i := range fields {
		fieldsByName[fields[i].StructFieldName] = &fields[i]
	}

	return &StructMetadata{
		Fields:       fields,
		fieldsByName: fieldsByName,
	}, nil
}

// BodyField returns the body field if one exists, nil otherwise.
func (m *StructMetadata) BodyField() *FieldMetadata {
	for i := range m.Fields {
		if m.Fields[i].IsBody {
			return &m.Fields[i]
		}
	}

	return nil
}

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

func IsLocationValid(location ParameterLocation) bool {
	return slices.Contains(validLocations, location)
}

func IsBodyTypeValid(bodyType BodyType) bool {
	return slices.Contains(validBodyTypes, bodyType)
}

func DefaultExplode(style Style) bool {
	return style == StyleForm || style == StyleDeepObject
}

// validateField validates a single FieldMetadata and returns an error if invalid.
func validateField(field FieldMetadata) error {
	var errs []error

	errs = append(errs, validateBasicField(field)...)
	errs = append(errs, validateParameterBodyExclusivity(field)...)

	if field.IsParameter {
		errs = append(errs, validateParameterField(field)...)
	}

	if field.IsBody {
		errs = append(errs, validateBodyField(field)...)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// validateBasicField validates basic field properties.
func validateBasicField(field FieldMetadata) []error {
	var errs []error

	if field.StructFieldName == "" {
		errs = append(errs, fmt.Errorf("structFieldName cannot be empty"))
	}

	if field.Type == nil {
		errs = append(errs, fmt.Errorf("type cannot be nil"))
	}

	if field.Index < 0 {
		errs = append(errs, fmt.Errorf("index must be non-negative, got %d", field.Index))
	}

	return errs
}

// validateParameterBodyExclusivity validates that field is either parameter or body, but not both.
func validateParameterBodyExclusivity(field FieldMetadata) []error {
	var errs []error

	if !field.IsParameter && !field.IsBody {
		errs = append(errs, fmt.Errorf("field must be either parameter or body"))
	}

	if field.IsParameter && field.IsBody {
		errs = append(errs, fmt.Errorf("field cannot be both parameter and body"))
	}

	return errs
}

// validateParameterField validates parameter-specific field properties.
func validateParameterField(field FieldMetadata) []error {
	var errs []error

	if field.ParamName == "" {
		errs = append(errs, fmt.Errorf("paramName cannot be empty for parameter fields"))
	}

	if field.MapKey == "" {
		errs = append(errs, fmt.Errorf("mapKey cannot be empty for parameter fields"))
	}

	if !IsLocationValid(field.Location) {
		errs = append(errs, fmt.Errorf("invalid location %q", field.Location))
	}

	// Validate style is allowed for location
	if !IsStyleAllowed(field.Location, field.Style) {
		errs = append(errs, fmt.Errorf(
			"style %q is not allowed for location %q",
			field.Style,
			field.Location,
		))
	}

	return errs
}

// validateBodyField validates body-specific field properties.
func validateBodyField(field FieldMetadata) []error {
	var errs []error

	if !IsBodyTypeValid(field.BodyType) {
		errs = append(errs, fmt.Errorf(
			"bodyType %q is not allowed for body fields",
			field.BodyType,
		))
	}

	return errs
}

func filterByLocation(fields []FieldMetadata, location ParameterLocation) []FieldMetadata {
	var result []FieldMetadata
	for _, field := range fields {
		if field.Location == location {
			result = append(result, field)
		}
	}

	return result
}

func groupByStyle(fields []FieldMetadata) map[styleGroup][]FieldMetadata {
	styleGroups := make(map[styleGroup][]FieldMetadata)
	for _, field := range fields {
		sg := styleGroup{
			Style:   field.Style,
			Explode: field.Explode,
		}
		styleGroups[sg] = append(styleGroups[sg], field)
	}

	return styleGroups
}
