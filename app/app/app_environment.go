package app

import (
	"context"
	"time"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/log"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
)

type TestEnvironment struct {
	Postgres bool
	Redis    bool
	Valkey   bool
}

func (e *TestEnvironment) Start(ctx context.Context) (func(), error) {
	l := log.NewLogger("testEnvironment")
	var conts []testcontainers.Container
	if e.Postgres {
		cont := containers.StartPostgres(ctx)
		conts = append(conts, cont)
	}

	if e.Redis {
		redisCont := containers.StartRedis(ctx)
		conts = append(conts, redisCont)
	}

	if e.Valkey {
		valkeyCont := containers.StartValkey(ctx)
		conts = append(conts, valkeyCont)
	}

	return func() {
		for _, c := range conts {
			timeout := time.Second * 30
			err := c.Stop(ctx, &timeout)
			if err != nil {
				l.Error("failed to stop container", zap.Error(err))
			}
		}
	}, nil
}
