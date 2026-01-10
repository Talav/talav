package fxemail

import (
	"log/slog"

	"github.com/talav/talav/pkg/component/email"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
)

const ModuleName = "email"

// FxEmailModule is the [Fx] email module.
var FxEmailModule = fx.Module(
	ModuleName,
	fxconfig.AsConfigWithDefaults("email", email.DefaultEmailConfig(), email.EmailConfig{}),
	fx.Provide(NewFxEmailService),
)

// NewFxEmailService returns a new [email.EmailService].
func NewFxEmailService(cfg email.EmailConfig, logger *slog.Logger) (*email.EmailService, error) {
	return email.NewEmailServiceFromConfig(cfg, logger)
}
