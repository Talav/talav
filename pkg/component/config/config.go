package config

import (
	"github.com/knadh/koanf/v2"
)

// Config allows to access the application configuration using koanf.
type Config struct {
	k *koanf.Koanf
}

// UnmarshalKey unmarshals the configuration at the given key path into the provided struct.
// Uses "config" struct tag by default (instead of "koanf").
func (c *Config) UnmarshalKey(key string, dest any) error {
	return c.k.UnmarshalWithConf(key, dest, koanf.UnmarshalConf{Tag: "config"})
}

type AppConfig struct {
	Env     string `config:"env"`
	Name    string `config:"name"`
	Version string `config:"version"`
}
