// Package cache keeps gob-encoded values in a Store under a namespace, with
// one TTL per namespace. Stores (in-memory, Redis, Valkey) only move bytes;
// key layout, encoding, expiry and clearing are decided here once.
package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"time"
)

// ErrNotFound is returned by Get when a key is missing or expired.
var ErrNotFound = errors.New("cache: not found")

// IsCacheMissErr reports whether err means the key wasn't in the cache.
func IsCacheMissErr(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// Register makes concrete types known to gob, for values stored behind
// interfaces.
func Register(types ...any) {
	for _, t := range types {
		if t != nil {
			gob.Register(t)
		}
	}
}

// Store is where cached bytes live. Keys passed to a Store are already
// namespaced.
type Store interface {
	// Get returns ErrNotFound for a missing or expired key.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores value for ttl; a ttl of 0 means no expiry.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes key. Deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error
	// DeletePrefix removes every key that starts with prefix, taken literally.
	DeletePrefix(ctx context.Context, prefix string) error
}

// Cache is one namespace of a Store. Values are gob-encoded, so a Cache can
// hold values of different types under different keys.
type Cache struct {
	store  Store
	prefix string
	ttl    time.Duration
}

// New returns the namespace in store. Entries expire after ttl; 0 means
// never.
func New(store Store, namespace string, ttl time.Duration) *Cache {
	return &Cache{store: store, prefix: namespace + "/", ttl: ttl}
}

// Get decodes the value at key into target, which must be a pointer. It
// returns ErrNotFound on a miss.
func (c *Cache) Get(ctx context.Context, key string, target any) error {
	b, err := c.store.Get(ctx, c.prefix+key)
	if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewReader(b)).Decode(target)
}

// Set stores value at key for the Cache's TTL.
func (c *Cache) Set(ctx context.Context, key string, value any) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}
	return c.store.Set(ctx, c.prefix+key, buf.Bytes(), c.ttl)
}

// Delete removes key.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, c.prefix+key)
}

// Clear removes every key in this namespace and nothing else.
func (c *Cache) Clear(ctx context.Context) error {
	return c.store.DeletePrefix(ctx, c.prefix)
}
