package validator

import (
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
)

// Ensure PasswordTranslation implements TranslationDefinition.
var _ validator.TranslationDefinition = (*PasswordTranslation)(nil)

// PasswordTranslation defines the translation for the "password" validator.
type PasswordTranslation struct{}

// NewPasswordTranslation creates a new password translation definition.
func NewPasswordTranslation() *PasswordTranslation {
	return &PasswordTranslation{}
}

// Tag returns the validator tag name.
func (t *PasswordTranslation) Tag() string {
	return "password"
}

// Register registers the translation for the "password" validator.
func (t *PasswordTranslation) Register(validator *v.Validate, trans ut.Translator) error {
	return validator.RegisterTranslation("password", trans,
		func(ut ut.Translator) error {
			return ut.Add("password", "{0} is too weak", false)
		},
		func(ut ut.Translator, fe v.FieldError) string {
			t, _ := ut.T(fe.Tag(), fe.Field())
			return t
		})
}
