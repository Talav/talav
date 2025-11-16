package validator

import (
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
)

// AliasDefinition defines a validator alias.
type AliasDefinition interface {
	Alias() string
	Tags() string
}

// ValidationDefinition defines a field-level validator without context.
type ValidationDefinition interface {
	Tag() string
	Fn() v.Func
	CallEvenIfNull() bool
}

// ValidationDefinitionCtx defines a field-level validator with context.
type ValidationDefinitionCtx interface {
	Tag() string
	Fn() v.FuncCtx
	CallEvenIfNull() bool
}

// StructValidationDefinition defines a struct-level validator.
type StructValidationDefinition interface {
	Fn() v.StructLevelFuncCtx
	Types() []any
}

// CustomTypeDefinition defines a custom type validator.
type CustomTypeDefinition interface {
	Fn() v.CustomTypeFunc
	Types() []any
}

// TranslationDefinition defines a validator translation.
type TranslationDefinition interface {
	Tag() string
	Register(validator *v.Validate, trans ut.Translator) error
}
