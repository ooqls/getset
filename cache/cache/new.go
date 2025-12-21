package cache

import (
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	valkey_store "github.com/eko/gocache/store/valkey/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"
)

func NewMemCache() *cache.Cache[[]byte] {
	memCache := gocache.New(time.Minute*5, time.Minute*10)
	memStore := gocache_store.NewGoCache(memCache)
	return cache.New[[]byte](memStore)
}

func NewRedisCache(rc redis.Client, ttl time.Duration) *cache.Cache[[]byte] {
	redisStore := redis_store.NewRedis(rc, store.WithExpiration(ttl))
	return cache.New[[]byte](redisStore)
}

func NewValkeyCache(valkey valkey.Client, ttl time.Duration) *cache.Cache[[]byte] {
	valkeyStore := valkey_store.NewValkey(valkey, store.WithExpiration(ttl))
	return cache.New[[]byte](valkeyStore)
}
