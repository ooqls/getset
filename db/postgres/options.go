package postgres

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"

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
		Host:   net.JoinHostPort(opt.Host, strconv.Itoa(opt.Port)),
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
	resolved, err := reg.Postgres.Server.Resolve(passwordOverride)
	if err != nil {
		return Options{}, fmt.Errorf("resolve Postgres connection: %w", err)
	}

	db := reg.Postgres.Database
	if db == "" {
		db = defaultDB
	}

	return Options{
		Host: resolved.Host,
		Port: resolved.Port,
		User: resolved.Username,
		Pw:   resolved.Password,
		DB:   db,
		Tls:  resolved.TLS,
	}, nil
}
