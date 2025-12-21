package mapstructure

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

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

//nolint:forcetypeassert,thelper // Test code - type assertions are expected to succeed
func TestUnmarshaler_Unmarshal_DirectAssignment(t *testing.T) {
	type Inner struct {
		Value int
		Name  string
	}

	type WithReader struct {
		Reader io.Reader
	}
	type WithReadCloser struct {
		Body io.ReadCloser
	}
	type WithWriter struct {
		Writer io.Writer
	}
	type WithBytes struct {
		Data []byte
	}
	type WithTime struct {
		CreatedAt time.Time
	}
	type WithInner struct {
		Inner Inner
	}
	type WithInnerPtr struct {
		Inner *Inner
	}
	type WithInnerSlice struct {
		Items []Inner
	}
	type WithMap struct {
		Metadata map[string]string
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		validate func(t *testing.T, target any)
	}{
		{
			name:   "[]byte direct assignment",
			data:   map[string]any{"Data": []byte{0x01, 0x02, 0x03}},
			target: &WithBytes{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithBytes)
				assert.Equal(t, []byte{0x01, 0x02, 0x03}, r.Data)
			},
		},
		{
			name:   "[]byte empty",
			data:   map[string]any{"Data": []byte{}},
			target: &WithBytes{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithBytes)
				assert.Equal(t, []byte{}, r.Data)
			},
		},
		{
			name:   "io.Reader direct assignment",
			data:   map[string]any{"Reader": bytes.NewReader([]byte("hello"))},
			target: &WithReader{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithReader)
				require.NotNil(t, r.Reader)
				content, err := io.ReadAll(r.Reader)
				require.NoError(t, err)
				assert.Equal(t, []byte("hello"), content)
			},
		},
		{
			name:   "io.ReadCloser direct assignment",
			data:   map[string]any{"Body": io.NopCloser(bytes.NewReader([]byte("body")))},
			target: &WithReadCloser{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithReadCloser)
				require.NotNil(t, r.Body)
				content, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, []byte("body"), content)
			},
		},
		{
			name:   "io.Writer interface satisfaction",
			data:   map[string]any{"Writer": &bytes.Buffer{}},
			target: &WithWriter{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithWriter)
				require.NotNil(t, r.Writer)
				_, err := r.Writer.Write([]byte("test"))
				assert.NoError(t, err)
			},
		},
		{
			name:   "custom struct direct assignment",
			data:   map[string]any{"Inner": Inner{Value: 42, Name: "test"}},
			target: &WithInner{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInner)
				assert.Equal(t, Inner{Value: 42, Name: "test"}, r.Inner)
			},
		},
		{
			name:   "custom struct pointer direct assignment",
			data:   map[string]any{"Inner": &Inner{Value: 99, Name: "ptr"}},
			target: &WithInnerPtr{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInnerPtr)
				require.NotNil(t, r.Inner)
				assert.Equal(t, 99, r.Inner.Value)
				assert.Equal(t, "ptr", r.Inner.Name)
			},
		},
		{
			name:   "slice of custom structs",
			data:   map[string]any{"Items": []Inner{{Value: 1, Name: "a"}, {Value: 2, Name: "b"}}},
			target: &WithInnerSlice{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithInnerSlice)
				require.Len(t, r.Items, 2)
				assert.Equal(t, "a", r.Items[0].Name)
				assert.Equal(t, "b", r.Items[1].Name)
			},
		},
		{
			name:   "map[string]string direct assignment",
			data:   map[string]any{"Metadata": map[string]string{"key": "value"}},
			target: &WithMap{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithMap)
				assert.Equal(t, "value", r.Metadata["key"])
			},
		},
		{
			name:   "time.Time direct assignment",
			data:   map[string]any{"CreatedAt": now},
			target: &WithTime{},
			validate: func(t *testing.T, target any) {
				r := target.(*WithTime)
				assert.Equal(t, now, r.CreatedAt)
			},
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			tt.validate(t, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_DirectAssignment_FallbackToConverter(t *testing.T) {
	type Target struct {
		Count int
	}

	tests := []struct {
		name     string
		data     map[string]any
		expected int
	}{
		{
			name:     "string to int uses converter",
			data:     map[string]any{"Count": "42"},
			expected: 42,
		},
		{
			name:     "int64 to int uses converter",
			data:     map[string]any{"Count": int64(42)},
			expected: 42,
		},
		{
			name:     "float64 to int uses converter",
			data:     map[string]any{"Count": float64(42.0)},
			expected: 42,
		},
	}

	u := testUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Target
			err := u.Unmarshal(tt.data, &result)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Count)
		})
	}
}

func TestUnmarshaler_Unmarshal_DefaultValues(t *testing.T) {
	type WithDefaults struct {
		Name   string  `schema:"name" default:"anonymous"`
		Count  int     `schema:"count" default:"10"`
		Score  float64 `schema:"score" default:"99.5"`
		Active bool    `schema:"active" default:"true"`
	}

	type WithPointerDefault struct {
		Value *int `schema:"value" default:"42"`
	}

	intPtr := func(v int) *int { return &v }

	type Mixed struct {
		Required string `schema:"required"`
		Optional string `schema:"optional" default:"default_value"`
	}

	tests := []struct {
		name     string
		data     map[string]any
		target   any
		expected any
	}{
		{
			name:   "all defaults applied",
			data:   map[string]any{},
			target: &WithDefaults{},
			expected: &WithDefaults{
				Name:   "anonymous",
				Count:  10,
				Score:  99.5,
				Active: true,
			},
		},
		{
			name:   "explicit values override defaults",
			data:   map[string]any{"name": "custom", "count": 99},
			target: &WithDefaults{},
			expected: &WithDefaults{
				Name:   "custom",
				Count:  99,
				Score:  99.5,
				Active: true,
			},
		},
		{
			name:     "pointer default",
			data:     map[string]any{},
			target:   &WithPointerDefault{},
			expected: &WithPointerDefault{Value: intPtr(42)},
		},
		{
			name:   "mixed required and optional",
			data:   map[string]any{"required": "provided"},
			target: &Mixed{},
			expected: &Mixed{
				Required: "provided",
				Optional: "default_value",
			},
		},
	}

	u := NewDefaultUnmarshaler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := u.Unmarshal(tt.data, tt.target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.target)
		})
	}
}

func TestUnmarshaler_Unmarshal_DefaultValues_CustomConverter(t *testing.T) {
	type Status int

	const (
		StatusPending Status = iota
		StatusActive
		StatusClosed
	)

	type WithCustomDefault struct {
		Status Status `schema:"status" default:"active"`
	}

	// Custom converter for Status type
	statusConverter := func(v any) (reflect.Value, error) {
		s, ok := v.(string)
		if !ok {
			return reflect.Value{}, nil
		}

		switch s {
		case "pending":
			return reflect.ValueOf(StatusPending), nil
		case "active":
			return reflect.ValueOf(StatusActive), nil
		case "closed":
			return reflect.ValueOf(StatusClosed), nil
		default:
			return reflect.Value{}, nil
		}
	}

	converters := map[reflect.Type]Converter{
		reflect.TypeOf(Status(0)): statusConverter,
	}

	// Create unmarshaler with custom converters
	cache := NewStructMetadataCache(DefaultCacheBuilder)
	convertersRegistry := NewDefaultConverterRegistry(converters)
	u := NewUnmarshaler(cache, convertersRegistry)

	var result WithCustomDefault
	err := u.Unmarshal(map[string]any{}, &result)

	require.NoError(t, err)
	assert.Equal(t, StatusActive, result.Status)
}
