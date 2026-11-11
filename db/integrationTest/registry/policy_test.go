//go:build integration

package registrytest

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/postgres"
	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/db/sqlx"
	"github.com/ooqls/getset/db/valkey"
	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type endpoint struct {
	host string
	port int
}

type postgresSecurity struct {
	encrypted         bool
	clientCertificate bool
}

func writeFixture(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, data, 0600))
	return path
}

func certificates(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Registry policy test"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return writeFixture(t, "cert.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		writeFixture(t, "key.pem", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
}

func start(t *testing.T, request testcontainers.ContainerRequest, portName string) endpoint {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: request, Started: true})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		assert.NoError(t, c.Terminate(ctx))
	})
	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, portName)
	require.NoError(t, err)
	return endpoint{host, int(port.Num())}
}

func preserveRegistry(t *testing.T) {
	t.Helper()
	if registry.IsSet() {
		old := *registry.Get()
		t.Cleanup(func() { registry.Set(old) })
	}
}

func TestPostgresRegistryPolicy(t *testing.T) {
	preserveRegistry(t)
	cert, key := certificates(t)
	const password = "registry-policy-password"
	script := writeFixture(t, "start.sh", []byte("#!/bin/sh\nset -eu\nchown postgres:postgres /tmp/server.key\nchmod 600 /tmp/server.key\nexec docker-entrypoint.sh postgres -c ssl=on -c ssl_cert_file=/tmp/server.crt -c ssl_key_file=/tmp/server.key -c ssl_ca_file=/tmp/server.crt\n"))
	tlsEndpoint := start(t, testcontainers.ContainerRequest{
		Image: "postgres:17-alpine", ExposedPorts: []string{"5432/tcp"},
		Env:        map[string]string{"POSTGRES_PASSWORD": password},
		Entrypoint: []string{"sh", "/tmp/start.sh"},
		Files: []testcontainers.ContainerFile{
			{HostFilePath: cert, ContainerFilePath: "/tmp/server.crt", FileMode: 0644},
			{HostFilePath: key, ContainerFilePath: "/tmp/server.key", FileMode: 0600},
			{HostFilePath: script, ContainerFilePath: "/tmp/start.sh", FileMode: 0644},
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}, "5432/tcp")
	plainEndpoint := start(t, testcontainers.ContainerRequest{
		Image: "postgres:17-alpine", ExposedPorts: []string{"5432/tcp"},
		Env:        map[string]string{"POSTGRES_PASSWORD": password},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}, "5432/tcp")
	passwordFile := writeFixture(t, "password", []byte(password))
	passfile := writeFixture(t, "pgpass", []byte("*:*:*:*:"+password+"\n"))
	t.Setenv("PGPASSFILE", passfile)
	t.Setenv("PGPASSWORD", password)
	t.Setenv("PGSSLMODE", "prefer")
	for _, adapter := range []struct {
		name  string
		query func(context.Context, postgres.Options) (postgresSecurity, error)
	}{
		{name: "SQLX", query: func(ctx context.Context, opts postgres.Options) (postgresSecurity, error) {
			db, err := sqlx.Init(opts)
			if err != nil {
				return postgresSecurity{}, err
			}
			var security postgresSecurity
			err = db.QueryRowContext(ctx, "SELECT ssl, client_dn IS NOT NULL FROM pg_stat_ssl WHERE pid = pg_backend_pid()").Scan(&security.encrypted, &security.clientCertificate)
			return security, err
		}},
		{name: "PGX", query: func(ctx context.Context, opts postgres.Options) (postgresSecurity, error) {
			db, err := pgx.Init(ctx, opts)
			if err != nil {
				return postgresSecurity{}, err
			}
			var security postgresSecurity
			err = db.QueryRow(ctx, "SELECT ssl, client_dn IS NOT NULL FROM pg_stat_ssl WHERE pid = pg_backend_pid()").Scan(&security.encrypted, &security.clientCertificate)
			return security, err
		}},
	} {
		t.Run(adapter.name, func(t *testing.T) {
			for _, tt := range []struct {
				name                  string
				endpoint              endpoint
				tls                   *registry.TLSConfig
				emptyPassword         bool
				wantTLS               bool
				wantClientCertificate bool
				wrongServerName       bool
				wantError             bool
			}{
				{name: "trusted TLS", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, CaPath: cert}, wantTLS: true},
				{name: "client certificate", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, CaPath: cert, CertPath: cert, KeyPath: key}, wantTLS: true, wantClientCertificate: true},
				{name: "wrong peer hostname", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, CaPath: cert}, wrongServerName: true, wantError: true},
				{name: "untrusted TLS", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true}, wantError: true},
				{name: "explicit verification bypass", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, InsecureSkipTLSVerify: true}, wantTLS: true},
				{name: "disabled is plaintext even when server offers TLS", endpoint: tlsEndpoint},
				{name: "no plaintext fallback", endpoint: plainEndpoint, tls: &registry.TLSConfig{Enabled: true, InsecureSkipTLSVerify: true}, wantError: true},
				{name: "no ambient password", endpoint: plainEndpoint, emptyPassword: true, wantError: true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					auth := registry.Auth{Enabled: true, Username: "postgres", PasswordFile: passwordFile}
					if tt.emptyPassword {
						auth.PasswordFile = ""
					}
					registry.Set(registry.Registry{Postgres: &registry.Database{Database: "postgres", Server: registry.Server{Host: tt.endpoint.host, Port: tt.endpoint.port, TLS: tt.tls, Auth: auth}}})
					opts, err := postgres.GetRegistryOptions(nil)
					require.NoError(t, err)
					if tt.wrongServerName {
						opts.Tls.ServerName = "wrong.invalid"
					}
					ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
					defer cancel()
					security, err := adapter.query(ctx, opts)
					if tt.wantError {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					assert.Equal(t, tt.wantTLS, security.encrypted)
					assert.Equal(t, tt.wantClientCertificate, security.clientCertificate)
				})
			}
		})
	}
}

func TestRedisAndValkeyRegistryPolicy(t *testing.T) {
	preserveRegistry(t)
	cert, key := certificates(t)
	const password = "registry-policy-password"
	tlsEndpoint := start(t, testcontainers.ContainerRequest{
		Image: "redis:7-alpine", ExposedPorts: []string{"6379/tcp"},
		Cmd: []string{"redis-server", "--port", "0", "--tls-port", "6379", "--tls-cert-file", "/tmp/server.crt", "--tls-key-file", "/tmp/server.key", "--tls-ca-cert-file", "/tmp/server.crt", "--tls-auth-clients", "no", "--requirepass", password},
		Files: []testcontainers.ContainerFile{
			{HostFilePath: cert, ContainerFilePath: "/tmp/server.crt", FileMode: 0644},
			{HostFilePath: key, ContainerFilePath: "/tmp/server.key", FileMode: 0644},
		},
		WaitingFor: wait.ForListeningPort("6379/tcp"),
	}, "6379/tcp")
	plainEndpoint := start(t, testcontainers.ContainerRequest{Image: "redis:7-alpine", ExposedPorts: []string{"6379/tcp"}, WaitingFor: wait.ForListeningPort("6379/tcp")}, "6379/tcp")
	passwordFile := writeFixture(t, "password", []byte(password))
	for _, adapter := range []struct {
		name string
		ping func(context.Context, registry.Database) error
	}{
		{name: "Redis", ping: func(ctx context.Context, db registry.Database) error {
			if err := redis.Init(db); err != nil {
				return err
			}
			return redis.GetConnection(ctx).Ping(ctx).Err()
		}},
		{name: "Valkey", ping: func(ctx context.Context, db registry.Database) error {
			if err := valkey.Init(&db); err != nil {
				return err
			}
			return valkey.GetConnection(ctx).Do(ctx, valkey.GetConnection(ctx).B().Ping().Build()).Error()
		}},
	} {
		t.Run(adapter.name, func(t *testing.T) {
			for _, tt := range []struct {
				name      string
				endpoint  endpoint
				tls       *registry.TLSConfig
				disabled  bool
				wantError bool
			}{
				{name: "trusted TLS", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, CaPath: cert}},
				{name: "untrusted TLS", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true}, wantError: true},
				{name: "explicit verification bypass", endpoint: tlsEndpoint, tls: &registry.TLSConfig{Enabled: true, InsecureSkipTLSVerify: true}},
				{name: "no plaintext fallback", endpoint: plainEndpoint, tls: &registry.TLSConfig{Enabled: true, InsecureSkipTLSVerify: true}, wantError: true},
				{name: "disabled auth skips configured credentials", endpoint: plainEndpoint, disabled: true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					auth := registry.Auth{Enabled: true, Username: "default", PasswordFile: passwordFile}
					if tt.disabled {
						auth.Enabled = false
						auth.PasswordFile = filepath.Join(t.TempDir(), "missing")
					}
					db := registry.Database{Database: "0", Server: registry.Server{Host: tt.endpoint.host, Port: tt.endpoint.port, TLS: tt.tls, Auth: auth}}
					ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
					defer cancel()
					err := adapter.ping(ctx, db)
					if tt.wantError {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
				})
			}
		})
	}
	// Feature credentials are explicit opt-ins even when Registry auth is off.
	for _, name := range []string{"Redis", "Valkey"} {
		t.Run(name+" Feature override", func(t *testing.T) {
			db := &registry.Database{Database: "0", Server: registry.Server{Host: tlsEndpoint.host, Port: tlsEndpoint.port, TLS: &registry.TLSConfig{Enabled: true, CaPath: cert}, Auth: registry.Auth{PasswordFile: "missing"}}}
			registry.Set(registry.Registry{Redis: db, Valkey: db})
			features := app.Features{}
			if name == "Redis" {
				features.Redis = app.Redis(app.WithRedisPasswordFile(passwordFile))
			} else {
				features.Valkey = app.Valkey(app.WithValkeyPasswordFile(passwordFile))
			}
			a := app.New("Registry policy", features)
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			a.OnStartup(func(ctx *app.AppContext) error {
				defer cancel()
				if name == "Redis" {
					return redis.GetConnection(ctx).Ping(ctx).Err()
				}
				return valkey.GetConnection(ctx).Do(ctx, valkey.GetConnection(ctx).B().Ping().Build()).Error()
			})
			require.NoError(t, a.Run(ctx), fmt.Sprintf("%s Feature must enable explicit credentials", name))
			assert.False(t, db.Auth.Enabled, "Feature overrides must not mutate Registry entries")
		})
	}
}
