package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// ConfigSource defines a config source with path, patterns, and parser
type ConfigSource struct {
	Path     string       // Directory path (prepended to patterns)
	Patterns []string     // File patterns with placeholder: {env}
	Parser   koanf.Parser // Parser for this source
}

// ConfigFactory is the interface for [Config] factories.
type ConfigFactory interface {
	Create(sources ...ConfigSource) (*Config, error)
}

// DefaultConfigFactory is the default [ConfigFactory] implementation.
type DefaultConfigFactory struct{}

// NewDefaultConfigFactory returns a [DefaultConfigFactory], implementing [ConfigFactory].
func NewDefaultConfigFactory() ConfigFactory {
	return &DefaultConfigFactory{}
}

// Create returns a new [Config].
//
// Loads configs in order specified by sources (later sources override earlier).
// Environment variables are always loaded last (highest priority).
//
// Example:
//
//	// Default behavior (mimics current YAML + ENV patterns)
//	var cfg, _ = factory.Create()
//
//	// Custom sources with full control
//	var cfg, _ = factory.Create(
//		ConfigSource{
//			Path:     "./config",
//			Patterns: []string{"config.yaml", "config_{env}.yaml"},
//			Parser:   yaml.Parser(),
//		},
//	)
func (f *DefaultConfigFactory) Create(sources ...ConfigSource) (*Config, error) {
	// Default sources mimic current behavior (YAML → ENV → env vars)
	if len(sources) == 0 {
		sources = []ConfigSource{
			{
				Path:     "./configs",
				Patterns: []string{"config.yaml", "config_{env}.yaml", "config_{env}_local.yaml"},
				Parser:   yaml.Parser(),
			},
			{
				Path:     ".",
				Patterns: []string{".env", ".env.{env}", ".env.local", ".env.{env}.local"},
				Parser:   NewDotenvParser(),
			},
		}
	}

	k := koanf.New(".")

	// Load all sources in order (order = priority, later sources override earlier)
	for _, source := range sources {
		if err := f.loadSource(k, source); err != nil {
			return nil, err
		}
	}

	// Load environment variables (highest priority, always last)
	if err := k.Load(env.Provider(".", env.Opt{
		TransformFunc: EnvTransformFunc(),
	}), nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Expand environment variable placeholders
	f.expandEnvPlaceholders(k)

	return &Config{k}, nil
}

// expandPatterns expands file patterns by replacing placeholders
func (f *DefaultConfigFactory) expandPatterns(path string, patterns []string) []string {
	env := f.getEnvironment()
	files := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		file := pattern
		file = strings.ReplaceAll(file, "{env}", env)
		// Prepend path if pattern is not absolute
		if !filepath.IsAbs(file) {
			file = filepath.Join(path, file)
		}
		files = append(files, file)
	}
	return files
}

// loadSource loads a config source by expanding patterns and loading files
func (f *DefaultConfigFactory) loadSource(k *koanf.Koanf, source ConfigSource) error {
	files := f.expandPatterns(source.Path, source.Patterns)
	for _, filePath := range files {
		if err := f.loadFile(k, filePath, source.Parser); err != nil {
			return err
		}
	}
	return nil
}

// loadFile loads a single config file with the given parser
func (f *DefaultConfigFactory) loadFile(k *koanf.Koanf, filePath string, parser koanf.Parser) error {
	if err := k.Load(file.Provider(filePath), parser); err != nil {
		// File not found → skip (safe!)
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load config file %s: %w", filePath, err)
	}
	return nil
}

// expandEnvPlaceholders expands ${VAR} placeholders in string values
func (f *DefaultConfigFactory) expandEnvPlaceholders(k *koanf.Koanf) {
	for _, key := range k.Keys() {
		val := k.String(key)
		if strings.Contains(val, "${") {
			_ = k.Set(key, os.ExpandEnv(val))
		}
	}
}

func (f *DefaultConfigFactory) getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	return env
}
