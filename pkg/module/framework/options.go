package framework

import "go.uber.org/fx"

// Option is a function that configures an Application.
type Option func(*Application)

// WithName sets the application name.
func WithName(name string) Option {
	return func(a *Application) {
		a.name = name
	}
}

// WithVersion sets the application version.
func WithVersion(version string) Option {
	return func(a *Application) {
		a.version = version
	}
}

// WithEnvironment sets the application environment.
func WithEnvironment(env string) Option {
	return func(a *Application) {
		a.environment = env
	}
}

// WithModules adds FX modules to the application.
func WithModules(modules ...fx.Option) Option {
	return func(a *Application) {
		a.modules = append(a.modules, modules...)
	}
}
