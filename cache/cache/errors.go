package cache

import (
	"errors"

	"github.com/eko/gocache/lib/v4/store"
	"github.com/redis/go-redis/v9"
)

func IsCacheMissErr(err error) bool {
	return errors.Is(err, redis.Nil) || errors.Is(err, store.NotFound{})
}
