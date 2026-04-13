package app

const (
	sqlite_databasesOpt string = "opt-sqlite-databases"
)

type sqliteOpt struct{ featureOpt }

// SQLiteDB describes a single SQLite database file and the schema
// statements to execute against it on startup.
type SQLiteDB struct {
	Name   string
	Path   string
	Schema []string
}

func WithSQLiteDatabase(name, path string, schema ...string) sqliteOpt {
	return sqliteOpt{featureOpt{key: sqlite_databasesOpt, value: SQLiteDB{
		Name:   name,
		Path:   path,
		Schema: schema,
	}}}
}

type SQLiteFeature struct {
	Enabled   bool
	Databases []SQLiteDB
}

func (f *SQLiteFeature) apply(opt sqliteOpt) {
	switch opt.key {
	case sqlite_databasesOpt:
		f.Databases = append(f.Databases, opt.value.(SQLiteDB))
	}
}

func SQLite(opts ...sqliteOpt) SQLiteFeature {
	f := SQLiteFeature{Enabled: true}
	for _, opt := range opts {
		f.apply(opt)
	}
	return f
}
