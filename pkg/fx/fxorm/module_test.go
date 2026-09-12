//go:build cgo

package fxorm

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/talav/talav/pkg/component/orm"
	"go.uber.org/fx"
)

type testLifecycle struct {
	hooks []fx.Hook
}

func (l *testLifecycle) Append(hook fx.Hook) {
	l.hooks = append(l.hooks, hook)
}

func TestNewORMClosesConnection(t *testing.T) {
	lifecycle := &testLifecycle{}
	db, err := newORM(ormParams{
		Lifecycle: lifecycle,
		Config: orm.ORMConfig{
			Driver: orm.DriverSQLite,
			DSN:    filepath.Join(t.TempDir(), "database.sqlite"),
		},
		Factory: orm.NewDefaultORMFactory(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("create database: %v", err)
	}
	if len(lifecycle.hooks) != 1 {
		t.Fatalf("lifecycle hooks: got %d, want 1", len(lifecycle.hooks))
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	if err := lifecycle.hooks[0].OnStop(context.Background()); err != nil {
		t.Fatalf("stop database: %v", err)
	}
	if err := sqlDB.Ping(); err == nil {
		t.Fatal("database connection remains open after shutdown")
	}
}
