package validator

import (
	v "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
	pv "github.com/wagslane/go-password-validator"
)

// Ensure PasswordValidator implements ValidationDefinition.
var _ validator.ValidationDefinition = (*PasswordValidator)(nil)

const minEntropyBits = 60

// PasswordValidator validates password strength.
type PasswordValidator struct{}

// NewPasswordValidator creates a new password validator.
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{}
}

// Tag returns the validator tag name.
func (p *PasswordValidator) Tag() string {
	return "password"
}

// Fn returns the validation function.
func (p *PasswordValidator) Fn() v.Func {
	return func(fl v.FieldLevel) bool {
		password := fl.Field().String()
		return pv.GetEntropy(password) >= minEntropyBits
	}
}

// CallEvenIfNull returns false.
func (p *PasswordValidator) CallEvenIfNull() bool {
	return false
}
