package fxorm

import (
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
)

// Ensure UniqueTranslation implements TranslationDefinition
var _ validator.TranslationDefinition = (*UniqueTranslation)(nil)

// UniqueTranslation defines the translation for the "unique" validator.
type UniqueTranslation struct{}

// NewUniqueTranslation creates a new unique translation definition.
func NewUniqueTranslation() *UniqueTranslation {
	return &UniqueTranslation{}
}

// Tag returns the validator tag name.
func (t *UniqueTranslation) Tag() string {
	return "unique"
}

// Register registers the translation for the "unique" validator.
func (t *UniqueTranslation) Register(validator *v.Validate, trans ut.Translator) error {
	return validator.RegisterTranslation("unique", trans,
		func(ut ut.Translator) error {
			return ut.Add("unique", "{0} already exists", true)
		},
		func(ut ut.Translator, fe v.FieldError) string {
			t, _ := ut.T(fe.Tag(), fe.Field())
			return t
		})
}
