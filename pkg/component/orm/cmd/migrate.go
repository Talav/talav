package cmd

import (
	"github.com/spf13/cobra"
)

// NewMigrateCmd creates a new migrate command
func NewMigrateCmd(
	createCmd *cobra.Command,
	upCmd *cobra.Command,
	downCmd *cobra.Command,
	resetCmd *cobra.Command,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  `Manage database schema migrations.`,
	}

	cmd.AddCommand(createCmd)
	cmd.AddCommand(upCmd)
	cmd.AddCommand(downCmd)
	cmd.AddCommand(resetCmd)

	return cmd
}
