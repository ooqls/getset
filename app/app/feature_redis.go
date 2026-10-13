package app

import (
	"fmt"

	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Database     string `yaml:"database"`
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"password_file"`
}

type RedisOption func(*RedisFeature)

func WithRedisAddress(host string) RedisOption {
	return func(f *RedisFeature) { f.Host = host }
}

func WithRedisPort(port int) RedisOption {
	return func(f *RedisFeature) { f.Port = port }
}

func WithRedisDB(db int) RedisOption {
	return func(f *RedisFeature) { f.Database = fmt.Sprintf("%d", db) }
}

func WithRedisUsername(username string) RedisOption {
	return func(f *RedisFeature) { f.Username = username }
}

func WithRedisPasswordFile(path string) RedisOption {
	return func(f *RedisFeature) { f.PasswordFile = path }
}

// RedisFeature connects to Redis. It starts from the Registry's Redis entry
// when a Registry is loaded; any field set here overrides it.
type RedisFeature struct {
	Enabled      bool
	Host         string
	Port         int
	Database     string
	Username     string
	PasswordFile string
}

func Redis(opts ...RedisOption) RedisFeature {
	f := RedisFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func redisFromConfig(c RedisConfig) RedisFeature {
	return RedisFeature(c)
}

func (f *RedisFeature) name() string  { return "redis" }
func (f *RedisFeature) enabled() bool { return f.Enabled }

func (f *RedisFeature) start(ctx *AppContext, fs *Features) error {
	db := registry.Database{}
	if registry.IsSet() && registry.Get().Redis != nil {
		db = *registry.Get().Redis
	}
	if f.Host != "" {
		db.Host = f.Host
	}
	if f.Port != 0 {
		db.Port = f.Port
	}
	if f.Database != "" {
		db.Database = f.Database
	}
	if f.Username != "" {
		db.Auth.Username = f.Username
	}
	if f.PasswordFile != "" {
		db.Auth.PasswordFile = f.PasswordFile
	}
	if db.Database == "" {
		db.Database = "0"
	}
	if db.Host == "" {
		return fmt.Errorf("no redis host: set one on the Redis feature or in the registry")
	}

	ctx.L().Debug("[Startup Redis] connecting", zap.String("host", db.Host), zap.Int("port", db.Port))
	return redis.Init(db)
}
