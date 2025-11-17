package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/orm"
)

// NewMigrateResetCmd creates the migrate reset command
func NewMigrateResetCmd(migration *orm.Migration, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset database",
		Long:  `Drop all tables and schema. WARNING: This will delete all data!`,
		RunE: func(c *cobra.Command, args []string) error {
			logger.Warn("Resetting database - dropping all tables...")
			if err := migration.Reset(); err != nil {
				logger.Error("Failed to reset database", "error", err)
				return err
			}

			logger.Info("Database reset successfully")
			return nil
		},
	}
}
