package orm

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
)

// UniqueValidator validates uniqueness of values in database.
type UniqueValidator struct {
	registry *RepositoryRegistry
}

// NewUniqueValidator creates a new unique validator with injected dependencies.
func NewUniqueValidator(registry *RepositoryRegistry) *UniqueValidator {
	return &UniqueValidator{
		registry: registry,
	}
}

// Validate checks if a value is unique in the database
// Tag format: unique=EntityName.fieldName
// Example: unique=User.email.
func (v *UniqueValidator) Validate(ctx context.Context, fl validator.FieldLevel) bool {
	// Parse the tag parameter (e.g., "User.email")
	param := fl.Param()
	if param == "" {
		return false // Invalid tag format
	}

	// Split into entity and field
	parts := strings.Split(param, ".")
	if len(parts) != 2 {
		return false // Invalid format
	}

	entityName := strings.ToLower(parts[0]) // Normalize to lowercase
	fieldName := parts[1]

	// Get the field value
	fieldValue := fl.Field().Interface()
	if fieldValue == "" || fieldValue == nil {
		return true // Empty values are handled by required validator
	}

	// Get repository from registry
	checker, err := v.registry.GetExistsChecker(entityName)
	if err != nil {
		return false // Repository not found
	}

	exists, err := checker.Exists(ctx, map[string]any{
		fieldName: fieldValue,
	})
	if err != nil {
		return false // Database error
	}

	// Return true if unique (doesn't exist), false if duplicate (exists)
	return !exists
}
