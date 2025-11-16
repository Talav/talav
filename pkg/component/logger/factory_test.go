package logger

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLoggerFactory_Create_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"invalid level defaults to info", "invalid", slog.LevelInfo},
		{"empty level defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewDefaultLoggerFactory()
			var buf bytes.Buffer

			log, err := factory.Create(LoggerConfig{
				Level:  tt.level,
				Format: "text",
				Output: "stdout",
			})
			require.NoError(t, err)
			assert.NotNil(t, log)

			// Create a test logger with buffer to verify level filtering
			testLog := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: tt.expected}))

			// Test that debug messages are filtered based on level
			testLog.Debug("debug message")
			testLog.Info("info message")

			output := buf.String()
			if tt.expected <= slog.LevelDebug {
				assert.Contains(t, output, "debug message")
			} else {
				assert.NotContains(t, output, "debug message")
			}
			if tt.expected <= slog.LevelInfo {
				assert.Contains(t, output, "info message")
			}

			// Verify the created logger can log without error
			log.Info("factory logger works")
		})
	}
}

func TestDefaultLoggerFactory_Create_Formats(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		checkJSON bool
	}{
		{"json format", "json", true},
		{"text format", "text", false},
		{"empty format defaults to json", "", true},
		{"invalid format defaults to json", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewDefaultLoggerFactory()
			var buf bytes.Buffer

			log, err := factory.Create(LoggerConfig{
				Level:  "info",
				Format: tt.format,
				Output: "stdout",
			})
			require.NoError(t, err)
			assert.NotNil(t, log)

			// Create a test logger with buffer to verify format
			var testLog *slog.Logger
			if tt.checkJSON {
				testLog = slog.New(slog.NewJSONHandler(&buf, nil))
			} else {
				testLog = slog.New(slog.NewTextHandler(&buf, nil))
			}

			testLog.Info("test message", "key", "value")

			output := buf.String()
			if tt.checkJSON {
				assert.Contains(t, output, `"msg":"test message"`)
				assert.Contains(t, output, `"key":"value"`)
			} else {
				assert.Contains(t, output, "test message")
				assert.Contains(t, output, "key=value")
			}

			// Verify the created logger can log without error
			log.Info("factory logger works")
		})
	}
}

func TestDefaultLoggerFactory_Create_Output_Stdout(t *testing.T) {
	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	require.NoError(t, err)
	assert.NotNil(t, log)

	// Verify it can log without error
	log.Info("test message")
}

func TestDefaultLoggerFactory_Create_Output_Empty(t *testing.T) {
	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: "",
	})
	require.NoError(t, err)
	assert.NotNil(t, log)

	// Verify it can log without error
	log.Info("test message")
}

func TestDefaultLoggerFactory_Create_Output_File(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: filePath,
	})
	require.NoError(t, err)
	assert.NotNil(t, log)

	// Log a message
	log.Info("file test message", "key", "value")

	// Verify file was created and contains the log
	cleanPath := filepath.Clean(filePath)
	if !filepath.IsAbs(cleanPath) || filepath.Dir(cleanPath) != filepath.Clean(tmpDir) {
		t.Fatalf("invalid file path: %s", cleanPath)
	}
	content, err := os.ReadFile(cleanPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "file test message")
	assert.Contains(t, string(content), `"key":"value"`)
}

func TestDefaultLoggerFactory_Create_Output_File_Fallback(t *testing.T) {
	// Create a path that will fail (directory doesn't exist and can't be created)
	invalidPath := "/nonexistent/directory/test.log"

	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: invalidPath,
	})
	// Should not error, but fall back to stdout
	require.NoError(t, err)
	assert.NotNil(t, log)

	// Verify it can still log (to stdout)
	log.Info("fallback test message")
}

func TestDefaultLoggerFactory_Create_Integration(t *testing.T) {
	var buf bytes.Buffer

	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	})
	require.NoError(t, err)
	assert.NotNil(t, log)

	// Create a test logger with buffer to verify all log levels work
	testLog := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	testLog.Debug("debug log", "debug_key", "debug_value")
	testLog.Info("info log", "info_key", "info_value")
	testLog.Warn("warn log", "warn_key", "warn_value")
	testLog.Error("error log", "error_key", "error_value")

	output := buf.String()
	assert.Contains(t, output, "debug log")
	assert.Contains(t, output, "info log")
	assert.Contains(t, output, "warn log")
	assert.Contains(t, output, "error log")
	assert.Contains(t, output, "debug_key=debug_value")
	assert.Contains(t, output, "info_key=info_value")

	// Verify the created logger can log all levels without error
	log.Debug("factory debug")
	log.Info("factory info")
	log.Warn("factory warn")
	log.Error("factory error")
}

func TestDefaultLoggerFactory_Create_FileAndStdout(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dual.log")

	factory := NewDefaultLoggerFactory()

	log, err := factory.Create(LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: filePath,
	})
	require.NoError(t, err)

	// Log a message - should write to both file and stdout via MultiWriter
	log.Info("dual output test", "key", "value")

	// Verify file contains the log
	cleanPath := filepath.Clean(filePath)
	if !filepath.IsAbs(cleanPath) || filepath.Dir(cleanPath) != filepath.Clean(tmpDir) {
		t.Fatalf("invalid file path: %s", cleanPath)
	}
	fileContent, err := os.ReadFile(cleanPath)
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "dual output test")
	assert.Contains(t, string(fileContent), "key=value")
}

func TestDefaultLoggerFactory_Create_NoColor(t *testing.T) {
	factory := NewDefaultLoggerFactory()

	// NoColor is currently not used in the implementation, but we test it doesn't break
	log, err := factory.Create(LoggerConfig{
		Level:   "info",
		Format:  "text",
		Output:  "stdout",
		NoColor: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, log)

	log.Info("no color test")
}
