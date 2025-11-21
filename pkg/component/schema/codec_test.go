package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodec_WithConverter(t *testing.T) {
	type ID int

	codec := NewCodec(WithConverter(ID(0), func(s string) (reflect.Value, error) {
		// Custom converter logic
		return reflect.ValueOf(ID(42)), nil
	}))

	// Converter is registered, but this test just verifies the option works
	assert.NotNil(t, codec)
}

func TestCodec_Decode_Encode_Integration(t *testing.T) {
	type User struct {
		Name string `schema:"name"`
		Age  int    `schema:"age"`
	}

	codec := NewCodec()
	opts, _ := NewOptions(LocationQuery, StyleForm)

	// Test Decode - decode string to struct
	var user User
	err := codec.Decode("name=John&age=30", opts, &user)
	require.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 30, user.Age)

	// Test Encode - encode struct to string
	encoded, err := codec.Encode(user, opts)
	require.NoError(t, err)
	assert.Contains(t, encoded, "name=John")
	assert.Contains(t, encoded, "age=30")
}

func TestCodec_WithDecoder_WithEncoder(t *testing.T) {
	type TestStruct struct {
		Key string `schema:"key"`
	}

	customDecoder := &mockDecoder{}
	customEncoder := &mockEncoder{}

	codec := NewCodec(
		WithDecoder(customDecoder),
		WithEncoder(customEncoder),
	)

	opts, _ := NewOptions(LocationQuery, StyleForm)
	var result TestStruct
	err := codec.Decode("test", opts, &result)
	require.NoError(t, err)
	assert.True(t, customDecoder.called)

	testStruct := TestStruct{Key: "value"}
	_, err = codec.Encode(testStruct, opts)
	require.NoError(t, err)
	assert.True(t, customEncoder.called)
}

type mockDecoder struct {
	called bool
}

func (m *mockDecoder) Decode(value string, opts Options) (map[string]any, error) {
	m.called = true

	return map[string]any{"test": "value"}, nil
}

type mockEncoder struct {
	called bool
}

func (m *mockEncoder) Encode(values map[string]any, opts Options) (string, error) {
	m.called = true

	return "encoded", nil
}
