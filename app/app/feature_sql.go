package app

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/postgres"
	gosqlx "github.com/ooqls/getset/db/sqlx"
	"go.uber.org/zap"
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
// Postgres entry, then runs the given SQL files and statements. A file that
// fails to load is logged, not fatal.
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
	if f.SQLPackage != SQLXPackage && f.SQLPackage != PGXPackage {
		return fmt.Errorf("unknown sql package %q", f.SQLPackage)
	}

	files := f.sqlFiles(ctx)
	if len(files) > 0 {
		if err := f.connect(); err != nil {
			return err
		}
		f.seedFiles(ctx, files)
	}

	if len(f.CreateTableStatements) > 0 || len(f.CreateIndexStatements) > 0 {
		ctx.L().Debug("[Startup SQL] seeding with SQL statements")
		if f.SQLPackage == SQLXPackage {
			gosqlx.SeedSQLX(f.CreateTableStatements, f.CreateIndexStatements)
		} else {
			pgx.SeedPGX(ctx, f.CreateTableStatements, f.CreateIndexStatements)
		}
	}

	return nil
}

// sqlFiles returns SQLFiles followed by every *.sql file in SQLDirs.
func (f *SQLFeature) sqlFiles(ctx *AppContext) []string {
	files := append([]string{}, f.SQLFiles...)
	for _, dir := range f.SQLDirs {
		pattern := path.Join(dir, "*.sql")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			ctx.L().Error("[Startup SQL] failed to glob SQL directory", zap.String("dir", pattern), zap.Error(err))
			continue
		}
		files = append(files, matches...)
	}
	return files
}

func (f *SQLFeature) connect() error {
	var opts *postgres.Options
	if f.PasswordFile != "" {
		b, err := os.ReadFile(f.PasswordFile)
		if err != nil {
			return fmt.Errorf("failed to read password file: %w", err)
		}
		o := postgres.GetRegistryOptions()
		o.Pw = string(b)
		opts = &o
	}

	if f.SQLPackage == SQLXPackage {
		if opts != nil {
			_, err := gosqlx.Init(*opts)
			return err
		}
		return gosqlx.InitDefault()
	}

	if opts != nil {
		_, err := pgx.Init(context.Background(), *opts)
		return err
	}
	return pgx.InitDefault()
}

func (f *SQLFeature) seedFiles(ctx *AppContext, files []string) {
	l := ctx.L()
	l.Debug("[Startup SQL] loading SQL files", zap.Strings("sql_files", files))
	for _, file := range files {
		var err error
		if f.SQLPackage == SQLXPackage {
			_, err = sqlx.LoadFile(gosqlx.GetSQLX(), file)
		} else {
			err = pgx.SeedPGXFile(ctx, file)
		}
		if err != nil {
			l.Error("[Startup SQL] failed to load file", zap.String("file", file), zap.Error(err))
		}
	}
}
