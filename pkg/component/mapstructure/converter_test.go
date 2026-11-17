package mapstructure

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	goFieldCount               = "Count"
	goFieldData                = "Data"
	goFieldInner               = "Inner"
	goFieldMyField             = "MyField"
	goFieldName                = "Name"
	goFieldValue               = "Value"
	caseBoolFalse              = "bool false"
	caseBoolTrue               = "bool true"
	caseBytes                  = "bytes"
	caseBytesEmpty             = "bytes empty"
	keyCustomName              = "custom_name"
	caseFloat                  = "float"
	caseFloatNegative          = "float negative"
	caseFloatPositive          = "float positive"
	caseFloat32                = "float32"
	valueHello                 = "hello"
	caseIntNegative            = "int negative"
	caseIntPositive            = "int positive"
	caseIntZero                = "int zero"
	caseInt64                  = "int64"
	valueInvalid               = "invalid"
	caseNil                    = "nil"
	caseString                 = "string"
	caseStringNaN              = "string NaN"
	caseStringEmpty            = "string empty"
	caseStringInfinity         = "string infinity"
	caseStringInvalid          = "string invalid"
	caseStringNegative         = "string negative"
	caseStringNegativeInfinity = "string negative infinity"
	caseStringOverflow         = "string overflow"
	caseStringUnderflow        = "string underflow"
	caseStringValid            = "string valid"
	caseStringZero             = "string zero"
	valueTest                  = "test"
	caseUint                   = "uint"
	caseUint64                 = "uint64"
)

func TestConverter_convertBool(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      bool
		wantError bool
	}{
		// Native bool
		{caseBoolTrue, true, true, false},
		{caseBoolFalse, false, false, false},
		// Native int
		{caseIntPositive, 42, true, false},
		{caseIntZero, 0, false, false},
		{caseIntNegative, -1, true, false},
		{"int8", int8(1), true, false},
		{caseInt64, int64(0), false, false},
		// Native uint
		{"uint positive", uint(42), true, false},
		{"uint zero", uint(0), false, false},
		{"uint8", uint8(1), true, false},
		// Native float
		{"float64 positive", 3.14, true, false},
		{"float64 zero", 0.0, false, false},
		{caseFloat32, float32(1.5), true, false},
		// String parsing
		{"string true", "true", true, false},
		{"string false", "false", false, false},
		{"string TRUE", "TRUE", true, false},
		{"string 1", "1", true, false},
		{"string 0", "0", false, false},
		{caseStringEmpty, "", false, false},
		{caseStringInvalid, valueInvalid, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBool(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Bool())
		})
	}
}

func TestConverter_convertString(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      string
		wantError bool
	}{
		// Native string
		{caseStringValid, valueHello, valueHello, false},
		{caseStringEmpty, "", "", false},
		{"string unicode", "世界", "世界", false},
		{"string with spaces", "hello world", "hello world", false},
		// Native bool
		{caseBoolTrue, true, "1", false},
		{caseBoolFalse, false, "0", false},
		// Native int
		{caseIntPositive, 42, "42", false},
		{caseIntNegative, -42, "-42", false},
		{caseIntZero, 0, "0", false},
		{caseInt64, int64(1234567890), "1234567890", false},
		// Native uint
		{caseUint, uint(42), "42", false},
		{caseUint64, uint64(1234567890), "1234567890", false},
		// Native float
		{caseFloatPositive, 3.14, "3.14", false},
		{caseFloatNegative, -3.14, "-3.14", false},
		{"float integer", 42.0, "42", false},
		// []byte
		{caseBytes, []byte(valueHello), valueHello, false},
		{caseBytesEmpty, []byte{}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertString(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

func TestConverter_convertInt(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int
		wantError bool
	}{
		// Native int
		{caseIntPositive, 42, 42, false},
		{caseIntNegative, -42, -42, false},
		{caseIntZero, 0, 0, false},
		{caseInt64, int64(100), 100, false},
		// Native uint
		{caseUint, uint(42), 42, false},
		{caseUint64, uint64(100), 100, false},
		// Native float
		{caseFloatPositive, 42.9, 42, false},
		{caseFloatNegative, -42.9, -42, false},
		// Native bool
		{caseBoolTrue, true, 1, false},
		{caseBoolFalse, false, 0, false},
		// String parsing
		{caseStringValid, "42", 42, false},
		{caseStringNegative, "-42", -42, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, int(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int8
		wantError bool
	}{
		// Native types
		{caseIntPositive, 127, 127, false},
		{caseIntNegative, -128, -128, false},
		{caseUint, uint(100), 100, false},
		{caseFloat, 50.9, 50, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "127", 127, false},
		{caseStringNegative, "-128", -128, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "128", 0, true},
		{caseStringUnderflow, "-129", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int8(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int16
		wantError bool
	}{
		// Native types
		{caseIntPositive, 32767, 32767, false},
		{caseIntNegative, -32768, -32768, false},
		{caseUint, uint(30000), 30000, false},
		{caseFloat, 1000.9, 1000, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "32767", 32767, false},
		{caseStringNegative, "-32768", -32768, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "32768", 0, true},
		{caseStringUnderflow, "-32769", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int16(result.Int()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertInt32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int32
		wantError bool
	}{
		// Native types
		{caseIntPositive, 1000000, 1000000, false},
		{caseIntNegative, -1000000, -1000000, false},
		{caseUint, uint(1000000), 1000000, false},
		{caseFloat, 1000000.9, 1000000, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "2147483647", 2147483647, false},
		{caseStringNegative, "-2147483648", -2147483648, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "2147483648", 0, true},
		{caseStringUnderflow, "-2147483649", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, int32(result.Int()))
		})
	}
}

func TestConverter_convertInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      int64
		wantError bool
	}{
		// Native types
		{caseIntPositive, 1000000, 1000000, false},
		{caseIntNegative, -1000000, -1000000, false},
		{caseInt64, int64(9223372036854775807), 9223372036854775807, false},
		{caseUint, uint(1000000), 1000000, false},
		{caseFloat, 1000000.9, 1000000, false},
		{caseBoolTrue, true, 1, false},
		{caseBoolFalse, false, 0, false},
		// String
		{caseStringValid, "9223372036854775807", 9223372036854775807, false},
		{caseStringNegative, "-9223372036854775808", -9223372036854775808, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "9223372036854775808", 0, true},
		{caseStringUnderflow, "-9223372036854775809", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertInt64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Int())
		})
	}
}

func TestConverter_convertUint(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint
		wantError bool
	}{
		// Native int
		{caseIntPositive, 42, 42, false},
		{caseIntZero, 0, 0, false},
		{caseIntNegative, -1, 0, true},
		// Native uint
		{caseUint, uint(42), 42, false},
		{caseUint64, uint64(100), 100, false},
		// Native float
		{caseFloatPositive, 42.9, 42, false},
		{caseFloatNegative, -1.5, 0, true},
		// Native bool
		{caseBoolTrue, true, 1, false},
		{caseBoolFalse, false, 0, false},
		// String parsing
		{caseStringValid, "42", 42, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringNegative, "-1", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, uint(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint8(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint8
		wantError bool
	}{
		// Native types
		{caseIntPositive, 255, 255, false},
		{caseIntNegative, -1, 0, true},
		{caseUint, uint(200), 200, false},
		{caseFloat, 100.9, 100, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "255", 255, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "256", 0, true},
		{caseStringNegative, "-1", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint8(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint8(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint16(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint16
		wantError bool
	}{
		// Native types
		{caseIntPositive, 65535, 65535, false},
		{caseIntNegative, -1, 0, true},
		{caseUint, uint(50000), 50000, false},
		{caseFloat, 1000.9, 1000, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "65535", 65535, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "65536", 0, true},
		{caseStringNegative, "-1", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint16(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint16(result.Uint()))
		})
	}
}

//nolint:dupl // Test cases are intentionally similar
func TestConverter_convertUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint32
		wantError bool
	}{
		// Native types
		{caseIntPositive, 1000000, 1000000, false},
		{caseIntNegative, -1, 0, true},
		{caseUint, uint(4294967295), 4294967295, false},
		{caseFloat, 1000000.9, 1000000, false},
		{caseBoolTrue, true, 1, false},
		// String
		{caseStringValid, "4294967295", 4294967295, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "4294967296", 0, true},
		{caseStringNegative, "-1", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:gosec // Intentional conversion for testing
			assert.Equal(t, tt.want, uint32(result.Uint()))
		})
	}
}

func TestConverter_convertUint64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      uint64
		wantError bool
	}{
		// Native types
		{caseIntPositive, 1000000, 1000000, false},
		{caseIntNegative, -1, 0, true},
		{caseUint, uint(1000000), 1000000, false},
		{caseUint64, uint64(18446744073709551615), 18446744073709551615, false},
		{caseFloat, 1000000.9, 1000000, false},
		{caseBoolTrue, true, 1, false},
		{caseBoolFalse, false, 0, false},
		// String
		{caseStringValid, "18446744073709551615", 18446744073709551615, false},
		{caseStringZero, "0", 0, false},
		{caseStringEmpty, "", 0, false},
		{caseStringOverflow, "18446744073709551616", 0, true},
		{caseStringNegative, "-1", 0, true},
		{caseStringInvalid, valueInvalid, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertUint64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Uint())
		})
	}
}

func TestConverter_convertFloat32(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      float32
		wantError bool
	}{
		// Native int
		{caseIntPositive, 42, 42.0, false},
		{caseIntNegative, -42, -42.0, false},
		{caseIntZero, 0, 0.0, false},
		{caseInt64, int64(100), 100.0, false},
		// Native uint
		{caseUint, uint(42), 42.0, false},
		{caseUint64, uint64(100), 100.0, false},
		// Native float
		{caseFloat32, float32(3.14), 3.14, false},
		{"float64", 3.14, 3.14, false},
		// Native bool
		{caseBoolTrue, true, 1.0, false},
		{caseBoolFalse, false, 0.0, false},
		// String parsing
		{"string positive", "3.14", 3.14, false},
		{caseStringNegative, "-3.14", -3.14, false},
		{caseStringZero, "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e10", 1e10, false},
		{caseStringEmpty, "", 0.0, false},
		{caseStringInvalid, valueInvalid, 0, true},
		{caseStringInfinity, "+Inf", float32(math.Inf(1)), false},
		{caseStringNegativeInfinity, "-Inf", float32(math.Inf(-1)), false},
		{caseStringNaN, "NaN", float32(math.NaN()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat32(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := float32(result.Float())
			switch tt.name {
			case caseStringNaN:
				assert.True(t, math.IsNaN(float64(got)))
			case caseStringInfinity, caseStringNegativeInfinity:
				assert.Equal(t, math.IsInf(float64(tt.want), 0), math.IsInf(float64(got), 0))
				if math.IsInf(float64(tt.want), 1) {
					assert.True(t, math.IsInf(float64(got), 1))
				} else if math.IsInf(float64(tt.want), -1) {
					assert.True(t, math.IsInf(float64(got), -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}

func TestConverter_convertFloat64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		want      float64
		wantError bool
	}{
		// Native int
		{caseIntPositive, 42, 42.0, false},
		{caseIntNegative, -42, -42.0, false},
		{caseIntZero, 0, 0.0, false},
		{caseInt64, int64(100), 100.0, false},
		// Native uint
		{caseUint, uint(42), 42.0, false},
		{caseUint64, uint64(100), 100.0, false},
		// Native float
		{caseFloat32, float32(3.14), 3.14, false},
		{"float64", 3.141592653589793, 3.141592653589793, false},
		// Native bool
		{caseBoolTrue, true, 1.0, false},
		{caseBoolFalse, false, 0.0, false},
		// String parsing
		{"string positive", "3.141592653589793", 3.141592653589793, false},
		{caseStringNegative, "-3.141592653589793", -3.141592653589793, false},
		{caseStringZero, "0.0", 0.0, false},
		{"string integer", "42", 42.0, false},
		{"string scientific", "1e100", 1e100, false},
		{caseStringEmpty, "", 0.0, false},
		{caseStringInvalid, valueInvalid, 0, true},
		{caseStringInfinity, "+Inf", math.Inf(1), false},
		{caseStringNegativeInfinity, "-Inf", math.Inf(-1), false},
		{caseStringNaN, "NaN", math.NaN(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertFloat64(tt.input)
			if tt.wantError {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			got := result.Float()
			switch tt.name {
			case caseStringNaN:
				assert.True(t, math.IsNaN(got))
			case caseStringInfinity, caseStringNegativeInfinity:
				assert.Equal(t, math.IsInf(tt.want, 0), math.IsInf(got, 0))
				if math.IsInf(tt.want, 1) {
					assert.True(t, math.IsInf(got, 1))
				} else if math.IsInf(tt.want, -1) {
					assert.True(t, math.IsInf(got, -1))
				}
			default:
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}

func TestConverter_convertBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{name: caseNil, input: nil, want: nil},
		{name: caseBytes, input: []byte{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: caseBytesEmpty, input: []byte{}, want: []byte{}},
		{name: caseString, input: valueHello, want: []byte(valueHello)},
		{name: caseStringEmpty, input: "", want: []byte("")},
		{name: "any slice", input: []any{1, 2, 3}, want: []byte{1, 2, 3}},
		{name: "any slice empty", input: []any{}, want: []byte{}},
		{name: "any slice with ints", input: []any{int(65), int(66), int(67)}, want: []byte{65, 66, 67}},
		{name: "reader", input: io.NopCloser(strings.NewReader(valueTest)), want: []byte(valueTest)},
		{name: "reader empty", input: io.NopCloser(strings.NewReader("")), want: []byte{}},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
		{name: "any slice invalid element", input: []any{1, valueInvalid, 3}, wantErr: true},
		{name: "any slice element over a byte", input: []any{1, 300}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBytes(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			//nolint:forcetypeassert // Test code
			assert.Equal(t, tt.want, result.Interface().([]byte))
		})
	}
}

func TestConverter_convertReadCloser(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		wantContent string
		wantNil     bool
		wantErr     bool
	}{
		{name: caseNil, input: nil, wantNil: true},
		{name: "readcloser", input: io.NopCloser(strings.NewReader(valueHello)), wantContent: valueHello},
		{name: "reader", input: strings.NewReader("world"), wantContent: "world"},
		{name: caseBytes, input: []byte(caseBytes), wantContent: caseBytes},
		{name: caseBytesEmpty, input: []byte{}, wantContent: ""},
		{name: caseString, input: caseString, wantContent: caseString},
		{name: caseStringEmpty, input: "", wantContent: ""},
		{name: "invalid int", input: 42, wantErr: true},
		{name: "invalid bool", input: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertReadCloser(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)

			if tt.wantNil {
				assert.True(t, result.IsNil())

				return
			}

			//nolint:forcetypeassert // Test code
			rc := result.Interface().(io.ReadCloser)
			defer func() { _ = rc.Close() }()

			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(content))
		})
	}
}
