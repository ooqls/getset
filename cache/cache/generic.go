package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/eko/gocache/lib/v4/cache"
)

func NewGenericCache(cacheKey string, c *cache.Cache[[]byte]) *GenericCache {
	return &GenericCache{
		cacheKey: cacheKey,
		c:        c,
	}
}

type GenericCache struct {
	cacheKey string
	c        *cache.Cache[[]byte]
}

func (c *GenericCache) getKey(localKey string) string {
	return fmt.Sprintf("%s/%s", c.cacheKey, localKey)
}

func (c *GenericCache) Set(ctx context.Context, key string, val any) error {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(val)
	if err != nil {
		return err
	}
	k := c.getKey(key)
	return c.c.Set(ctx, k, buff.Bytes())
}

func (c *GenericCache) Get(ctx context.Context, key string, target any) error {
	b, err := c.c.Get(ctx, c.getKey(key))
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	return dec.Decode(target)
}

func (c *GenericCache) Delete(ctx context.Context, key string) error {
	return c.c.Delete(ctx, c.getKey(key))
}

func (c *GenericCache) Clear(ctx context.Context) error {
	return c.c.Clear(ctx)
}
