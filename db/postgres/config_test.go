package postgres_test

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"github.com/ooqls/getset/db/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Config(t *testing.T) {
	// These deliberately conflicting settings must not affect either adapter.
	t.Setenv("PGSERVICE", "")
	t.Setenv("PGSSLMODE", "prefer")
	t.Setenv("PGSSLROOTCERT", filepath.Join(t.TempDir(), "missing-ca"))
	t.Setenv("PGSSLCERT", filepath.Join(t.TempDir(), "missing-cert"))
	t.Setenv("PGSSLKEY", filepath.Join(t.TempDir(), "missing-key"))
	t.Setenv("PGUSER", "ambient-user")
	t.Setenv("PGPASSWORD", "ambient-password")
	t.Setenv("PGHOST", "ambient-host")
	t.Setenv("PGPORT", "9999")
	passfile := filepath.Join(t.TempDir(), "pgpass")
	require.NoError(t, os.WriteFile(passfile, []byte("*:*:*:*:ambient-password\n"), 0600))
	t.Setenv("PGPASSFILE", passfile)
	for _, tt := range []struct {
		name string
		tls  *tls.Config
	}{
		{name: "plaintext"},
		{name: "verified TLS", tls: &tls.Config{MinVersion: tls.VersionTLS12}},
		{name: "explicit insecure TLS", tls: &tls.Config{InsecureSkipVerify: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			opt := postgres.Options{Host: "db.example", Port: 5432, DB: "app", Tls: tt.tls}
			cfg, err := opt.Config()
			require.NoError(t, err)
			conn := cfg.ConnConfig
			assert.Equal(t, opt.Host, conn.Host)
			assert.Equal(t, uint16(opt.Port), conn.Port)
			assert.Equal(t, "app", conn.Database)
			assert.Empty(t, conn.User)
			assert.Empty(t, conn.Password)
			assert.Empty(t, conn.Fallbacks, "TLS must never downgrade")
			if tt.tls == nil {
				assert.Nil(t, conn.TLSConfig)
			} else {
				require.NotNil(t, conn.TLSConfig)
				assert.Equal(t, "db.example", conn.TLSConfig.ServerName)
				assert.Equal(t, tt.tls.InsecureSkipVerify, conn.TLSConfig.InsecureSkipVerify)
				conn.TLSConfig.ServerName = "changed"
				assert.Empty(t, tt.tls.ServerName, "driver configuration must own its TLS snapshot")
			}
		})
	}
}

func TestOptions_ConfigRejectsAmbientService(t *testing.T) {
	t.Setenv("PGSERVICE", "ambient-service")
	_, err := (postgres.Options{Host: "localhost", Port: 5432}).Config()
	require.ErrorContains(t, err, "PGSERVICE is not supported")
}

func TestOptions_ConfigRejectsInvalidAddress(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		options postgres.Options
	}{
		{name: "missing host", options: postgres.Options{Port: 5432}},
		{name: "multiple hosts", options: postgres.Options{Host: "one,two", Port: 5432}},
		{name: "invalid port", options: postgres.Options{Host: "localhost", Port: 65536}},
		{name: "TLS over unix socket", options: postgres.Options{Host: "/socket", Port: 5432, Tls: &tls.Config{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.options.Config()
			require.Error(t, err)
		})
	}
}
