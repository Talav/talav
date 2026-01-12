package preset

import (
	"fmt"

	"github.com/talav/talav/pkg/component/media"
	"github.com/talav/talav/pkg/component/media/app/provider"
)

// Registry manages presets and their configurations.
type Registry interface {
	// GetPresetConfig returns the full preset configuration
	GetPresetConfig(preset string) (media.PresetConfig, error)
	// GetPresetFormats returns the formats configured for a preset
	GetPresetFormats(preset string) (map[string]provider.FormatConfig, error)
	// IsProviderAllowed checks if a provider is allowed for a preset
	IsProviderAllowed(preset, providerName string) (bool, error)
}

// DefaultRegistry is the default implementation of Registry.
type DefaultRegistry struct {
	presets          map[string]media.PresetConfig
	allowedProviders map[string]map[string]bool // preset -> provider name -> true
}

// NewDefaultRegistry creates a new DefaultRegistry.
func NewDefaultRegistry(presets map[string]media.PresetConfig) Registry {
	allowedProviders := make(map[string]map[string]bool)
	for presetName, config := range presets {
		providerSet := make(map[string]bool)
		for _, providerName := range config.Providers {
			providerSet[providerName] = true
		}
		allowedProviders[presetName] = providerSet
	}

	return &DefaultRegistry{
		presets:          presets,
		allowedProviders: allowedProviders,
	}
}

// GetPresetConfig returns the full preset configuration.
func (r *DefaultRegistry) GetPresetConfig(preset string) (media.PresetConfig, error) {
	config, exists := r.presets[preset]
	if !exists {
		return media.PresetConfig{}, fmt.Errorf("preset %q not found", preset)
	}

	return config, nil
}

// GetPresetFormats returns the formats configured for a preset.
func (r *DefaultRegistry) GetPresetFormats(preset string) (map[string]provider.FormatConfig, error) {
	config, err := r.GetPresetConfig(preset)
	if err != nil {
		return nil, err
	}

	return config.Formats, nil
}

// IsProviderAllowed checks if a provider is allowed for a preset.
func (r *DefaultRegistry) IsProviderAllowed(preset, providerName string) (bool, error) {
	providerSet, exists := r.allowedProviders[preset]
	if !exists {
		return false, fmt.Errorf("preset %q not found", preset)
	}

	return providerSet[providerName], nil
}
