package fxlogger

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/talav/pkg/component/config"
	"github.com/talav/talav/pkg/fx/fxconfig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestModule_FxLoggerModule_ProvidesLogger(t *testing.T) {
	testdataDir := filepath.Join("testdata", "testmodule_fxloggermodule_provideslogger")
	t.Setenv("APP_ENV", "dev")

	var log *slog.Logger

	fxtest.New(
		t,
		fx.NopLogger,
		fxconfig.FxConfigModule,
		fxconfig.AsConfigSource(config.ConfigSource{
			Path:     testdataDir,
			Patterns: []string{"config.yaml"},
			Parser:   yaml.Parser(),
		}),
		FxLoggerModule,
		fx.Populate(&log),
	).RequireStart().RequireStop()

	require.NotNil(t, log)

	// Verify logger can log without error
	log.Info("test message", "key", "value")
	assert.NotNil(t, log.Handler())
}
