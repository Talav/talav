package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDotenvParser_KeyNormalizationAndInterface(t *testing.T) {
	parser := NewDotenvParser()

	envContent := `APP_NAME=test
DATABASE_HOST=localhost
APP_DATABASE_PORT=5432`

	result, err := parser.Unmarshal([]byte(envContent))
	require.NoError(t, err)

	// After normalization and unflattening, keys are nested
	app, ok := result["app"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test", app["name"])

	database, ok := result["database"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "localhost", database["host"])

	appDatabase, ok := app["database"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "5432", appDatabase["port"])
}
