package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultUnmarshaler returns a default Unmarshaler instance for testing.
func defaultUnmarshaler() Unmarshaler {
	fieldCache := NewFieldCache()
	converters := NewConverterRegistry()
	unmarshaler := NewDefaultUnmarshaler("schema", fieldCache, converters)

	// Register default converters - need to access the concrete type
	if du, ok := unmarshaler.(*DefaultUnmarshaler); ok {
		du.converters.Register(reflect.TypeOf(bool(false)), convertBool)
		du.converters.Register(reflect.TypeOf(string("")), convertString)
		du.converters.Register(reflect.TypeOf(int(0)), convertInt)
		du.converters.Register(reflect.TypeOf(int8(0)), convertInt8)
		du.converters.Register(reflect.TypeOf(int16(0)), convertInt16)
		du.converters.Register(reflect.TypeOf(int32(0)), convertInt32)
		du.converters.Register(reflect.TypeOf(int64(0)), convertInt64)
		du.converters.Register(reflect.TypeOf(uint(0)), convertUint)
		du.converters.Register(reflect.TypeOf(uint8(0)), convertUint8)
		du.converters.Register(reflect.TypeOf(uint16(0)), convertUint16)
		du.converters.Register(reflect.TypeOf(uint32(0)), convertUint32)
		du.converters.Register(reflect.TypeOf(uint64(0)), convertUint64)
		du.converters.Register(reflect.TypeOf(float32(0)), convertFloat32)
		du.converters.Register(reflect.TypeOf(float64(0)), convertFloat64)
	}

	return unmarshaler
}

func TestUnmarshal_IntTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected int
	}{
		{
			name:     "from string",
			data:     map[string]any{"Value": "123"},
			expected: 123,
		},
		// Converters only work on strings, so non-string conversions are not supported
	}

	unmarshaler := defaultUnmarshaler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value int
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

func TestUnmarshal_IntVariants(t *testing.T) {
	type TestStruct struct {
		IntVal   int
		Int8Val  int8
		Int16Val int16
		Int32Val int32
		Int64Val int64
	}

	data := map[string]any{
		"IntVal":   "42",
		"Int8Val":  "127",
		"Int16Val": "32767",
		"Int32Val": "2147483647",
		"Int64Val": "9223372036854775807",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result.IntVal)
	assert.Equal(t, int8(127), result.Int8Val)
	assert.Equal(t, int16(32767), result.Int16Val)
	assert.Equal(t, int32(2147483647), result.Int32Val)
	assert.Equal(t, int64(9223372036854775807), result.Int64Val)
}

func TestUnmarshal_UintVariants(t *testing.T) {
	type TestStruct struct {
		UintVal   uint
		Uint8Val  uint8
		Uint16Val uint16
		Uint32Val uint32
		Uint64Val uint64
	}

	data := map[string]any{
		"UintVal":   "42",
		"Uint8Val":  "255",
		"Uint16Val": "65535",
		"Uint32Val": "4294967295",
		"Uint64Val": "18446744073709551615",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, uint(42), result.UintVal)
	assert.Equal(t, uint8(255), result.Uint8Val)
	assert.Equal(t, uint16(65535), result.Uint16Val)
	assert.Equal(t, uint32(4294967295), result.Uint32Val)
	assert.Equal(t, uint64(18446744073709551615), result.Uint64Val)
}

func TestUnmarshal_Float32TypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected float32
	}{
		{
			name:     "from string",
			data:     map[string]any{"Value": "1.5"},
			expected: 1.5,
		},
		// Converters only work on strings, so non-string conversions are not supported
	}

	unmarshaler := defaultUnmarshaler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value float32
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

func TestUnmarshal_Float64(t *testing.T) {
	unmarshaler := defaultUnmarshaler()
	tests := []struct {
		name     string
		data     map[string]any
		expected float64
	}{
		{"from string", map[string]any{"Value": "3.14159"}, 3.14159},
		// Converters only work on strings, so non-string conversions are not supported
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value float64
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.InDelta(t, tt.expected, result.Value, 0.0001)
		})
	}
}

func TestUnmarshal_Bool(t *testing.T) {
	unmarshaler := defaultUnmarshaler()
	tests := []struct {
		name     string
		data     map[string]any
		expected bool
	}{
		{"from string true", map[string]any{"Value": "true"}, true},
		{"from string false", map[string]any{"Value": "false"}, false},
		// Converters only work on strings, so non-string conversions are not supported
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value bool
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

func TestUnmarshal_String(t *testing.T) {
	unmarshaler := defaultUnmarshaler()
	tests := []struct {
		name     string
		data     map[string]any
		expected string
	}{
		{"from string", map[string]any{"Value": "hello"}, "hello"},
		// Converters only work on strings, so non-string to string conversion is not supported
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value string
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

func TestUnmarshal_Pointer_NilHandling(t *testing.T) {
	unmarshaler := defaultUnmarshaler()
	tests := []struct {
		name     string
		data     map[string]any
		expected *int
	}{
		{
			name:     "nil value",
			data:     map[string]any{"Value": nil},
			expected: nil,
		},
		{
			name:     "missing value",
			data:     map[string]any{},
			expected: nil,
		},
		{
			name:     "with value",
			data:     map[string]any{"Value": "42"},
			expected: Ptr(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value *int
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, result.Value)
			} else {
				require.NotNil(t, result.Value)
				assert.Equal(t, *tt.expected, *result.Value)
			}
		})
	}
}

func TestUnmarshal_Slice_ArrayHandling(t *testing.T) {
	unmarshaler := defaultUnmarshaler()
	tests := []struct {
		name     string
		data     map[string]any
		expected []int
	}{
		{
			name:     "from []any",
			data:     map[string]any{"Value": []any{"1", "2", "3"}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "empty slice",
			data:     map[string]any{"Value": []any{}},
			expected: []int{},
		},
		{
			name:     "nil value",
			data:     map[string]any{"Value": nil},
			expected: nil,
		},
		{
			name:     "missing value",
			data:     map[string]any{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value []int
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Value)
		})
	}
}

func TestUnmarshal_Struct_NestedStruct(t *testing.T) {
	type NestedStruct struct {
		Field int
	}

	type TestStruct struct {
		Age    int
		Nested NestedStruct
	}

	data := map[string]any{
		"Age": "30",
		"Nested": map[string]any{
			"Field": "42",
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 30, result.Age)
	assert.Equal(t, 42, result.Nested.Field)
}

func TestUnmarshal_Struct_Tags(t *testing.T) {
	type TestStruct struct {
		Value1  int     `schema:"custom_name"`
		Value2  int     `schema:"years"`
		Ignored float32 `schema:"-"`
		Value3  int
	}

	data := map[string]any{
		"custom_name": "10",
		"years":       "25",
		"ignored":     "99.9",
		"Value3":      "42",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Value1)
	assert.Equal(t, 25, result.Value2)
	assert.Equal(t, float32(0), result.Ignored, "Ignored field should remain zero value")
	assert.Equal(t, 42, result.Value3)
}

func TestUnmarshal_Struct_AnonymousEmbedded(t *testing.T) {
	type Embedded struct {
		Field1 int
		Field2 float32
	}

	type TestStruct struct {
		Embedded
		Field3 int
	}

	data := map[string]any{
		"Field1": "10",
		"Field2": "42.5",
		"Field3": "99",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Field1)
	assert.Equal(t, float32(42.5), result.Field2)
	assert.Equal(t, 99, result.Field3)
}

func TestUnmarshal_Struct_NamedEmbedded(t *testing.T) {
	type Embedded struct {
		Field1 int
		Field2 float32
	}

	type TestStruct struct {
		Embedded Embedded
		Field3   int
	}

	data := map[string]any{
		"Embedded": map[string]any{
			"Field1": "10",
			"Field2": "42.5",
		},
		"Field3": "99",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Embedded.Field1)
	assert.Equal(t, float32(42.5), result.Embedded.Field2)
	assert.Equal(t, 99, result.Field3)
}

func TestUnmarshal_Struct_ComplexNested(t *testing.T) {
	type Address struct {
		StreetNumber int
		ZipCode      int
	}

	type Person struct {
		Age     int
		Address Address
		Tags    []int
	}

	data := map[string]any{
		"Age": "28",
		"Address": map[string]any{
			"StreetNumber": "123",
			"ZipCode":      "10001",
		},
		"Tags": []any{"1", "2", "3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result Person
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 28, result.Age)
	assert.Equal(t, 123, result.Address.StreetNumber)
	assert.Equal(t, 10001, result.Address.ZipCode)
	assert.Len(t, result.Tags, 3)
}

func TestUnmarshal_Struct_UnsupportedTypes(t *testing.T) {
	type TestStruct struct {
		Value complex64 // complex64 is not supported, should return error
		Age   int
	}

	data := map[string]any{
		"Value": complex64(1 + 2i),
		"Age":   "30",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no converter registered")
	assert.Contains(t, err.Error(), "complex64")
}

func TestUnmarshal_Struct_MissingFields(t *testing.T) {
	type TestStruct struct {
		Value1 int
		Value2 int
	}

	data := map[string]any{
		"Value1": "10",
		// Value2 is missing
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Value1)
	assert.Zero(t, result.Value2, "Value2 should remain zero value")
}

func TestUnmarshal_Struct_Cache(t *testing.T) {
	// Create codec to test caching behavior
	unmarshaler := defaultUnmarshaler()

	type TestStruct struct {
		Field1 int
		Field2 float32
	}

	data := map[string]any{
		"Field1": "1",
		"Field2": "2.5",
	}

	// First unmarshal - should build cache
	var result1 TestStruct
	err := unmarshaler.Unmarshal(data, &result1)
	require.NoError(t, err)

	// Second unmarshal - should use cache
	var result2 TestStruct
	err = unmarshaler.Unmarshal(data, &result2)
	require.NoError(t, err)

	assert.Equal(t, result1.Field1, result2.Field1, "Cache not working: Field1 differs")
	assert.Equal(t, result1.Field2, result2.Field2, "Cache not working: Field2 differs")
}

func TestUnmarshal_Slice_Ints(t *testing.T) {
	type TestStruct struct {
		Values []int
	}

	data := map[string]any{
		"Values": []any{"1", "2", "3", "4", "5"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 5)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result.Values)
}

func TestUnmarshal_Slice_Float32(t *testing.T) {
	type TestStruct struct {
		Values []float32
	}

	data := map[string]any{
		"Values": []any{"1.1", "2.2", "3.3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 3)
	assert.Equal(t, []float32{1.1, 2.2, 3.3}, result.Values)
}

func TestUnmarshal_Slice_Float64(t *testing.T) {
	type TestStruct struct {
		Values []float64
	}

	data := map[string]any{
		"Values": []any{"1.1", "2.2", "3.3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 3)
	assert.InDeltaSlice(t, []float64{1.1, 2.2, 3.3}, result.Values, 0.0001)
}

func TestUnmarshal_Slice_Bool(t *testing.T) {
	type TestStruct struct {
		Values []bool
	}

	data := map[string]any{
		"Values": []any{"true", "false", "true"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 3)
	assert.Equal(t, []bool{true, false, true}, result.Values)
}

func TestUnmarshal_Slice_String(t *testing.T) {
	type TestStruct struct {
		Values []string
	}

	data := map[string]any{
		"Values": []any{"a", "b", "c"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 3)
	assert.Equal(t, []string{"a", "b", "c"}, result.Values)
}

func TestUnmarshal_Slice_Uint(t *testing.T) {
	type TestStruct struct {
		Values []uint
	}

	data := map[string]any{
		"Values": []any{"1", "2", "3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result.Values, 3)
	assert.Equal(t, []uint{1, 2, 3}, result.Values)
}

func TestUnmarshal_Pointer_Int(t *testing.T) {
	type TestStruct struct {
		Value *int
	}

	data := map[string]any{
		"Value": "42",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, 42, *result.Value)
}

func TestUnmarshal_Pointer_Float32(t *testing.T) {
	type TestStruct struct {
		Value *float32
	}

	data := map[string]any{
		"Value": "3.14",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, float32(3.14), *result.Value)
}

func TestUnmarshal_Pointer_Float64(t *testing.T) {
	type TestStruct struct {
		Value *float64
	}

	data := map[string]any{
		"Value": "3.14159",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.InDelta(t, 3.14159, *result.Value, 0.0001)
}

func TestUnmarshal_Pointer_Bool(t *testing.T) {
	type TestStruct struct {
		Value *bool
	}

	data := map[string]any{
		"Value": "true",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.True(t, *result.Value)
}

func TestUnmarshal_Pointer_String(t *testing.T) {
	type TestStruct struct {
		Value *string
	}

	data := map[string]any{
		"Value": "hello",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, "hello", *result.Value)
}

func TestUnmarshal_Pointer_Uint(t *testing.T) {
	type TestStruct struct {
		Value *uint
	}

	data := map[string]any{
		"Value": "42",
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, uint(42), *result.Value)
}

func TestUnmarshal_ErrorCases(t *testing.T) {
	t.Run("unsupported type complex64", func(t *testing.T) {
		type TestStruct struct {
			Value complex64
		}
		data := map[string]any{"Value": complex64(1 + 2i)}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no converter registered")
		assert.Contains(t, err.Error(), "complex64")
	})

	t.Run("cannot convert map to int", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}
		data := map[string]any{"Value": map[string]any{"key": "val"}}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no converter registered")
	})

	t.Run("cannot convert slice to int", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}
		data := map[string]any{"Value": []any{1, 2, 3}}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no converter registered")
	})

	t.Run("cannot convert string to int (invalid)", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}
		data := map[string]any{"Value": "not a number"}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("cannot convert negative to uint", func(t *testing.T) {
		type TestStruct struct {
			Value uint
		}
		data := map[string]any{"Value": "-1"}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("cannot convert string to struct", func(t *testing.T) {
		type Nested struct {
			Field int
		}
		type TestStruct struct {
			Value Nested
		}
		data := map[string]any{"Value": "not a map"}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("cannot convert non-slice to slice", func(t *testing.T) {
		type TestStruct struct {
			Value []int
		}
		data := map[string]any{"Value": "not a slice"}
		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

// TestUnmarshal_SliceOfPointers tests slice of pointers ([]*int).
func TestUnmarshal_SliceOfPointers(t *testing.T) {
	type TestStruct struct {
		Values []*int
	}

	data := map[string]any{
		"Values": []any{"1", "2", "3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Len(t, result.Values, 3)
	require.NotNil(t, result.Values[0])
	require.NotNil(t, result.Values[1])
	require.NotNil(t, result.Values[2])
	assert.Equal(t, 1, *result.Values[0])
	assert.Equal(t, 2, *result.Values[1])
	assert.Equal(t, 3, *result.Values[2])
}

// TestUnmarshal_PointerToSlice tests pointer to slice (*[]int).
func TestUnmarshal_PointerToSlice(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected []int
		wantNil  bool
	}{
		{
			name:     "with values",
			data:     map[string]any{"Value": []any{"1", "2", "3"}},
			expected: []int{1, 2, 3},
			wantNil:  false,
		},
		{
			name:     "nil value",
			data:     map[string]any{"Value": nil},
			expected: nil,
			wantNil:  true,
		},
		{
			name:     "missing value",
			data:     map[string]any{},
			expected: nil,
			wantNil:  true,
		},
		{
			name:     "empty slice",
			data:     map[string]any{"Value": []any{}},
			expected: []int{},
			wantNil:  false,
		},
	}

	unmarshaler := defaultUnmarshaler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value *[]int
			}

			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, result.Value)
			} else {
				require.NotNil(t, result.Value)
				assert.Equal(t, tt.expected, *result.Value)
			}
		})
	}
}

// TestUnmarshal_PointerToSliceOfPointers tests pointer to slice of pointers (*[]*int).
func TestUnmarshal_PointerToSliceOfPointers(t *testing.T) {
	type TestStruct struct {
		Values *[]*int
	}

	data := map[string]any{
		"Values": []any{"1", "2", "3"},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Values)
	require.Len(t, *result.Values, 3)
	require.NotNil(t, (*result.Values)[0])
	require.NotNil(t, (*result.Values)[1])
	require.NotNil(t, (*result.Values)[2])
	assert.Equal(t, 1, *(*result.Values)[0])
	assert.Equal(t, 2, *(*result.Values)[1])
	assert.Equal(t, 3, *(*result.Values)[2])
}

// TestUnmarshal_SliceOfStructs tests slice of structs ([]S1).
func TestUnmarshal_SliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	type TestStruct struct {
		Items []Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{"ID": "1", "Name": "first"},
			map[string]any{"ID": "2", "Name": "second"},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Items[0].ID)
	assert.Equal(t, "first", result.Items[0].Name)
	assert.Equal(t, 2, result.Items[1].ID)
	assert.Equal(t, "second", result.Items[1].Name)
}

// TestUnmarshal_SliceOfPointersToStructs tests slice of pointers to structs ([]*S1).
func TestUnmarshal_SliceOfPointersToStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	type TestStruct struct {
		Items []*Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{"ID": "1", "Name": "first"},
			map[string]any{"ID": "2", "Name": "second"},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	require.NotNil(t, result.Items[0])
	require.NotNil(t, result.Items[1])
	assert.Equal(t, 1, result.Items[0].ID)
	assert.Equal(t, "first", result.Items[0].Name)
	assert.Equal(t, 2, result.Items[1].ID)
	assert.Equal(t, "second", result.Items[1].Name)
}

// TestUnmarshal_PointerToSliceOfStructs tests pointer to slice of structs (*[]S1).
func TestUnmarshal_PointerToSliceOfStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	type TestStruct struct {
		Items *[]Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{"ID": "1", "Name": "first"},
			map[string]any{"ID": "2", "Name": "second"},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Items)
	require.Len(t, *result.Items, 2)
	assert.Equal(t, 1, (*result.Items)[0].ID)
	assert.Equal(t, "first", (*result.Items)[0].Name)
	assert.Equal(t, 2, (*result.Items)[1].ID)
	assert.Equal(t, "second", (*result.Items)[1].Name)
}

// TestUnmarshal_PointerToSliceOfPointersToStructs tests pointer to slice of pointers to structs (*[]*S1).
func TestUnmarshal_PointerToSliceOfPointersToStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	type TestStruct struct {
		Items *[]*Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{"ID": "1", "Name": "first"},
			map[string]any{"ID": "2", "Name": "second"},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Items)
	require.Len(t, *result.Items, 2)
	require.NotNil(t, (*result.Items)[0])
	require.NotNil(t, (*result.Items)[1])
	assert.Equal(t, 1, (*result.Items)[0].ID)
	assert.Equal(t, "first", (*result.Items)[0].Name)
	assert.Equal(t, 2, (*result.Items)[1].ID)
	assert.Equal(t, "second", (*result.Items)[1].Name)
}

// TestUnmarshal_DeeplyNestedStructs tests 3+ levels of nesting.
func TestUnmarshal_DeeplyNestedStructs(t *testing.T) {
	type Level3 struct {
		Value int
	}

	type Level2 struct {
		Level3 Level3
	}

	type Level1 struct {
		Level2 Level2
	}

	type TestStruct struct {
		Level1 Level1
	}

	data := map[string]any{
		"Level1": map[string]any{
			"Level2": map[string]any{
				"Level3": map[string]any{
					"Value": "42",
				},
			},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result.Level1.Level2.Level3.Value)
}

// TestUnmarshal_NestedStructsInSlices tests nested structs within slices.
func TestUnmarshal_NestedStructsInSlices(t *testing.T) {
	type Nested struct {
		Value int
	}

	type Item struct {
		Nested Nested
	}

	type TestStruct struct {
		Items []Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{
				"Nested": map[string]any{"Value": "1"},
			},
			map[string]any{
				"Nested": map[string]any{"Value": "2"},
			},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Items[0].Nested.Value)
	assert.Equal(t, 2, result.Items[1].Nested.Value)
}

// TestUnmarshal_SlicesInNestedStructs tests slices within nested structs within slices.
func TestUnmarshal_SlicesInNestedStructs(t *testing.T) {
	type Tag struct {
		Name string
	}

	type Item struct {
		Tags []Tag
	}

	type TestStruct struct {
		Items []Item
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{
				"Tags": []any{
					map[string]any{"Name": "tag1"},
					map[string]any{"Name": "tag2"},
				},
			},
			map[string]any{
				"Tags": []any{
					map[string]any{"Name": "tag3"},
				},
			},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	require.Len(t, result.Items[0].Tags, 2)
	require.Len(t, result.Items[1].Tags, 1)
	assert.Equal(t, "tag1", result.Items[0].Tags[0].Name)
	assert.Equal(t, "tag2", result.Items[0].Tags[1].Name)
	assert.Equal(t, "tag3", result.Items[1].Tags[0].Name)
}

// TestUnmarshal_TypeAlias tests type alias with explicit converter registration.
func TestUnmarshal_TypeAlias(t *testing.T) {
	type IntAlias int

	// Register converter for type alias (required, as per gorilla/schema behavior)
	fieldCache := NewFieldCache()
	converters := NewConverterRegistry()
	converters.Register(reflect.TypeOf(bool(false)), convertBool)
	converters.Register(reflect.TypeOf(string("")), convertString)
	converters.Register(reflect.TypeOf(int(0)), convertInt)
	converters.Register(reflect.TypeOf(int8(0)), convertInt8)
	converters.Register(reflect.TypeOf(int16(0)), convertInt16)
	converters.Register(reflect.TypeOf(int32(0)), convertInt32)
	converters.Register(reflect.TypeOf(int64(0)), convertInt64)
	converters.Register(reflect.TypeOf(uint(0)), convertUint)
	converters.Register(reflect.TypeOf(uint8(0)), convertUint8)
	converters.Register(reflect.TypeOf(uint16(0)), convertUint16)
	converters.Register(reflect.TypeOf(uint32(0)), convertUint32)
	converters.Register(reflect.TypeOf(uint64(0)), convertUint64)
	converters.Register(reflect.TypeOf(float32(0)), convertFloat32)
	converters.Register(reflect.TypeOf(float64(0)), convertFloat64)
	converters.Register(reflect.TypeOf(IntAlias(0)), func(value string) (reflect.Value, error) {
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(IntAlias(v)), nil
	})
	unmarshaler := NewDefaultUnmarshaler("schema", fieldCache, converters)

	t.Run("basic type alias", func(t *testing.T) {
		type TestStruct struct {
			Value IntAlias
		}

		data := map[string]any{"Value": "42"}
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, IntAlias(42), result.Value)
	})

	t.Run("pointer to type alias", func(t *testing.T) {
		type TestStruct struct {
			Value *IntAlias
		}

		data := map[string]any{"Value": "42"}
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Value)
		assert.Equal(t, IntAlias(42), *result.Value)
	})

	t.Run("slice of type alias", func(t *testing.T) {
		type TestStruct struct {
			Values []IntAlias
		}

		data := map[string]any{"Values": []any{"1", "2", "3"}}
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.NoError(t, err)
		require.Len(t, result.Values, 3)
		assert.Equal(t, IntAlias(1), result.Values[0])
		assert.Equal(t, IntAlias(2), result.Values[1])
		assert.Equal(t, IntAlias(3), result.Values[2])
	})
}

// TestUnmarshal_EmptyZeroValues tests empty and zero value edge cases.
func TestUnmarshal_EmptyZeroValues(t *testing.T) {
	type TestStruct struct {
		EmptyString string
		ZeroInt     int
		ZeroFloat   float32
		FalseBool   bool
		NilSlice    []int
		EmptySlice  []int
	}

	data := map[string]any{
		"emptystring": "",
		"zeroint":     "0",
		"zerofloat":   "0.0",
		"falsebool":   "false",
		"emptyslice":  []any{},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Equal(t, "", result.EmptyString)
	assert.Equal(t, 0, result.ZeroInt)
	assert.Equal(t, float32(0), result.ZeroFloat)
	assert.False(t, result.FalseBool)
	assert.Nil(t, result.NilSlice)
	assert.Empty(t, result.EmptySlice)
}

// TestUnmarshal_ZeroValueVsMissing tests distinction between zero values and missing fields.
func TestUnmarshal_ZeroValueVsMissing(t *testing.T) {
	type TestStruct struct {
		ExplicitZero int
		Missing      int
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected struct {
			explicitZero int
			missing      int
		}
	}{
		{
			name: "explicit zero vs missing",
			data: map[string]any{
				"explicitzero": "0",
			},
			expected: struct {
				explicitZero int
				missing      int
			}{
				explicitZero: 0,
				missing:      0, // Both are zero, but one was explicitly set
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unmarshaler := defaultUnmarshaler()
			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.explicitZero, result.ExplicitZero)
			assert.Equal(t, tt.expected.missing, result.Missing)
		})
	}
}

// TestUnmarshal_PointerToStruct tests pointer to struct.
func TestUnmarshal_PointerToStruct(t *testing.T) {
	type Nested struct {
		Value int
	}

	tests := []struct {
		name     string
		data     map[string]any
		wantNil  bool
		expected *Nested
	}{
		{
			name: "with value",
			data: map[string]any{
				"Nested": map[string]any{"Value": "42"},
			},
			wantNil:  false,
			expected: &Nested{Value: 42},
		},
		{
			name:    "nil value",
			data:    map[string]any{"Nested": nil},
			wantNil: true,
		},
		{
			name:    "missing value",
			data:    map[string]any{},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Nested *Nested
			}

			unmarshaler := defaultUnmarshaler()
			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, result.Nested)
			} else {
				require.NotNil(t, result.Nested)
				assert.Equal(t, tt.expected.Value, result.Nested.Value)
			}
		})
	}
}

// TestUnmarshal_PointerToStructWithPointerFields tests pointer to struct with pointer fields.
func TestUnmarshal_PointerToStructWithPointerFields(t *testing.T) {
	type Nested struct {
		Value *int
	}

	type TestStruct struct {
		Nested *Nested
	}

	data := map[string]any{
		"Nested": map[string]any{
			"Value": "42",
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Nested)
	require.NotNil(t, result.Nested.Value)
	assert.Equal(t, 42, *result.Nested.Value)
}

// TestUnmarshal_FieldPathErrors tests error messages include proper field paths.
func TestUnmarshal_FieldPathErrors(t *testing.T) {
	t.Run("nested field path", func(t *testing.T) {
		type Nested struct {
			Value int
		}
		type TestStruct struct {
			Nested Nested
		}

		data := map[string]any{
			"Nested": map[string]any{
				"Value": "not a number",
			},
		}

		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Nested.Value")
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("deeply nested field path", func(t *testing.T) {
		type Level3 struct {
			Value int
		}
		type Level2 struct {
			Level3 Level3
		}
		type Level1 struct {
			Level2 Level2
		}
		type TestStruct struct {
			Level1 Level1
		}

		data := map[string]any{
			"Level1": map[string]any{
				"Level2": map[string]any{
					"Level3": map[string]any{
						"Value": "invalid",
					},
				},
			},
		}

		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Level1.Level2.Level3.Value")
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("slice index in error path", func(t *testing.T) {
		type TestStruct struct {
			Items []int
		}

		data := map[string]any{
			"Items": []any{"1", "invalid", "3"},
		}

		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Items[1]")
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

// TestUnmarshal_ConversionOverflow tests overflow cases.
func TestUnmarshal_ConversionOverflow(t *testing.T) {
	t.Run("uint64 overflow to int64", func(t *testing.T) {
		type TestStruct struct {
			Value int64
		}

		// uint64 value larger than int64 max
		data := map[string]any{
			"Value": fmt.Sprintf("%d", uint64(1<<63+1)),
		}

		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("negative to uint", func(t *testing.T) {
		type TestStruct struct {
			Value uint
		}

		data := map[string]any{
			"Value": "-1",
		}

		unmarshaler := defaultUnmarshaler()
		var result TestStruct
		err := unmarshaler.Unmarshal(data, &result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert")
	})
}

// TestUnmarshal_StringConversionEdgeCases tests string to number edge cases.
func TestUnmarshal_StringConversionEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name:      "empty string to int",
			data:      map[string]any{"Value": ""},
			wantError: true,
		},
		{
			name:      "whitespace string to int",
			data:      map[string]any{"Value": "   "},
			wantError: true,
		},
		{
			name:      "invalid format to int",
			data:      map[string]any{"Value": "abc"},
			wantError: true,
		},
		{
			name:      "empty string to float",
			data:      map[string]any{"Value": ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Value int
			}

			unmarshaler := defaultUnmarshaler()
			var result TestStruct
			err := unmarshaler.Unmarshal(tt.data, &result)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot convert")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestUnmarshal_ComplexCombinedScenario tests all features combined.
func TestUnmarshal_ComplexCombinedScenario(t *testing.T) {
	type Tag struct {
		Name string
	}

	type Metadata struct {
		Tags []Tag
	}

	type Item struct {
		ID       int
		Metadata Metadata
	}

	type Container struct {
		Items *[]*Item
	}

	type TestStruct struct {
		Container Container
	}

	data := map[string]any{
		"Container": map[string]any{
			"Items": []any{
				map[string]any{
					"ID": "1",
					"Metadata": map[string]any{
						"Tags": []any{
							map[string]any{"Name": "tag1"},
						},
					},
				},
				map[string]any{
					"ID": "2",
					"Metadata": map[string]any{
						"Tags": []any{
							map[string]any{"Name": "tag2"},
						},
					},
				},
			},
		},
	}

	unmarshaler := defaultUnmarshaler()
	var result TestStruct
	err := unmarshaler.Unmarshal(data, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Container.Items)
	require.Len(t, *result.Container.Items, 2)
	require.NotNil(t, (*result.Container.Items)[0])
	require.NotNil(t, (*result.Container.Items)[1])
	assert.Equal(t, 1, (*result.Container.Items)[0].ID)
	assert.Equal(t, 2, (*result.Container.Items)[1].ID)
	require.Len(t, (*result.Container.Items)[0].Metadata.Tags, 1)
	require.Len(t, (*result.Container.Items)[1].Metadata.Tags, 1)
	assert.Equal(t, "tag1", (*result.Container.Items)[0].Metadata.Tags[0].Name)
	assert.Equal(t, "tag2", (*result.Container.Items)[1].Metadata.Tags[0].Name)
}

func TestBuildFieldPath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		field    string
		expected string
	}{
		{"root", "", "name", "name"},
		{"nested", "user", "name", "user.name"},
		{"deep", "user.profile", "email", "user.profile.email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFieldPath(tt.base, tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}
