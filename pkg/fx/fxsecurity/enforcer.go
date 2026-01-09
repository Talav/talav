package fxsecurity

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/talav/talav/pkg/component/security"
)

// CasbinEnforcer is a Casbin-based implementation of SecurityEnforcer.
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewCasbinEnforcer creates a new CasbinEnforcer.
func NewCasbinEnforcer(e *casbin.Enforcer) *CasbinEnforcer {
	return &CasbinEnforcer{enforcer: e}
}

// Enforce checks if the user meets all security requirements.
func (c *CasbinEnforcer) Enforce(ctx context.Context, user *security.AuthUser, requirements *security.SecurityRequirements) (bool, error) {
	if user == nil {
		return false, nil
	}

	return c.checkAllRequirements(user, requirements)
}

// checkAllRequirements checks all security requirements sequentially.
// If resource check passes, returns early without checking permissions/roles.
// If permissions check passes, returns early without checking roles.
func (c *CasbinEnforcer) checkAllRequirements(user *security.AuthUser, requirements *security.SecurityRequirements) (bool, error) {
	if requirements.Resource != "" {
		allowed, err := c.checkResourceAccess(user, requirements.Resource, requirements.Action)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	if len(requirements.Permissions) > 0 {
		allowed, err := c.checkPermissions(user, requirements.Permissions)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	if len(requirements.Roles) > 0 {
		if hasRole, err := c.checkRoles(user, requirements.Roles); err != nil || !hasRole {
			return false, err
		}
	}

	return true, nil
}

// checkResourceAccess checks if the user has access to a resource via any of their roles.
func (c *CasbinEnforcer) checkResourceAccess(user *security.AuthUser, resource, action string) (bool, error) {
	for _, role := range user.Roles {
		ok, err := c.enforcer.Enforce(user.ID, role, resource, action)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

// checkPermissions checks if the user has all required permissions.
func (c *CasbinEnforcer) checkPermissions(user *security.AuthUser, permissions []string) (bool, error) {
	for _, perm := range permissions {
		ok, err := c.enforcer.Enforce(user.ID, perm)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// checkRoles checks if the user has at least one of the required roles.
func (c *CasbinEnforcer) checkRoles(user *security.AuthUser, requiredRoles []string) (bool, error) {
	for _, required := range requiredRoles {
		ok, err := c.enforcer.HasRoleForUser(user.ID, required)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}
