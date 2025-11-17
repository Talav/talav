package fxorm

import (
	playgroundvalidator "github.com/go-playground/validator/v10"
	"github.com/talav/talav/pkg/component/orm"
)

type uniqueValidatorDefinition struct {
	validator *orm.UniqueValidator
}

func (d *uniqueValidatorDefinition) Tag() string {
	return "unique"
}

func (d *uniqueValidatorDefinition) Fn() playgroundvalidator.FuncCtx {
	return d.validator.Validate
}

func (d *uniqueValidatorDefinition) CallEvenIfNull() bool {
	return false
}
