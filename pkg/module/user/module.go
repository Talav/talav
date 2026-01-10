package user

import (
	"github.com/talav/talav/pkg/component/zorya"
	"github.com/talav/talav/pkg/module/user/handler"
	"go.uber.org/fx"
)

const ModuleName = "user"

// Module provides HTTP handlers and routes for the user module.
// It depends on fxuser.Module which provides service handlers.
//
// Dependencies:
//   - fxuser.Module (for service handlers)
//   - zorya.API (for route registration)
//
// Usage:
//
//	fx.New(
//	    fxuser.Module,
//	    userhttp.Module,
//	    fx.Invoke(userhttp.RegisterRoutesInvoker),
//	    // ... application modules
//	)
var Module = fx.Module(
	ModuleName,
	fx.Provide(
		handler.NewCreateUserHandler,
		handler.NewGetUserHandler,
		handler.NewListUsersHandler,
	),
)

// RegisterRoutesInvoker registers user routes with the Zorya API.
// This should be invoked after all modules are loaded.
func RegisterRoutesInvoker(
	api zorya.API,
	createUserHandler *handler.CreateUserHandler,
	getUserHandler *handler.GetUserHandler,
	listUsersHandler *handler.ListUsersHandler,
) {
	RegisterRoutes(api, createUserHandler, getUserHandler, listUsersHandler)
}
