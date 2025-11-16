package fxvalidator

import (
	ut "github.com/go-playground/universal-translator"
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
	"go.uber.org/fx"
)

const ModuleName = "validator"

// FxValidatorModule is the [Fx] validator module.
var FxValidatorModule = fx.Module(
	ModuleName,
	fx.Provide(
		validator.NewDefaultValidatorFactory,
		validator.NewDefaultTranslatorFactory,
		NewFxValidator,
		NewFxTranslator,
	),
)

// FxValidatorParam allows injection of the required dependencies in [NewFxValidator].
type FxValidatorParam struct {
	fx.In
	Factory                     validator.ValidatorFactory
	AliasDefinitions            []validator.AliasDefinition            `group:"talav-validator-aliases"`
	ValidationDefinitions       []validator.ValidationDefinition       `group:"talav-validator-validations"`
	ValidationDefinitionsCtx    []validator.ValidationDefinitionCtx    `group:"talav-validator-validations-ctx"`
	StructValidationDefinitions []validator.StructValidationDefinition `group:"talav-validator-struct-validations"`
	CustomTypeDefinitions       []validator.CustomTypeDefinition       `group:"talav-validator-custom-types"`
}

// NewFxValidator returns a [playgroundvalidator.Validate].
func NewFxValidator(p FxValidatorParam) (*playgroundvalidator.Validate, error) {
	// Filter out nil definitions (e.g., when optional dependencies are missing)
	validDefs := make([]validator.ValidationDefinition, 0, len(p.ValidationDefinitions))
	for _, def := range p.ValidationDefinitions {
		if def != nil {
			validDefs = append(validDefs, def)
		}
	}

	validDefsCtx := make([]validator.ValidationDefinitionCtx, 0, len(p.ValidationDefinitionsCtx))
	for _, def := range p.ValidationDefinitionsCtx {
		if def != nil {
			validDefsCtx = append(validDefsCtx, def)
		}
	}

	structDefs := make([]validator.StructValidationDefinition, 0, len(p.StructValidationDefinitions))
	for _, def := range p.StructValidationDefinitions {
		if def != nil {
			structDefs = append(structDefs, def)
		}
	}

	customTypeDefs := make([]validator.CustomTypeDefinition, 0, len(p.CustomTypeDefinitions))
	for _, def := range p.CustomTypeDefinitions {
		if def != nil {
			customTypeDefs = append(customTypeDefs, def)
		}
	}

	return p.Factory.Create(
		p.AliasDefinitions,
		validDefs,
		validDefsCtx,
		structDefs,
		customTypeDefs,
	)
}

// FxTranslatorParam allows injection of the required dependencies in [NewFxTranslator].
type FxTranslatorParam struct {
	fx.In
	Factory                validator.TranslatorFactory
	TranslationDefinitions []validator.TranslationDefinition `group:"talav-validator-translations"`
	Validator              *playgroundvalidator.Validate
}

// NewFxTranslator returns a [ut.Translator] and registers translations.
func NewFxTranslator(p FxTranslatorParam) (ut.Translator, error) {
	translator, err := p.Factory.Create()
	if err != nil {
		return nil, err
	}

	// Register translations with the validator
	if err := validator.RegisterTranslations(p.Validator, translator, p.TranslationDefinitions); err != nil {
		return nil, err
	}

	return translator, nil
}
