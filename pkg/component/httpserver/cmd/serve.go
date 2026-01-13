package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/httpserver"
)

// NewServeHTTPCmd creates the serve-http command.
// The command manages the HTTP server lifecycle directly.
// It's a self-sufficient component command with no framework dependencies.
func NewServeHTTPCmd(
	server *httpserver.Server,
	logger *slog.Logger,
) *cobra.Command {
	return &cobra.Command{
		Use:   "serve-http",
		Short: "Start the HTTP server",
		Long:  "Start the HTTP server",
		Example: `  myapp serve-http
  APP_ENV=prod myapp serve-http`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("starting HTTP server...")

			if err := server.Start(cmd.Context()); err != nil {
				return err
			}

			logger.Info("HTTP server shutdown complete")

			return nil
		},
	}
}
