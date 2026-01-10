# FxSeeder Module

Uber FX integration for the seeder component. Provides automatic seeder registration, environment filtering, and CLI command integration.

## Features

- **Automatic Registration**: Collect all seeders using FX groups
- **Environment Filtering**: Automatically filters seeders by `APP_ENV`
- **CLI Integration**: Registers `seed` command with FX Core
- **Simple API**: One helper function to register seeders

## Installation

```bash
go get github.com/talav/talav/pkg/fx/fxseeder
```

## Quick Start

### 1. Add Module to Application

```go
package main

import (
    "github.com/talav/talav/pkg/fx/fxcore"
    "github.com/talav/talav/pkg/fx/fxseeder"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        fxcore.FxCoreModule,      // Required for CLI
        fxseeder.FxSeederModule,   // Seeder module
        // Your modules with seeders...
    )
    app.Run()
}
```

### 2. Register Seeders in Your Module

```go
package fxuser

import (
    "github.com/talav/talav/pkg/fx/fxseeder"
    "go.uber.org/fx"
)

var Module = fx.Module(
    "user",
    // ... other providers ...
    
    // Register seeders
    fxseeder.AsSeeder(NewUserSeeder),
    fxseeder.AsSeeder(NewTestDataSeeder),
)

// Seeder constructor
func NewUserSeeder(db *gorm.DB) seeder.Seeder {
    return &UserSeeder{db: db}
}
```

### 3. Run from CLI

```bash
# Run all seeders
./myapp seed

# Set environment
APP_ENV=prod ./myapp seed
```

## Module Configuration

The module provides:

1. **SeederRegistry**: Automatically created with all registered seeders
2. **Seed Command**: CLI command registered with FX Core
3. **Environment Filtering**: Uses `APP_ENV` to filter seeders

```go
var FxSeederModule = fx.Module(
    ModuleName,
    fx.Provide(NewFxSeederRegistry),           // Registry provider
    fxcore.AsRootCommand(cmd.NewSeedCmd),      // CLI command
)
```

## Registering Seeders

### Basic Registration

```go
fxseeder.AsSeeder(NewMySeeder)
```

### With Additional Annotations

```go
fxseeder.AsSeeder(
    NewMySeeder,
    fx.ParamTags(`name:"my-dependency"`),
)
```

### Seeder Implementation

```go
type MySeeder struct {
    db     *gorm.DB
    logger *slog.Logger
}

func NewMySeeder(db *gorm.DB, logger *slog.Logger) seeder.Seeder {
    return &MySeeder{db: db, logger: logger}
}

func (s *MySeeder) Seed(ctx context.Context) error {
    s.logger.Info("seeding data")
    // Seeding logic here
    return nil
}

func (s *MySeeder) Environments() []string {
    return []string{"dev", "test"} // Only in dev and test
}
```

## Environment Detection

The module automatically detects the environment from `APP_ENV`:

```bash
APP_ENV=dev ./myapp seed    # Runs dev and base seeders
APP_ENV=test ./myapp seed   # Runs test and base seeders
APP_ENV=prod ./myapp seed   # Runs prod and base seeders
APP_ENV="" ./myapp seed     # Defaults to "dev"
```

## Complete Example

### Project Structure

```
myapp/
├── main.go
└── internal/
    ├── seeder/
    │   ├── user_seeder.go
    │   └── testdata_seeder.go
    └── module/
        └── user/
            └── module.go
```

### main.go

```go
package main

import (
    "github.com/talav/talav/pkg/fx/fxcore"
    "github.com/talav/talav/pkg/fx/fxconfig"
    "github.com/talav/talav/pkg/fx/fxlogger"
    "github.com/talav/talav/pkg/fx/fxorm"
    "github.com/talav/talav/pkg/fx/fxseeder"
    "myapp/internal/module/user"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // Core infrastructure
        fxcore.FxCoreModule,
        fxlogger.FxLoggerModule,
        fxconfig.FxConfigModule,
        fxorm.FxORMModule,
        fxseeder.FxSeederModule,
        
        // Application modules
        user.Module,
    )
    app.Run()
}
```

### internal/seeder/user_seeder.go

```go
package seeder

import (
    "context"
    "github.com/talav/talav/pkg/component/seeder"
    "gorm.io/gorm"
)

type UserSeeder struct {
    db *gorm.DB
}

func NewUserSeeder(db *gorm.DB) seeder.Seeder {
    return &UserSeeder{db: db}
}

func (s *UserSeeder) Seed(ctx context.Context) error {
    // Check if admin exists
    var count int64
    if err := s.db.Model(&User{}).Where("email = ?", "admin@example.com").Count(&count).Error; err != nil {
        return err
    }
    if count > 0 {
        return nil // Already seeded
    }
    
    // Create admin
    admin := &User{
        Email: "admin@example.com",
        Name:  "Admin User",
        Role:  "admin",
    }
    return s.db.Create(admin).Error
}

// Run in all environments
func (s *UserSeeder) Environments() []string {
    return []string{}
}
```

### internal/module/user/module.go

```go
package user

import (
    "github.com/talav/talav/pkg/fx/fxseeder"
    "myapp/internal/seeder"
    "go.uber.org/fx"
)

var Module = fx.Module(
    "user",
    fx.Provide(
        // ... repositories, services ...
    ),
    
    // Register seeders
    fxseeder.AsSeeder(seeder.NewUserSeeder),
    fxseeder.AsSeeder(seeder.NewTestDataSeeder),
)
```

## Dependency Injection

Seeders can inject any FX-provided dependencies:

```go
type ComplexSeeder struct {
    db       *gorm.DB
    logger   *slog.Logger
    hasher   security.PasswordHasher
    userRepo repository.UserRepository
    roleRepo repository.RoleRepository
}

func NewComplexSeeder(
    db *gorm.DB,
    logger *slog.Logger,
    hasher security.PasswordHasher,
    userRepo repository.UserRepository,
    roleRepo repository.RoleRepository,
) seeder.Seeder {
    return &ComplexSeeder{
        db:       db,
        logger:   logger,
        hasher:   hasher,
        userRepo: userRepo,
        roleRepo: roleRepo,
    }
}
```

## API Reference

### AsSeeder

```go
func AsSeeder(constructor any, annotations ...fx.Annotation) fx.Option
```

Registers a seeder constructor to the seeders group.

**Parameters:**
- `constructor` - Seeder constructor function
- `annotations` - Optional FX annotations (e.g., `fx.ParamTags`)

**Example:**
```go
fxseeder.AsSeeder(NewUserSeeder)
fxseeder.AsSeeder(NewUserSeeder, fx.ParamTags(`name:"my-db"`))
```

### Module Providers

The module automatically provides:

- `*seeder.SeederRegistry` - Registry with filtered seeders
- `*cobra.Command` - Seed command (registered with FX Core)

## Testing

For integration tests, you can invoke the registry directly:

```go
func TestMyFeature_Integration(t *testing.T) {
    var registry *seeder.SeederRegistry
    
    app := fxtest.New(
        t,
        fx.NopLogger,
        fxorm.FxORMModule,
        fxseeder.FxSeederModule,
        mymodule.Module,
        fx.Populate(&registry),
    ).RequireStart()
    defer app.RequireStop()
    
    // Run seeders
    err := registry.SeedAll(context.Background())
    require.NoError(t, err)
    
    // Test with seeded data
    // ...
}
```

## Best Practices

### 1. Organize Seeders by Module

Keep seeders close to the domain they seed:

```
internal/
├── module/
│   ├── user/
│   │   ├── module.go       # Registers UserSeeder
│   │   └── seeder.go       # UserSeeder implementation
│   └── product/
│       ├── module.go       # Registers ProductSeeder
│       └── seeder.go       # ProductSeeder implementation
```

### 2. Use Base Seeders for Required Data

Required data (roles, permissions) should run in all environments:

```go
func (s *RoleSeeder) Environments() []string {
    return []string{} // All environments
}
```

### 3. Environment-Specific Test Data

Test/demo data should only run in dev/test:

```go
func (s *TestDataSeeder) Environments() []string {
    return []string{"dev", "test"}
}
```

### 4. Make Seeders Idempotent

Always check if data exists before creating:

```go
func (s *MySeeder) Seed(ctx context.Context) error {
    exists, _ := s.repo.Exists(ctx, criteria)
    if exists {
        return nil
    }
    return s.repo.Create(ctx, data)
}
```
