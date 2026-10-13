package valkey

import (
	"context"
	"fmt"
	"time"

	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/registry"
	"github.com/valkey-io/valkey-go"
)

var c = conn.New("valkey", func() (valkey.Client, error) {
	reg := registry.Get()
	if reg.Valkey == nil {
		return nil, fmt.Errorf("valkey not found in registry")
	}
	return open(reg.Valkey)
}, func(vc valkey.Client) error {
	vc.Close()
	return nil
})

func open(db *registry.Database) (valkey.Client, error) {
	cliOpts := valkey.ClientOption{
		InitAddress:      []string{fmt.Sprintf("%s:%d", db.Host, db.Port)},
		ConnWriteTimeout: 5 * time.Second,
	}

	if db.TLS != nil && db.TLS.Enabled {
		tlsCfg, err := db.TLS.TLSConfig()
		if err != nil {
			return nil, err
		}
		cliOpts.TLSConfig = tlsCfg
	}

	if db.Auth.Enabled {
		pw, err := db.Server.ResolvePassword()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve valkey password: %v", err)
		}
		cliOpts.Username = db.Auth.Username
		cliOpts.Password = pw
	}

	return valkey.NewClient(cliOpts)
}

// Init opens a client for db and makes it the one GetConnection returns.
func Init(db *registry.Database) error {
	return c.Init(func() (valkey.Client, error) { return open(db) })
}

// InitDefault opens the client from the registry unless one is already open.
func InitDefault() error {
	return c.InitDefault()
}

// GetConnection returns the client, opening it from the registry on first
// use. It panics if the client can't be opened. The client reconnects on its
// own.
func GetConnection(ctx context.Context) valkey.Client {
	return c.MustGet()
}
