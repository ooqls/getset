package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/redis/go-redis/v9"
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

func NewRedisStore(storeKey string, db redis.Client, ttl time.Duration) GenericInterface {
	return &RedisStore{db: db, ttl: ttl, storeKey: storeKey}
}

func (s *RedisStore) getKey(key string) string {
	return fmt.Sprintf("%s/%s", s.storeKey, key)
}

func (s *RedisStore) Set(ctx context.Context, key string, value any) error {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(value)
	if err != nil {
		return err
	}

	return s.db.Set(ctx, s.getKey(key), buff.Bytes(), s.ttl).Err()
}

func (s *RedisStore) Get(ctx context.Context, key string, target any) error {
	res, err := s.db.Get(ctx, s.getKey(key)).Result()
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(bytes.NewBuffer([]byte(res)))
	return dec.Decode(target)
}

func (s *RedisStore) Update(ctx context.Context, key string, fn func(func(target any) error) (any, error)) error {
	return s.db.Watch(ctx, func(tx *redis.Tx) error {
		res, err := tx.Get(ctx, s.getKey(key)).Result()
		if err != nil {
			return err
		}

		target, err := fn(func(target any) error {
			dec := gob.NewDecoder(bytes.NewReader([]byte(res)))
			return dec.Decode(target)
		})
		if err != nil {
			return err
		}

		var buff bytes.Buffer
		enc := gob.NewEncoder(&buff)
		err = enc.Encode(target)
		if err != nil {
			return err
		}

		_, err = tx.Set(ctx, s.getKey(key), buff.Bytes(), s.ttl).Result()
		return err
	}, key)
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.db.Del(ctx, s.getKey(key)).Err()
}
