package cache

import (
	"crypto/tls"

	"github.com/ooqls/getset/log"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("cache")

type RedisOptions struct {
	Username string
	Password string
	Addr     string
	DB       int
	TLS      *TLS
}

func (o *RedisOptions) GetGoRedisOptions() *goredis.Options {
	var tlsCfg *tls.Config
	var err error

	if o.TLS != nil {
		tlsCfg, err = o.TLS.GetTLSConfig()
		if err != nil {
			l.Info("failed to get tls config", zap.Error(err))
		}
	}

	return &goredis.Options{
		Addr:      o.Addr,
		Username:  o.Username,
		Password:  o.Password,
		DB:        o.DB,
		TLSConfig: tlsCfg,
	}
}
