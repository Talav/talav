package seeder

import (
	"context"
	"fmt"
	"slices"
)

// SeederRegistry holds all registered seeders filtered by environment.
type SeederRegistry struct {
	seeders []Seeder
}

// NewSeederRegistry creates a new seeder registry with seeders filtered by environment.
// Only seeders that should run in the current environment are included.
// A seeder is included if:
//   - Its Environments() returns an empty slice (runs in all environments), or
//   - Its Environments() contains the currentEnv string
//
// Seeders are sorted so that seeders running in all environments (base seeders) run first.
func NewSeederRegistry(seeders []Seeder, currentEnv string) *SeederRegistry {
	baseSeeders := make([]Seeder, 0)
	envSeeders := make([]Seeder, 0)

	for _, s := range seeders {
		envs := s.Environments()
		// Empty slice means all environments - these run first
		if len(envs) == 0 {
			baseSeeders = append(baseSeeders, s)

			continue
		}
		// Check if current environment is in the list
		if slices.Contains(envs, currentEnv) {
			envSeeders = append(envSeeders, s)
		}
	}

	// Combine: base seeders first, then environment-specific seeders
	return &SeederRegistry{
		seeders: append(baseSeeders, envSeeders...),
	}
}

// SeedAll executes all registered seeders in order.
func (r *SeederRegistry) SeedAll(ctx context.Context) error {
	for _, s := range r.seeders {
		if err := s.Seed(ctx); err != nil {
			return fmt.Errorf("seeding failed: %w", err)
		}
	}

	return nil
}
