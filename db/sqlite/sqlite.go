package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ooqls/getset/db/conn"

	_ "modernc.org/sqlite"
)

var (
	dbs = make(map[string]*conn.Conn[*sql.DB])
	m   sync.Mutex
)

// open opens the SQLite file at path and runs every schema statement.
func open(name, path string, schema []string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db %q (%s): %v", name, path, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite db %q: %v", name, err)
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to exec schema on sqlite db %q: %v\nstatement: %s", name, err, stmt)
		}
	}

	return db, nil
}

// Init opens the SQLite file at path, runs all schema statements, and
// registers the database under name for Get. Calling Init again for the
// same name replaces (and closes) the previous database; if that fails, the
// previous database stays.
func Init(name, path string, schema []string) error {
	opener := func() (*sql.DB, error) { return open(name, path, schema) }

	m.Lock()
	c, ok := dbs[name]
	if !ok {
		c = conn.New("sqlite/"+name, opener, (*sql.DB).Close)
		dbs[name] = c
	}
	m.Unlock()

	if !ok {
		if err := c.InitDefault(); err != nil {
			m.Lock()
			delete(dbs, name)
			m.Unlock()
			return err
		}
		return nil
	}
	return c.Init(opener)
}

// Get returns the database registered under name. The database is shared:
// callers must not close it.
func Get(name string) (*sql.DB, error) {
	m.Lock()
	c, ok := dbs[name]
	m.Unlock()
	if !ok {
		return nil, fmt.Errorf("sqlite db %q not found; ensure it was configured as a SQLiteFeature database", name)
	}
	return c.Get()
}

// MustGet returns the named database and panics if it does not exist.
func MustGet(name string) *sql.DB {
	db, err := Get(name)
	if err != nil {
		panic(err)
	}
	return db
}
