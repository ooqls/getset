package postgres

import (
	"crypto/tls"
	"fmt"

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

func (opt *Options) ConnectionString() string {
	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", opt.User, opt.Pw, opt.Host, opt.Port, opt.DB)
	return url
}

var dbName string = "postgres"

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

	return Options{
		Host: reg.Postgres.Host,
		Port: reg.Postgres.Port,
		User: reg.Postgres.Auth.Username,
		Pw:   pw,
		DB:   dbName,
		Tls:  tlsCfg,
	}, nil
}
