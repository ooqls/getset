package valkey

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/cache/cachetest"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/valkey"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	timeout := time.Second * 30
	cont := containers.StartValkey(context.Background())

	code := m.Run()
	cont.Stop(context.Background(), &timeout)
	os.Exit(code)
}

func TestConnectValkey(t *testing.T) {
	ctx := context.Background()
	con := valkey.GetConnection(ctx)
	assert.NotNilf(t, con, "valkey client should not be nil")
}

func TestValkeyCacheStore(t *testing.T) {
	cachetest.Run(t, cache.NewValkeyStore(valkey.GetConnection(context.Background())))
}
