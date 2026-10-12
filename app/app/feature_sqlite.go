package app

import (
	dbsqlite "github.com/ooqls/getset/db/sqlite"
	"go.uber.org/zap"
)

type SQLiteDBConfig struct {
	Name   string   `yaml:"name"`
	Path   string   `yaml:"path"`
	Schema []string `yaml:"schema"`
}

type SQLiteConfig struct {
	Enabled   bool             `yaml:"enabled"`
	Databases []SQLiteDBConfig `yaml:"databases"`
}

// SQLiteDB describes a single SQLite database file and the schema
// statements to execute against it on startup.
type SQLiteDB struct {
	Name   string
	Path   string
	Schema []string
}

type SQLiteOption func(*SQLiteFeature)

func WithSQLiteDatabase(name, path string, schema ...string) SQLiteOption {
	return func(f *SQLiteFeature) {
		f.Databases = append(f.Databases, SQLiteDB{Name: name, Path: path, Schema: schema})
	}
}

type SQLiteFeature struct {
	Enabled   bool
	Databases []SQLiteDB
}

func SQLite(opts ...SQLiteOption) SQLiteFeature {
	f := SQLiteFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func sqliteFromConfig(c SQLiteConfig) SQLiteFeature {
	f := SQLiteFeature{Enabled: c.Enabled}
	for _, db := range c.Databases {
		f.Databases = append(f.Databases, SQLiteDB(db))
	}
	return f
}

func (f *SQLiteFeature) name() string  { return "sqlite" }
func (f *SQLiteFeature) enabled() bool { return f.Enabled }

func (f *SQLiteFeature) start(ctx *AppContext, fs *Features) error {
	l := ctx.L()
	for _, db := range f.Databases {
		l.Debug("[Startup SQLite] opening database", zap.String("name", db.Name), zap.String("path", db.Path))
		if err := dbsqlite.Init(db.Name, db.Path, db.Schema); err != nil {
			return err
		}
	}
	return nil
}
