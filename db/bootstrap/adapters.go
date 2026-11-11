package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlxdriver "github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/postgres"
	"github.com/ooqls/getset/db/sqlx"
)

// SQLX selects the shared SQLX connection. A nonempty passwordFile overrides
// the Registry password and replaces the connection through db/conn.
// An empty passwordFile preserves any previously initialized connection.
func SQLX(passwordFile string) Open {
	return func(ctx context.Context) (Execute, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var db *sqlxdriver.DB
		var err error
		if passwordFile != "" {
			opts, optionsErr := overrideOptions(passwordFile)
			if optionsErr != nil {
				return nil, optionsErr
			}
			db, err = sqlx.Init(*opts)
		} else {
			db, err = sqlx.Get()
		}
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, sql string) error {
			_, err := db.ExecContext(ctx, sql)
			return err
		}, nil
	}
}

// PGX is SQLX's PGX counterpart, using the caller's context for execution and
// for opening an explicitly configured pool.
func PGX(passwordFile string) Open {
	return func(ctx context.Context) (Execute, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var db *pgxpool.Pool
		var err error
		if passwordFile != "" {
			opts, optionsErr := overrideOptions(passwordFile)
			if optionsErr != nil {
				return nil, optionsErr
			}
			db, err = pgx.Init(ctx, *opts)
		} else {
			db, err = pgx.Get()
		}
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, sql string) error {
			_, err := db.Exec(ctx, sql)
			return err
		}, nil
	}
}

func overrideOptions(file string) (*postgres.Options, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read SQL password file %q: %w", file, err)
	}
	pw := string(b)
	opts, err := postgres.GetRegistryOptions(&pw)
	if err != nil {
		return nil, err
	}
	return &opts, nil
}
