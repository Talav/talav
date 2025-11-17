package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/orm"
)

// NewMigrateCreateCmd creates the migrate create command
func NewMigrateCreateCmd(migration *orm.Migration, logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration",
		Long:  `Generate timestamp-based SQL migration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := migration.Create(args[0]); err != nil {
				logger.Error("Failed to create migration", "error", err, "name", args[0])
				return err
			}

			logger.Info("Migration created successfully", "name", args[0])
			return nil
		},
	}
}
