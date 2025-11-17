package fxcore

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule_AsRootCommand_RegistersCommand(t *testing.T) {
	var param FxCommandsParam

	fxtest.New(t,
		AsRootCommand(func() *cobra.Command {
			return &cobra.Command{
				Use: "test-root",
			}
		}),
		fx.Populate(&param),
	).RequireStart()

	require.Len(t, param.Commands, 1, "Should register one command")
	assert.Equal(t, "test-root", param.Commands[0].Use, "Command should have correct Use value")
}

func TestModule_AsNamedCommand_RegistersCommand(t *testing.T) {
	var cmd struct {
		fx.In
		Command *cobra.Command `name:"test-named-cmd"`
	}

	fxtest.New(t,
		AsNamedCommand("test-named-cmd", func() *cobra.Command {
			return &cobra.Command{
				Use: "test-named",
			}
		}),
		fx.Populate(&cmd),
	).RequireStart()

	require.NotNil(t, cmd.Command, "Command should be provided")
	assert.Equal(t, "test-named", cmd.Command.Use, "Command should have correct Use value")
}
