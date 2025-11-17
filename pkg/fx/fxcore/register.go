package fxcore

import (
	"go.uber.org/fx"
)

// AsRootCommand registers a top-level command constructor to the cobra-commands group.
// Top-level commands are added directly to the root command.
// It wraps the constructor with fx.Annotate to add the group tag.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	var FxMyModule = fx.Module(
//		"my",
//		AsRootCommand(cmd.NewMyCmd),
//	)
//
// Example with parameter tags:
//
//	var FxMyModule = fx.Module(
//		"my",
//		AsRootCommand(
//			cmd.NewUserCmd,
//			fx.ParamTags(`name:"create-user-cmd"`, `name:"search-user-cmd"`),
//		),
//	)
func AsRootCommand(constructor any, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-core-commands"`))

	return fx.Provide(
		fx.Annotate(
			constructor,
			annotations...,
		),
	)
}

// AsNamedCommand registers a command constructor with a named tag.
// Use this for subcommands that will be injected into parent commands.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
//
// Example:
//
//	var FxMyModule = fx.Module(
//		"my",
//		AsNamedCommand("create-user-cmd", cmd.NewCreateUserCmd),
//	)
func AsNamedCommand(name string, constructor any, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`name:"`+name+`"`))

	return fx.Provide(
		fx.Annotate(
			constructor,
			annotations...,
		),
	)
}
