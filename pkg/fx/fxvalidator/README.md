# Fx Validator Module

> [Fx](https://uber-go.github.io/fx/) module for [validator](../component/validator).

## Overview

The `fxvalidator` module provides seamless integration between the Talav validation system and [Uber's Fx dependency injection framework](https://github.com/uber-go/fx). It handles validator and translator initialization during application startup and makes them available throughout your application via dependency injection.

## Features

- **Automatic Validator Initialization**: Creates and configures validators during Fx application startup
- **Dependency Injection**: Makes `*validator.Validate` and `ut.Translator` available for injection
- **Custom Validators**: Register field-level, struct-level, and custom type validators using helper functions
- **Validator Aliases**: Create aliases for common validation tag combinations
- **Translation Support**: Register custom error message translations
- **Factory Override**: Support for custom validator and translator factories via `fx.Decorate()`
- **Error Handling**: Fails fast on validator registration errors during application construction

## Installation

```bash
go get github.com/talav/talav/pkg/fx/fxvalidator
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"

	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
)

type User struct {
	Email string `validate:"required,email"`
	Name  string `validate:"required,min=3"`
}

func main() {
	fx.New(
		fxvalidator.FxValidatorModule,              // Load the validator module
		fx.Invoke(func(v *playgroundvalidator.Validate) {
			user := User{
				Email: "invalid-email",
				Name:  "ab",
			}
			
			err := v.Struct(user)
			if err != nil {
				fmt.Printf("Validation errors: %v\n", err)
			}
		}),
	).Run()
}
```

## Dependencies

The `FxValidatorModule` has no external dependencies. It can be used standalone or alongside other Fx modules:

```go
fx.New(
	fxvalidator.FxValidatorModule, // Can be used standalone
	// ...
)
```

## Registering Custom Validators

### Field-Level Validators (Non-Context)

Register a simple field validator without context:

```go
package main

import (
	"strings"

	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
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

func main() {
	fx.New(
		fxvalidator.FxValidatorModule,
		fxvalidator.AsValidator(&uppercaseValidator{tag: "uppercase"}),
		fx.Invoke(func(v *playgroundvalidator.Validate) {
			type Code struct {
				Value string `validate:"uppercase"`
			}
			
			err := v.Struct(Code{Value: "ABC"}) // Valid
			fmt.Println(err) // nil
			
			err = v.Struct(Code{Value: "abc"}) // Invalid
			fmt.Println(err) // validation error
		}),
	).Run()
}
```

### Field-Level Validators (With Context)

Register a field validator that receives context:

```go
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

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidatorCtx(&uppercaseValidatorCtx{tag: "uppercase_ctx"}),
	// ...
)
```

### Validator Constructors

Use constructors when validators need dependency injection:

```go
type configurableValidator struct {
	tag    string
	minLen int
}

func (v *configurableValidator) Tag() string {
	return v.tag
}

func (v *configurableValidator) Fn() playgroundvalidator.Func {
	return func(fl playgroundvalidator.FieldLevel) bool {
		val := fl.Field().String()
		return len(val) >= v.minLen
	}
}

func (v *configurableValidator) CallEvenIfNull() bool {
	return false
}

func NewConfigurableValidator() validator.ValidationDefinition {
	// Can inject dependencies via Fx
	return &configurableValidator{
		tag:    "minlen",
		minLen: 5, // default or from injected config
	}
}

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidatorConstructor(NewConfigurableValidator),
	// ...
)
```

### Struct-Level Validators

Register validators that validate entire structs:

```go
import (
	"context"
	"reflect"

	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
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

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsStructValidator(&passwordMatchValidator{}),
	// ...
)
```

### Custom Type Validators

Register validators for specific types:

```go
type customTypeValidator struct {
	types []reflect.Type
}

func (v *customTypeValidator) Fn() playgroundvalidator.CustomTypeFunc {
	return func(fl playgroundvalidator.FieldLevel) bool {
		// Custom validation logic for the type
		return true
	}
}

func (v *customTypeValidator) Types() []reflect.Type {
	return v.types
}

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsCustomTypeValidator(&customTypeValidator{
		types: []reflect.Type{reflect.TypeOf(MyCustomType{})},
	}),
	// ...
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

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidatorAlias(&notEmptyAlias{}),
	fx.Invoke(func(v *playgroundvalidator.Validate) {
		type User struct {
			Name string `validate:"notempty"` // Equivalent to "required"
		}
		// ...
	}),
)
```

## Translation Support

### Registering Translations

Register custom error message translations:

```go
import (
	ut "github.com/go-playground/universal-translator"
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/validator"
)

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
			t, _ := ut.T(t.tag, fe.Field())
			return t
		},
	)
}

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidator(&uppercaseValidator{tag: "uppercase"}),
	fxvalidator.AsTranslation(&uppercaseTranslation{tag: "uppercase"}),
	fx.Invoke(func(v *playgroundvalidator.Validate, trans ut.Translator) {
		type Code struct {
			Value string `validate:"uppercase"`
		}
		
		err := v.Struct(Code{Value: "abc"})
		if err != nil {
			errs := err.(playgroundvalidator.ValidationErrors)
			for _, e := range errs {
				fmt.Println(e.Translate(trans)) // "Value must be uppercase"
			}
		}
	}),
)
```

### Translation Constructors

Use constructors when translations need dependency injection:

```go
func NewUppercaseTranslation(cfg *config.Config) validator.TranslationDefinition {
	// Translation can be configured from config
	return &uppercaseTranslation{tag: "uppercase"}
}

fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsTranslationConstructor(NewUppercaseTranslation),
	// ...
)
```

## Module Structure

### Provided Dependencies

The module provides the following for injection:

- `*validator.Validate` - Main validator instance (from `github.com/go-playground/validator/v10`)
- `ut.Translator` - Translator instance for error messages (from `github.com/go-playground/universal-translator`)
- `validator.ValidatorFactory` - Validator factory (can be decorated)
- `validator.TranslatorFactory` - Translator factory (can be decorated)

### Required Dependencies

The module automatically collects:

- `[]validator.AliasDefinition` (group: `"talav-validator-aliases"`) - Validator aliases
- `[]validator.ValidationDefinition` (group: `"talav-validator-validations"`) - Field validators (non-context)
- `[]validator.ValidationDefinitionCtx` (group: `"talav-validator-validations-ctx"`) - Field validators (with context)
- `[]validator.StructValidationDefinition` (group: `"talav-validator-struct-validations"`) - Struct validators
- `[]validator.CustomTypeDefinition` (group: `"talav-validator-custom-types"`) - Custom type validators
- `[]validator.TranslationDefinition` (group: `"talav-validator-translations"`) - Translation definitions

### Module Name

The module is registered with the name `"validator"`.

## API Reference

### FxValidatorModule

```go
var FxValidatorModule = fx.Module(
	"validator",
	fx.Provide(
		validator.NewDefaultValidatorFactory,
		validator.NewDefaultTranslatorFactory,
		NewFxValidator,
		NewFxTranslator,
	),
)
```

The main Fx module. Include this in your `fx.New()` call.

### AsValidatorAlias

```go
func AsValidatorAlias(alias validator.AliasDefinition, annotations ...fx.Annotation) fx.Option
```

Registers a validator alias definition.

**Parameters:**
- `alias` - Alias definition implementing `validator.AliasDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsValidatorAlias(&notEmptyAlias{})
```

### AsValidator

```go
func AsValidator(def validator.ValidationDefinition, annotations ...fx.Annotation) fx.Option
```

Registers a field-level validation definition (non-context).

**Parameters:**
- `def` - Validation definition implementing `validator.ValidationDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsValidator(&uppercaseValidator{tag: "uppercase"})
```

### AsValidatorCtx

```go
func AsValidatorCtx(def validator.ValidationDefinitionCtx, annotations ...fx.Annotation) fx.Option
```

Registers a field-level validation definition (with context).

**Parameters:**
- `def` - Validation definition implementing `validator.ValidationDefinitionCtx`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsValidatorCtx(&uppercaseValidatorCtx{tag: "uppercase_ctx"})
```

### AsValidatorConstructor

```go
func AsValidatorConstructor(constructor any, annotations ...fx.Annotation) fx.Option
```

Registers a validation definition constructor (non-context). The constructor will be called by Fx with dependency injection.

**Parameters:**
- `constructor` - Constructor function returning `validator.ValidationDefinition`
- `annotations` - Optional Fx annotations (e.g., `fx.ParamTags`)

**Example:**
```go
fxvalidator.AsValidatorConstructor(NewConfigurableValidator)
```

### AsValidatorConstructorCtx

```go
func AsValidatorConstructorCtx(constructor any, annotations ...fx.Annotation) fx.Option
```

Registers a validation definition constructor (with context). The constructor will be called by Fx with dependency injection.

**Parameters:**
- `constructor` - Constructor function returning `validator.ValidationDefinitionCtx`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsValidatorConstructorCtx(NewConfigurableValidatorCtx)
```

### AsStructValidator

```go
func AsStructValidator(def validator.StructValidationDefinition, annotations ...fx.Annotation) fx.Option
```

Registers a struct-level validation definition.

**Parameters:**
- `def` - Struct validation definition implementing `validator.StructValidationDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsStructValidator(&passwordMatchValidator{})
```

### AsCustomTypeValidator

```go
func AsCustomTypeValidator(def validator.CustomTypeDefinition, annotations ...fx.Annotation) fx.Option
```

Registers a custom type validation definition.

**Parameters:**
- `def` - Custom type definition implementing `validator.CustomTypeDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsCustomTypeValidator(&customTypeValidator{types: [...]})
```

### AsTranslation

```go
func AsTranslation(def validator.TranslationDefinition, annotations ...fx.Annotation) fx.Option
```

Registers a translation definition.

**Parameters:**
- `def` - Translation definition implementing `validator.TranslationDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsTranslation(&uppercaseTranslation{tag: "uppercase"})
```

### AsTranslationConstructor

```go
func AsTranslationConstructor(constructor any, annotations ...fx.Annotation) fx.Option
```

Registers a translation definition constructor. The constructor will be called by Fx with dependency injection.

**Parameters:**
- `constructor` - Constructor function returning `validator.TranslationDefinition`
- `annotations` - Optional Fx annotations

**Example:**
```go
fxvalidator.AsTranslationConstructor(NewUppercaseTranslation)
```

## Custom Factories

Override the default factories with custom implementations:

```go
package main

import (
	"github.com/talav/talav/pkg/component/validator"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
)

type CustomValidatorFactory struct{}

func NewCustomValidatorFactory() validator.ValidatorFactory {
	return &CustomValidatorFactory{}
}

func (f *CustomValidatorFactory) Create(
	aliasDefs []validator.AliasDefinition,
	validationDefs []validator.ValidationDefinition,
	validationDefsCtx []validator.ValidationDefinitionCtx,
	structValidationDefs []validator.StructValidationDefinition,
	customTypeDefs []validator.CustomTypeDefinition,
) (*validator.Validate, error) {
	// Custom validator creation logic
	return validator.New(), nil
}

func main() {
	fx.New(
		fx.Decorate(NewCustomValidatorFactory),     // Override the factory
		fxvalidator.FxValidatorModule,
		fx.Invoke(func(v *validator.Validate) {
			// Uses custom factory
		}),
	).Run()
}
```

## Error Handling

Validation errors during module construction cause the application to fail:

```go
fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidator(&invalidValidator{}), // Invalid validator
).Run()
// Application will fail to start with validator registration error
```

This fail-fast behavior ensures validator issues are caught early rather than causing runtime errors.

## Best Practices

### 1. Use Constructors for Configurable Validators

When validators need configuration or other dependencies, use constructors:

```go
// Good
fxvalidator.AsValidatorConstructor(NewConfigurableValidator)

// Less ideal (no access to config/dependencies)
fxvalidator.AsValidator(&staticValidator{})
```

### 2. Group Related Validators

Create a module for related validators:

```go
package user

import (
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
)

var UserValidatorsModule = fx.Module(
	"user-validators",
	fxvalidator.AsValidator(&emailValidator{}),
	fxvalidator.AsValidator(&usernameValidator{}),
	fxvalidator.AsStructValidator(&passwordMatchValidator{}),
)
```

### 3. Provide Translations

Always provide translations for custom validators:

```go
fx.New(
	fxvalidator.FxValidatorModule,
	fxvalidator.AsValidator(&uppercaseValidator{tag: "uppercase"}),
	fxvalidator.AsTranslation(&uppercaseTranslation{tag: "uppercase"}), // Provide translation
	// ...
)
```

### 4. Use Context Validators When Needed

Use context validators when you need access to request context or other contextual data:

```go
// Use when you need context
fxvalidator.AsValidatorCtx(&contextAwareValidator{})

// Use when context is not needed
fxvalidator.AsValidator(&simpleValidator{})
```

## Testing

When testing components that depend on validators:

```go
import (
	"testing"

	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestMyComponent(t *testing.T) {
	var v *playgroundvalidator.Validate
	
	fxtest.New(
		t,
		fx.NopLogger,
		fxvalidator.FxValidatorModule,
		fx.Populate(&v),
	).RequireStart().RequireStop()
	
	// Test with injected validator
	if v == nil {
		t.Fatal("validator not injected")
	}
}
```

## Dependencies

- [go.uber.org/fx](https://github.com/uber-go/fx) v1.24.0+
- [github.com/talav/talav/pkg/component/validator](../component/validator)
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator)
- [github.com/go-playground/universal-translator](https://github.com/go-playground/universal-translator)

## See Also

- [Validator Package Documentation](../component/validator/README.md) - Detailed validator system documentation
- [Fx Documentation](https://uber-go.github.io/fx/) - Uber's Fx dependency injection framework
- [Go Playground Validator](https://github.com/go-playground/validator) - Underlying validation library

