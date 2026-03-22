# fxcore

Core FX utilities for command registration. Every Talav FX module that registers a CLI command uses helpers from this package.

## What it provides

- `FxCommandsParam` — an `fx.In` struct that collects all commands from the `talav-core-commands` group
- `AsRootCommand` — registers a constructor as a top-level Cobra command
- `AsNamedCommand` — registers a constructor with a named tag for injection into parent commands

## Usage

### Register a top-level command

```go
var FxMyModule = fx.Module(
    "my",
    fx.Provide(myservice.NewService),
    fxcore.AsRootCommand(cmd.NewServeCmd),
)
```

### Register subcommands

```go
var FxORMModule = fx.Module(
    "orm",
    fxcore.AsNamedCommand("migrate-up-cmd", cmd.NewMigrateUpCmd),
    fxcore.AsNamedCommand("migrate-down-cmd", cmd.NewMigrateDownCmd),
    fxcore.AsRootCommand(
        cmd.NewMigrateCmd,
        fx.ParamTags(`name:"migrate-up-cmd"`, `name:"migrate-down-cmd"`),
    ),
)
```

### Collect commands in application bootstrap

```go
fx.Invoke(func(p fxcore.FxCommandsParam) {
    root.AddCommand(p.Commands...)
})
```

`pkg/module/framework` does this automatically — most users never call it directly.
