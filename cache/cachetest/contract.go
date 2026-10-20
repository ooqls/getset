// Package cachetest is the contract every cache.Store must satisfy. Each
// adapter's tests call Run with a constructor for that adapter.
package cachetest

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type user struct {
	Name  string
	Roles []string
}

// Run checks store's behaviour through cache.Cache. Namespaces are derived
// from the test name, so Run can share a store with other tests.
func Run(t *testing.T, store cache.Store) {
	ctx := context.Background()
	ns := func(t *testing.T) string {
		return strings.ReplaceAll(t.Name(), "/", ".")
	}

	t.Run("miss", func(t *testing.T) {
		c := cache.New(store, ns(t), time.Minute)
		var got user
		err := c.Get(ctx, "missing", &got)
		assert.ErrorIs(t, err, cache.ErrNotFound)
		assert.True(t, cache.IsCacheMissErr(err))
	})

	t.Run("set then get", func(t *testing.T) {
		c := cache.New(store, ns(t), time.Minute)
		want := user{Name: "ada", Roles: []string{"admin"}}
		require.NoError(t, c.Set(ctx, "u1", want))

		var got user
		require.NoError(t, c.Get(ctx, "u1", &got))
		assert.Equal(t, want, got)
	})

	t.Run("overwrite", func(t *testing.T) {
		c := cache.New(store, ns(t), time.Minute)
		require.NoError(t, c.Set(ctx, "k", "first"))
		require.NoError(t, c.Set(ctx, "k", "second"))

		var got string
		require.NoError(t, c.Get(ctx, "k", &got))
		assert.Equal(t, "second", got)
	})

	t.Run("delete", func(t *testing.T) {
		c := cache.New(store, ns(t), time.Minute)
		require.NoError(t, c.Set(ctx, "k", "v"))
		require.NoError(t, c.Delete(ctx, "k"))
		assert.ErrorIs(t, c.Get(ctx, "k", new(string)), cache.ErrNotFound)
		assert.NoError(t, c.Delete(ctx, "never-set"), "deleting a missing key is not an error")
	})

	t.Run("entries expire after ttl", func(t *testing.T) {
		c := cache.New(store, ns(t), 300*time.Millisecond)
		require.NoError(t, c.Set(ctx, "k", "v"))
		require.NoError(t, c.Get(ctx, "k", new(string)))

		assert.Eventually(t, func() bool {
			return c.Get(ctx, "k", new(string)) == cache.ErrNotFound
		}, 3*time.Second, 50*time.Millisecond)
	})

	t.Run("zero ttl never expires", func(t *testing.T) {
		c := cache.New(store, ns(t), 0)
		require.NoError(t, c.Set(ctx, "k", "v"))
		time.Sleep(500 * time.Millisecond)
		assert.NoError(t, c.Get(ctx, "k", new(string)))
	})

	t.Run("namespaces are isolated", func(t *testing.T) {
		a := cache.New(store, ns(t)+".a", time.Minute)
		b := cache.New(store, ns(t)+".b", time.Minute)
		require.NoError(t, a.Set(ctx, "k", "from a"))
		require.NoError(t, b.Set(ctx, "k", "from b"))

		var got string
		require.NoError(t, a.Get(ctx, "k", &got))
		assert.Equal(t, "from a", got)
	})

	t.Run("clear removes only its own namespace", func(t *testing.T) {
		// Glob characters in the namespace must not widen what Clear deletes.
		target := cache.New(store, ns(t)+".a*[x]?", time.Minute)
		prefixTwin := cache.New(store, ns(t)+".a*[x]?-other", time.Minute)
		globMatch := cache.New(store, ns(t)+".abxZ", time.Minute)

		for i := range 250 {
			require.NoError(t, target.Set(ctx, fmt.Sprint(i), i))
		}
		require.NoError(t, prefixTwin.Set(ctx, "k", "keep"))
		require.NoError(t, globMatch.Set(ctx, "k", "keep"))

		require.NoError(t, target.Clear(ctx))

		for _, i := range []int{0, 125, 249} {
			assert.ErrorIs(t, target.Get(ctx, fmt.Sprint(i), new(int)), cache.ErrNotFound)
		}
		assert.NoError(t, prefixTwin.Get(ctx, "k", new(string)), "a namespace sharing the prefix must survive")
		assert.NoError(t, globMatch.Get(ctx, "k", new(string)), "a namespace matching the glob must survive")
	})
}
