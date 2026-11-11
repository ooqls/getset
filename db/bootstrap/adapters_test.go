package bootstrap_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ooqls/getset/db/bootstrap"
	"github.com/stretchr/testify/require"
)

func TestAdapters_EmptyAndPasswordReadFailure(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		adapter func(string) bootstrap.Open
	}{
		{name: "SQLX", adapter: bootstrap.SQLX},
		{name: "PGX", adapter: bootstrap.PGX},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			open := tt.adapter(filepath.Join(t.TempDir(), "missing-password"))
			require.NoError(t, bootstrap.Run(t.Context(), bootstrap.Plan{}, open))
			err := bootstrap.Run(t.Context(), bootstrap.Plan{Tables: []string{"SELECT 1"}}, open)
			require.ErrorContains(t, err, "read SQL password file")
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = open(ctx)
			require.ErrorIs(t, err, context.Canceled)
		})
	}
}
