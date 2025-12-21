package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"
)

func Register(types ...any) {
	for _, t := range types {
		if t == nil {
			continue
		}

		gob.Register(t)
	}
}

//go:generate mockgen -source=store.go -destination=store_mock.go -package=store GenericInterface
type GenericInterface interface {
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string, target any) error
	Update(ctx context.Context, key string, fn func(func(target any) error) (any, error)) error
	Delete(ctx context.Context, key string) error
}

type MemStore struct {
	c *cache.GenericCache
}

func NewMemStore(storeName string, ttl time.Duration) GenericInterface {
	return &MemStore{
		c: cache.NewGenericCache(storeName, cache.NewMemCache()),
	}
}

func (s *MemStore) Set(ctx context.Context, key string, value any) error {
	return s.c.Set(ctx, key, value)
}

func (s *MemStore) Get(ctx context.Context, key string, target any) error {
	return s.c.Get(ctx, key, target)
}

func (s *MemStore) Update(ctx context.Context, key string, fn func(func(target any) error) (any, error)) error {
	target, err := fn(func(target any) error {
		return s.c.Get(ctx, key, target)
	})
	if err != nil {
		return err
	}

	return s.c.Set(ctx, key, target)
}

func (s *MemStore) Delete(ctx context.Context, key string) error {
	return s.c.Delete(ctx, key)
}

type RedisStore struct {
	db       redis.Client
	ttl      time.Duration
	storeKey string
}

func encode(value any) ([]byte, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func decode(data []byte, target any) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	return dec.Decode(target)
}

func NewRedisStore(storeKey string, db redis.Client, ttl time.Duration) GenericInterface {
	return &RedisStore{db: db, ttl: ttl, storeKey: storeKey}
}

func (s *RedisStore) getKey(key string) string {
	return fmt.Sprintf("%s/%s", s.storeKey, key)
}

func (s *RedisStore) Set(ctx context.Context, key string, value any) error {
	buff, err := encode(value)
	if err != nil {
		return err
	}

	return s.db.Set(ctx, s.getKey(key), buff, s.ttl).Err()
}

func (s *RedisStore) Get(ctx context.Context, key string, target any) error {
	res, err := s.db.Get(ctx, s.getKey(key)).Result()
	if err != nil {
		return err
	}

	return decode([]byte(res), target)
}

func (s *RedisStore) Update(ctx context.Context, key string, fn func(func(target any) error) (any, error)) error {
	return s.db.Watch(ctx, func(tx *redis.Tx) error {
		res, err := tx.Get(ctx, s.getKey(key)).Result()
		if err != nil {
			return err
		}

		target, err := fn(func(target any) error {
			return decode([]byte(res), target)
		})
		if err != nil {
			return err
		}

		buff, err := encode(target)
		if err != nil {
			return err
		}

		_, err = tx.Set(ctx, s.getKey(key), buff, s.ttl).Result()
		return err
	}, key)
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.db.Del(ctx, s.getKey(key)).Err()
}

type ValkeyStore struct {
	valkey   valkey.Client
	ttl      time.Duration
	storeKey string
}

func NewValkeyStore(storeKey string, valkey valkey.Client, ttl time.Duration) GenericInterface {
	return &ValkeyStore{valkey: valkey, ttl: ttl, storeKey: storeKey}
}

func (s *ValkeyStore) getKey(key string) string {
	return fmt.Sprintf("%s/%s", s.storeKey, key)
}

func (s *ValkeyStore) Set(ctx context.Context, key string, value any) error {
	buff, err := encode(value)
	if err != nil {
		return err
	}

	return s.valkey.Do(ctx, s.valkey.B().Set().Key(s.getKey(key)).Value(string(buff)).Build()).Error()
}

func (s *ValkeyStore) Get(ctx context.Context, key string, target any) error {
	res := s.valkey.Do(ctx, s.valkey.B().Get().Key(s.getKey(key)).Build())
	if res.Error() != nil {
		return res.Error()
	}

	buff, err := res.AsBytes()
	if err != nil {
		return err
	}

	return decode(buff, target)
}

func (s *ValkeyStore) Delete(ctx context.Context, key string) error {
	return s.valkey.Do(ctx, s.valkey.B().Del().Key(s.getKey(key)).Build()).Error()
}

func (s *ValkeyStore) Update(ctx context.Context, key string, fn func(func(target any) error) (any, error)) error {
	err := s.valkey.Dedicated(func(dedicated valkey.DedicatedClient) error {
		dedicated.Do(ctx, dedicated.B().Watch().Key(s.getKey(key)).Build())
		res := dedicated.Do(ctx, dedicated.B().Get().Key(s.getKey(key)).Build())
		if res.Error() != nil {
			return res.Error()
		}

		buff, err := res.AsBytes()
		if err != nil {
			return err
		}

		target, err := fn(func(target any) error {
			return decode(buff, target)
		})
		if err != nil {
			return err
		}

		buff, err = encode(target)
		if err != nil {
			return err
		}

		return dedicated.Do(ctx, dedicated.B().Set().Key(s.getKey(key)).Value(string(buff)).Build()).Error()
	})
	if err != nil {
		return err
	}

	return nil
}
