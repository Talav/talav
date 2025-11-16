package fxvalidator

import (
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

// AsStructValidator registers a struct validation definition.
// Additional annotations can be passed as variadic arguments.
func AsStructValidator(def validator.StructValidationDefinition, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-struct-validations"`))

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
}

// AsCustomTypeValidator registers a custom type definition.
// Additional annotations can be passed as variadic arguments.
func AsCustomTypeValidator(def validator.CustomTypeDefinition, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.ResultTags(`group:"talav-validator-custom-types"`))

	return fx.Supply(
		fx.Annotate(
			def,
			annotations...,
		),
	)
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
