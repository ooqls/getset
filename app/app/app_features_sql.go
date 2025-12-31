package app

import "strings"

type sqlPackage string

// flags
var sqlFilesFlag string

// sql packages
const (
	PGXPackage  sqlPackage = "PGX"
	SQLXPackage sqlPackage = "SQLX"
)

// sql options
const (
	sql_createTableStatementsOpt string = "opt-create-table"
	sql_createIndexStatementsOpt string = "opt-create-index"
	sql_DirsOpt                  string = "opt-sql-dirs"
	sql_sqlFilesOpt              string = "opt-sql-files"
)

type sqlOpt struct {
	featureOpt
}

func WithCreateTableStatements(stmts []string) sqlOpt {
	return sqlOpt{
		featureOpt: featureOpt{
			key:   sql_createTableStatementsOpt,
			value: stmts,
		},
	}
}

func WithCreateIndexStatements(stmts []string) sqlOpt {
	return sqlOpt{
		featureOpt: featureOpt{
			key:   sql_createIndexStatementsOpt,
			value: stmts,
		},
	}
}

func WithSQLFiles(files []string) sqlOpt {
	return sqlOpt{
		featureOpt: featureOpt{
			key:   sql_sqlFilesOpt,
			value: files,
		},
	}
}

func WithSQLDirs(dirs []string) sqlOpt {
	return sqlOpt{
		featureOpt: featureOpt{
			key:   sql_DirsOpt,
			value: dirs,
		},
	}
}

func SQLX(opts ...sqlOpt) SQLFeature {
	return newSQLFeature(SQLXPackage, opts...)
}

func PGX(opts ...sqlOpt) SQLFeature {
	return newSQLFeature(SQLXPackage, opts...)
}

func newSQLFeature(sp sqlPackage, opts ...sqlOpt) SQLFeature {
	f := SQLFeature{
		Enabled:    true,
		SQLFiles:   strings.Split(sqlFilesFlag, ","),
		SQLPackage: sp,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}

type SQLFeature struct {
	Enabled               bool
	CreateTableStatements []string
	CreateIndexStatements []string
	SQLFiles              []string
	SQLDirs               []string
	SQLPackage            sqlPackage
}

func (f *SQLFeature) apply(opt sqlOpt) {
	switch opt.featureOpt.key {
	case sql_createIndexStatementsOpt:
		f.CreateIndexStatements = opt.featureOpt.value.([]string)
	case sql_createTableStatementsOpt:
		f.CreateTableStatements = opt.featureOpt.value.([]string)
	case sql_sqlFilesOpt:
		f.SQLFiles = opt.featureOpt.value.([]string)
	case sql_DirsOpt:
		f.SQLDirs = opt.featureOpt.value.([]string)
	}
}
