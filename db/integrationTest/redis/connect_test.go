package redis

import (
	"context"
	"log"
	"testing"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/redis"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Initialize the Redis container
	redisContainer := containers.StartRedis(context.Background())
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
