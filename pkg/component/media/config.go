package media

import (
	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/infra/cdn"
)

// ResizerConfig represents configuration for a resizer.
type ResizerConfig struct {
	Type string `config:"type"` // simple, crop, square
}

// PresetConfig represents configuration for a preset.
type PresetConfig struct {
	Providers []string                         `config:"providers"` // list of allowed provider names
	Formats   map[string]provider.FormatConfig `config:"formats"`   // thumbnail formats to generate
}

// MediaConfig represents the configuration for the media module.
type MediaConfig struct {
	Resizers  map[string]ResizerConfig           `config:"resizers"`
	CDN       map[string]cdn.CDNSpec             `config:"cdn"` // map of CDN name to CDN spec
	Providers map[string]provider.ProviderConfig `config:"providers"`
	Presets   map[string]PresetConfig            `config:"presets"`
}
