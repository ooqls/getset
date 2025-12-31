package cache

import (
	"context"

	"github.com/eko/gocache/lib/v4/cache"
)

type MultiKeyObject interface {
	GetKeys() []string
}

type MultiKeyCache[V MultiKeyObject] struct {
	c cache.Cache[V]
}

func NewMultiKeyCache[V MultiKeyObject](c cache.Cache[V]) *MultiKeyCache[V] {
	return &MultiKeyCache[V]{c: c}
}

func (m *MultiKeyCache[V]) Get(ctx context.Context, key string) (V, error) {
	return m.c.Get(ctx, key)
}

func (m *MultiKeyCache[V]) Set(ctx context.Context, value V) error {
	keys := value.GetKeys()
	for _, key := range keys {
		if err := m.c.Set(ctx, key, value); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiKeyCache[V]) Delete(ctx context.Context, key string) error {
	return m.c.Delete(ctx, key)
}
