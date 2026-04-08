package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var errValidatableSentinel = errors.New("validatable sentinel")

type validatableConfig struct {
	Reject bool `config:"reject"`
}

func (c *validatableConfig) Validate() error {
	if c.Reject {
		return errValidatableSentinel
	}

	return nil
}

type plainConfig struct {
	X string `config:"x"`
}

func TestValidate_ValidatableNilError(t *testing.T) {
	t.Parallel()

	cfg := validatableConfig{}
	require.NoError(t, Validate(&cfg))
}

func TestValidate_ValidatableReturnsError(t *testing.T) {
	t.Parallel()

	cfg := validatableConfig{Reject: true}
	err := Validate(&cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, errValidatableSentinel)
}

func TestValidate_NonValidatable(t *testing.T) {
	t.Parallel()

	cfg := plainConfig{X: "y"}
	require.NoError(t, Validate(&cfg))
}

func TestValidate_NilPointer(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		require.NoError(t, Validate((*validatableConfig)(nil)))
	})
}
