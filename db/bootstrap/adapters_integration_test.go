//go:build integration

package bootstrap_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/ooqls/getset/db/bootstrap"
	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAdapters_Postgres(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	const password = "bootstrap-password"
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres:17-alpine", ExposedPorts: []string{"5432/tcp"},
			Env:        map[string]string{"POSTGRES_PASSWORD": password},
			WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		assert.NoError(t, container.Terminate(cleanupCtx))
	})
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	dsn := &url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port.Port()), User: url.UserPassword("postgres", password), Path: "/postgres"}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	observer, err := sql.Open("postgres", dsn.String())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, observer.Close()) })
	require.NoError(t, observer.PingContext(ctx))
	if registry.IsSet() {
		old := *registry.Get()
		t.Cleanup(func() { registry.Set(old) })
	}
	registry.Set(registry.Registry{Postgres: &registry.Database{
		Database: "postgres",
		Server: registry.Server{
			Host: host, Port: int(port.Num()),
			Auth: registry.Auth{Enabled: true, Username: "postgres", PasswordFile: filepath.Join(t.TempDir(), "missing-registry-password")},
		},
	}})
	passwordFile := writeSQL(t, t.TempDir(), "password", password)
	for _, tt := range []struct {
		name    string
		adapter func(string) bootstrap.Open
	}{
		{name: "sqlx", adapter: bootstrap.SQLX},
		{name: "pgx", adapter: bootstrap.PGX},
	} {
		t.Run(tt.name, func(t *testing.T) {
			table := "bootstrap_" + tt.name
			insert := func(n int) string { return fmt.Sprintf("INSERT INTO %s VALUES (%d)", table, n) }
			explicit := writeSQL(t, t.TempDir(), "init.sql", fmt.Sprintf("CREATE TABLE %s (n integer); %s;", table, insert(1)))
			dir := t.TempDir()
			writeSQL(t, dir, "b.sql", insert(3))
			writeSQL(t, dir, "a.sql", insert(2))
			// Files and statements share an override and execution order.
			require.NoError(t, bootstrap.Run(ctx, bootstrap.Plan{
				Files: []string{explicit}, Dirs: []string{dir},
				Tables: []string{insert(4)}, Indexes: []string{fmt.Sprintf("CREATE INDEX %s_idx ON %s (n)", table, table), insert(5)},
			}, tt.adapter(passwordFile)))
			rows, err := observer.QueryContext(ctx, "SELECT n FROM "+table+" ORDER BY ctid")
			require.NoError(t, err)
			defer rows.Close()
			var got []int
			for rows.Next() {
				var n int
				require.NoError(t, rows.Scan(&n))
				got = append(got, n)
			}
			require.NoError(t, rows.Err())
			assert.Equal(t, []int{1, 2, 3, 4, 5}, got)
			// A statement-only override also bypasses the broken Registry password.
			require.NoError(t, bootstrap.Run(ctx, bootstrap.Plan{Tables: []string{insert(6)}}, tt.adapter(passwordFile)))
			// Default selection preserves the explicitly initialized connection.
			// Registry password is still invalid; success requires reusing the
			// connection selected by the preceding explicit initialization.
			require.NoError(t, bootstrap.Run(ctx, bootstrap.Plan{Tables: []string{insert(7)}}, tt.adapter("")))
			// A failure neither rolls back prior statements nor runs later ones.
			err = bootstrap.Run(ctx, bootstrap.Plan{Tables: []string{insert(8), "invalid SQL", insert(9)}}, tt.adapter(""))
			require.ErrorContains(t, err, "table statement 2")
			var count int
			require.NoError(t, observer.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE n = 8").Scan(&count))
			assert.Equal(t, 1, count)
			require.NoError(t, observer.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE n = 9").Scan(&count))
			assert.Zero(t, count)
			// A running statement uses the caller's cancellation.
			deadlineCtx, deadlineCancel := context.WithTimeout(ctx, 250*time.Millisecond)
			defer deadlineCancel()
			err = bootstrap.Run(deadlineCtx, bootstrap.Plan{Tables: []string{"SELECT pg_sleep(10)", insert(10)}}, tt.adapter(""))
			require.ErrorIs(t, err, context.DeadlineExceeded)
			assert.ErrorIs(t, deadlineCtx.Err(), context.DeadlineExceeded)
			require.NoError(t, observer.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE n = 10").Scan(&count))
			assert.Zero(t, count)
		})
	}
}
