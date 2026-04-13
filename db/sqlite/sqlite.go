package sqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ooqls/getset/log"
	"go.uber.org/zap"

	_ "modernc.org/sqlite"
)

var (
	dbs map[string]*sql.DB = make(map[string]*sql.DB)
	m   sync.Mutex
	l   = log.NewLogger("sqlite")
)

// Init opens the SQLite file at path, runs all schema statements,
// and registers the connection under name for later retrieval with Get.
func Init(name, path string, schema []string) error {
	m.Lock()
	defer m.Unlock()

	l.Debug("[SQLite] opening database", zap.String("name", name), zap.String("path", path))
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db %q (%s): %v", name, path, err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping sqlite db %q: %v", name, err)
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to exec schema on sqlite db %q: %v\nstatement: %s", name, err, stmt)
		}
	}

	dbs[name] = db
	l.Debug("[SQLite] database ready", zap.String("name", name))
	return nil
}

// Get returns the named database connection. Returns an error if Init was
// not called for that name.
func Get(name string) (*sql.DB, error) {
	m.Lock()
	defer m.Unlock()

	db, ok := dbs[name]
	if !ok {
		return nil, fmt.Errorf("sqlite db %q not found; ensure it was configured as a SQLiteFeature database", name)
	}
	return db, nil
}

// MustGet returns the named connection and panics if it does not exist.
func MustGet(name string) *sql.DB {
	db, err := Get(name)
	if err != nil {
		panic(err)
	}
	return db
}
