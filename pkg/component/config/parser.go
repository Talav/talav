package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/v2"
)

// NewDotenvParser returns a koanf.Parser configured for .env files with key normalization.
//
// The parser normalizes environment variable keys by:
//   - Converting underscores to dots (e.g., APP_NAME -> app.name)
//   - Lowercasing keys (e.g., DATABASE_HOST -> database.host)
//
// This matches the behavior used throughout the application and ensures consistent
// key naming between .env files and environment variables.
//
// Example:
//
//	parser := NewDotenvParser()
//	cfg, err := factory.Create(
//		ConfigSource{
//			Path:     ".",
//			Patterns: []string{".env", ".env.{env}"},
//			Parser:   parser,
//		},
//	)
func NewDotenvParser() koanf.Parser {
	return dotenv.ParserEnv("", ".", normalizeKey)
}

// EnvTransformFunc returns a TransformFunc for env.Provider that normalizes keys
// using the same logic as NewDotenvParser, ensuring consistency between .env files
// and environment variables.
func EnvTransformFunc() func(k, v string) (string, any) {
	return func(k, v string) (string, any) {
		return normalizeKey(k), v
	}
}

// normalizeKey normalizes environment variable keys by converting underscores to dots
// and lowercasing (e.g., APP_NAME -> app.name, DATABASE_HOST -> database.host).
func normalizeKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", "."))
}
