package schema

import (
	"fmt"
	"reflect"
	"strconv"
)

// convertBool converts a string value to bool.
// Handles numeric strings: non-zero numbers are true, zero is false.
func convertBool(value string) (reflect.Value, error) {
	if v, err := strconv.ParseBool(value); err == nil {
		return reflect.ValueOf(v), nil
	}
	// Try parsing as number: non-zero is true, zero is false
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return reflect.ValueOf(i != 0), nil
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return reflect.ValueOf(f != 0), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to bool", value)
}

// convertString converts a string value to string.
func convertString(value string) (reflect.Value, error) {
	return reflect.ValueOf(value), nil
}

// convertInt converts a string value to int.
// Handles float strings by truncating (e.g., "12.7" -> 12).
func convertInt(value string) (reflect.Value, error) {
	if v, err := strconv.ParseInt(value, 10, 0); err == nil {
		return reflect.ValueOf(int(v)), nil
	}
	// Try parsing as float and truncating (for cases like "12.7")
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return reflect.ValueOf(int(f)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to int", value)
}

// convertInt8 converts a string value to int8.
func convertInt8(value string) (reflect.Value, error) {
	if v, err := strconv.ParseInt(value, 10, 8); err == nil {
		return reflect.ValueOf(int8(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to int8", value)
}

// convertInt16 converts a string value to int16.
func convertInt16(value string) (reflect.Value, error) {
	if v, err := strconv.ParseInt(value, 10, 16); err == nil {
		return reflect.ValueOf(int16(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to int16", value)
}

// convertInt32 converts a string value to int32.
func convertInt32(value string) (reflect.Value, error) {
	if v, err := strconv.ParseInt(value, 10, 32); err == nil {
		return reflect.ValueOf(int32(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to int32", value)
}

// convertInt64 converts a string value to int64.
func convertInt64(value string) (reflect.Value, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return reflect.ValueOf(v), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to int64", value)
}

// convertUint converts a string value to uint.
func convertUint(value string) (reflect.Value, error) {
	if v, err := strconv.ParseUint(value, 10, 0); err == nil {
		return reflect.ValueOf(uint(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to uint", value)
}

// convertUint8 converts a string value to uint8.
func convertUint8(value string) (reflect.Value, error) {
	if v, err := strconv.ParseUint(value, 10, 8); err == nil {
		return reflect.ValueOf(uint8(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to uint8", value)
}

// convertUint16 converts a string value to uint16.
func convertUint16(value string) (reflect.Value, error) {
	if v, err := strconv.ParseUint(value, 10, 16); err == nil {
		return reflect.ValueOf(uint16(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to uint16", value)
}

// convertUint32 converts a string value to uint32.
func convertUint32(value string) (reflect.Value, error) {
	if v, err := strconv.ParseUint(value, 10, 32); err == nil {
		return reflect.ValueOf(uint32(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to uint32", value)
}

// convertUint64 converts a string value to uint64.
func convertUint64(value string) (reflect.Value, error) {
	if v, err := strconv.ParseUint(value, 10, 64); err == nil {
		return reflect.ValueOf(v), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to uint64", value)
}

// convertFloat32 converts a string value to float32.
func convertFloat32(value string) (reflect.Value, error) {
	if v, err := strconv.ParseFloat(value, 32); err == nil {
		return reflect.ValueOf(float32(v)), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to float32", value)
}

// convertFloat64 converts a string value to float64.
func convertFloat64(value string) (reflect.Value, error) {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return reflect.ValueOf(v), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %q to float64", value)
}
