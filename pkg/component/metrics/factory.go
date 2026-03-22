package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// MetricsRegistryFactory is the interface for [prometheus.Registry] factories.
type MetricsRegistryFactory interface {
	Create(cfg MetricsConfig) (*prometheus.Registry, error)
}

// DefaultMetricsRegistryFactory is the default [MetricsRegistryFactory] implementation.
type DefaultMetricsRegistryFactory struct {
	// newRegistry is the function used to create a fresh registry.
	// Defaults to prometheus.NewRegistry; overridable in tests.
	newRegistry func() *prometheus.Registry
}

// NewDefaultMetricsRegistryFactory returns a [DefaultMetricsRegistryFactory], implementing [MetricsRegistryFactory].
func NewDefaultMetricsRegistryFactory() MetricsRegistryFactory {
	return &DefaultMetricsRegistryFactory{
		newRegistry: prometheus.NewRegistry,
	}
}

// Create returns a new [prometheus.Registry] configured from the given [MetricsConfig].
//
// Optional collectors (build info, Go runtime, process) are registered when enabled via config.
// Using an isolated registry (rather than the global default) prevents data race conditions in tests.
//
// Example:
//
//	factory := NewDefaultMetricsRegistryFactory()
//	registry, err := factory.Create(MetricsConfig{
//		Collect: CollectConfig{Go: true},
//	})
func (f *DefaultMetricsRegistryFactory) Create(cfg MetricsConfig) (*prometheus.Registry, error) {
	registry := f.newRegistry()

	if cfg.Collect.Build {
		if err := registry.Register(collectors.NewBuildInfoCollector()); err != nil {
			return nil, err
		}
	}

	if cfg.Collect.Go {
		if err := registry.Register(collectors.NewGoCollector()); err != nil {
			return nil, err
		}
	}

	if cfg.Collect.Process {
		if err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
