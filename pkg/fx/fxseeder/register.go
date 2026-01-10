package fxseeder

import (
	"github.com/talav/talav/pkg/component/seeder"
	"go.uber.org/fx"
)

// AsSeeder registers a seeder constructor to the seeders group.
// The seeder will be automatically collected and filtered by environment.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	var FxMyModule = fx.Module(
//		"my",
//		fxseeder.AsSeeder(NewUserSeeder),
//	)
func AsSeeder(constructor any, annotations ...fx.Annotation) fx.Option {
	allAnnotations := make([]fx.Annotation, 0, len(annotations)+2)
	allAnnotations = append(allAnnotations, annotations...)
	allAnnotations = append(allAnnotations,
		fx.ResultTags(`group:"seeders"`),
		fx.As(new(seeder.Seeder)),
	)

	return fx.Provide(
		fx.Annotate(
			constructor,
			allAnnotations...,
		),
	)
}
