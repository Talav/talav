package fxmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/talav/talav/pkg/component/metrics"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

// ModuleName is the module name.
const ModuleName = "metrics"

// FxMetricsModule is the [fx] metrics module.
//
// It provides a [prometheus.Registry] populated with any collectors registered
// via [AsMetricsCollector] or [AsMetricsCollectors].
var FxMetricsModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("metrics", metrics.DefaultMetricsConfig(), metrics.MetricsConfig{}),
	fx.Provide(
		metrics.NewDefaultMetricsRegistryFactory,
		newFxMetricsRegistry,
	),
)

type fxMetricsRegistryParams struct {
	fx.In

	Factory    metrics.MetricsRegistryFactory
	Config     metrics.MetricsConfig
	Collectors []prometheus.Collector `group:"metrics-collectors"`
}

func newFxMetricsRegistry(p fxMetricsRegistryParams) (*prometheus.Registry, error) {
	registry, err := p.Factory.Create(p.Config)
	if err != nil {
		return nil, err
	}

	for _, collector := range p.Collectors {
		if err := registry.Register(collector); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
