package schema

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverterRegistry_Find_Found(t *testing.T) {
	registry := NewConverterRegistry()

	typ := reflect.TypeOf(int(0))
	conv := func(s string) (reflect.Value, error) {
		return reflect.ValueOf(42), nil
	}

	registry.Register(typ, conv)

	found, ok := registry.Find(typ)
	require.True(t, ok)
	assert.NotNil(t, found)
}

func TestConverterRegistry_Find_NotFoundInEmptyRegistry(t *testing.T) {
	registry := NewConverterRegistry()

	typ := reflect.TypeOf(int(0))
	_, ok := registry.Find(typ)
	assert.False(t, ok)
}
