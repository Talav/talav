package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldCache_GetCachedFields(t *testing.T) {
	type User struct {
		Name string `schema:"name"`
		Age  int    `schema:"age"`
		ID   int    // No tag
	}

	cache := NewFieldCache()
	typ := reflect.TypeOf(User{})
	fields := cache.GetCachedFields(typ, "schema")

	require.Len(t, fields, 3)

	// Check field with tag
	nameField := fields[0]
	assert.Equal(t, "Name", nameField.name)
	assert.Equal(t, "name", nameField.tagName)
	assert.Equal(t, "name", nameField.mapKey)

	// Check field without tag
	idField := fields[2]
	assert.Equal(t, "ID", idField.name)
	assert.Equal(t, "", idField.tagName)
	assert.Equal(t, "ID", idField.mapKey)
}

func TestFieldCache_CacheReuse(t *testing.T) {
	type User struct {
		Name string
	}

	cache := NewFieldCache()
	typ := reflect.TypeOf(User{})

	fields1 := cache.GetCachedFields(typ, "schema")
	fields2 := cache.GetCachedFields(typ, "schema")

	// Should return same slice (cached)
	assert.Equal(t, fields1, fields2)
}

func TestGetFieldName(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{"simple", "name", "name"},
		{"with options", "name,omitempty", "name"},
		{"empty", "", ""},
		{"only options", ",omitempty", ""},
	}

	cache := NewFieldCache()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.getFieldName(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}
