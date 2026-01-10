package cdn

import (
	"fmt"
	"log/slog"
)

// CDNSpec represents configuration for a single CDN.
type CDNSpec struct {
	Type   string          `config:"type"` // server
	Server ServerCDNConfig `config:"server"`
}

// ServerCDNConfig represents configuration for server CDN.
type ServerCDNConfig struct {
	Path string `config:"path"` // base URL for server CDN
}

// Factory is the interface for CDN factories.
type Factory interface {
	Create(cfg CDNSpec, logger *slog.Logger) (CDN, error)
}

// DefaultFactory is the default [Factory] implementation.
type DefaultFactory struct{}

// NewDefaultFactory returns a [DefaultFactory], implementing [Factory].
func NewDefaultFactory() Factory {
	return &DefaultFactory{}
}

// Create returns a new [CDN] based on the configuration.
func (f *DefaultFactory) Create(cfg CDNSpec, logger *slog.Logger) (CDN, error) {
	switch cfg.Type {
	case "server":
		if cfg.Server.Path == "" {
			return nil, fmt.Errorf("cdn.server.path is required for server CDN")
		}

		return NewServer(cfg.Server.Path), nil
	default:
		if cfg.Type == "" {
			return nil, fmt.Errorf("cdn.type is required")
		}
		logger.Error("Unknown CDN type", "type", cfg.Type)

		return nil, fmt.Errorf("unknown CDN type: %s", cfg.Type)
	}
}
