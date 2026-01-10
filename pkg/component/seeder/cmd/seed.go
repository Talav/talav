package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/talav/talav/pkg/component/seeder"
)

// NewSeedCmd creates a new seed command.
func NewSeedCmd(registry *seeder.SeederRegistry) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Run database seeders",
		Long:  `Execute all registered seeders to populate the database with initial data.`,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := context.Background()

			if err := registry.SeedAll(ctx); err != nil {
				return fmt.Errorf("seeding failed: %w", err)
			}

			return nil
		},
	}

	return cmd
}
