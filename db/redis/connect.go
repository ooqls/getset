package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/registry"
	"github.com/redis/go-redis/v9"
)

var c = conn.New("redis", func() (*redis.Client, error) {
	reg := registry.Get()
	if reg.Redis == nil {
		return nil, fmt.Errorf("no redis server found in registry")
	}
	return open(*reg.Redis)
}, func(rc *redis.Client) error {
	return rc.Close()
})

func open(db registry.Database) (*redis.Client, error) {
	var tlsCfg *tls.Config
	if db.TLS != nil {
		var err error
		tlsCfg, err = db.TLS.TLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load tls config for redis: %v", err)
		}
	}

	redisDb, err := strconv.Atoi(db.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to convert database string to int: %v", err)
	}

	pw, err := db.Server.ResolvePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve redis password: %v", err)
	}

	return redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%d", db.Host, db.Port),
		Password:  pw,
		Username:  db.Auth.Username,
		DB:        redisDb,
		TLSConfig: tlsCfg,
		PoolSize:  3,
	}), nil
}

// Init opens a client for db and makes it the one GetConnection returns.
func Init(db registry.Database) error {
	return c.Init(func() (*redis.Client, error) { return open(db) })
}

// InitDefault opens the client from the registry unless one is already open.
func InitDefault() error {
	return c.InitDefault()
}

// GetConnection returns the client, opening it from the registry on first
// use. It panics if the client can't be opened. The client reconnects on its
// own, so a returned client may fail commands while Redis is unreachable.
func GetConnection(ctx context.Context) *redis.Client {
	return c.MustGet()
}
