package valkey

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/registry"
	"github.com/valkey-io/valkey-go"
)

var c = conn.New("valkey", func() (valkey.Client, error) {
	if !registry.IsSet() {
		return nil, fmt.Errorf("valkey Registry is not initialized")
	}
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
	if db == nil {
		return nil, fmt.Errorf("missing valkey Registry entry")
	}
	resolved, err := db.Server.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve valkey connection: %w", err)
	}
	cliOpts := valkey.ClientOption{
		InitAddress:      []string{net.JoinHostPort(resolved.Host, strconv.Itoa(resolved.Port))},
		ConnWriteTimeout: 5 * time.Second,
		TLSConfig:        resolved.TLS,
		Username:         resolved.Username,
		Password:         resolved.Password,
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
