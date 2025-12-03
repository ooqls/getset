package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ooqls/getset/registry"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var pool *redis.Client
var redisCfg *registry.Database
var m sync.Mutex = sync.Mutex{}

func Init(db registry.Database) error {
	m.Lock()
	defer m.Unlock()

	return initRedis(&db)
}

func initRedis(db *registry.Database) error {
	redisCfg = db

	if pool != nil {
		return nil
	}

	var tlsCfg *tls.Config
	if db.TLS != nil {
		var err error
		tlsCfg, err = db.TLS.TLSConfig()
		if err != nil {
			return fmt.Errorf("failed to load tls config for redis: %v", err)
		}
	}

	redisDb, err := strconv.Atoi(db.Database)
	if err != nil {
		return fmt.Errorf("failed to convert database string to int: %v", err)
	}

	pool = redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", db.Host, db.Port),
		Password:  db.Auth.Password,
		Username:  db.Auth.Username,
		DB:        redisDb,
		TLSConfig: tlsCfg,
		PoolSize:  3,
	})

	return nil

}
func InitDefault() error {
	m.Lock()
	defer m.Unlock()

	reg := registry.Get()
	if reg.Redis == nil {
		return fmt.Errorf("no redis server found in registry")
	}

	redisCfg = reg.Redis

	return initRedis(redisCfg)
}

func GetConnection(ctx context.Context) *redis.Client {
	if pool == nil {
		if err := InitDefault(); err != nil {
			panic(err)
		}

	} else {
		if err := pool.Ping(ctx).Err(); err != nil {
			pool = nil
			for err != nil {
				if err := initRedis(redisCfg); err != nil {
					zap.L().Error("redis client was diconnected, reconnecting...")
					time.Sleep(time.Second * 3)
				}
			}
		}
	}
	return pool
}
