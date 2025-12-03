package postgres

import (
	"crypto/tls"
	"fmt"

	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

type Options struct {
	Host string
	Port int
	User string
	DB   string
	Pw   string
	Tls  *tls.Config
}

func (opt *Options) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", opt.User, opt.Pw, opt.Host, opt.Port, opt.DB)
}

var dbName string = "postgres"

func GetRegistryOptions() Options {
	reg := registry.Get()
	var tlsCfg *tls.Config
	var err error
	if reg.Postgres.TLS != nil {
		tlsCfg, err = reg.Postgres.TLS.TLSConfig()
		if err != nil {
			l.Error("failed to get TLS config", zap.Error(err))
			panic(err)
		}
	}

	return Options{
		Host: reg.Postgres.Host,
		Port: reg.Postgres.Port,
		User: reg.Postgres.Auth.Username,
		Pw:   reg.Postgres.Auth.Password,
		DB:   dbName,
		Tls:  tlsCfg,
	}
}
