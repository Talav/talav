package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()

	assert.False(t, cfg.Collect.Build)
	assert.False(t, cfg.Collect.Go)
	assert.False(t, cfg.Collect.Process)
}

func TestNewDefaultMetricsRegistryFactory(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	require.NotNil(t, factory)
	assert.IsType(t, &DefaultMetricsRegistryFactory{}, factory)
}

func TestDefaultMetricsRegistryFactory_Create_EmptyConfig(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	registry, err := factory.Create(DefaultMetricsConfig())

	require.NoError(t, err)
	require.NotNil(t, registry)
}

func TestDefaultMetricsRegistryFactory_Create_WithBuildCollector(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	cfg := MetricsConfig{
		Collect: CollectConfig{Build: true},
	}

	registry, err := factory.Create(cfg)

	require.NoError(t, err)
	require.NotNil(t, registry)

	mfs, err := registry.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, mfs, "expected build info metrics to be gathered")
}

func TestDefaultMetricsRegistryFactory_Create_WithGoCollector(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	cfg := MetricsConfig{
		Collect: CollectConfig{Go: true},
	}

	registry, err := factory.Create(cfg)

	require.NoError(t, err)
	require.NotNil(t, registry)

	mfs, err := registry.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, mfs, "expected go runtime metrics to be gathered")
}

func TestDefaultMetricsRegistryFactory_Create_WithProcessCollector(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	cfg := MetricsConfig{
		Collect: CollectConfig{Process: true},
	}

	registry, err := factory.Create(cfg)

	require.NoError(t, err)
	require.NotNil(t, registry)
}

func TestDefaultMetricsRegistryFactory_Create_WithAllCollectors(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	cfg := MetricsConfig{
		Collect: CollectConfig{
			Build:   true,
			Go:      true,
			Process: true,
		},
	}

	registry, err := factory.Create(cfg)

	require.NoError(t, err)
	require.NotNil(t, registry)

	mfs, err := registry.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, mfs, "expected metrics to be gathered")
}

func TestDefaultMetricsRegistryFactory_Create_BuildCollectorRegistrationError(t *testing.T) {
	prePopulated := prometheus.NewRegistry()
	prePopulated.MustRegister(prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_build_info",
		Help: "conflicting metric to trigger registration error",
	}))

	factory := &DefaultMetricsRegistryFactory{
		newRegistry: func() *prometheus.Registry { return prePopulated },
	}

	_, err := factory.Create(MetricsConfig{Collect: CollectConfig{Build: true}})
	require.Error(t, err)
}

func TestDefaultMetricsRegistryFactory_Create_GoCollectorRegistrationError(t *testing.T) {
	prePopulated := prometheus.NewRegistry()
	prePopulated.MustRegister(prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_goroutines",
		Help: "conflicting metric to trigger registration error",
	}))

	factory := &DefaultMetricsRegistryFactory{
		newRegistry: func() *prometheus.Registry { return prePopulated },
	}

	_, err := factory.Create(MetricsConfig{Collect: CollectConfig{Go: true}})
	require.Error(t, err)
}

func TestDefaultMetricsRegistryFactory_Create_ProcessCollectorRegistrationError(t *testing.T) {
	prePopulated := prometheus.NewRegistry()
	prePopulated.MustRegister(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "process_cpu_seconds_total",
		Help: "conflicting metric to trigger registration error",
	}))

	factory := &DefaultMetricsRegistryFactory{
		newRegistry: func() *prometheus.Registry { return prePopulated },
	}

	_, err := factory.Create(MetricsConfig{Collect: CollectConfig{Process: true}})
	require.Error(t, err)
}

func TestDefaultMetricsRegistryFactory_Create_IsolatedRegistry(t *testing.T) {
	factory := NewDefaultMetricsRegistryFactory()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_isolation_total",
		Help: "tests registry isolation",
	})

	registry, err := factory.Create(DefaultMetricsConfig())
	require.NoError(t, err)

	require.NoError(t, registry.Register(counter))

	counter.Add(42)

	mfs, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, mfs, 1)
	assert.Equal(t, "test_isolation_total", mfs[0].GetName())
	assert.InDelta(t, 42.0, mfs[0].GetMetric()[0].GetCounter().GetValue(), 0.001)
}
