package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// scanCount is how many keys DeletePrefix asks Redis/Valkey for per SCAN.
const scanCount = 100

type redisStore struct {
	rc *redis.Client
}

// NewRedisStore returns a Store backed by rc.
func NewRedisStore(rc *redis.Client) Store {
	return &redisStore{rc: rc}
}

func (s *redisStore) Get(ctx context.Context, key string) ([]byte, error) {
	b, err := s.rc.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	return b, err
}

func (s *redisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.rc.Set(ctx, key, value, ttl).Err()
}

func (s *redisStore) Delete(ctx context.Context, key string) error {
	return s.rc.Del(ctx, key).Err()
}

func (s *redisStore) DeletePrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	for {
		keys, next, err := s.rc.Scan(ctx, cursor, globPrefix(prefix), scanCount).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.rc.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

// globPrefix returns a SCAN MATCH pattern for keys starting with prefix,
// escaping glob characters so prefix is matched literally.
func globPrefix(prefix string) string {
	var b strings.Builder
	for _, r := range prefix {
		switch r {
		case '*', '?', '[', ']', '\\':
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	b.WriteRune('*')
	return b.String()
}
