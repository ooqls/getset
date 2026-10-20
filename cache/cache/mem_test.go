package cache_test

import (
	"testing"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/cache/cachetest"
)

func TestMemStore(t *testing.T) {
	cachetest.Run(t, cache.NewMemStore())
}
