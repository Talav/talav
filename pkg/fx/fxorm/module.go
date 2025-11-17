package fxorm

import (
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
		func(cfg orm.ORMConfig, factory orm.ORMFactory, logger *slog.Logger) (*gorm.DB, error) {
			return factory.Create(cfg, logger)
		},
		NewFxMigration,
	),
	// Register unique validator to validator group
	fxvalidator.AsValidatorConstructorCtx(NewFxUniqueValidatorDefinition),
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

// FxUniqueValidatorDefinitionParam allows injection of the required dependencies in [NewFxUniqueValidatorDefinition].
type FxUniqueValidatorDefinitionParam struct {
	fx.In
	Repositories []orm.ExistsChecker `group:"repository-checkers"`
}

// NewFxUniqueValidatorDefinition returns a [validator.ValidationDefinitionCtx] for the unique validator.
// It creates a repository registry from the provided repositories, creates the validator, and wraps it as a definition.
func NewFxUniqueValidatorDefinition(p FxUniqueValidatorDefinitionParam) validator.ValidationDefinitionCtx {
	registry := orm.NewRepositoryRegistryFromRepos(p.Repositories)
	validator := orm.NewUniqueValidator(registry)
	return &uniqueValidatorDefinition{
		validator: validator,
	}
}

// FxMigrationParam allows injection of the required dependencies in [NewFxMigration].
type FxMigrationParam struct {
	fx.In
	Factory orm.MigrationFactory
	DB      *gorm.DB
}

// NewFxMigration returns a [orm.Migration] instance.
func NewFxMigration(p FxMigrationParam) (*orm.Migration, error) {
	return p.Factory.Create(p.DB)
}
