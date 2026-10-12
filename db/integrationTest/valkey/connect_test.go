package valkey

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/valkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestValkeyCacheRoundTrip(t *testing.T) {
	ctx := context.Background()
	c := cache.NewGenericCache("test", cache.NewValkeyCache(valkey.GetConnection(ctx), time.Minute))
	require.NoError(t, c.Set(ctx, "key", "value"))

	var got string
	require.NoError(t, c.Get(ctx, "key", &got))
	assert.Equal(t, "value", got)
}
