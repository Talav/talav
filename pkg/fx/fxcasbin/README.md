# fxcasbin

Fx module for Casbin RBAC enforcer.

## Overview

The `fxcasbin` module provides a Casbin enforcer factory that accepts a `persist.Adapter` interface, making it decoupled from any specific storage implementation (GORM, file, memory, etc.).

## Usage

```go
import (
	"github.com/talav/talav/pkg/fx/fxcasbin"
	"github.com/casbin/gorm-adapter/v3"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fxcasbin.FxCasbinModule,
		// Provide adapter (e.g., GORM adapter)
		fx.Provide(func(db *gorm.DB) (persist.Adapter, error) {
			return gormadapter.NewAdapterByDB(db)
		}),
		fx.Invoke(func(enforcer *casbin.Enforcer) {
			// Use enforcer
		}),
	).Run()
}
```

## Configuration

```yaml
casbin:
  model_path: "configs/casbin.conf"
```

The adapter is injected separately, allowing you to use any Casbin adapter implementation.

