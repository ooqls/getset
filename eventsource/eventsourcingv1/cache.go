package eventsourcingv1

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	lib_store "github.com/eko/gocache/lib/v4/store"
	ekogocache "github.com/eko/gocache/store/go_cache/v4"
	"github.com/eko/gocache/store/redis/v4"
	goredis "github.com/ooqls/getset/db/redis"
	gocache "github.com/patrickmn/go-cache"
)

type JsonCache[T any] struct {
	cache.CacheInterface[string]
}

func NewJsonCache[T any](c cache.CacheInterface[string], opts ...lib_store.Option) *JsonCache[T] {
	return &JsonCache[T]{c}
}

func (c *JsonCache[T]) Get(ctx context.Context, key any) (T, error) {
	var dest T
	bytes, err := c.CacheInterface.Get(ctx, key)
	if err != nil {
		return dest, err
	}

	err = json.Unmarshal([]byte(bytes), &dest)
	if err != nil {
		return dest, err
	}

	return dest, nil
}

func (c *JsonCache[T]) Set(ctx context.Context, key any, value T, options ...lib_store.Option) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.CacheInterface.Set(ctx, key, string(bytes), options...)
}

func NewGoCache[T any](opts ...lib_store.Option) cache.CacheInterface[T] {

	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	cachedStore := ekogocache.NewGoCache(gocacheClient, opts...)
	cached := cache.New[string](cachedStore)

	return NewJsonCache[T](cached)
}

func NewRedisCache[T any](opts ...lib_store.Option) cache.CacheInterface[T] {
	c := goredis.GetConnection(context.Background())
	redisStore := redis.NewRedis(c, opts...)
	redisCached := cache.New[string](redisStore)
	jsonCache := NewJsonCache[T](redisCached, opts...)
	return jsonCache
}
