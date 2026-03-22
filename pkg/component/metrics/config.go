package metrics

// CollectConfig controls which optional Prometheus collectors are registered.
type CollectConfig struct {
	Build   bool `config:"build"`
	Go      bool `config:"go"`
	Process bool `config:"process"`
}

// MetricsConfig holds configuration for the metrics component.
type MetricsConfig struct {
	Collect CollectConfig `config:"collect"`
}

// DefaultMetricsConfig returns a [MetricsConfig] with all optional collectors disabled.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Collect: CollectConfig{
			Build:   false,
			Go:      false,
			Process: false,
		},
	}
}
