package app

import (
	"fmt"

	"github.com/ooqls/getset/registry"
)

type redisOpt struct {
	featureOpt
}

const (
	redis_addressOpt  string = "opt-redis-address"
	redis_portOpt     string = "opt-redis-port"
	redis_dbOpt       string = "opt-redis-db"
	redis_userOpt     string = "opt-redis-user"
	redis_passwordOpt string = "opt-redis-password"
)

func WithRedisAddress(address string) redisOpt {
	return redisOpt{featureOpt: featureOpt{key: redis_addressOpt, value: address}}
}

func WithRedisPort(port int) redisOpt {
	return redisOpt{featureOpt: featureOpt{key: redis_portOpt, value: port}}
}

func WithRedisDB(db int) redisOpt {
	return redisOpt{featureOpt: featureOpt{key: redis_dbOpt, value: db}}
}

func WithUsername(username string) redisOpt {
	return redisOpt{featureOpt: featureOpt{key: redis_userOpt, value: username}}
}

func WithPassword(password string) redisOpt {
	return redisOpt{featureOpt: featureOpt{key: redis_passwordOpt, value: password}}
}

type RedisFeature struct {
	Enabled bool
	redisDB registry.Database
}

func (f *RedisFeature) apply(opt redisOpt) {
	switch opt.key {
	case redis_addressOpt:
		f.redisDB.Server.Host = opt.value.(string)
	case redis_portOpt:
		f.redisDB.Server.Port = opt.value.(int)
	case redis_dbOpt:
		f.redisDB.Database = fmt.Sprintf("%d", opt.value.(int))
	case redis_userOpt:
		f.redisDB.Auth.Username = opt.value.(string)
	case redis_passwordOpt:
		f.redisDB.Auth.Password = opt.value.(string)

	}
}

func Redis(opts ...redisOpt) RedisFeature {
	r := registry.Get()

	redisDb := r.Redis
	f := RedisFeature{
		Enabled: true,
		redisDB: *redisDb,
	}

	for _, opt := range opts {
		f.apply(opt)
	}

	return f
}
