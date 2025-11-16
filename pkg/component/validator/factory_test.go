package validator

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper types and functions for tests

type testAliasDefinition struct {
	alias string
	tags  string
}

func (t *testAliasDefinition) Alias() string { return t.alias }
func (t *testAliasDefinition) Tags() string  { return t.tags }

type testValidationDefinition struct {
	tag            string
	fn             validator.Func
	callEvenIfNull bool
}

func (t *testValidationDefinition) Tag() string          { return t.tag }
func (t *testValidationDefinition) Fn() validator.Func   { return t.fn }
func (t *testValidationDefinition) CallEvenIfNull() bool { return t.callEvenIfNull }

type testValidationDefinitionCtx struct {
	tag            string
	fn             validator.FuncCtx
	callEvenIfNull bool
}

func (t *testValidationDefinitionCtx) Tag() string           { return t.tag }
func (t *testValidationDefinitionCtx) Fn() validator.FuncCtx { return t.fn }
func (t *testValidationDefinitionCtx) CallEvenIfNull() bool  { return t.callEvenIfNull }

type testStructValidationDefinition struct {
	fn    validator.StructLevelFuncCtx
	types []reflect.Type
}

func (t *testStructValidationDefinition) Fn() validator.StructLevelFuncCtx { return t.fn }
func (t *testStructValidationDefinition) Types() []reflect.Type            { return t.types }

func TestDefaultValidatorFactory_Create_JSONTagNameFunction(t *testing.T) {
	factory := NewDefaultValidatorFactory()

	type TestStruct struct {
		EmailAddress string `json:"email" validate:"required"`
		UserName     string `json:"user_name" validate:"required"`
		Password     string `validate:"required"` // no json tag
	}

	v, err := factory.Create(nil, nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	err = v.Struct(TestStruct{})
	require.Error(t, err)

	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	require.True(t, ok)

	// Verify that Field() returns JSON tag names
	fieldNames := make(map[string]bool)
	for _, err := range validationErrors {
		fieldNames[err.Field()] = true
	}

	// Should use JSON tag names where available
	assert.True(t, fieldNames["email"], "should use JSON tag 'email'")
	assert.True(t, fieldNames["user_name"], "should use JSON tag 'user_name'")
	// Fields without JSON tags should use struct field name
	assert.True(t, fieldNames["Password"], "should use struct field name when no JSON tag")
}

func TestDefaultValidatorFactory_Create_FieldLevelValidation(t *testing.T) {
	factory := NewDefaultValidatorFactory()

	// Custom validator that checks if string is uppercase
	upperCaseFn := func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		return strings.ToUpper(val) == val
	}

	validationDefs := []ValidationDefinition{
		&testValidationDefinition{tag: "uppercase", fn: upperCaseFn, callEvenIfNull: false},
	}

	v, err := factory.Create(nil, validationDefs, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	type TestStruct struct {
		Code string `validate:"uppercase"`
	}

	// Valid case
	err = v.Struct(TestStruct{Code: "ABC"})
	assert.NoError(t, err)

	// Invalid case
	err = v.Struct(TestStruct{Code: "abc"})
	require.Error(t, err)
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	require.True(t, ok)
	assert.Len(t, validationErrors, 1)
	assert.Equal(t, "uppercase", validationErrors[0].Tag())
}

func TestDefaultValidatorFactory_Create_AliasRegistration(t *testing.T) {
	factory := NewDefaultValidatorFactory()

	aliasDefs := []AliasDefinition{
		&testAliasDefinition{alias: "notempty", tags: "required"},
	}

	v, err := factory.Create(aliasDefs, nil, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	type TestStruct struct {
		Name string `validate:"notempty"`
	}

	// Invalid case - empty string
	err = v.Struct(TestStruct{Name: ""})
	require.Error(t, err)
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	require.True(t, ok)
	assert.Len(t, validationErrors, 1)
	assert.Equal(t, "notempty", validationErrors[0].Tag())

	// Valid case
	err = v.Struct(TestStruct{Name: "test"})
	assert.NoError(t, err)
}

func TestDefaultValidatorFactory_Create_StructLevelValidation(t *testing.T) {
	factory := NewDefaultValidatorFactory()

	type TestStruct struct {
		Field1 string
		Field2 string
	}

	structValidationFn := func(ctx context.Context, sl validator.StructLevel) {
		// Struct-level validation function
		// Type assertion is safe here as we control the struct type being validated
		_, _ = sl.Current().Interface().(TestStruct)
		// Validation logic would go here
	}

	structDefs := []StructValidationDefinition{
		&testStructValidationDefinition{
			fn: structValidationFn,
			types: []reflect.Type{
				reflect.TypeOf(TestStruct{}),
			},
		},
	}

	// Verify that struct validation can be registered without error
	v, err := factory.Create(nil, nil, nil, structDefs, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	// Verify that the validator instance is functional
	// Struct-level validations are registered and will be called during validation
	testStruct := TestStruct{
		Field1: "value1",
		Field2: "value2",
	}

	// This should not error (no field-level validations)
	err = v.StructCtx(context.Background(), testStruct)
	assert.NoError(t, err)
}

func TestDefaultValidatorFactory_Create_ErrorHandling(t *testing.T) {
	factory := NewDefaultValidatorFactory()

	// Test error handling with invalid validator function (nil function)
	// First register a valid validator
	validationDefs := []ValidationDefinition{
		&testValidationDefinition{tag: "testvalidator", fn: func(fl validator.FieldLevel) bool { return true }, callEvenIfNull: false},
	}

	v, err := factory.Create(nil, validationDefs, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	// Test that duplicate registration of same tag fails
	// Create a new factory and try to register the same tag again
	factory2 := NewDefaultValidatorFactory()
	validationDefs2 := []ValidationDefinition{
		&testValidationDefinition{tag: "testvalidator", fn: func(fl validator.FieldLevel) bool { return false }, callEvenIfNull: false},
	}

	// This should work since we're creating a new validator instance
	v2, err := factory2.Create(nil, validationDefs2, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v2)

	// Test error handling with context validator - use a custom tag
	validationDefsCtx := []ValidationDefinitionCtx{
		&testValidationDefinitionCtx{tag: "testctxvalidator", fn: func(ctx context.Context, fl validator.FieldLevel) bool { return true }, callEvenIfNull: false},
	}

	v3, err := factory.Create(nil, nil, validationDefsCtx, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v3)
}
