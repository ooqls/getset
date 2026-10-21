package postgres

import (
	"net/url"
	"testing"

	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionStringEscapesCredentials(t *testing.T) {
	opt := Options{Host: "localhost", Port: 5432, User: "us@er", Pw: "p/a+s:s@w ord=", DB: "app"}

	u, err := url.Parse(opt.ConnectionString())
	require.NoError(t, err)

	pw, _ := u.User.Password()
	assert.Equal(t, "postgres", u.Scheme)
	assert.Equal(t, "us@er", u.User.Username())
	assert.Equal(t, "p/a+s:s@w ord=", pw)
	assert.Equal(t, "localhost:5432", u.Host)
	assert.Equal(t, "/app", u.Path)
}

func TestGetRegistryOptionsDatabase(t *testing.T) {
	t.Cleanup(func() { registry.Set(registry.Registry{}) })

	registry.Set(registry.Registry{Postgres: &registry.Database{Database: "app"}})
	assert.Equal(t, "app", GetRegistryOptions().DB)

	registry.Set(registry.Registry{Postgres: &registry.Database{}})
	assert.Equal(t, "postgres", GetRegistryOptions().DB)
}
