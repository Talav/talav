package httpserver

import "github.com/talav/talav/pkg/component/zorya"

// Config represents the HTTP server configuration.
type Config struct {
	// Server configuration
	Server ServerConfig `config:"server"`

	// Logging configuration for HTTP requests
	Logging LoggingConfig `config:"logging"`

	// API contains Zorya behavioral configuration (paths, formats).
	API APIConfig `config:"api"`

	// OpenAPI contains metadata for the OpenAPI specification (title, version, etc.).
	OpenAPI OpenAPIConfig `config:"openapi"`
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	// Host is the server host (default: "localhost").
	Host string `config:"host"`

	// Port is the server port (default: 8080).
	Port int `config:"port"`

	// ReadTimeout is the maximum duration for reading the entire request (default: "15s").
	ReadTimeout string `config:"readTimeout"`

	// WriteTimeout is the maximum duration before timing out writes of the response (default: "15s").
	WriteTimeout string `config:"writeTimeout"`

	// IdleTimeout is the maximum amount of time to wait for the next request (default: "60s").
	IdleTimeout string `config:"idleTimeout"`

	// ReadHeaderTimeout is the amount of time allowed to read request headers (default: "").
	ReadHeaderTimeout string `config:"readHeaderTimeout"`

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys and values (default: 0).
	MaxHeaderBytes int `config:"maxHeaderBytes"`

	// ShutdownTimeout is the maximum duration for graceful shutdown (default: "10s").
	ShutdownTimeout string `config:"shutdownTimeout"`
}

// LoggingConfig contains HTTP request logging settings.
type LoggingConfig struct {
	// Enabled enables HTTP request logging.
	Enabled bool `config:"enabled"`

	// Level defines httplog verbosity filter (separate from logger level).
	// This controls which HTTP requests get logged by httplog middleware.
	Level string `config:"level"` // debug, info, warn, error

	// Schema defines format: standard, ecs, otel, gcp (default: "standard").
	Schema string `config:"schema"`

	// RecoverPanics recovers from panics and returns HTTP 500.
	RecoverPanics bool `config:"recoverPanics"`

	// LogRequestBody enables conditional request body logging.
	LogRequestBody bool `config:"logRequestBody"`

	// LogResponseBody enables conditional response body logging.
	LogResponseBody bool `config:"logResponseBody"`

	// SkipPaths contains paths to exclude from logging (e.g., health checks).
	SkipPaths []string `config:"skipPaths"`
}

// APIConfig contains Zorya behavioral configuration.
type APIConfig struct {
	// SpecPath is the path to the OpenAPI spec without extension.
	// Default: "/openapi" (serves /openapi.json and /openapi.yaml).
	SpecPath string `config:"specPath"`

	// DocsPath is the path to the API documentation UI.
	// Default: "/docs".
	DocsPath string `config:"docsPath"`

	// SchemasPath is the path to the API schemas.
	// Default: "/schemas".
	SchemasPath string `config:"schemasPath"`

	// DefaultFormat specifies the default content type.
	// Default: "application/json".
	DefaultFormat string `config:"defaultFormat"`

	// NoFormatFallback disables fallback to application/json.
	// Default: false.
	NoFormatFallback bool `config:"noFormatFallback"`
}

// OpenAPIConfig contains metadata for the OpenAPI specification.
type OpenAPIConfig struct {
	// Title of the API (required).
	Title string `config:"title"`

	// Description of the API.
	Description string `config:"description"`

	// Version of the API (required).
	Version string `config:"version"`

	// TermsOfService URL.
	TermsOfService string `config:"termsOfService"`

	// Contact information.
	Contact ContactConfig `config:"contact"`

	// License information.
	License LicenseConfig `config:"license"`

	// Tags for grouping operations.
	Tags []TagConfig `config:"tags"`

	// ExternalDocs for additional documentation.
	ExternalDocs *ExternalDocsConfig `config:"externalDocs"`

	// Security schemes (simplified: just names).
	Security []string `config:"security"`
}

// ContactConfig is simplified contact information.
type ContactConfig struct {
	Name  string `config:"name"`
	Email string `config:"email"`
	URL   string `config:"url"`
}

// LicenseConfig is simplified license information.
type LicenseConfig struct {
	Name       string `config:"name"`
	URL        string `config:"url"`
	Identifier string `config:"identifier"` // SPDX identifier
}

// TagConfig is simplified tag metadata.
type TagConfig struct {
	Name        string `config:"name"`
	Description string `config:"description"`
}

// ExternalDocsConfig is simplified external documentation reference.
type ExternalDocsConfig struct {
	Description string `config:"description"`
	URL         string `config:"url"`
}

// DefaultConfig returns a Config with all default values set.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     "15s",
			WriteTimeout:    "15s",
			IdleTimeout:     "60s",
			ShutdownTimeout: "10s",
		},
		Logging: LoggingConfig{
			Schema: "standard",
		},
		API: APIConfig{
			SpecPath:      "/openapi",
			DocsPath:      "/docs",
			SchemasPath:   "/schemas",
			DefaultFormat: "application/json",
		},
	}
}

// ToZoryaOpenAPI converts the OpenAPI config to a Zorya OpenAPI spec.
func (c *Config) ToZoryaOpenAPI() *zorya.OpenAPI {
	// Build OpenAPI spec from config
	spec := &zorya.OpenAPI{
		OpenAPI:  "3.1.0", // Default to latest
		Info:     c.buildInfo(),
		Tags:     c.buildTags(),
		Security: c.buildSecurity(),
	}

	// Add external docs if provided
	if c.OpenAPI.ExternalDocs != nil {
		spec.ExternalDocs = c.buildExternalDocs()
	}

	return spec
}

// ToZoryaConfig converts the API config to a Zorya Config.
func (c *Config) ToZoryaConfig() *zorya.Config {
	cfg := zorya.DefaultConfig()

	// Apply API config (defaults are already applied via AsConfigWithDefaults)
	cfg.OpenAPIPath = c.API.SpecPath
	cfg.DocsPath = c.API.DocsPath
	cfg.SchemasPath = c.API.SchemasPath
	cfg.DefaultFormat = c.API.DefaultFormat
	cfg.NoFormatFallback = c.API.NoFormatFallback

	return cfg
}

// buildInfo constructs the Info section of OpenAPI spec.
func (c *Config) buildInfo() *zorya.Info {
	info := &zorya.Info{
		Title:       c.OpenAPI.Title,
		Description: c.OpenAPI.Description,
		Version:     c.OpenAPI.Version,
	}

	// Add terms of service if provided
	if c.OpenAPI.TermsOfService != "" {
		info.TermsOfService = c.OpenAPI.TermsOfService
	}

	// Add contact if any field is provided
	if c.OpenAPI.Contact.Name != "" || c.OpenAPI.Contact.Email != "" || c.OpenAPI.Contact.URL != "" {
		info.Contact = &zorya.Contact{
			Name:  c.OpenAPI.Contact.Name,
			Email: c.OpenAPI.Contact.Email,
			URL:   c.OpenAPI.Contact.URL,
		}
	}

	// Add license if name is provided
	if c.OpenAPI.License.Name != "" {
		info.License = &zorya.License{
			Name:       c.OpenAPI.License.Name,
			URL:        c.OpenAPI.License.URL,
			Identifier: c.OpenAPI.License.Identifier,
		}
	}

	return info
}

// buildTags converts tag configs to Zorya tags.
func (c *Config) buildTags() []*zorya.Tag {
	tags := make([]*zorya.Tag, len(c.OpenAPI.Tags))
	for i, tag := range c.OpenAPI.Tags {
		tags[i] = &zorya.Tag{
			Name:        tag.Name,
			Description: tag.Description,
		}
	}

	return tags
}

// buildExternalDocs converts external docs to Zorya format.
func (c *Config) buildExternalDocs() *zorya.ExternalDocs {
	return &zorya.ExternalDocs{
		Description: c.OpenAPI.ExternalDocs.Description,
		URL:         c.OpenAPI.ExternalDocs.URL,
	}
}

// buildSecurity converts security config to OpenAPI format.
// This creates a simple security requirement with empty scopes for each scheme.
func (c *Config) buildSecurity() []map[string][]string {
	security := make([]map[string][]string, len(c.OpenAPI.Security))
	for i, scheme := range c.OpenAPI.Security {
		security[i] = map[string][]string{
			scheme: {}, // Empty scopes
		}
	}

	return security
}
