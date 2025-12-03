package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetGeneric(t *testing.T) {

	c := NewGenericCache("test", NewMemCache())

	err := c.Set(context.Background(), "key", "value")
	assert.Nil(t, err)

	var val string
	err = c.Get(context.Background(), "key", &val)
	assert.Nil(t, err)
	assert.Equal(t, "value", val)

	type Obj struct {
		V string
	}

	o := Obj{
		V: "10",
	}

	err = c.Set(context.Background(), "key", o)
	assert.Nil(t, err)

	var val2 Obj
	err = c.Get(context.Background(), "key", &val2)
	assert.Nil(t, err)
	assert.Equal(t, o, val2)
}
