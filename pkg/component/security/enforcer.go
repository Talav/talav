package security

import (
	"context"
	"fmt"
	"slices"
)

// SecurityEnforcer checks if a user meets all security requirements.
// This is the core authorization interface - framework-agnostic.
type SecurityEnforcer interface {
	// Enforce checks if the user meets all security requirements.
	// Returns true if all requirements are met, false otherwise.
	Enforce(ctx context.Context, user *AuthUser, requirements *SecurityRequirements) (bool, error)
}

// SecurityRequirements represents generic authorization requirements.
// This is the input type for SecurityEnforcer - it's framework-agnostic
// and contains only the essential data needed for authorization decisions.
// Authentication is implied if any of Roles, Permissions, or Resource are set.
type SecurityRequirements struct {
	Roles       []string // Required roles (user needs at least one)
	Permissions []string // Required permissions (user needs all)
	Resource    string   // Resolved resource identifier (e.g., "organizations/123")
	Action      string   // Action being performed (e.g., "view", "edit", "POST")
}

// SimpleEnforcer checks roles stored in the AuthUser context.
// No external policy storage needed.
type SimpleEnforcer struct{}

// NewSimpleEnforcer creates a new SimpleEnforcer.
func NewSimpleEnforcer() *SimpleEnforcer {
	return &SimpleEnforcer{}
}

// Enforce checks if the user meets all security requirements.
// SimpleEnforcer only supports role-based authorization.
// Returns an error if permissions or resources are specified (use a custom enforcer for those).
//
// Note: requirements must have at least one requirement (roles, permissions, or resource).
// Empty requirements should not reach the enforcer (enforced by Secure() validation).
func (s *SimpleEnforcer) Enforce(ctx context.Context, user *AuthUser, requirements *SecurityRequirements) (bool, error) {
	if user == nil {
		return false, nil
	}

	// SimpleEnforcer only supports roles
	if len(requirements.Permissions) > 0 {
		return false, fmt.Errorf("SimpleEnforcer does not support permissions - use a custom enforcer")
	}

	if requirements.Resource != "" {
		return false, fmt.Errorf("SimpleEnforcer does not support resource-based RBAC - use a custom enforcer")
	}

	// Check roles (user needs at least one)
	// Note: len(requirements.Roles) == 0 should not happen due to Secure() validation,
	// but we handle it defensively
	if len(requirements.Roles) == 0 {
		return false, fmt.Errorf("SimpleEnforcer: no security requirements specified")
	}

	return s.checkRoles(user, requirements.Roles), nil
}

// checkRoles checks if the user has at least one of the required roles.
func (s *SimpleEnforcer) checkRoles(user *AuthUser, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		if slices.Contains(user.Roles, required) {
			return true
		}
	}

	return false
}
