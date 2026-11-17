package user

import (
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/user/handler"
)

const (
	// tagUsers groups the user routes in the OpenAPI document.
	tagUsers = "Users"

	// roleAdmin is the role allowed to manage any user.
	roleAdmin = "admin"

	// roleUser is the role of a regular signed-in user.
	roleUser = "user"
)

// RegisterRoutes registers all user-related routes.
func RegisterRoutes(
	api zorya.API,
	createUserHandler *handler.CreateUserHandler,
	getUserHandler *handler.GetUserHandler,
	listUsersHandler *handler.ListUsersHandler,
) {
	// User group
	userGroup := zorya.NewGroup(api, "/users")

	// POST /users - Create user (public)
	zorya.Post(userGroup, "", createUserHandler.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Create User",
				Description: "Create a new user account",
				Tags:        []string{tagUsers},
				OperationID: "createUser",
			}
		},
	)

	// GET /users - List users (admin only)
	zorya.Get(userGroup, "", listUsersHandler.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "List Users",
				Description: "Retrieve a paginated list of users with optional filtering",
				Tags:        []string{tagUsers},
				OperationID: "listUsers",
			}
		},
		zorya.Secure(
			func(s *zorya.RouteSecurity) {
				s.Roles = []string{roleAdmin}
			},
		),
	)

	// GET /users/me - Get current user (authenticated)
	zorya.Get(userGroup, "/me", getUserHandler.HandleCurrentUser,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Get Current User",
				Description: "Retrieve the authenticated user's profile information",
				Tags:        []string{tagUsers},
				OperationID: "getCurrentUser",
			}
		},
		zorya.Secure(
			func(s *zorya.RouteSecurity) {
				s.Roles = []string{roleUser, roleAdmin}
			},
		),
	)

	// GET /users/{id} - Get user by ID (owner or admin)
	zorya.Get(userGroup, "/{id}", getUserHandler.Handle,
		func(r *zorya.BaseRoute) {
			r.Operation = &zorya.Operation{
				Summary:     "Get User by ID",
				Description: "Retrieve a single user by their unique ID",
				Tags:        []string{tagUsers},
				OperationID: "getUser",
			}
		},
		zorya.Secure(
			func(s *zorya.RouteSecurity) {
				s.Roles = []string{roleUser, roleAdmin}
				s.Resource = "users/{id}"
				s.Action = "read"
			},
		),
	)
}
