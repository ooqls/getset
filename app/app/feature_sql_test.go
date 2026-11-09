package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSQLFeature_BootstrapFailureStopsStartup(t *testing.T) {
	for _, tt := range []struct {
		name    string
		feature SQLFeature
	}{
		{name: "SQLX", feature: SQLX(WithCreateTableStatements([]string{"SELECT 1"}))},
		{name: "PGX", feature: PGX(WithCreateTableStatements([]string{"SELECT 1"}))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.feature.PasswordFile = filepath.Join(t.TempDir(), "missing-password")
			app := New("bootstrap-failure", Features{SQL: tt.feature})
			started := false
			app.OnStartup(func(*AppContext) error {
				started = true
				return nil
			})
			err := app.Run(t.Context())
			require.ErrorContains(t, err, "failed to start sql")
			assert.ErrorContains(t, err, "read SQL password file")
			assert.False(t, started, "bootstrap failure must precede the user's startup hook")
		})
	}
}

func TestSQLFeature_EmptyBootstrapDoesNotReadPassword(t *testing.T) {
	for _, tt := range []struct {
		name    string
		feature SQLFeature
	}{
		{name: "SQLX", feature: SQLX()},
		{name: "PGX", feature: PGX()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.feature.PasswordFile = filepath.Join(t.TempDir(), "missing-password")
			ctx := NewAppContext(context.Background(), zap.NewNop())
			require.NoError(t, tt.feature.start(ctx, &Features{}))
		})
	}
}

func TestSQLFeature_DiscoveryFailureStopsStartup(t *testing.T) {
	for _, tt := range []struct {
		name    string
		feature SQLFeature
	}{
		{name: "SQLX", feature: SQLX(WithSQLDirs([]string{"["}))},
		{name: "PGX", feature: PGX(WithSQLDirs([]string{"["}))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := New("discovery-failure", Features{SQL: tt.feature})
			started := false
			app.OnStartup(func(*AppContext) error {
				started = true
				return nil
			})
			err := app.Run(t.Context())
			require.ErrorIs(t, err, filepath.ErrBadPattern)
			assert.ErrorContains(t, err, "failed to start sql")
			assert.False(t, started)
		})
	}
}
