package fxorm

import (
	"github.com/talav/talav/pkg/component/orm"
	"go.uber.org/fx"
)

// AsRepository registers a repository to the repository-checkers group.
// Any repository implementing orm.BaseRepositoryInterface[T] can be registered.
// The registry will automatically collect all registered repositories.
// The repository is also provided as its concrete type for direct injection.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	var FxMyModule = fx.Module(
//		"my",
//		AsRepository(repo.NewUserRepository),
//	)
func AsRepository[T any](constructor any, annotations ...fx.Annotation) fx.Option {
	// Provide to the group as ExistsChecker
	groupAnnotations := append(annotations,
		fx.ResultTags(`group:"repository-checkers"`),
		fx.As(new(orm.ExistsChecker)),
	)

	// Provide as concrete type for direct injection
	// Note: We need to provide the constructor twice - once for the group,
	// and once as the concrete type. FX will only call the constructor once.
	return fx.Options(
		// Provide to group
		fx.Provide(
			fx.Annotate(
				constructor,
				groupAnnotations...,
			),
		),
		// Provide as concrete type
		fx.Provide(constructor),
	)
}
