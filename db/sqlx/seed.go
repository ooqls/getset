package sqlx

import (
	"os"

	"go.uber.org/zap"
)

func SeedSQLXFile(sqlFilePath string) error {
	db := GetSQLX()
	sqlB, err := os.ReadFile(sqlFilePath)
	if err != nil {
		l.Error("failed to read sql file", zap.String("file", sqlFilePath), zap.Error(err))
		return err
	}

	_, err = db.Exec(string(sqlB))
	if err != nil {
		l.Error("failed to prepare sql statement", zap.Error(err))
		return err
	}

	return nil
}

func SeedSQLX(tableStmts []string, indexStmts []string) {
	// db.Get().Exec(tables.GetDropTableStmt())

	for _, stmt := range tableStmts {
		l.Info("executing table statement", zap.String("stmt", stmt))
		_, err := GetSQLX().Exec(stmt) // Changed to GetSQLX() for SQLX usage
		if err != nil {
			l.Error("failed to execute table statement", zap.String("stmt", stmt), zap.Error(err))
			panic(err)
		}
	}

	for _, stmt := range indexStmts {
		l.Info("executing index statement", zap.String("stmt", stmt))
		_, err := GetSQLX().Exec(stmt) // Changed to GetSQLX() for SQLX usage
		if err != nil {
			l.Error("failed to execute index statement", zap.String("stmt", stmt), zap.Error(err))
			panic(err)
		}
	}
}
