package validator

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// TranslatorFactory is the interface for [ut.Translator] factories.
type TranslatorFactory interface {
	Create() (ut.Translator, error)
}

// DefaultTranslatorFactory is the default [TranslatorFactory] implementation.
type DefaultTranslatorFactory struct{}

// NewDefaultTranslatorFactory returns a [DefaultTranslatorFactory], implementing [TranslatorFactory].
func NewDefaultTranslatorFactory() TranslatorFactory {
	return &DefaultTranslatorFactory{}
}

// Create returns a new [ut.Translator].
func (f *DefaultTranslatorFactory) Create() (ut.Translator, error) {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, found := uni.GetTranslator("en")
	if !found {
		return nil, fmt.Errorf("translator 'en' not found")
	}

	return trans, nil
}

// RegisterTranslations registers default and custom validator translations.
func RegisterTranslations(v *validator.Validate, trans ut.Translator, definitions []TranslationDefinition) error {
	// Register all default translations (required, email, min, max, etc.)
	if err := en_translations.RegisterDefaultTranslations(v, trans); err != nil {
		return fmt.Errorf("failed to register default translations: %w", err)
	}

	// Register custom translations
	for _, def := range definitions {
		if err := def.Register(v, trans); err != nil {
			return fmt.Errorf("failed to register translation for tag %s: %w", def.Tag(), err)
		}
	}

	return nil
}
