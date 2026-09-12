package fxorm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/talav/talav/pkg/component/orm"
	"github.com/talav/talav/pkg/component/orm/cmd"
	"github.com/talav/talav/pkg/component/validator"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"github.com/talav/talav/pkg/fx/fxcore"
	"github.com/talav/talav/pkg/fx/fxvalidator"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const ModuleName = "orm"

// FxORMModule is the [Fx] ORM module.
var FxORMModule = fx.Module(
	ModuleName,
	fxconfig.AsConfig("database", orm.ORMConfig{}),
	fx.Provide(
		orm.NewDefaultORMFactory,
		orm.NewDefaultMigrationFactory,
		newORM,
		newMigration,
	),
	// Register unique validator to validator group
	fxvalidator.AsValidatorConstructorCtx(newUniqueValidatorDefinition),
	// Register unique translation to validator group
	fxvalidator.AsTranslationConstructor(NewUniqueTranslation),
	// Register migrate command subcommands with named tags
	fxcore.AsNamedCommand("migrate-create-cmd", cmd.NewMigrateCreateCmd),
	fxcore.AsNamedCommand("migrate-up-cmd", cmd.NewMigrateUpCmd),
	fxcore.AsNamedCommand("migrate-down-cmd", cmd.NewMigrateDownCmd),
	fxcore.AsNamedCommand("migrate-reset-cmd", cmd.NewMigrateResetCmd),
	// Register main migrate command as top-level command
	fxcore.AsRootCommand(
		cmd.NewMigrateCmd,
		fx.ParamTags(`name:"migrate-create-cmd"`, `name:"migrate-up-cmd"`, `name:"migrate-down-cmd"`, `name:"migrate-reset-cmd"`),
	),
)

type ormParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Config    orm.ORMConfig
	Factory   orm.ORMFactory
	Logger    *slog.Logger
}

func newORM(p ormParams) (*gorm.DB, error) {
	db, err := p.Factory.Create(p.Config, p.Logger)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection: %w", err)
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}

type uniqueValidatorDefinitionParams struct {
	fx.In
	Repositories []orm.ExistsChecker `group:"repository-checkers"`
}

func newUniqueValidatorDefinition(p uniqueValidatorDefinitionParams) validator.ValidationDefinitionCtx {
	registry := orm.NewRepositoryRegistryFromRepos(p.Repositories)
	validator := orm.NewUniqueValidator(registry)

	return &uniqueValidatorDefinition{
		validator: validator,
	}
}

type migrationParams struct {
	fx.In
	Factory orm.MigrationFactory
	DB      *gorm.DB
}

func newMigration(p migrationParams) (*orm.Migration, error) {
	return p.Factory.Create(p.DB)
}
