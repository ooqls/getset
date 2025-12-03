package sqlx

import (
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/ooqls/getset/db/postgres"
	log "github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var db *sqlx.DB
var l *zap.Logger = log.NewLogger("db")
var m sync.Mutex = sync.Mutex{}

func GetSQLX() *sqlx.DB {
	var err error
	if db != nil {
		err = postgres.Retry(func() error {
			return db.Ping()
		})
		if err != nil {
			l.Error("failed to ensure connection", zap.Error(err))
		}
	}

	if db == nil || err != nil {
		if db != nil {
			db.Close()
		}

		db, err = connectSqlx(postgres.GetRegistryOptions())
		if err != nil {
			panic(err)
		}
	}

	l.Info("SQLX database connection established")
	return db
}

func connectSqlx(opt postgres.Options) (*sqlx.DB, error) {
	conStr := fmt.Sprintf("host=%s password=%s port=%d user=%s dbname=%s sslmode=disable",
		opt.Host, opt.Pw, opt.Port, opt.User, opt.DB)
	l.Info("connection string", zap.String("con", conStr))
	dbCon, err := sqlx.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql: %v", err)
	}

	db = dbCon

	return dbCon, nil
}

func Init(opt postgres.Options) (*sqlx.DB, error) {
	m.Lock()
	defer m.Unlock()

	err := initSQLX(opt)
	return db, err
}

func InitDefault() error {
	m.Lock()
	defer m.Unlock()

	err := initSQLX(postgres.GetRegistryOptions()) // Updated to use GetOptions()
	if err != nil {
		l.Error("failed to initialize default options", zap.Error(err))
		return err
	}
	l.Info("default options initialized successfully")
	return nil
}

func initSQLX(opt postgres.Options) error {
	var err error
	if db != nil {
		err = postgres.Retry(func() error {
			return db.Ping()
		})
		if err != nil {
			l.Error("failed to ensure connection", zap.Error(err))
		}
	}

	if db == nil || err != nil {
		if db != nil {
			db.Close()
		}

		db, err = connectSqlx(opt)
		if err != nil {
			return err
		}
	}

	l.Info("SQLX database connection established")
	return nil
}
