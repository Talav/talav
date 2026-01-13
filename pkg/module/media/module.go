package media

import (
	"github.com/talav/talav/pkg/module/media/handler"
	"go.uber.org/fx"
)

const ModuleName = "media-module"

// FxMediaHTTPModule provides HTTP handlers and routes for the media module.
// It depends on fxmedia.FxMediaModule which provides commands and queries.
var FxMediaHTTPModule = fx.Module(
	ModuleName,
	fx.Provide(
		handler.NewCreateMediaHandler,
		handler.NewGetMediaHandler,
		handler.NewListMediaHandler,
		handler.NewUpdateMediaHandler,
		handler.NewDeleteMediaHandler,
	),
	fx.Invoke(RegisterRoutes),
)
