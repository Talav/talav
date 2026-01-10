package fxmedia

import (
	"github.com/talav/talav/pkg/component/media/app/provider"
	"github.com/talav/talav/pkg/component/media/infra/resizer"
	"go.uber.org/fx"
)

// AsResizer registers a resizer with a named tag for dependency injection.
func AsResizer(name string, constructor any, annotations ...fx.Annotation) fx.Option {
	// Create a copy to avoid modifying the input slice
	interfaceAnnotations := make([]fx.Annotation, len(annotations), len(annotations)+2)
	copy(interfaceAnnotations, annotations)
	interfaceAnnotations = append(interfaceAnnotations,
		fx.ResultTags(`name:"media-resizer-`+name+`"`),
		fx.As(new(resizer.Resizer)),
	)

	return fx.Options(
		// Provide as interface type with name tag
		fx.Provide(
			fx.Annotate(
				constructor,
				interfaceAnnotations...,
			),
		),
		// Register to resizers group
		fx.Provide(
			fx.Annotate(
				func(r resizer.Resizer) ResizerEntry {
					return ResizerEntry{Name: name, Resizer: r}
				},
				fx.ParamTags(`name:"media-resizer-`+name+`"`),
				fx.ResultTags(`group:"resizers"`),
			),
		),
	)
}

// AsProvider registers a provider with a named tag for dependency injection.
func AsProvider(name string, constructor any, annotations ...fx.Annotation) fx.Option {
	// Create a copy to avoid modifying the input slice
	interfaceAnnotations := make([]fx.Annotation, len(annotations), len(annotations)+2)
	copy(interfaceAnnotations, annotations)
	interfaceAnnotations = append(interfaceAnnotations,
		fx.ResultTags(`name:"media-provider-`+name+`"`),
		fx.As(new(provider.Provider)),
	)

	// Provide as concrete type for direct injection
	// Note: We need to provide the constructor twice - once for the interface,
	// and once as the concrete type. FX will only call the constructor once.
	return fx.Options(
		// Provide as interface type
		fx.Provide(
			fx.Annotate(
				constructor,
				interfaceAnnotations...,
			),
		),
		// Provide as concrete type
		fx.Provide(constructor),
	)
}
