package cache

import (
	"context"
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
	return cache.New[[]byte](bytesStore{redisStore})
}

func NewValkeyCache(valkey valkey.Client, ttl time.Duration) *cache.Cache[[]byte] {
	valkeyStore := valkey_store.NewValkey(valkey, store.WithExpiration(ttl))
	return cache.New[[]byte](bytesStore{valkeyStore})
}

// bytesStore adapts a store whose Get returns strings (Redis, Valkey) to the
// []byte values cache.Cache[[]byte] expects. Without it, Cache.Get's type
// assertion fails and it silently returns an empty value.
type bytesStore struct {
	store.StoreInterface
}

func (s bytesStore) Get(ctx context.Context, key any) (any, error) {
	v, err := s.StoreInterface.Get(ctx, key)
	return stringToBytes(v), err
}

func (s bytesStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	v, ttl, err := s.StoreInterface.GetWithTTL(ctx, key)
	return stringToBytes(v), ttl, err
}

func stringToBytes(v any) any {
	if s, ok := v.(string); ok {
		return []byte(s)
	}
	return v
}
