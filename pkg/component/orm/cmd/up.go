package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/orm"
)

// NewMigrateUpCmd creates the migrate up command
func NewMigrateUpCmd(migration *orm.Migration, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run pending migrations",
		Long:  `Apply all pending migrations to update the database schema.`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := migration.Up(); err != nil {
				logger.Error("Failed to apply migrations", "error", err)
				return err
			}

			logger.Info("Migrations applied successfully")
			return nil
		},
	}
}
