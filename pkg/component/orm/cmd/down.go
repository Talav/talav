package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/orm"
)

// NewMigrateDownCmd creates the migrate down command
func NewMigrateDownCmd(migration *orm.Migration, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Rollback last migration",
		Long:  `Rollback the most recently applied migration.`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := migration.Down(); err != nil {
				logger.Error("Failed to rollback migration", "error", err)
				return err
			}

			logger.Info("Migration rolled back successfully")
			return nil
		},
	}
}
