# Validator Package

A framework-agnostic validation package built on top of [go-playground/validator/v10](https://github.com/go-playground/validator). Provides a factory pattern for creating and configuring validators with support for custom validators, translations, and structured error handling.

## Features

- **Factory Pattern**: Easy validator creation via `ValidatorFactory` interface
- **Multiple Validator Types**: Support for field-level, struct-level, and custom type validators
- **Context Support**: Validators can optionally receive context for advanced validation scenarios
- **Translation Support**: Built-in support for error message translations via `TranslatorFactory`
- **JSON Tag Integration**: Automatic use of JSON tag names in error messages
- **Validator Aliases**: Create aliases for common validation tag combinations
- **Custom Type Validators**: Register validators for specific types
- **Framework Agnostic**: Use independently or integrate with dependency injection frameworks

## Installation

```bash
go get github.com/talav/talav/pkg/component/validator
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/talav/talav/pkg/component/validator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

type User struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required,min=3"`
}

func main() {
	factory := validator.NewDefaultValidatorFactory()
	
	v, err := factory.Create(nil, nil, nil, nil, nil)
	if err != nil {
		panic(err)
	}
	
	user := User{
		Email: "invalid-email",
		Name:  "ab",
	}
	
	err = v.Struct(user)
	if err != nil {
		fmt.Printf("Validation errors: %v\n", err)
	}
}
```

## Validator Types

### Field-Level Validators (Non-Context)

Simple validators that validate individual fields without context:

```go
import (
	"strings"
	"github.com/talav/talav/pkg/component/validator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

type uppercaseValidator struct {
	tag string
}

func (v *uppercaseValidator) Tag() string {
	return v.tag
}

func (v *uppercaseValidator) Fn() playgroundvalidator.Func {
	return func(fl playgroundvalidator.FieldLevel) bool {
		val := fl.Field().String()
		return strings.ToUpper(val) == val
	}
}

func (v *uppercaseValidator) CallEvenIfNull() bool {
	return false
}

factory := validator.NewDefaultValidatorFactory()
v, _ := factory.Create(
	nil,
	[]validator.ValidationDefinition{&uppercaseValidator{tag: "uppercase"}},
	nil,
	nil,
	nil,
)

type Code struct {
	Value string `validate:"uppercase"`
}

err := v.Struct(Code{Value: "ABC"}) // Valid
err = v.Struct(Code{Value: "abc"})  // Invalid
```

### Field-Level Validators (With Context)

Validators that receive context for advanced scenarios:

```go
import (
	"context"
	"github.com/talav/talav/pkg/component/validator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

type uppercaseValidatorCtx struct {
	tag string
}

func (v *uppercaseValidatorCtx) Tag() string {
	return v.tag
}

func (v *uppercaseValidatorCtx) Fn() playgroundvalidator.FuncCtx {
	return func(ctx context.Context, fl playgroundvalidator.FieldLevel) bool {
		val := fl.Field().String()
		return strings.ToUpper(val) == val
	}
}

func (v *uppercaseValidatorCtx) CallEvenIfNull() bool {
	return false
}

factory := validator.NewDefaultValidatorFactory()
v, _ := factory.Create(
	nil,
	nil,
	[]validator.ValidationDefinitionCtx{&uppercaseValidatorCtx{tag: "uppercase_ctx"}},
	nil,
	nil,
)
```

### Struct-Level Validators

Validators that validate entire structs:

```go
import (
	"context"
	"reflect"
	"github.com/talav/talav/pkg/component/validator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

type passwordMatchValidator struct{}

func (v *passwordMatchValidator) Fn() playgroundvalidator.StructLevelFuncCtx {
	return func(ctx context.Context, sl playgroundvalidator.StructLevel) {
		user := sl.Current().Interface().(struct {
			Password        string
			PasswordConfirm string
		})
		
		if user.Password != user.PasswordConfirm {
			sl.ReportError(
				reflect.ValueOf(user.PasswordConfirm),
				"PasswordConfirm",
				"PasswordConfirm",
				"password_match",
				"",
			)
		}
	}
}

func (v *passwordMatchValidator) Types() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(struct {
			Password        string
			PasswordConfirm string
		}{}),
	}
}

factory := validator.NewDefaultValidatorFactory()
v, _ := factory.Create(
	nil,
	nil,
	nil,
	[]validator.StructValidationDefinition{&passwordMatchValidator{}},
	nil,
)
```

### Custom Type Validators

Register validators for specific types:

```go
import (
	"reflect"
	"github.com/talav/talav/pkg/component/validator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

type MyString string

type customTypeValidator struct {
	types []reflect.Type
}

func (v *customTypeValidator) Fn() playgroundvalidator.CustomTypeFunc {
	return func(field reflect.Value) interface{} {
		if field.Kind() == reflect.String {
			return field.String()
		}
		return field.Interface()
	}
}

func (v *customTypeValidator) Types() []reflect.Type {
	return v.types
}

factory := validator.NewDefaultValidatorFactory()
v, _ := factory.Create(
	nil,
	nil,
	nil,
	nil,
	[]validator.CustomTypeDefinition{&customTypeValidator{
		types: []reflect.Type{reflect.TypeOf(MyString(""))},
	}},
)
```

### Validator Aliases

Create aliases for common validation tag combinations:

```go
type notEmptyAlias struct{}

func (a *notEmptyAlias) Alias() string {
	return "notempty"
}

func (a *notEmptyAlias) Tags() string {
	return "required"
}

factory := validator.NewDefaultValidatorFactory()
v, _ := factory.Create(
	[]validator.AliasDefinition{&notEmptyAlias{}},
	nil,
	nil,
	nil,
	nil,
)

type User struct {
	Name string `validate:"notempty"` // Equivalent to "required"
}
```

## Translation Support

### Creating a Translator

```go
import (
	"github.com/talav/talav/pkg/component/validator"
	ut "github.com/go-playground/universal-translator"
	playgroundvalidator "github.com/go-playground/validator/v10"
)

factory := validator.NewDefaultTranslatorFactory()
trans, err := factory.Create()
if err != nil {
	panic(err)
}
```

### Registering Translations

```go
type uppercaseTranslation struct {
	tag string
}

func (t *uppercaseTranslation) Tag() string {
	return t.tag
}

func (t *uppercaseTranslation) Register(v *playgroundvalidator.Validate, trans ut.Translator) error {
	return v.RegisterTranslation(
		t.tag,
		trans,
		func(ut ut.Translator) error {
			return ut.Add(t.tag, "{0} must be uppercase", true)
		},
		func(ut ut.Translator, fe playgroundvalidator.FieldError) string {
			translated, _ := ut.T(t.tag, fe.Field())
			return translated
		},
	)
}

// Register translations
validatorFactory := validator.NewDefaultValidatorFactory()
v, _ := validatorFactory.Create(
	nil,
	[]validator.ValidationDefinition{&uppercaseValidator{tag: "uppercase"}},
	nil,
	nil,
	nil,
)

translatorFactory := validator.NewDefaultTranslatorFactory()
trans, _ := translatorFactory.Create()

err := validator.RegisterTranslations(
	v,
	trans,
	[]validator.TranslationDefinition{&uppercaseTranslation{tag: "uppercase"}},
)
if err != nil {
	panic(err)
}

// Use translator
type Code struct {
	Value string `validate:"uppercase" json:"value"`
}

err = v.Struct(Code{Value: "abc"})
if err != nil {
	validationErrors := err.(playgroundvalidator.ValidationErrors)
	for _, e := range validationErrors {
		fmt.Println(e.Translate(trans)) // "value must be uppercase"
	}
}
```

## JSON Tag Integration

The validator automatically uses JSON tag names in error messages when available:

```go
type User struct {
	EmailAddress string `json:"email" validate:"required"`
	UserName     string `json:"user_name" validate:"required"`
	Password     string `validate:"required"` // no json tag
}

err := v.Struct(User{})
// Error fields will be: "email", "user_name", "Password"
```

Fields without JSON tags fall back to struct field names.

## API Reference

### ValidatorFactory

```go
type ValidatorFactory interface {
	Create(
		aliasDefs []AliasDefinition,
		validationDefs []ValidationDefinition,
		validationDefsCtx []ValidationDefinitionCtx,
		structValidationDefs []StructValidationDefinition,
		customTypeDefs []CustomTypeDefinition,
	) (*validator.Validate, error)
}
```

Interface for creating validators.

### DefaultValidatorFactory

```go
func NewDefaultValidatorFactory() ValidatorFactory
```

Returns a new `DefaultValidatorFactory` instance.

### TranslatorFactory

```go
type TranslatorFactory interface {
	Create() (ut.Translator, error)
}
```

Interface for creating translators.

### DefaultTranslatorFactory

```go
func NewDefaultTranslatorFactory() TranslatorFactory
```

Returns a new `DefaultTranslatorFactory` instance.

### RegisterTranslations

```go
func RegisterTranslations(
	v *validator.Validate,
	trans ut.Translator,
	definitions []TranslationDefinition,
) error
```

Registers default and custom validator translations.

### Definition Interfaces

#### AliasDefinition

```go
type AliasDefinition interface {
	Alias() string
	Tags() string
}
```

Defines a validator alias.

#### ValidationDefinition

```go
type ValidationDefinition interface {
	Tag() string
	Fn() v.Func
	CallEvenIfNull() bool
}
```

Defines a field-level validator without context.

#### ValidationDefinitionCtx

```go
type ValidationDefinitionCtx interface {
	Tag() string
	Fn() v.FuncCtx
	CallEvenIfNull() bool
}
```

Defines a field-level validator with context.

#### StructValidationDefinition

```go
type StructValidationDefinition interface {
	Fn() v.StructLevelFuncCtx
	Types() []reflect.Type
}
```

Defines a struct-level validator.

#### CustomTypeDefinition

```go
type CustomTypeDefinition interface {
	Fn() v.CustomTypeFunc
	Types() []reflect.Type
}
```

Defines a custom type validator.

#### TranslationDefinition

```go
type TranslationDefinition interface {
	Tag() string
	Register(validator *v.Validate, trans ut.Translator) error
}
```

Defines a validator translation.

## Error Handling

Validation errors are returned as `validator.ValidationErrors`:

```go
err := v.Struct(user)
if err != nil {
	var validationErrors playgroundvalidator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, ve := range validationErrors {
			fmt.Printf("Field: %s, Tag: %s, Error: %s\n",
				ve.Field(),
				ve.Tag(),
				ve.Error(),
			)
		}
	}
}
```

## Best Practices

### 1. Use JSON Tags

Always use JSON tags for better error messages:

```go
type User struct {
	Email string `json:"email" validate:"required,email"`
}
```

### 2. Provide Translations

Always provide translations for custom validators:

```go
factory.Create(
	nil,
	[]validator.ValidationDefinition{&uppercaseValidator{tag: "uppercase"}},
	nil,
	nil,
	nil,
)

validator.RegisterTranslations(
	v,
	trans,
	[]validator.TranslationDefinition{&uppercaseTranslation{tag: "uppercase"}},
)
```

### 3. Use Context Validators When Needed

Use context validators when you need access to request context or other contextual data:

```go
// Use when you need context
validationDefsCtx := []validator.ValidationDefinitionCtx{&contextAwareValidator{}}

// Use when context is not needed
validationDefs := []validator.ValidationDefinition{&simpleValidator{}}
```

### 4. Group Related Validators

Create helper functions or structs to group related validators:

```go
func NewUserValidators() []validator.ValidationDefinition {
	return []validator.ValidationDefinition{
		&emailValidator{tag: "email_format"},
		&usernameValidator{tag: "username_format"},
	}
}
```

## Dependencies

- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) - Core validation library
- [github.com/go-playground/universal-translator](https://github.com/go-playground/universal-translator) - Translation support
- [github.com/go-playground/locales](https://github.com/go-playground/locales) - Locale support

## See Also

- [Fx Validator Module](../../fx/fxvalidator/README.md) - Fx integration for validators
- [Go Playground Validator](https://github.com/go-playground/validator) - Underlying validation library
- [Universal Translator](https://github.com/go-playground/universal-translator) - Translation library

