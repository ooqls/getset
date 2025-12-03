package cache

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/eko/gocache/lib/v4/cache"
)

func New[V any](c *cache.Cache[[]byte]) *Cache[V] {
	return &Cache[V]{
		c: cache.Cache[[]byte](*c),
	}
}

type Cache[T any] struct {
	c cache.Cache[[]byte]
}

func (c *Cache[T]) Set(ctx context.Context, key string, val T) error {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	return c.c.Set(ctx, key, buff.Bytes())
}

func (c *Cache[T]) Get(ctx context.Context, key string) (*T, error) {
	payload, err := c.c.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(payload))
	var obj T
	err = dec.Decode(&obj)
	return &obj, err

}

func (c *Cache[T]) Clear(ctx context.Context) error {
	return c.c.Clear(ctx)
}

func (c *Cache[T]) Delete(ctx context.Context, key string) error {
	return c.c.Delete(ctx, key)
}
