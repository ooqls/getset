package redis

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

var redisContainer testcontainers.Container

func TestMain(m *testing.M) {
	// Initialize the Redis container
	redisContainer = containers.StartRedis(context.Background())
	defer func() {
		if err := redisContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate redis container: %v", err)
		}
	}()

	m.Run()

}

func TestConnectRedis(t *testing.T) {
	ctx := context.Background()
	con := redis.GetConnection(ctx)
	res := con.Ping(ctx)
	assert.Nilf(t, res.Err(), "should not have gotten an error when pinging redis")
}

func TestRedisCacheRoundTrip(t *testing.T) {
	ctx := context.Background()
	c := cache.NewGenericCache("test", cache.NewRedisCache(*redis.GetConnection(ctx), time.Minute))
	require.NoError(t, c.Set(ctx, "key", "value"))

	var got string
	require.NoError(t, c.Get(ctx, "key", &got))
	assert.Equal(t, "value", got)
}

// Runs last: it stops the shared container.
func TestGetConnectionReturnsWhenRedisIsDown(t *testing.T) {
	ctx := context.Background()
	redis.GetConnection(ctx)

	timeout := time.Second * 5
	require.NoError(t, redisContainer.Stop(ctx, &timeout))

	done := make(chan struct{})
	go func() {
		redis.GetConnection(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("GetConnection should return when redis is down, not spin")
	}
}
