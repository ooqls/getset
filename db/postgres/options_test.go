package postgres

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Registry is process-global, so these cases deliberately run sequentially.
func TestGetRegistryOptions(t *testing.T) {
	oldSet := registry.IsSet()
	if oldSet {
		old := *registry.Get()
		t.Cleanup(func() { registry.Set(old) })
	} else {
		_, err := GetRegistryOptions(nil)
		require.ErrorContains(t, err, "Registry is not initialized")
	}
	dir := t.TempDir()
	passwordFile := filepath.Join(dir, "password")
	require.NoError(t, os.WriteFile(passwordFile, []byte("password\n"), 0600))
	override := "override\n"
	empty := ""
	for _, tt := range []struct {
		name         string
		db           *registry.Database
		override     *string
		wantPassword string
		wantError    string
	}{
		{name: "missing entry", wantError: "entry is missing"},
		{name: "plaintext", db: &registry.Database{Server: registry.Server{Auth: registry.Auth{Enabled: true, PasswordFile: passwordFile}}}, wantPassword: "password\n"},
		{name: "missing password", db: &registry.Database{Server: registry.Server{Auth: registry.Auth{Enabled: true, PasswordFile: filepath.Join(dir, "missing")}}}, wantError: "resolve password"},
		{name: "disabled skips missing password", db: &registry.Database{Server: registry.Server{Auth: registry.Auth{Username: "ignored", PasswordFile: filepath.Join(dir, "missing")}}}},
		{name: "override bypasses missing password", db: &registry.Database{Server: registry.Server{Auth: registry.Auth{PasswordFile: filepath.Join(dir, "missing")}}}, override: &override, wantPassword: override},
		{name: "empty override is explicit", db: &registry.Database{Server: registry.Server{Auth: registry.Auth{PasswordFile: passwordFile}}}, override: &empty},
		{name: "TLS error", db: &registry.Database{Server: registry.Server{TLS: &registry.TLSConfig{Enabled: true, CaPath: filepath.Join(dir, "missing-ca")}}}, wantError: "resolve TLS"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			registry.Set(registry.Registry{Postgres: tt.db})
			opts, err := GetRegistryOptions(tt.override)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPassword, opts.Pw)
		})
	}
}

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
	opts, err := GetRegistryOptions(nil)
	require.NoError(t, err)
	assert.Equal(t, "app", opts.DB)

	registry.Set(registry.Registry{Postgres: &registry.Database{}})
	opts, err = GetRegistryOptions(nil)
	require.NoError(t, err)
	assert.Equal(t, "postgres", opts.DB)
}
