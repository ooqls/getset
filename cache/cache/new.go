package cache

import (
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
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
