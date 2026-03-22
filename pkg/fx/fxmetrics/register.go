package fxmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// AsMetricsCollector registers a [prometheus.Collector] instance into the metrics registry.
//
// The collector is added to the group consumed by [FxMetricsModule] and registered on the
// shared [prometheus.Registry].
//
// Using this helper (rather than promauto or the global registry) avoids data race conditions
// when running parallel tests, because each test can spin up its own isolated FX app.
//
// Example:
//
//	var RequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
//		Name: "http_requests_total",
//		Help: "Total number of HTTP requests.",
//	}, []string{"method", "status"})
//
//	var FxMyModule = fx.Module(
//		"my",
//		fxmetrics.AsMetricsCollector(RequestsTotal),
//	)
func AsMetricsCollector(collector prometheus.Collector) fx.Option {
	return fx.Supply(
		fx.Annotate(
			collector,
			fx.As(new(prometheus.Collector)),
			fx.ResultTags(`group:"metrics-collectors"`),
		),
	)
}

// AsMetricsCollectors registers multiple [prometheus.Collector] instances into the metrics registry.
//
// This is a convenience wrapper around [AsMetricsCollector] for registering several collectors at once.
//
// Example:
//
//	var FxMyModule = fx.Module(
//		"my",
//		fxmetrics.AsMetricsCollectors(RequestsTotal, ResponseDuration),
//	)
func AsMetricsCollectors(collectors ...prometheus.Collector) fx.Option {
	opts := make([]fx.Option, 0, len(collectors))
	for _, c := range collectors {
		opts = append(opts, AsMetricsCollector(c))
	}

	return fx.Options(opts...)
}
