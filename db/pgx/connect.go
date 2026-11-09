package pgx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/db/postgres"
)

var c = conn.New("pgx", func() (*pgxpool.Pool, error) {
	opts, err := postgres.GetRegistryOptions(nil)
	if err != nil {
		return nil, err
	}
	return open(context.Background(), opts)
}, func(p *pgxpool.Pool) error {
	p.Close()
	return nil
})

func open(ctx context.Context, opt postgres.Options) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, opt.ConnectionString())
}

// GetPGX returns the pool, opening it from the registry on first use. It
// panics if the pool can't be opened.
func GetPGX() *pgxpool.Pool {
	return c.MustGet()
}

// Get is the error-returning counterpart of GetPGX.
func Get() (*pgxpool.Pool, error) {
	return c.Get()
}

// Init opens a pool with opt and makes it the one GetPGX returns.
func Init(ctx context.Context, opt postgres.Options) (*pgxpool.Pool, error) {
	if err := c.Init(func() (*pgxpool.Pool, error) { return open(ctx, opt) }); err != nil {
		return nil, err
	}
	return c.Get()
}

// InitDefault opens the pool from the registry unless one is already open.
func InitDefault() error {
	return c.InitDefault()
}
