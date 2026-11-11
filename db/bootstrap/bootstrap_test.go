package bootstrap_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooqls/getset/db/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSQL(t *testing.T, dir, name, sql string) string {
	t.Helper()
	file := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(file, []byte(sql), 0600))
	return file
}

func TestRun_OrderAndSingleConnection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	first := writeSQL(t, t.TempDir(), "first.sql", "explicit; whole file;")
	writeSQL(t, dir, "b.sql", "directory-b")
	writeSQL(t, dir, "a.sql", "directory-a")
	writeSQL(t, dir, "ignored.txt", "not SQL")
	secondDir := t.TempDir()
	writeSQL(t, secondDir, "a.sql", "second-directory")
	plan := bootstrap.Plan{
		Files: []string{first, first}, Dirs: []string{dir, secondDir},
		Tables: []string{"table-1", "table-2"}, Indexes: []string{"index-1", "index-2"},
	}
	ctx := t.Context()
	var got []string
	opens := 0
	err := bootstrap.Run(ctx, plan, func(openCtx context.Context) (bootstrap.Execute, error) {
		opens++
		assert.Equal(t, ctx, openCtx)
		return func(execCtx context.Context, sql string) error {
			assert.Equal(t, ctx, execCtx)
			got = append(got, sql)
			return nil
		}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, opens)
	assert.Equal(t, []string{
		"explicit; whole file;", "explicit; whole file;", "directory-a", "directory-b", "second-directory",
		"table-1", "table-2", "index-1", "index-2",
	}, got)
	assert.Equal(t, []string{first, first}, plan.Files)
}

func TestRun_EmptyDoesNotOpen(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		plan bootstrap.Plan
	}{
		{name: "empty"},
		{name: "no directory matches", plan: bootstrap.Plan{Dirs: []string{t.TempDir()}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.NoError(t, bootstrap.Run(t.Context(), tt.plan, func(context.Context) (bootstrap.Execute, error) {
				t.Fatal("an empty bootstrap must not open a connection")
				return nil, nil
			}))
		})
	}
}

func TestRun_StopsAtFirstFailure(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("execution failed")
	file := writeSQL(t, t.TempDir(), "init.sql", "file")
	for _, tt := range []struct {
		name string
		fail string
		want []string
		step string
	}{
		{name: "file", fail: "file", want: []string{"file"}, step: file},
		{name: "table", fail: "table-1", want: []string{"file", "table-1"}, step: "table statement 1"},
		{name: "index", fail: "index-1", want: []string{"file", "table-1", "table-2", "index-1"}, step: "index statement 1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			err := bootstrap.Run(t.Context(), bootstrap.Plan{
				Files: []string{file}, Tables: []string{"table-1", "table-2"}, Indexes: []string{"index-1", "index-2"},
			}, func(context.Context) (bootstrap.Execute, error) {
				return func(_ context.Context, sql string) error {
					got = append(got, sql)
					if sql == tt.fail {
						return sentinel
					}
					return nil
				}, nil
			})
			assert.ErrorIs(t, err, sentinel)
			assert.ErrorContains(t, err, tt.step)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRun_DiscoveryReadAndConnectionErrors(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "missing.sql")
	for _, tt := range []struct {
		name     string
		plan     bootstrap.Plan
		openErr  error
		wantErr  error
		wantOpen int
		step     string
	}{
		{name: "invalid directory pattern", plan: bootstrap.Plan{Dirs: []string{"["}}, wantErr: filepath.ErrBadPattern, step: "discover SQL directory"},
		{name: "missing file", plan: bootstrap.Plan{Files: []string{missing}}, wantErr: os.ErrNotExist, wantOpen: 1, step: missing},
		{name: "connection", plan: bootstrap.Plan{Tables: []string{"never execute"}}, openErr: os.ErrPermission, wantErr: os.ErrPermission, wantOpen: 1, step: "open SQL bootstrap connection"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opens := 0
			err := bootstrap.Run(t.Context(), tt.plan, func(context.Context) (bootstrap.Execute, error) {
				opens++
				return func(context.Context, string) error {
					t.Fatal("must stop before execution")
					return nil
				}, tt.openErr
			})
			assert.ErrorIs(t, err, tt.wantErr)
			assert.ErrorContains(t, err, tt.step)
			assert.Equal(t, tt.wantOpen, opens)
		})
	}
}

func TestRun_Cancellation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		cancelBefore bool
		cancelOnOpen bool
		execErr      error
		last         bool
		wantCalls    int
	}{
		{name: "before opening", cancelBefore: true},
		{name: "during opening", cancelOnOpen: true},
		{name: "during execution", execErr: context.Canceled, wantCalls: 1},
		{name: "driver cancellation error", execErr: errors.New("driver interrupted"), wantCalls: 1},
		{name: "between statements", wantCalls: 1},
		{name: "after last statement", last: true, wantCalls: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tt.cancelBefore {
				cancel()
			}
			calls := 0
			stmts := []string{"first", "never execute"}
			if tt.last {
				stmts = stmts[:1]
			}
			err := bootstrap.Run(ctx, bootstrap.Plan{Tables: stmts}, func(context.Context) (bootstrap.Execute, error) {
				if tt.cancelBefore {
					t.Fatal("canceled bootstrap must not open")
				}
				if tt.cancelOnOpen {
					cancel()
				}
				return func(context.Context, string) error {
					calls++
					cancel()
					return tt.execErr
				}, nil
			})
			assert.ErrorIs(t, err, context.Canceled)
			if tt.execErr != nil {
				assert.ErrorIs(t, err, tt.execErr)
			}
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func ExampleRun() {
	// Real callers select bootstrap.SQLX(passwordFile) or bootstrap.PGX(passwordFile).
	open := func(context.Context) (bootstrap.Execute, error) {
		return func(_ context.Context, sql string) error {
			fmt.Println(sql)
			return nil
		}, nil
	}
	err := bootstrap.Run(context.Background(), bootstrap.Plan{
		Tables:  []string{"CREATE TABLE IF NOT EXISTS users (id integer)"},
		Indexes: []string{"CREATE INDEX IF NOT EXISTS users_id ON users (id)"},
	}, open)
	fmt.Println(err)
	// Output:
	// CREATE TABLE IF NOT EXISTS users (id integer)
	// CREATE INDEX IF NOT EXISTS users_id ON users (id)
	// <nil>
}
