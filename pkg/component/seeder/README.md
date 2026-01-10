# Seeder Component

A framework-agnostic database seeding component with environment-aware filtering. Allows defining seeders that run in specific environments (dev, test, prod) or all environments.

## Features

- **Environment Filtering**: Seeders can specify which environments they should run in
- **Ordered Execution**: Base seeders (all environments) run first, then environment-specific seeders
- **CLI Integration**: Includes a Cobra command for running seeders
- **Framework Agnostic**: Core component has no dependency injection dependencies

## Quick Start

### Define a Seeder

```go
package myapp

import (
    "context"
    "github.com/talav/talav/pkg/component/seeder"
)

type UserSeeder struct {
    userRepo UserRepository
}

func NewUserSeeder(userRepo UserRepository) seeder.Seeder {
    return &UserSeeder{userRepo: userRepo}
}

// Seed creates initial users
func (s *UserSeeder) Seed(ctx context.Context) error {
    // Create admin user
    admin := &User{
        Email: "admin@example.com",
        Role:  "admin",
    }
    return s.userRepo.Create(ctx, admin)
}

// Environments returns which environments this seeder runs in
// Empty slice = all environments
func (s *UserSeeder) Environments() []string {
    return []string{} // Run in all environments
}
```

### Environment-Specific Seeder

```go
type TestDataSeeder struct {
    db *gorm.DB
}

func NewTestDataSeeder(db *gorm.DB) seeder.Seeder {
    return &TestDataSeeder{db: db}
}

func (s *TestDataSeeder) Seed(ctx context.Context) error {
    // Seed test data
    return nil
}

// Only run in dev and test environments
func (s *TestDataSeeder) Environments() []string {
    return []string{"dev", "test"}
}
```

### Manual Usage

```go
package main

import (
    "context"
    "github.com/talav/talav/pkg/component/seeder"
)

func main() {
    // Create seeders
    seeders := []seeder.Seeder{
        NewUserSeeder(userRepo),
        NewTestDataSeeder(db),
    }
    
    // Create registry filtered by environment
    currentEnv := "dev" // From config or environment variable
    registry := seeder.NewSeederRegistry(seeders, currentEnv)
    
    // Run all seeders
    ctx := context.Background()
    if err := registry.SeedAll(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Seeder Interface

```go
type Seeder interface {
    // Seed executes the seeding logic
    Seed(ctx context.Context) error
    
    // Environments returns which environments this seeder runs in
    // Return empty slice []string{} to run in all environments
    // Example: []string{"dev", "test"} runs only in dev and test
    Environments() []string
}
```

## Environment Filtering

The `SeederRegistry` filters seeders based on the current environment:

1. **Base Seeders** (empty `Environments()`) - Run in ALL environments
   - Always executed first
   - Example: Creating required roles, default admin user

2. **Environment-Specific Seeders** - Run only in specified environments
   - Executed after base seeders
   - Example: Test data for dev/test, production data for prod

### Execution Order

```go
// Base seeders (all environments) run FIRST
AdminSeeder{}.Environments() => []string{}
RoleSeeder{}.Environments() => []string{}

// Then environment-specific seeders
TestDataSeeder{}.Environments() => []string{"dev", "test"}
ProdDataSeeder{}.Environments() => []string{"prod"}
```

## CLI Command

The component includes a Cobra command for CLI integration:

```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/talav/talav/pkg/component/seeder"
    "github.com/talav/talav/pkg/component/seeder/cmd"
)

func main() {
    // Create registry
    registry := seeder.NewSeederRegistry(seeders, currentEnv)
    
    // Create seed command
    seedCmd := cmd.NewSeedCmd(registry)
    
    // Add to root command
    rootCmd := &cobra.Command{Use: "myapp"}
    rootCmd.AddCommand(seedCmd)
    
    rootCmd.Execute()
}
```

Usage:
```bash
myapp seed
```

## Best Practices

### 1. Idempotent Seeders

Make seeders idempotent (safe to run multiple times):

```go
func (s *UserSeeder) Seed(ctx context.Context) error {
    // Check if admin exists
    if exists, _ := s.userRepo.ExistsByEmail(ctx, "admin@example.com"); exists {
        return nil // Already seeded
    }
    
    // Create admin
    return s.userRepo.Create(ctx, admin)
}
```

### 2. Environment Detection

Use `APP_ENV` environment variable:

```go
currentEnv := os.Getenv("APP_ENV")
if currentEnv == "" {
    currentEnv = "dev" // Default
}
```

### 3. Order Dependencies

Base seeders run first, so use them for foundational data:

```go
// Base seeder - creates roles (runs first)
type RoleSeeder struct {}
func (s *RoleSeeder) Environments() []string { return []string{} }

// Env-specific seeder - creates users with roles (runs after)
type UserSeeder struct {}
func (s *UserSeeder) Environments() []string { return []string{"dev", "test"} }
```

### 4. Use Faker for Test Data

For test/dev seeders, use faker to generate realistic data:

```go
import "github.com/jaswdr/faker/v2"

type TestUserSeeder struct {
    faker    *faker.Faker
    userRepo UserRepository
}

func (s *TestUserSeeder) Seed(ctx context.Context) error {
    for i := 0; i < 10; i++ {
        user := &User{
            Email: s.faker.Internet().Email(),
            Name:  s.faker.Person().Name(),
        }
        if err := s.userRepo.Create(ctx, user); err != nil {
            return err
        }
    }
    return nil
}

func (s *TestUserSeeder) Environments() []string {
    return []string{"dev", "test"}
}
```

## Error Handling

Seeding stops on first error:

```go
err := registry.SeedAll(ctx)
if err != nil {
    // Error message includes which seeder failed
    log.Fatal(err) // "seeding failed: <seeder error>"
}
```

## API Reference

### Registry

```go
// NewSeederRegistry creates a filtered registry
func NewSeederRegistry(seeders []Seeder, currentEnv string) *SeederRegistry

// SeedAll executes all seeders in order
func (r *SeederRegistry) SeedAll(ctx context.Context) error
```

### Command

```go
// NewSeedCmd creates a Cobra command for seeding
func NewSeedCmd(registry *SeederRegistry) *cobra.Command
```

## Integration with Testing

Seeders are useful for integration tests:

```go
func TestUserService_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    
    // Seed test data
    seeders := []seeder.Seeder{
        NewRoleSeeder(db),
        NewTestUserSeeder(db),
    }
    registry := seeder.NewSeederRegistry(seeders, "test")
    require.NoError(t, registry.SeedAll(context.Background()))
    
    // Run tests with seeded data
    // ...
}
```
