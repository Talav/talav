package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Encode Query tests.
func TestEncodeQuery_FormStyle(t *testing.T) {
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, false)
		require.NoError(t, err)

		values := map[string]any{
			"ids":  []any{"1", "2", "3"},
			"name": "John",
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		// URL encoding: comma becomes %2C
		assert.Contains(t, result, "ids=1%2C2%2C3")
		assert.Contains(t, result, "name=John")
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, true)
		require.NoError(t, err)

		values := map[string]any{
			"ids":  []any{"1", "2", "3"},
			"name": "John",
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		assert.Contains(t, result, "ids=1")
		assert.Contains(t, result, "ids=2")
		assert.Contains(t, result, "ids=3")
		assert.Contains(t, result, "name=John")
	})

	t.Run("nested structure", func(t *testing.T) {
		opts, err := NewOptions(LocationQuery, StyleForm, true)
		require.NoError(t, err)

		values := map[string]any{
			"filter": map[string]any{
				"type":  "car",
				"color": "red",
			},
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		assert.Contains(t, result, "filter.type=car")
		assert.Contains(t, result, "filter.color=red")
	})
}

func TestEncodeQuery_DeepObject(t *testing.T) {
	opts, err := NewOptions(LocationQuery, StyleDeepObject)
	require.NoError(t, err)

	values := map[string]any{
		"filter": map[string]any{
			"type":  "car",
			"color": "red",
		},
	}

	encoder := NewDefaultEncoder()
	result, err := encoder.Encode(values, opts)
	require.NoError(t, err)
	// URL encoding: [ becomes %5B, ] becomes %5D
	assert.Contains(t, result, "filter%5Btype%5D=car")
	assert.Contains(t, result, "filter%5Bcolor%5D=red")
}

// Encode Path tests.
func TestEncodePath_Matrix(t *testing.T) {
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleMatrix, false)
		require.NoError(t, err)

		values := map[string]any{
			"ids": []any{"1", "2", "3"},
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		// URL encoding: comma becomes %2C
		assert.Equal(t, ";ids=1%2C2%2C3", result)
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleMatrix, true)
		require.NoError(t, err)

		values := map[string]any{
			"ids": []any{"1", "2", "3"},
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		assert.Contains(t, result, ";ids=1")
		assert.Contains(t, result, ";ids=2")
		assert.Contains(t, result, ";ids=3")
	})
}

func TestEncodePath_Label(t *testing.T) {
	t.Run("explode false", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleLabel, false)
		require.NoError(t, err)

		values := map[string]any{
			"": []any{"1", "2", "3"},
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		assert.Equal(t, ".1,2,3", result)
	})

	t.Run("explode true", func(t *testing.T) {
		opts, err := NewOptions(LocationPath, StyleLabel, true)
		require.NoError(t, err)

		values := map[string]any{
			"": []any{"1", "2", "3"},
		}

		encoder := NewDefaultEncoder()
		result, err := encoder.Encode(values, opts)
		require.NoError(t, err)
		assert.Equal(t, ".1.2.3", result)
	})
}

// Encode Header tests.
func TestEncodeHeader_Simple(t *testing.T) {
	opts, err := NewOptions(LocationHeader, StyleSimple)
	require.NoError(t, err)

	values := map[string]any{
		"": []any{"1", "2", "3"},
	}

	encoder := NewDefaultEncoder()
	result, err := encoder.Encode(values, opts)
	require.NoError(t, err)
	assert.Equal(t, "1,2,3", result)
}
