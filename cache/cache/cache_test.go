package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Obj map[string]string

func TestMemCache_map(t *testing.T) {

	o := Obj{
		"a": "10",
	}

	c := New[Obj](NewMemCache())
	assert.Nilf(t, c.Set(context.Background(), "key", o), "should not get an error when setting cache")

	cachedO, err := c.Get(context.Background(), "key")
	derefO := *cachedO
	assert.Nilf(t, err, "should not get an error when getting from cache")
	assert.Equal(t, derefO["a"], o["a"])

	err = c.Delete(context.Background(), "key")
	assert.Nilf(t, err, "should not get an error when deleting from cache")

	_, err = c.Get(context.Background(), "key")
	assert.NotNilf(t, err, "should get an error when getting from cache")
	assert.Truef(t, IsCacheMissErr(err), "should get a cache miss error")
}

func TestMemCache_bool(t *testing.T) {
	c := New[bool](NewMemCache())
	assert.Nilf(t, c.Set(context.Background(), "key", true), "should not get an error when setting cache")

	val, err := c.Get(context.Background(), "key")
	assert.Nilf(t, err, "should not get an error when getting from cache")

	assert.Truef(t, *val, "retrieved value should be true")
}
