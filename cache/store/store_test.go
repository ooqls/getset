package store

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/redis"
	"github.com/stretchr/testify/assert"
)

type OtherObj struct {
	Id uuid.UUID
	Ts time.Time
	V  string
}

type MyAlias = OtherObj

type Obj struct {
	V     string
	Ts    time.Time
	id    uuid.UUID
	Other MyAlias
}

func TestRedisStore(t *testing.T) {
	ctx := context.Background()
	redisContainer := containers.StartRedis(context.Background())
	defer func() {
		if err := redisContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate redis container: %v", err)
		}
	}()
	Register(Obj{}, MyAlias{})

	var store GenericInterface = NewRedisStore("test", *redis.GetConnection(ctx), 450*time.Second)

	err := store.Get(context.Background(), "key", &Obj{})
	assert.NotNilf(t, err, "expected cache miss error, got %v", err)
	assert.True(t, cache.IsCacheMissErr(err))

	obj := Obj{V: "value", Ts: time.Now(), id: uuid.New()}
	err = store.Set(context.Background(), "key", obj)
	assert.Nil(t, err)

	var updatedObj Obj
	err = store.Get(context.Background(), "key", &updatedObj)
	assert.Nil(t, err)
	assert.Equal(t, obj.V, updatedObj.V)

	err = store.Update(context.Background(), "key", func(fn func(target any) error) (any, error) {
		var obj Obj

		err := fn(&obj)
		assert.Nil(t, err)

		obj.V = "updated value"
		return obj, nil
	})
	assert.Nil(t, err)

	err = store.Get(context.Background(), "key", &updatedObj)
	assert.Nil(t, err)
	assert.Equal(t, "updated value", updatedObj.V)

	err = store.Get(context.Background(), uuid.New().String(), &updatedObj)
	assert.NotNil(t, err)
	assert.True(t, cache.IsCacheMissErr(err))
}

func TestMemStore(t *testing.T) {
	store := NewMemStore("test", 10*time.Second)

	obj := Obj{V: "value"}
	err := store.Set(context.Background(), "key", obj)
	assert.Nil(t, err)

	var updatedObj Obj
	err = store.Get(context.Background(), "key", &updatedObj)
	assert.Nil(t, err)
	assert.Equal(t, obj.V, updatedObj.V)

	err = store.Update(context.Background(), "key", func(fn func(target any) error) (any, error) {
		var obj Obj

		err := fn(&obj)
		assert.Nil(t, err)

		obj.V = "updated value"
		return obj, nil
	})
	assert.Nil(t, err)

	err = store.Get(context.Background(), "key", &updatedObj)
	assert.Nil(t, err)
	assert.Equal(t, "updated value", updatedObj.V)
}

func TestMemStore_Get_NotFound(t *testing.T) {
	store := NewMemStore("test", 10*time.Second)

	var obj Obj
	err := store.Get(context.Background(), "key", &obj)
	assert.True(t, cache.IsCacheMissErr(err))
}
