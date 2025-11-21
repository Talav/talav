package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// Marshaler interface for marshaling Go structs to maps.
type Marshaler interface {
	Marshal(v any) (map[string]any, error)
}

// DefaultMarshaler handles marshaling of Go structs to maps.
type DefaultMarshaler struct {
	tagName    string
	fieldCache *FieldCache
}

// NewMarshaler creates a new marshaler.
func NewDefaultMarshaler(tagName string, fieldCache *FieldCache) Marshaler {
	return &DefaultMarshaler{
		tagName:    tagName,
		fieldCache: fieldCache,
	}
}

// Marshal transforms a Go struct into map[string]any.
func (m *DefaultMarshaler) Marshal(v any) (map[string]any, error) {
	rv := reflect.ValueOf(v)

	// Handle pointer to struct
	if rv.Kind() == kindPtr {
		if rv.IsNil() {
			return nil, fmt.Errorf("cannot marshal nil pointer")
		}
		rv = rv.Elem()
	}

	// Use reflection-based marshaling
	if rv.Kind() != kindStruct {
		return nil, fmt.Errorf("cannot marshal non-struct type %v", rv.Type())
	}

	result := make(map[string]any)
	if err := m.marshalValue(rv, result, ""); err != nil {
		return nil, err
	}

	return result, nil
}

// marshalValue recursively marshals a reflect.Value into the result map.
//
//nolint:cyclop // Complex switch statement by design - handles multiple types
func (m *DefaultMarshaler) marshalValue(rv reflect.Value, result map[string]any, prefix string) error {
	if !rv.IsValid() {
		return nil
	}

	kind := rv.Kind()

	//nolint:exhaustive // Unsupported types are handled in default case
	switch kind {
	case kindPtr:
		return m.marshalPtr(rv, result, prefix)
	case kindSlice:
		return m.marshalSlice(rv, result, prefix)
	case kindStruct:
		return m.marshalStruct(rv, result, prefix)
	case kindBool:
		return m.marshalBool(rv, result, prefix)
	case kindString:
		return m.marshalString(rv, result, prefix)
	case kindInt, kindInt8, kindInt16, kindInt32, kindInt64:
		return m.marshalInt(rv, result, prefix)
	case kindUint, kindUint8, kindUint16, kindUint32, kindUint64:
		return m.marshalUint(rv, result, prefix)
	case kindFloat32, kindFloat64:
		return m.marshalFloat(rv, result, prefix)
	default:
		return fmt.Errorf("%w %v for marshaling", ErrUnsupportedType, kind)
	}
}

// marshalBool marshals a bool value.
func (m *DefaultMarshaler) marshalBool(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("bool value requires a field name")
	}
	result[prefix] = rv.Bool()

	return nil
}

// marshalString marshals a string value.
func (m *DefaultMarshaler) marshalString(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("string value requires a field name")
	}
	result[prefix] = rv.String()

	return nil
}

// marshalInt marshals an int value.
func (m *DefaultMarshaler) marshalInt(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("int value requires a field name")
	}
	result[prefix] = rv.Interface()

	return nil
}

// marshalUint marshals a uint value.
func (m *DefaultMarshaler) marshalUint(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("uint value requires a field name")
	}
	result[prefix] = rv.Interface()

	return nil
}

// marshalFloat marshals a float value.
func (m *DefaultMarshaler) marshalFloat(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("float value requires a field name")
	}
	result[prefix] = rv.Interface()

	return nil
}

// marshalPtr marshals a pointer value.
func (m *DefaultMarshaler) marshalPtr(rv reflect.Value, result map[string]any, prefix string) error {
	if rv.IsNil() {
		if prefix != "" {
			result[prefix] = nil
		}

		return nil
	}

	return m.marshalValue(rv.Elem(), result, prefix)
}

// marshalSlice marshals a slice value.
func (m *DefaultMarshaler) marshalSlice(rv reflect.Value, result map[string]any, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("slice value requires a field name")
	}

	if rv.IsNil() {
		result[prefix] = nil

		return nil
	}

	// Empty slice (not nil) should be marshaled as empty array
	if rv.Len() == 0 {
		result[prefix] = []any{}

		return nil
	}

	arr := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		val, err := m.marshalElement(elem)
		if err != nil {
			return err
		}
		arr[i] = val
	}

	result[prefix] = arr

	return nil
}

// marshalElement marshals a single element (used in slices).
func (m *DefaultMarshaler) marshalElement(elem reflect.Value) (any, error) {
	if !elem.IsValid() {
		return nil, ErrInvalidElement
	}

	//nolint:exhaustive // Only specific kinds are supported for slice elements
	switch elem.Kind() {
	case kindBool, kindString, kindInt, kindInt8, kindInt16, kindInt32, kindInt64,
		kindUint, kindUint8, kindUint16, kindUint32, kindUint64,
		kindFloat32, kindFloat64:
		return elem.Interface(), nil
	case kindPtr:
		if elem.IsNil() {
			return nil, fmt.Errorf("nil pointer element")
		}

		return m.marshalElement(elem.Elem())
	case kindStruct:
		nestedMap := make(map[string]any)
		if err := m.marshalValue(elem, nestedMap, ""); err != nil {
			return nil, err
		}

		return nestedMap, nil
	default:
		return nil, fmt.Errorf("%w %v", ErrUnsupportedSliceElementType, elem.Kind())
	}
}

// marshalStruct marshals a struct value using cached field metadata.
//
//nolint:cyclop // Complex function by design - handles multiple struct field types and embedded structs
func (m *DefaultMarshaler) marshalStruct(rv reflect.Value, result map[string]any, prefix string) error {
	typ := rv.Type()
	fields := m.fieldCache.GetCachedFields(typ, m.tagName)

	for _, field := range fields {
		fieldValue := rv.Field(field.index)
		if !fieldValue.IsValid() {
			continue
		}

		// Check for omitempty
		tag := typ.Field(field.index).Tag.Get(m.tagName)
		omitEmpty := strings.Contains(tag, "omitempty")

		// Handle embedded structs
		if field.embedded {
			if err := m.marshalEmbeddedField(fieldValue, field, result, prefix, omitEmpty); err != nil {
				return err
			}

			continue
		}

		// Check if field should be omitted
		if omitEmpty && m.isZeroValue(fieldValue) {
			continue
		}

		// For slices, handle nil as empty slice if not omitted
		if field.kind == kindSlice && fieldValue.IsNil() && !omitEmpty {
			result[field.mapKey] = []any{}

			continue
		}

		// For nested structs or pointers to structs, create nested map
		if field.kind == kindStruct || (field.kind == kindPtr && fieldValue.Type().Elem().Kind() == kindStruct) {
			nestedMap := make(map[string]any)
			if err := m.marshalValue(fieldValue, nestedMap, ""); err != nil {
				return err
			}
			result[field.mapKey] = nestedMap

			continue
		}

		// Marshal the field value
		if err := m.marshalValue(fieldValue, result, field.mapKey); err != nil {
			return err
		}
	}

	return nil
}

// marshalEmbeddedField handles marshaling of embedded struct fields.
func (m *DefaultMarshaler) marshalEmbeddedField(fieldValue reflect.Value, field cachedField, result map[string]any, prefix string, omitEmpty bool) error {
	if field.kind != kindStruct {
		return nil
	}

	// Check if field should be omitted
	if omitEmpty && m.isZeroValue(fieldValue) {
		return nil
	}

	// Anonymous embedded: promote fields to parent level
	if field.embedded {
		return m.marshalValue(fieldValue, result, prefix)
	}

	// Named embedded: create nested map
	nestedMap := make(map[string]any)
	if err := m.marshalValue(fieldValue, nestedMap, ""); err != nil {
		return err
	}
	result[field.name] = nestedMap

	return nil
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
//
//nolint:cyclop // Complex function by design - handles multiple types
func (m *DefaultMarshaler) isZeroValue(rv reflect.Value) bool {
	if !rv.IsValid() {
		return true
	}

	//nolint:exhaustive // Only specific kinds need zero value checking
	switch rv.Kind() {
	case kindPtr, kindSlice, kindMap, kindInterface, kindChan, kindFunc:
		return rv.IsNil()
	case kindBool:
		return !rv.Bool()
	case kindString:
		return rv.String() == ""
	case kindInt, kindInt8, kindInt16, kindInt32, kindInt64:
		return rv.Int() == 0
	case kindUint, kindUint8, kindUint16, kindUint32, kindUint64:
		return rv.Uint() == 0
	case kindFloat32, kindFloat64:
		return rv.Float() == 0
	case kindStruct:
		// Check if all fields are zero
		for i := 0; i < rv.NumField(); i++ {
			if !m.isZeroValue(rv.Field(i)) {
				return false
			}
		}

		return true
	default:
		return false
	}
}
