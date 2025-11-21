package schema

import (
	"fmt"
	"reflect"
)

// Unmarshaler interface for unmarshaling maps to Go structs.
type Unmarshaler interface {
	Unmarshal(data map[string]any, result any) error
}

// DefaultUnmarshaler handles unmarshaling of maps to Go structs.
type DefaultUnmarshaler struct {
	tagName    string
	fieldCache *FieldCache
	converters *ConverterRegistry
}

// NewUnmarshaler creates a new unmarshaler.
func NewDefaultUnmarshaler(tagName string, fieldCache *FieldCache, converters *ConverterRegistry) Unmarshaler {
	return &DefaultUnmarshaler{
		tagName:    tagName,
		fieldCache: fieldCache,
		converters: converters,
	}
}

// Unmarshal transforms map[string]any into a Go struct pointed to by result.
// result must be a pointer to the target type.
func (u *DefaultUnmarshaler) Unmarshal(data map[string]any, result any) error {
	rv := reflect.ValueOf(result)
	if rv.Kind() != kindPtr {
		return fmt.Errorf("result must be a pointer")
	}

	if rv.IsNil() {
		return fmt.Errorf("result pointer is nil")
	}

	rv = rv.Elem()

	// Use reflection-based unmarshaling
	return u.unmarshalValue(data, rv, "")
}

// unmarshalValue recursively unmarshals a value into the reflect.Value.
func (u *DefaultUnmarshaler) unmarshalValue(data any, rv reflect.Value, fieldPath string) error {
	if !rv.CanSet() {
		return nil
	}

	kind := rv.Kind()
	typ := rv.Type()

	// Converters only applicable to strings
	str, useConverter := data.(string)

	// Try converter only if data is not already in correct shape
	if useConverter {
		if conv, ok := u.converters.Find(typ); ok {
			converted, err := conv(str)
			if err != nil {
				return fmt.Errorf("%s: cannot convert %T to %v: %w", fieldPath, data, typ, err)
			}
			rv.Set(converted)

			return nil
		}
	}

	//nolint:exhaustive // Unsupported types are handled in default case with error
	switch kind {
	case kindPtr:
		return u.unmarshalPtr(data, rv, fieldPath)
	case kindSlice:
		return u.unmarshalSlice(data, rv, fieldPath)
	case kindStruct:
		return u.unmarshalStruct(data, rv, fieldPath)
	default:
		return fmt.Errorf("%s: no converter registered for type %v", fieldPath, typ)
	}
}

// unmarshalPtr unmarshals a pointer value.
func (u *DefaultUnmarshaler) unmarshalPtr(data any, rv reflect.Value, fieldPath string) error {
	// If data is nil or missing, set pointer to nil
	if data == nil {
		rv.Set(reflect.Zero(rv.Type()))

		return nil
	}

	// If pointer is nil, allocate new instance
	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	// Recursively unmarshal the pointed-to type
	return u.unmarshalValue(data, rv.Elem(), fieldPath)
}

// unmarshalSlice unmarshals a slice value.
func (u *DefaultUnmarshaler) unmarshalSlice(data any, rv reflect.Value, fieldPath string) error {
	// nil is acceptable for slices
	if data == nil {
		rv.Set(reflect.Zero(rv.Type()))

		return nil
	}

	// Expect []any in map data
	arr, ok := data.([]any)
	if !ok {
		return fmt.Errorf("%s: cannot convert %T to slice", fieldPath, data)
	}

	// Create new slice with appropriate length
	slice := reflect.MakeSlice(rv.Type(), len(arr), len(arr))

	// Recursively unmarshal each element
	for i := range len(arr) {
		elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
		if err := u.unmarshalValue(arr[i], slice.Index(i), elemPath); err != nil {
			return err
		}
	}
	rv.Set(slice)

	return nil
}

// unmarshalStruct unmarshals a struct value using cached field metadata.
func (u *DefaultUnmarshaler) unmarshalStruct(data any, rv reflect.Value, fieldPath string) error {
	// Expect map[string]any for struct data
	dataMap, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: cannot convert %T to struct", fieldPath, data)
	}

	// Get cached fields
	typ := rv.Type()
	fields := u.fieldCache.GetCachedFields(typ, u.tagName)

	// Process each cached field
	for _, field := range fields {
		fieldValue := rv.Field(field.index)

		// Handle embedded structs
		if field.embedded {
			if err := u.unmarshalEmbeddedField(dataMap, fieldValue, field, fieldPath); err != nil {
				return err
			}

			continue
		}

		// Get value from map using precomputed key
		value, exists := dataMap[field.mapKey]
		if !exists {
			continue // Skip fields that don't exist in the map
		}

		// Unmarshal the field value (handles converters and built-in conversion)
		fullPath := buildFieldPath(fieldPath, field.mapKey)
		if err := u.unmarshalValue(value, fieldValue, fullPath); err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
	}

	return nil
}

// unmarshalEmbeddedField handles unmarshaling of embedded struct fields.
func (u *DefaultUnmarshaler) unmarshalEmbeddedField(dataMap map[string]any, fieldValue reflect.Value, field cachedField, fieldPath string) error {
	if field.kind != kindStruct {
		return nil
	}

	// Check if there's a nested map with the field name (named embedded)
	if nestedMap, exists := dataMap[field.name]; exists {
		if nestedData, ok := nestedMap.(map[string]any); ok {
			// Named embedded: unmarshal from nested map
			return u.unmarshalValue(nestedData, fieldValue, fieldPath)
		}
	}

	// Anonymous embedded: pass entire data map (promoted fields)
	return u.unmarshalValue(dataMap, fieldValue, fieldPath)
}

// buildFieldPath builds a field path for error messages.
func buildFieldPath(base, field string) string {
	if base == "" {
		return field
	}

	return base + "." + field
}
