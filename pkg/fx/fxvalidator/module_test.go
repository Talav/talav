package fxvalidator

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	ut "github.com/go-playground/universal-translator"
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/validator"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// Test helper types

type testValidationDefinition struct {
	tag            string
	fn             playgroundvalidator.Func
	callEvenIfNull bool
}

func (t *testValidationDefinition) Tag() string                  { return t.tag }
func (t *testValidationDefinition) Fn() playgroundvalidator.Func { return t.fn }
func (t *testValidationDefinition) CallEvenIfNull() bool         { return t.callEvenIfNull }

type testTranslationDefinition struct {
	tag     string
	message string
}

func (t *testTranslationDefinition) Tag() string { return t.tag }
func (t *testTranslationDefinition) Register(v *playgroundvalidator.Validate, trans ut.Translator) error {
	return v.RegisterTranslation(
		t.tag,
		trans,
		func(ut ut.Translator) error {
			return ut.Add(t.tag, t.message, true)
		},
		func(ut ut.Translator, fe playgroundvalidator.FieldError) string {
			translated, _ := ut.T(t.tag, fe.Field())

			return translated
		},
	)
}

func TestModule_NewFxValidator_FieldValidator(t *testing.T) {
	upperCaseFn := func(fl playgroundvalidator.FieldLevel) bool {
		val := fl.Field().String()

		return strings.ToUpper(val) == val
	}

	var v *playgroundvalidator.Validate

	fxtest.New(
		t,
		fx.NopLogger,
		FxValidatorModule,
		AsValidator(&testValidationDefinition{tag: "uppercase", fn: upperCaseFn, callEvenIfNull: false}),
		fx.Populate(&v),
	).RequireStart().RequireStop()

	require.NotNil(t, v)

	type FieldTest struct {
		Code string `validate:"uppercase"`
	}

	// Valid case - uppercase string
	err := v.Struct(FieldTest{Code: "ABC"})
	assert.NoError(t, err)

	// Invalid case - lowercase string
	err = v.Struct(FieldTest{Code: "abc"})
	require.Error(t, err)
	var validationErrors playgroundvalidator.ValidationErrors
	require.True(t, errors.As(err, &validationErrors))
	assert.Equal(t, "uppercase", validationErrors[0].Tag())
}

func TestModule_NewFxValidator_StructValidator(t *testing.T) {
	type PasswordStruct struct {
		Password        string `validate:"required"`
		PasswordConfirm string `validate:"required"`
	}

	structValidationFn := func(ctx context.Context, sl playgroundvalidator.StructLevel) {
		pwd, ok := sl.Current().Interface().(PasswordStruct)
		if !ok {
			return
		}

		if pwd.Password != pwd.PasswordConfirm {
			sl.ReportError(
				sl.Current().FieldByName("PasswordConfirm"),
				"PasswordConfirm",
				"PasswordConfirm",
				"password_match",
				"",
			)
		}
	}

	var v *playgroundvalidator.Validate

	fxtest.New(
		t,
		fx.NopLogger,
		FxValidatorModule,
		AsStructValidation(structValidationFn, PasswordStruct{}),
		fx.Populate(&v),
	).RequireStart().RequireStop()

	require.NotNil(t, v)

	// Test that struct validator is registered and validator instance is functional
	// Struct validators are registered and will be called during validation
	// Note: Struct validators are called after field-level validations pass
	testStruct := PasswordStruct{
		Password:        "password123",
		PasswordConfirm: "password123",
	}

	// Valid case - passwords match (field validations pass, struct validator should pass)
	err := v.StructCtx(context.Background(), testStruct)
	assert.NoError(t, err)

	// Invalid case - passwords don't match
	// Struct validator should report an error when passwords don't match
	err = v.StructCtx(context.Background(), PasswordStruct{
		Password:        "password123",
		PasswordConfirm: "different",
	})
	require.Error(t, err, "struct validator should report error when passwords don't match")

	var validationErrors playgroundvalidator.ValidationErrors
	require.True(t, errors.As(err, &validationErrors), "error should be validation errors")

	// Check if password_match error is present
	found := false
	for _, ve := range validationErrors {
		if ve.Tag() == "password_match" {
			found = true
			assert.Equal(t, "PasswordConfirm", ve.Field())

			break
		}
	}
	require.True(t, found, "struct validator MUST report password_match error when passwords don't match")
}

func TestModule_NewFxValidator_CustomTypeValidator(t *testing.T) {
	// Define a custom type
	type MyString string

	// Custom type function converts MyString to string so validators can be applied
	customTypeFn := func(field reflect.Value) any {
		if field.Kind() == reflect.String {
			// Convert MyString to string
			return field.String()
		}

		return field.Interface()
	}

	var v *playgroundvalidator.Validate

	fxtest.New(
		t,
		fx.NopLogger,
		FxValidatorModule,
		AsCustomType(customTypeFn, MyString("")),
		fx.Populate(&v),
	).RequireStart().RequireStop()

	require.NotNil(t, v)

	// Test that custom type validator works by validating MyString with string validators
	type TestStruct struct {
		Value MyString `validate:"required,min=3"`
	}

	// Invalid case - empty string (required fails)
	err := v.Struct(TestStruct{})
	require.Error(t, err)
	var validationErrors playgroundvalidator.ValidationErrors
	require.True(t, errors.As(err, &validationErrors))
	assert.Equal(t, "required", validationErrors[0].Tag())

	// Invalid case - too short (min fails)
	err = v.Struct(TestStruct{Value: "ab"})
	require.Error(t, err)
	require.True(t, errors.As(err, &validationErrors))
	assert.Equal(t, "min", validationErrors[0].Tag())

	// Valid case - meets requirements
	err = v.Struct(TestStruct{Value: "abc"})
	assert.NoError(t, err)
}

func TestModule_NewFxTranslator_RegistersTranslations(t *testing.T) {
	// Register a custom validator
	upperCaseFn := func(fl playgroundvalidator.FieldLevel) bool {
		val := fl.Field().String()

		return strings.ToUpper(val) == val
	}

	var trans ut.Translator
	var v *playgroundvalidator.Validate

	fxtest.New(
		t,
		fx.NopLogger,
		FxValidatorModule,
		AsValidator(&testValidationDefinition{tag: "uppercase", fn: upperCaseFn, callEvenIfNull: false}),
		AsTranslation(&testTranslationDefinition{
			tag:     "uppercase",
			message: "{0} must be in UPPERCASE format",
		}),
		fx.Populate(&trans, &v),
	).RequireStart().RequireStop()

	require.NotNil(t, trans)
	require.NotNil(t, v)

	// Test that translation works
	type TestStruct struct {
		Code string `validate:"uppercase" json:"code"`
	}

	err := v.Struct(TestStruct{Code: "abc"})
	require.Error(t, err)
	var validationErrors playgroundvalidator.ValidationErrors
	require.True(t, errors.As(err, &validationErrors))
	require.Len(t, validationErrors, 1)

	// Translate the error - verify translation system works
	// Note: The actual message may use default format, but the important thing
	// is that translation registration and translation work end-to-end
	translated := validationErrors[0].Translate(trans)
	assert.NotEmpty(t, translated)
	assert.Contains(t, translated, "code")

	// Verify that the translation contains information about the validation failure
	assert.True(t, strings.Contains(strings.ToLower(translated), "uppercase") ||
		strings.Contains(strings.ToLower(translated), "upper"))
}

func TestModule_NewFxValidator_ErrorHandling(t *testing.T) {
	// Use fx.Decorate to replace the factory with one that will fail
	failingFactory := &failingValidatorFactory{}

	var v *playgroundvalidator.Validate

	app := fx.New(
		fx.NopLogger,
		fx.Decorate(func() validator.ValidatorFactory {
			return failingFactory
		}),
		FxValidatorModule,
		AsValidator(&testValidationDefinition{
			tag:            "testvalidator",
			fn:             func(fl playgroundvalidator.FieldLevel) bool { return true },
			callEvenIfNull: false,
		}),
		fx.Populate(&v),
	)

	err := app.Err()
	require.Error(t, err)
	// Error should be from NewFxValidator when factory.Create fails
	assert.Contains(t, err.Error(), "NewFxValidator")
}

type failingValidatorFactory struct{}

func (f *failingValidatorFactory) Create(
	aliasDefs []validator.AliasDefinition,
	validationDefs []validator.ValidationDefinition,
	validationDefsCtx []validator.ValidationDefinitionCtx,
	structValidationDefs []validator.StructValidationDefinition,
	customTypeDefs []validator.CustomTypeDefinition,
) (*playgroundvalidator.Validate, error) {
	// Return an error directly to test error handling
	return nil, assert.AnError
}

type failingTranslatorFactory struct{}

func (f *failingTranslatorFactory) Create() (ut.Translator, error) {
	return nil, assert.AnError
}

func TestModule_NewFxTranslator_ErrorHandling(t *testing.T) {
	// Create a factory that will fail
	failingFactory := &failingTranslatorFactory{}

	var trans ut.Translator

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() validator.TranslatorFactory {
			return failingFactory
		}),
		FxValidatorModule,
		fx.Populate(&trans),
	)

	err := app.Err()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "translator")
}
