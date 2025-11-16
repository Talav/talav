package fxvalidator

import (
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
	"go.uber.org/fx"
)

// AsValidatorAlias registers an alias definition.
// Additional annotations can be passed as variadic arguments.
func AsValidatorAlias(alias validator.AliasDefinition, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-aliases"`))

	return fx.Supply(
		fx.Annotate(
			alias,
			annotations...,
		),
	)
}

// AsValidator registers a validation definition (non-context).
// Additional annotations can be passed as variadic arguments.
func AsValidator(def validator.ValidationDefinition, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-validations"`))

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

// AsValidatorCtx registers a validation definition (context).
// Additional annotations can be passed as variadic arguments.
func AsValidatorCtx(def validator.ValidationDefinitionCtx, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-validations-ctx"`))

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

// AsValidatorConstructor registers a validation definition constructor (non-context).
// The constructor will be called by FX with dependency injection.
// The constructor must return a type that implements validator.ValidationDefinition.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
func AsValidatorConstructor(constructor any, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations,
		fx.ResultTags(`group:"talav-validator-validations"`),
		fx.As(new(validator.ValidationDefinition)),
	)

	return fx.Provide(
		fx.Annotate(
			constructor,
			annotations...,
		),
	)
}

// AsValidatorConstructorCtx registers a validation definition constructor (context).
// The constructor will be called by FX with dependency injection.
// The constructor must return a type that implements validator.ValidationDefinitionCtx.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
func AsValidatorConstructorCtx(constructor any, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations,
		fx.ResultTags(`group:"talav-validator-validations-ctx"`),
		fx.As(new(validator.ValidationDefinitionCtx)),
	)

	return fx.Provide(
		fx.Annotate(
			constructor,
			annotations...,
		),
	)
}

// AsStructValidation registers a struct level validation.
func AsStructValidation(fn playgroundvalidator.StructLevelFuncCtx, types ...any) fx.Option {
	def := &structValidationDefinition{
		fn:    fn,
		types: types,
	}
	annotations := []fx.Annotation{
		fx.ResultTags(`group:"talav-validator-struct-validations"`),
		fx.As(new(validator.StructValidationDefinition)),
	}

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

type structValidationDefinition struct {
	fn    playgroundvalidator.StructLevelFuncCtx
	types []any
}

func (d *structValidationDefinition) Fn() playgroundvalidator.StructLevelFuncCtx {
	return d.fn
}

func (d *structValidationDefinition) Types() []any {
	return d.types
}

// AsCustomType registers a custom type.
func AsCustomType(fn playgroundvalidator.CustomTypeFunc, types ...any) fx.Option {
	def := &customTypeDefinition{
		fn:    fn,
		types: types,
	}
	annotations := []fx.Annotation{
		fx.ResultTags(`group:"talav-validator-custom-types"`),
		fx.As(new(validator.CustomTypeDefinition)),
	}

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

type customTypeDefinition struct {
	fn    playgroundvalidator.CustomTypeFunc
	types []any
}

func (d *customTypeDefinition) Fn() playgroundvalidator.CustomTypeFunc {
	return d.fn
}

func (d *customTypeDefinition) Types() []any {
	return d.types
}

// AsTranslation registers a translation definition.
// Additional annotations can be passed as variadic arguments.
func AsTranslation(def validator.TranslationDefinition, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-translations"`))

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

// AsTranslationConstructor registers a translation definition constructor.
// The constructor will be called by FX with dependency injection.
// The constructor must return a type that implements validator.TranslationDefinition.
// Additional annotations (like fx.ParamTags) can be passed as variadic arguments.
func AsTranslationConstructor(constructor any, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations,
		fx.ResultTags(`group:"talav-validator-translations"`),
		fx.As(new(validator.TranslationDefinition)),
	)

	return fx.Provide(
		fx.Annotate(
			constructor,
			annotations...,
		),
	)
}
