package seeder

import "context"

// Seeder is the interface that all seeders must implement.
type Seeder interface {
	Seed(ctx context.Context) error
	// Environments returns the list of environments this seeder should run in.
	// Return empty slice []string{} to run in all environments.
	// Example: []string{"dev", "test"} means run in dev and test only.
	Environments() []string
}
