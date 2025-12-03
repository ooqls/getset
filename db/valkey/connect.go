package valkey

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ooqls/getset/registry"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/zap"
)

var valkeyCfg *registry.Database
var m sync.Mutex = sync.Mutex{}
var c valkey.Client

func Init(db *registry.Database) error {
	m.Lock()
	defer m.Unlock()

	return initValkey(db)
}

func InitDefault() error {
	m.Lock()
	defer m.Unlock()

	if c != nil {
		return nil
	}

	reg := registry.Get()
	if reg.Valkey == nil {
		return fmt.Errorf("valkey not found in registry")
	}

	return initValkey(reg.Valkey)
}

func initValkey(db *registry.Database) error {
	if c != nil {
		return nil
	}

	valkeyCfg = db
	cliOpts := valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", db.Host, db.Port)},
	}

	if db.TLS != nil && db.TLS.Enabled {
		tlsCfg, err := db.TLS.TLSConfig()
		if err != nil {
			return err
		}

		cliOpts.TLSConfig = tlsCfg
	}

	if db.Auth.Enabled {
		cliOpts.Username = db.Auth.Username
		cliOpts.Password = db.Auth.Password
	}

	var err error
	c, err = valkey.NewClient(cliOpts)
	return err
}

func GetConnection(ctx context.Context) valkey.Client {
	if c == nil {
		err := InitDefault()
		if err != nil {
			panic(fmt.Errorf("failed to init valkey: %v", err))
		}
	} else {
		if err := c.Do(ctx, c.B().Ping().Build()).Error(); err != nil {
			c = nil
			for err != nil {
				m.Lock()
				err := initValkey(valkeyCfg)
				m.Unlock()

				if err != nil {
					zap.L().Error("valkey client was disconnected, reconnecting...")
					time.Sleep(time.Second * 3)
				}
			}
		}
	}

	return c
}
