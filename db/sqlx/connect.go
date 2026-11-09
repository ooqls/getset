package sqlx

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/db/postgres"
)

var c = conn.New("sqlx", func() (*sqlx.DB, error) {
	opts, err := postgres.GetRegistryOptions(nil)
	if err != nil {
		return nil, err
	}
	return open(opts)
}, func(db *sqlx.DB) error {
	return db.Close()
})

func open(opt postgres.Options) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s password=%s port=%d user=%s dbname=%s sslmode=disable",
		opt.Host, opt.Pw, opt.Port, opt.User, opt.DB)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql: %w", err)
	}
	return db, nil
}

// GetSQLX returns the database, opening it from the registry on first use.
// It panics if the database can't be opened.
func GetSQLX() *sqlx.DB {
	return c.MustGet()
}

// Get is the error-returning counterpart of GetSQLX.
func Get() (*sqlx.DB, error) {
	return c.Get()
}

// Init opens a database with opt and makes it the one GetSQLX returns.
func Init(opt postgres.Options) (*sqlx.DB, error) {
	if err := c.Init(func() (*sqlx.DB, error) { return open(opt) }); err != nil {
		return nil, err
	}
	return c.Get()
}

// InitDefault opens the database from the registry unless one is already
// open.
func InitDefault() error {
	return c.InitDefault()
}
