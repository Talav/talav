package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidatorFactory is the interface for [validator.Validate] factories.
type ValidatorFactory interface {
	Create(
		aliasDefs []AliasDefinition,
		validationDefs []ValidationDefinition,
		validationDefsCtx []ValidationDefinitionCtx,
		structValidationDefs []StructValidationDefinition,
		customTypeDefs []CustomTypeDefinition,
	) (*validator.Validate, error)
}

// DefaultValidatorFactory is the default [ValidatorFactory] implementation.
type DefaultValidatorFactory struct{}

// NewDefaultValidatorFactory returns a [DefaultValidatorFactory], implementing [ValidatorFactory].
func NewDefaultValidatorFactory() ValidatorFactory {
	return &DefaultValidatorFactory{}
}

// Create returns a new [validator.Validate] with all validators registered.
func (f *DefaultValidatorFactory) Create(
	aliasDefs []AliasDefinition,
	validationDefs []ValidationDefinition,
	validationDefsCtx []ValidationDefinitionCtx,
	structValidationDefs []StructValidationDefinition,
	customTypeDefs []CustomTypeDefinition,
) (*validator.Validate, error) {
	v := validator.New()

	f.registerTagNameFunc(v)
	f.registerAliases(v, aliasDefs)

	if err := f.registerFieldValidations(v, validationDefs); err != nil {
		return nil, err
	}

	if err := f.registerFieldValidationsCtx(v, validationDefsCtx); err != nil {
		return nil, err
	}

	f.registerStructValidations(v, structValidationDefs)
	f.registerCustomTypeValidations(v, customTypeDefs)

	return v, nil
}

func (f *DefaultValidatorFactory) registerTagNameFunc(v *validator.Validate) {
	// Register JSON tag name function
	// This makes err.Field() and err.Namespace() return JSON tag names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		jsonTag := fld.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			return ""
		}
		// Extract tag name before comma (e.g., "email" from "email,omitempty")
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			return jsonTag[:idx]
		}

		return jsonTag
	})
}

func (f *DefaultValidatorFactory) registerAliases(v *validator.Validate, aliasDefs []AliasDefinition) {
	for _, def := range aliasDefs {
		v.RegisterAlias(def.Alias(), def.Tags())
	}
}

func (f *DefaultValidatorFactory) registerFieldValidations(v *validator.Validate, validationDefs []ValidationDefinition) error {
	for _, def := range validationDefs {
		if err := v.RegisterValidation(def.Tag(), def.Fn(), def.CallEvenIfNull()); err != nil {
			return fmt.Errorf("failed to register validator %s: %w", def.Tag(), err)
		}
	}

	return nil
}

func (f *DefaultValidatorFactory) registerFieldValidationsCtx(v *validator.Validate, validationDefsCtx []ValidationDefinitionCtx) error {
	for _, def := range validationDefsCtx {
		if err := v.RegisterValidationCtx(def.Tag(), def.Fn(), def.CallEvenIfNull()); err != nil {
			return fmt.Errorf("failed to register context validator %s: %w", def.Tag(), err)
		}
	}

	return nil
}

func (f *DefaultValidatorFactory) registerStructValidations(v *validator.Validate, structValidationDefs []StructValidationDefinition) {
	for _, def := range structValidationDefs {
		types := def.Types()
		typeArgs := make([]any, len(types))
		for i, t := range types {
			typeArgs[i] = t
		}
		v.RegisterStructValidationCtx(def.Fn(), typeArgs...)
	}
}

func (f *DefaultValidatorFactory) registerCustomTypeValidations(v *validator.Validate, customTypeDefs []CustomTypeDefinition) {
	for _, def := range customTypeDefs {
		types := def.Types()
		typeArgs := make([]any, len(types))
		for i, t := range types {
			typeArgs[i] = t
		}
		v.RegisterCustomTypeFunc(def.Fn(), typeArgs...)
	}
}
