# fxmetrics

FX module for [pkg/component/metrics](../../component/metrics). Provides an isolated `*prometheus.Registry` and a mechanism for FX modules to register their own `prometheus.Collector` implementations.

## Quick start

```go
fx.New(
    fxconfig.FxConfigModule,
    fxmetrics.FxMetricsModule,
    fx.Invoke(func(registry *prometheus.Registry) {
        // registry is ready; mount promhttp.HandlerFor(registry, ...) on your HTTP server
    }),
)
```

## Configuration

```yaml
metrics:
  collect:
    build: true    # go_build_info
    go: true       # go_goroutines, go_gc_duration_seconds, etc.
    process: true  # process_cpu_seconds_total, process_open_fds, etc.
```

All collectors default to disabled.

## Registering custom collectors

```go
// In any FX module:
fxmetrics.AsMetricsCollector(func() prometheus.Collector {
    return prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests.",
    }, []string{"method", "status"})
})
```

All collectors provided via `AsMetricsCollector` are registered into the shared registry during FX startup.

## Exposing `/metrics`

This module does not mount an HTTP endpoint. Mount the handler manually on your server:

```go
r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
```

A dedicated ops server for `/metrics` (separate port from the main HTTP server) is on the roadmap.

## Injected types

- `*prometheus.Registry` — the isolated Prometheus registry
