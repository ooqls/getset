package app

import (
	"path"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/db/pgx"
	gosqlx "github.com/ooqls/getset/db/sqlx"
	"go.uber.org/zap"
)

func (a *app) _seed_pgx_files(ctx *AppContext, files []string) bool {
	l := ctx.L()
	sqlSeeded := len(files) > 0

	for _, file := range files {
		l.Debug("[Startup SQL] loading SQL file: " + file)
		if err := pgx.SeedPGXFile(ctx, file); err != nil {
			l.Error("failed to load file: "+file, zap.Error(err))
			sqlSeeded = false
		}
	}

	return sqlSeeded
}

func (a *app) _seed_sqlx_files(ctx *AppContext, files []string) bool {
	l := ctx.L()
	c := gosqlx.GetSQLX()
	sqlSeeded := len(files) > 0
	for _, file := range files {
		l.Debug("[Startup SQL] Loading sql file: " + file)
		if _, err := sqlx.LoadFile(c, file); err != nil {
			l.Error("[Startup SQL] failed to load file: "+file, zap.Error(err))
			sqlSeeded = false
		}
	}

	return sqlSeeded
}

func (a *app) _startup_sql(ctx *AppContext) error {
	l := ctx.L()
	sqlFiles := []string{}

	if len(a.features.SQL.SQLFiles) > 0 {
		sqlFiles = append(sqlFiles, a.features.SQL.SQLFiles...)
	}

	if len(a.features.SQL.SQLDirs) > 0 {
		for _, dir := range a.features.SQL.SQLDirs {
			sqlDir := path.Join(dir, "*.sql")
			files, err := filepath.Glob(sqlDir)
			if err != nil {
				l.Error("[Startup SQL] failed to glob SQL directory", zap.String("dir", sqlDir), zap.Error(err))
				continue
			}
			sqlFiles = append(sqlFiles, files...)
		}
	}

	if len(sqlFiles) > 0 {

		l.Debug("[Startup SQL] initializing SQL files", zap.Strings("sql_files", sqlFiles))

		if a.features.SQL.SQLPackage == SQLXPackage {
			err := gosqlx.InitDefault()
			if err != nil {
				l.Error("[Startup SQL] failed to initialize SQLX", zap.Error(err))
				return err
			}

			a.state.SQLSeeded = a._seed_sqlx_files(ctx, sqlFiles)
		} else if a.features.SQL.SQLPackage == PGXPackage {
			err := pgx.InitDefault()
			if err != nil {
				l.Error("[Startup SQL] failed to initialize PGX", zap.Error(err))
				return err
			}

			a.state.SQLSeeded = a._seed_pgx_files(ctx, sqlFiles)
		}
		l.Debug("[Startup SQL] SQL files initialized successfully")
	}

	tableStmts := []string{}
	indexStmts := []string{}

	if len(a.features.SQL.CreateTableStatements) > 0 {
		tableStmts = a.features.SQL.CreateTableStatements
	}

	if len(a.features.SQL.CreateIndexStatements) > 0 {
		indexStmts = a.features.SQL.CreateIndexStatements
	}

	if len(indexStmts) > 0 || len(tableStmts) > 0 {
		l.Debug("[Startup SQL] seeding with SQL statements")
		if a.features.SQL.SQLPackage == SQLXPackage {
			gosqlx.SeedSQLX(tableStmts, indexStmts)
		} else if a.features.SQL.SQLPackage == PGXPackage {
			pgx.SeedPGX(ctx, tableStmts, indexStmts)
		}

		l.Debug("[Startup SQL] finished seeding with SQL statements")
	} else {
		l.Debug("[Startup SQL] no SQL statements")
	}

	a.state.SQLInitialized = true
	return nil
}
