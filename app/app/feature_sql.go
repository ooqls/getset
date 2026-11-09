package app

import (
	"fmt"

	"github.com/ooqls/getset/db/bootstrap"
)

type sqlPackage string

const (
	PGXPackage  sqlPackage = "PGX"
	SQLXPackage sqlPackage = "SQLX"
)

type SQLConfig struct {
	Enabled          bool       `yaml:"enabled"`
	SQLPackage       sqlPackage `yaml:"sql_package"`
	SQLFilesDirs     []string   `yaml:"sql_files_dirs"`
	SQLFiles         []string   `yaml:"sql_files"`
	CreateTableStmts []string   `yaml:"create_table_stmts"`
	CreateIndexStmts []string   `yaml:"create_index_stmts"`
}

type SQLOption func(*SQLFeature)

func WithCreateTableStatements(stmts []string) SQLOption {
	return func(f *SQLFeature) { f.CreateTableStatements = stmts }
}

func WithCreateIndexStatements(stmts []string) SQLOption {
	return func(f *SQLFeature) { f.CreateIndexStatements = stmts }
}

func WithSQLFiles(files []string) SQLOption {
	return func(f *SQLFeature) { f.SQLFiles = files }
}

func WithSQLDirs(dirs []string) SQLOption {
	return func(f *SQLFeature) { f.SQLDirs = dirs }
}

func WithSQLPasswordFile(path string) SQLOption {
	return func(f *SQLFeature) { f.PasswordFile = path }
}

// SQLFeature connects to Postgres through SQLX or PGX, using the Registry's
// Postgres entry, then runs the given SQL files and statements. Any bootstrap
// failure stops startup; previously completed SQL is not rolled back.
type SQLFeature struct {
	Enabled               bool
	CreateTableStatements []string
	CreateIndexStatements []string
	SQLFiles              []string
	SQLDirs               []string
	SQLPackage            sqlPackage
	PasswordFile          string
}

func SQLX(opts ...SQLOption) SQLFeature {
	return newSQLFeature(SQLXPackage, opts...)
}

func PGX(opts ...SQLOption) SQLFeature {
	return newSQLFeature(PGXPackage, opts...)
}

func newSQLFeature(sp sqlPackage, opts ...SQLOption) SQLFeature {
	f := SQLFeature{Enabled: true, SQLPackage: sp}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func sqlFromConfig(c SQLConfig) SQLFeature {
	return SQLFeature{
		Enabled:               c.Enabled,
		SQLPackage:            c.SQLPackage,
		SQLFiles:              c.SQLFiles,
		SQLDirs:               c.SQLFilesDirs,
		CreateTableStatements: c.CreateTableStmts,
		CreateIndexStatements: c.CreateIndexStmts,
	}
}

func (f *SQLFeature) name() string  { return "sql" }
func (f *SQLFeature) enabled() bool { return f.Enabled }

func (f *SQLFeature) start(ctx *AppContext, fs *Features) error {
	var open bootstrap.Open
	switch f.SQLPackage {
	case SQLXPackage:
		open = bootstrap.SQLX(f.PasswordFile)
	case PGXPackage:
		open = bootstrap.PGX(f.PasswordFile)
	default:
		return fmt.Errorf("unknown sql package %q", f.SQLPackage)
	}
	return bootstrap.Run(ctx, bootstrap.Plan{
		Files: f.SQLFiles, Dirs: f.SQLDirs,
		Tables: f.CreateTableStatements, Indexes: f.CreateIndexStatements,
	}, open)
}
