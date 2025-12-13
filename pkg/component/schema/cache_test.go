package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaCache_GetCachedFields(t *testing.T) {
	type User struct {
		Name string `schema:"name"`
		Age  int    `schema:"age"`
		ID   int    // No tag - included by default (JSON-like behavior)
	}

	parser := newStructMetadataParser()
	cache := NewStructMetadataCache(parser)
	typ := reflect.TypeOf(User{})
	metadata, err := cache.getStructMetadata(typ)

	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.Len(t, metadata.Fields, 3) // All exported fields included

	// Check field with tag
	nameField := metadata.Fields[0]
	assert.Equal(t, "Name", nameField.StructFieldName)
	assert.Equal(t, "name", nameField.ParamName)
	assert.Equal(t, "name", nameField.MapKey)

	// Check untagged field (should have default metadata)
	idField := metadata.Fields[2]
	assert.Equal(t, "ID", idField.StructFieldName)
	assert.Equal(t, "ID", idField.ParamName)
	assert.Equal(t, "ID", idField.MapKey)
	assert.Equal(t, LocationQuery, idField.Location)
	assert.Equal(t, StyleForm, idField.Style)
}

func TestSchemaCache_SkipFieldWithDash(t *testing.T) {
	type User struct {
		Name string `schema:"name"`
		Age  int    `schema:"age"`
		ID   int    `schema:"-"` // Explicitly skipped
	}

	parser := newStructMetadataParser()
	cache := NewStructMetadataCache(parser)
	typ := reflect.TypeOf(User{})
	metadata, err := cache.getStructMetadata(typ)

	require.NoError(t, err)
	require.NotNil(t, metadata)
	require.Len(t, metadata.Fields, 2) // ID field skipped
}

func TestSchemaCache_CacheReuse(t *testing.T) {
	type User struct {
		Name string `schema:"name"`
	}

	parser := newStructMetadataParser()
	cache := NewStructMetadataCache(parser)
	typ := reflect.TypeOf(User{})

	metadata1, err1 := cache.getStructMetadata(typ)
	require.NoError(t, err1)

	metadata2, err2 := cache.getStructMetadata(typ)
	require.NoError(t, err2)

	// Should return same pointer (cached)
	assert.Equal(t, metadata1, metadata2)
}
