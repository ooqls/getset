package postgres

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/ooqls/getset/registry"
)

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

// GetRegistryOptions resolves Postgres settings. A nonnil passwordOverride
// bypasses Registry password resolution, including decryption prerequisites.
func GetRegistryOptions(passwordOverride *string) (Options, error) {
	if !registry.IsSet() {
		return Options{}, fmt.Errorf("Postgres Registry is not initialized")
	}
	reg := registry.Get()
	if reg.Postgres == nil {
		return Options{}, fmt.Errorf("Postgres Registry entry is missing")
	}
	var tlsCfg *tls.Config
	var err error
	if reg.Postgres.TLS != nil {
		tlsCfg, err = reg.Postgres.TLS.TLSConfig()
		if err != nil {
			return Options{}, fmt.Errorf("resolve Postgres TLS: %w", err)
		}
	}
	var pw string
	if passwordOverride != nil {
		pw = *passwordOverride
	} else {
		pw, err = reg.Postgres.Server.ResolvePassword()
		if err != nil {
			return Options{}, fmt.Errorf("resolve Postgres password: %w", err)
		}
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
	}, nil
}
