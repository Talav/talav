package provider

import (
	"fmt"
)

// Registry manages providers.
type Registry interface {
	// GetProvider returns a provider by name
	GetProvider(name string) (Provider, error)
}

// DefaultRegistry is the default implementation of Registry.
type DefaultRegistry struct {
	providers map[string]Provider
}

// NewDefaultRegistry creates a new DefaultRegistry.
func NewDefaultRegistry(providers map[string]Provider) Registry {
	return &DefaultRegistry{
		providers: providers,
	}
}

// GetProvider returns a provider by name.
func (r *DefaultRegistry) GetProvider(name string) (Provider, error) {
	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}

	return provider, nil
}
