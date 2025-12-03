package pgx

import (
	"context"
	"os"
)

func SeedPGXFile(ctx context.Context, sqlFile string) error {
	db := GetPGX()
	b, err := os.ReadFile(sqlFile)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, string(b))
	return err
}

func SeedPGX(ctx context.Context, tableStmts []string, indexStmts []string) {
	// db.Get().Exec(tables.GetDropTableStmt())

	for _, stmt := range tableStmts {
		_, err := GetPGX().Exec(ctx, stmt) // Changed to GetSQLX() for SQLX usage
		if err != nil {
			panic(err)
		}
	}

	for _, stmt := range indexStmts {
		_, err := GetPGX().Exec(ctx, stmt) // Changed to GetSQLX() for SQLX usage
		if err != nil {
			panic(err)
		}
	}
}
