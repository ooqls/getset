package app

import (
	"fmt"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/db/valkey"
)

type CacheConfig struct {
	Enabled   bool   `yaml:"enabled"`
	CacheType string `yaml:"cache_type"`
}

type cacheType string

const (
	cacheTypeRedis  cacheType = "redis"
	cacheTypeMem    cacheType = "mem"
	cacheTypeValkey cacheType = "valkey"
)

type CacheOption func(*CacheFeature)

func WithCacheType(t string) CacheOption {
	return func(f *CacheFeature) { f.CacheType = cacheType(t) }
}

// WithRedis backs the cache with the Redis feature, which must be enabled.
func WithRedis() CacheOption {
	return func(f *CacheFeature) { f.CacheType = cacheTypeRedis }
}

// WithValkey backs the cache with the Valkey feature, which must be enabled.
func WithValkey() CacheOption {
	return func(f *CacheFeature) { f.CacheType = cacheTypeValkey }
}

func WithMem() CacheOption {
	return func(f *CacheFeature) { f.CacheType = cacheTypeMem }
}

// CacheFeature puts a cache store on the AppContext. An empty CacheType
// means in-memory.
type CacheFeature struct {
	Enabled   bool
	CacheType cacheType
}

func Cache(opts ...CacheOption) CacheFeature {
	f := CacheFeature{Enabled: true, CacheType: cacheTypeMem}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func cacheFromConfig(c CacheConfig) CacheFeature {
	return CacheFeature{Enabled: c.Enabled, CacheType: cacheType(c.CacheType)}
}

func (f *CacheFeature) name() string  { return "cache" }
func (f *CacheFeature) enabled() bool { return f.Enabled }

func (f *CacheFeature) start(ctx *AppContext, fs *Features) error {
	switch f.CacheType {
	case "", cacheTypeMem:
		ctx.WithCacheStore(cache.NewMemStore())
	case cacheTypeRedis:
		if !fs.Redis.Enabled {
			return fmt.Errorf("cache type redis needs the Redis feature enabled")
		}
		ctx.WithCacheStore(cache.NewRedisStore(redis.GetConnection(ctx)))
	case cacheTypeValkey:
		if !fs.Valkey.Enabled {
			return fmt.Errorf("cache type valkey needs the Valkey feature enabled")
		}
		ctx.WithCacheStore(cache.NewValkeyStore(valkey.GetConnection(ctx)))
	default:
		return fmt.Errorf("unknown cache type %q", f.CacheType)
	}
	return nil
}
