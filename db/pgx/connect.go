package pgx

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ooqls/getset/db/postgres"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("pgx")
var db *pgxpool.Pool
var m sync.Mutex = sync.Mutex{}

func GetPGX() *pgxpool.Pool {
	m.Lock()
	defer m.Unlock()

	if db != nil {
		if err := db.Ping(context.Background()); err != nil {
			l.Error("pgxpool ping failed, replacing pool", zap.Error(err))
			db.Close()
			db = nil
		}
	}

	if db == nil {
		var err error
		db, err = connectPgx(context.Background(), postgres.GetRegistryOptions())
		if err != nil {
			panic(err)
		}
	}

	l.Info("PGX database connection established")
	return db
}

func connectPgx(ctx context.Context, opt postgres.Options) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, opt.ConnectionString())
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func Init(ctx context.Context, opt postgres.Options) (*pgxpool.Pool, error) {
	m.Lock()
	defer m.Unlock()

	return connectPgx(ctx, opt)
}

func InitDefault() error {
	m.Lock()
	defer m.Unlock()

	if db != nil {
		return nil
	}

	var err error
	db, err = connectPgx(context.Background(), postgres.GetRegistryOptions())
	if err != nil {
		l.Error("failed to initialize default options", zap.Error(err))
		return err
	}
	l.Info("default options initialized successfully")
	return nil
}
