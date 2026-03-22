package fxmetrics

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule_FxMetricsModule_ProvidesRegistry(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_fxmetricsmodule_providesregistry")
	t.Setenv("APP_ENV", "dev")

	var registry *prometheus.Registry

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxMetricsModule,
		fx.Populate(&registry),
	).RequireStart().RequireStop()

	require.NotNil(t, registry)
}

func TestModule_FxMetricsModule_WithCollector(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_fxmetricsmodule_withcollectors")
	t.Setenv("APP_ENV", "dev")

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_requests_total",
		Help: "Total number of test requests.",
	})

	var registry *prometheus.Registry

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxMetricsModule,
		AsMetricsCollector(counter),
		fx.Invoke(func() {
			counter.Add(5)
		}),
		fx.Populate(&registry),
	).RequireStart().RequireStop()

	require.NotNil(t, registry)

	expected := `
		# HELP test_requests_total Total number of test requests.
		# TYPE test_requests_total counter
		test_requests_total 5
	`

	err := testutil.GatherAndCompare(registry, strings.NewReader(expected), "test_requests_total")
	assert.NoError(t, err)
}

func TestModule_FxMetricsModule_WithMultipleCollectors(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_fxmetricsmodule_withcollectors")
	t.Setenv("APP_ENV", "dev")

	counterA := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_alpha_total",
		Help: "Alpha counter.",
	})
	counterB := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_beta_total",
		Help: "Beta counter.",
	})

	var registry *prometheus.Registry

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxMetricsModule,
		AsMetricsCollectors(counterA, counterB),
		fx.Invoke(func() {
			counterA.Add(1)
			counterB.Add(2)
		}),
		fx.Populate(&registry),
	).RequireStart().RequireStop()

	require.NotNil(t, registry)

	mfs, err := registry.Gather()
	require.NoError(t, err)
	assert.Len(t, mfs, 2)
}
