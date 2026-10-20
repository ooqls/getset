package cache

import (
	"context"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type memStore struct {
	c *gocache.Cache
}

// NewMemStore returns a Store in process memory. Expired entries are swept
// every minute.
func NewMemStore() Store {
	return &memStore{c: gocache.New(gocache.NoExpiration, time.Minute)}
}

func (s *memStore) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := s.c.Get(key)
	if !ok {
		return nil, ErrNotFound
	}
	return v.([]byte), nil
}

func (s *memStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = gocache.NoExpiration
	}
	s.c.Set(key, value, ttl)
	return nil
}

func (s *memStore) Delete(ctx context.Context, key string) error {
	s.c.Delete(key)
	return nil
}

func (s *memStore) DeletePrefix(ctx context.Context, prefix string) error {
	for key := range s.c.Items() {
		if strings.HasPrefix(key, prefix) {
			s.c.Delete(key)
		}
	}
	return nil
}
