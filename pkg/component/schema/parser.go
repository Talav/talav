package schema

import (
	"fmt"
	"reflect"

	"github.com/talav/talav/pkg/component/tagparser"
)

const (
	// DefaultSchemaTag is the default struct tag name for schema parameters.
	DefaultSchemaTag = "schema"
	DefaultBodyTag   = "body"
)

// structMetadataParser parses struct tags to extract field metadata for OpenAPI parameter and body handling.
type structMetadataParser struct {
	schemaTag string
	bodyTag   string
}

// StructMetadataParserOption configures a structMetadataParser.
type StructMetadataParserOption func(*structMetadataParser)

// newStructMetadataParser creates a new structMetadataParser with the given options.
func newStructMetadataParser(opts ...StructMetadataParserOption) *structMetadataParser {
	p := &structMetadataParser{
		schemaTag: DefaultSchemaTag,
		bodyTag:   DefaultBodyTag,
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// BuildStructMetadata parses the struct type and returns its metadata.
func (p *structMetadataParser) BuildStructMetadata(typ reflect.Type) (*StructMetadata, error) {
	var fields []FieldMetadata

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldMetadata, err := p.buildSchemaField(field, i)
		if err != nil {
			return nil, err
		}
		if fieldMetadata == nil {
			continue
		}

		fields = append(fields, *fieldMetadata)
	}

	return NewStructMetadata(fields)
}

// WithSchemaTag sets a custom schema tag name for the parser.
func WithSchemaTag(tag string) StructMetadataParserOption {
	return func(p *structMetadataParser) {
		p.schemaTag = tag
	}
}

// WithBodyTag sets a custom body tag name for the parser.
func WithBodyTag(tag string) StructMetadataParserOption {
	return func(p *structMetadataParser) {
		p.bodyTag = tag
	}
}

func (p *structMetadataParser) buildSchemaField(field reflect.StructField, index int) (*FieldMetadata, error) {
	// Single lookup for both tags - more efficient
	schemaTag, hasSchema := field.Tag.Lookup(p.schemaTag)
	bodyTag, hasBody := field.Tag.Lookup(p.bodyTag)

	// Early exit for skip tags or unexported fields
	if p.isSkipped(field, hasSchema, schemaTag, hasBody, bodyTag) {
		//nolint:nilnil // nil is a valid return value indicating field should be skipped
		return nil, nil
	}

	switch {
	case hasSchema && hasBody:
		return nil, fmt.Errorf("field %s cannot have both schema and body tags", field.Name)
	case hasSchema:
		return p.parseSchemaTag(field, index, schemaTag)
	case hasBody:
		return p.parseBodyTag(field, index, bodyTag)
	default:
		return p.createDefaultMetadata(field, index), nil
	}
}

func (p *structMetadataParser) isSkipped(field reflect.StructField, hasSchema bool, schemaTag string, hasBody bool, bodyTag string) bool {
	return !field.IsExported() || (hasSchema && schemaTag == "-") || (hasBody && bodyTag == "-")
}

// createDefaultMetadata creates default metadata for fields without tags.
func (p *structMetadataParser) createDefaultMetadata(field reflect.StructField, index int) *FieldMetadata {
	location := LocationQuery
	style := DefaultStyle(location)
	explode := DefaultExplode(style)

	return &FieldMetadata{
		StructFieldName: field.Name,
		ParamName:       field.Name,
		MapKey:          field.Name,
		Type:            field.Type,
		Embedded:        field.Anonymous,
		IsParameter:     true,
		Location:        location,
		Style:           style,
		Explode:         explode,
		Required:        false,
		Index:           index,
	}
}

func (p *structMetadataParser) parseSchemaTag(field reflect.StructField, index int, tagValue string) (*FieldMetadata, error) {
	// Empty tag means no schema tag
	if tagValue == "" {
		//nolint:nilnil // nil is a valid return value indicating no schema tag
		return nil, nil
	}

	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse schema tag: %w", field.Name, err)
	}

	// Use field name if tag name is empty (like JSON standard)
	paramName := tag.Name
	if paramName == "" {
		paramName = field.Name
	}

	// Parse options from tag.Options map
	location, style, explode, required := parseTagOptionsFromMap(tag.Options)

	if err := isValidLocationAndStyle(field.Name, location, style); err != nil {
		return nil, err
	}

	return buildParameterMetadata(field, index, paramName, location, style, explode, required), nil
}

const (
	optKeyLocation = "location"
	optKeyStyle    = "style"
	optKeyExplode  = "explode"
	optKeyRequired = "required"
	optValueTrue   = "true"
)

// parseTagOptionsFromMap parses tag options from a map[string]string (from tagparser).
func parseTagOptionsFromMap(options map[string]string) (location ParameterLocation, style Style, explode bool, required bool) {
	// process location first
	if options[optKeyLocation] != "" {
		location = ParameterLocation(options[optKeyLocation])
	} else {
		location = LocationQuery
	}

	// process style next, it depends on location
	if options[optKeyStyle] != "" {
		style = Style(options[optKeyStyle])
	} else {
		style = DefaultStyle(location)
	}

	explode = extractBoolean(options, optKeyExplode, DefaultExplode(style))
	required = extractBoolean(options, optKeyRequired, false)
	if location == LocationPath {
		required = true
	}

	return location, style, explode, required
}

// isValidLocationAndStyle validates style against location with field context.
func isValidLocationAndStyle(fieldName string, location ParameterLocation, style Style) error {
	// Validate location first
	if !IsLocationValid(location) {
		return fmt.Errorf("field %s: invalid location %q", fieldName, location)
	}

	// Then validate style for that location
	if !IsStyleAllowed(location, style) {
		return fmt.Errorf("field %s: invalid style %q for location %q", fieldName, style, location)
	}

	return nil
}

func (p *structMetadataParser) parseBodyTag(field reflect.StructField, index int, tagValue string) (*FieldMetadata, error) {
	tag, err := tagparser.Parse(tagValue)
	if err != nil {
		return nil, fmt.Errorf("field %s: failed to parse body tag: %w", field.Name, err)
	}

	bodyType, err := parseBodyType(tag.Name)
	if err != nil {
		return nil, fmt.Errorf("field %s: %w", field.Name, err)
	}

	required := extractBoolean(tag.Options, optKeyRequired, false)

	return buildBodyMetadata(field, index, bodyType, required), nil
}

// parseBodyType parses the body type from the tag name.
func parseBodyType(bodyTypeStr string) (BodyType, error) {
	switch bodyTypeStr {
	case "", "structured":
		return BodyTypeStructured, nil
	case "file":
		return BodyTypeFile, nil
	case "multipart":
		return BodyTypeMultipart, nil
	default:
		return "", fmt.Errorf("invalid body type %q (must be 'structured', 'file', or 'multipart')", bodyTypeStr)
	}
}

// buildParameterMetadata creates FieldMetadata for a parameter field.
func buildParameterMetadata(field reflect.StructField, index int, paramName string, location ParameterLocation, style Style, explode bool, required bool) *FieldMetadata {
	return &FieldMetadata{
		StructFieldName: field.Name,
		ParamName:       paramName,
		MapKey:          paramName,
		Type:            field.Type,
		Embedded:        field.Anonymous,
		IsParameter:     true,
		Location:        location,
		Style:           style,
		Explode:         explode,
		Required:        required,
		Index:           index,
	}
}

// buildBodyMetadata creates FieldMetadata for a body field.
func buildBodyMetadata(field reflect.StructField, index int, bodyType BodyType, required bool) *FieldMetadata {
	return &FieldMetadata{
		StructFieldName: field.Name,
		ParamName:       "",
		MapKey:          field.Name,
		Type:            field.Type,
		Embedded:        field.Anonymous,
		IsBody:          true,
		BodyType:        bodyType,
		Required:        required,
		Index:           index,
	}
}

func extractBoolean(options map[string]string, key string, defaultValue bool) bool {
	if value, exists := options[key]; exists {
		if value == "" {
			// Flag form: "required" means true
			return true
		}

		// Value form: "required=true/false"
		return value == optValueTrue
	}

	return defaultValue
}
