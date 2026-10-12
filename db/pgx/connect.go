package pgx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/db/postgres"
)

var c = conn.New("pgx", func() (*pgxpool.Pool, error) {
	return open(context.Background(), postgres.GetRegistryOptions())
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

// Init opens a pool with opt and makes it the one GetPGX returns.
func Init(ctx context.Context, opt postgres.Options) (*pgxpool.Pool, error) {
	if err := c.Init(func() (*pgxpool.Pool, error) { return open(ctx, opt) }); err != nil {
		return nil, err
	}
	return c.MustGet(), nil
}

// InitDefault opens the pool from the registry unless one is already open.
func InitDefault() error {
	return c.InitDefault()
}
