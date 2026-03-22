# metrics

Prometheus metrics registry factory. Provides an isolated `*prometheus.Registry` per application instance, avoiding the global default registry (which causes data races in parallel tests).

## Usage

```go
factory := metrics.NewDefaultMetricsRegistryFactory()

registry, err := factory.Create(metrics.MetricsConfig{
    Collect: metrics.CollectConfig{
        Build:   true,  // go_build_info
        Go:      true,  // go_goroutines, go_gc_duration_seconds, etc.
        Process: true,  // process_cpu_seconds_total, process_open_fds, etc.
    },
})
```

## Configuration

```go
type MetricsConfig struct {
    Collect CollectConfig `config:"collect"`
}

type CollectConfig struct {
    Build   bool `config:"build"`    // prometheus/client_golang BuildInfoCollector
    Go      bool `config:"go"`       // prometheus/client_golang GoCollector
    Process bool `config:"process"`  // prometheus/client_golang ProcessCollector
}
```

All collectors default to disabled. Enable via config or `DefaultMetricsConfig()`.

## Registering custom collectors

```go
registry, _ := factory.Create(cfg)

httpRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
    Name: "http_requests_total",
    Help: "Total HTTP requests by method and status.",
}, []string{"method", "status"})

registry.MustRegister(httpRequests)
```
