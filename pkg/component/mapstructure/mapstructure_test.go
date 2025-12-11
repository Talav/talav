package mapstructure

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCacheBuilder builds metadata for test structs.
func testCacheBuilder(typ reflect.Type) (*StructMetadata, error) {
	fields := make([]FieldMetadata, 0, typ.NumField())
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		fields = append(fields, FieldMetadata{
			StructFieldName: f.Name,
			MapKey:          f.Name,
			Index:           i,
			Type:            f.Type,
			Embedded:        f.Anonymous,
		})
	}

	return &StructMetadata{Fields: fields}, nil
}

func testUnmarshaler() *Unmarshaler {
	cache := NewStructMetadataCache(CacheBuilderFunc(testCacheBuilder))

	return &Unmarshaler{
		fieldCache: cache,
		converters: NewDefaultConverterRegistry(nil),
	}
}

func TestUnmarshaler_Unmarshal_Slices(t *testing.T) {
	type SliceInt struct {
		Values []int
	}
	type SliceByte struct {
		Data []byte
	}
	type SliceString struct {
		Names []string
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:     "slice of any to int",
			data:     map[string]any{"Values": []any{1, 2, 3}},
			target:   &SliceInt{},
			expected: &SliceInt{Values: []int{1, 2, 3}},
		},
		{
			name:     "slice of int to int",
			data:     map[string]any{"Values": []int{10, 20, 30}},
			target:   &SliceInt{},
			expected: &SliceInt{Values: []int{10, 20, 30}},
		},
		{
			name:     "slice of bytes direct",
			data:     map[string]any{"Data": []byte{1, 2, 3, 4, 5}},
			target:   &SliceByte{},
			expected: &SliceByte{Data: []byte{1, 2, 3, 4, 5}},
		},
		{
			name:     "slice of any to bytes",
			data:     map[string]any{"Data": []any{1, 2, 3, 4, 5}},
			target:   &SliceByte{},
			expected: &SliceByte{Data: []byte{1, 2, 3, 4, 5}},
		},
		{
			name:     "slice of strings",
			data:     map[string]any{"Names": []string{"alice", "bob", "charlie"}},
			target:   &SliceString{},
			expected: &SliceString{Names: []string{"alice", "bob", "charlie"}},
		},
		{
			name:     "nil slice",
			data:     map[string]any{"Data": nil},
			target:   &SliceByte{},
			expected: &SliceByte{Data: nil},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_BasicTypes(t *testing.T) {
	type Basic struct {
		Name   string
		Age    int
		Score  float64
		Active bool
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected Basic
	}{
		{
			name: "all fields",
			data: map[string]any{
				"Name":   "test",
				"Age":    25,
				"Score":  95.5,
				"Active": true,
			},
			expected: Basic{Name: "test", Age: 25, Score: 95.5, Active: true},
		},
		{
			name: "missing field uses zero value",
			data: map[string]any{
				"Name": "test",
			},
			expected: Basic{Name: "test", Age: 0, Score: 0, Active: false},
		},
		{
			name:     "empty map",
			data:     map[string]any{},
			expected: Basic{},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Basic
			err := u.Unmarshal(tt.data, &result)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnmarshaler_Unmarshal_NestedStruct(t *testing.T) {
	type Inner struct {
		Value int
	}
	type Outer struct {
		Inner Inner
	}
	type DeepNested struct {
		Level1 struct {
			Level2 struct {
				Value string
			}
		}
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name: "single level nested",
			data: map[string]any{
				"Inner": map[string]any{"Value": 42},
			},
			target:   &Outer{},
			expected: &Outer{Inner: Inner{Value: 42}},
		},
		{
			name: "deep nested",
			data: map[string]any{
				"Level1": map[string]any{
					"Level2": map[string]any{
						"Value": "deep",
					},
				},
			},
			target:   &DeepNested{},
			expected: &DeepNested{Level1: struct{ Level2 struct{ Value string } }{Level2: struct{ Value string }{Value: "deep"}}},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_Pointers(t *testing.T) {
	type WithPointer struct {
		Value *int
	}
	type WithPointerString struct {
		Name *string
	}

	intVal := 42
	strVal := "test"

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:     "non-nil pointer",
			data:     map[string]any{"Value": 42},
			target:   &WithPointer{},
			expected: &WithPointer{Value: &intVal},
		},
		{
			name:     "nil pointer",
			data:     map[string]any{"Value": nil},
			target:   &WithPointer{},
			expected: &WithPointer{Value: nil},
		},
		{
			name:     "string pointer",
			data:     map[string]any{"Name": "test"},
			target:   &WithPointerString{},
			expected: &WithPointerString{Name: &strVal},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_Errors(t *testing.T) {
	type Target struct {
		Data []int
		Name string
	}

	tests := []struct {
		name        string
		data        map[string]any
		target      any
		errContains string
	}{
		{
			name:        "invalid slice data",
			data:        map[string]any{"Data": "not a slice"},
			target:      &Target{},
			errContains: "cannot convert",
		},
		{
			name:        "non-pointer result",
			data:        map[string]any{"Name": "test"},
			target:      Target{},
			errContains: "must be a pointer",
		},
		{
			name:        "nil pointer result",
			data:        map[string]any{"Name": "test"},
			target:      (*Target)(nil),
			errContains: "nil",
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}
