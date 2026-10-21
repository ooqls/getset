package postgres

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("postgres")

type Options struct {
	Host string
	Port int
	User string
	DB   string
	Pw   string
	Tls  *tls.Config
}

// ConnectionString returns a postgres:// URL. The user, password and database
// are escaped, so they may contain any character.
func (opt *Options) ConnectionString() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(opt.User, opt.Pw),
		Host:   fmt.Sprintf("%s:%d", opt.Host, opt.Port),
		Path:   "/" + opt.DB,
	}
	return u.String()
}

// defaultDB is used when the registry's Postgres entry names no database.
const defaultDB = "postgres"

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
	pw, err := reg.Postgres.Server.ResolvePassword()
	if err != nil {
		l.Error("failed to resolve postgres password", zap.Error(err))
		panic(err)
	}

	db := reg.Postgres.Database
	if db == "" {
		db = defaultDB
	}

	return Options{
		Host: reg.Postgres.Host,
		Port: reg.Postgres.Port,
		User: reg.Postgres.Auth.Username,
		Pw:   pw,
		DB:   db,
		Tls:  tlsCfg,
	}
}
