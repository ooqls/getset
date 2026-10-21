package sqlx

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/db/postgres"
	"github.com/ooqls/getset/log"
)

var l = log.NewLogger("db")

var c = conn.New("sqlx", func() (*sqlx.DB, error) {
	return open(postgres.GetRegistryOptions())
}, func(db *sqlx.DB) error {
	return db.Close()
})

func open(opt postgres.Options) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", opt.ConnectionString()+"?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open sql: %v", err)
	}
	return db, nil
}

// GetSQLX returns the database, opening it from the registry on first use.
// It panics if the database can't be opened.
func GetSQLX() *sqlx.DB {
	return c.MustGet()
}

// Init opens a database with opt and makes it the one GetSQLX returns.
func Init(opt postgres.Options) (*sqlx.DB, error) {
	if err := c.Init(func() (*sqlx.DB, error) { return open(opt) }); err != nil {
		return nil, err
	}
	return c.MustGet(), nil
}

// InitDefault opens the database from the registry unless one is already
// open.
func InitDefault() error {
	return c.InitDefault()
}
