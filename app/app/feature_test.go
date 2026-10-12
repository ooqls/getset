package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	dbsqlite "github.com/ooqls/getset/db/sqlite"
	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// startFeatures runs the startup phase for every enabled feature in fs and
// stops everything when the test ends.
func startFeatures(t *testing.T, fs *Features) *AppContext {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	appCtx := NewAppContext(ctx, zap.NewNop())
	t.Cleanup(func() {
		cancel()
		appCtx.wg.Wait()
	})
	for _, f := range fs.ordered() {
		if f.enabled() {
			require.NoError(t, f.start(appCtx, fs), "start %s", f.name())
		}
	}
	return appCtx
}

// serveFeatures runs the serve phase for every enabled server in fs.
func serveFeatures(t *testing.T, ctx *AppContext, fs *Features) {
	t.Helper()
	for _, f := range fs.ordered() {
		if s, ok := f.(server); ok && s.enabled() {
			require.NoError(t, s.serve(ctx, fs), "serve %s", s.name())
		}
	}
}

func get(t *testing.T, client *http.Client, url string) (int, string) {
	t.Helper()
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(b)
}

func port(addr net.Addr) int {
	return addr.(*net.TCPAddr).Port
}

func tlsFeature(t *testing.T) TLSFeature {
	caPath, certPath, keyPath := writeTLS(t)
	return TLS(WithServerCAFile(caPath), WithServerCert(certPath), WithServerKey(keyPath))
}

func TestTLS_ServesGinAndHTTPOverTLS(t *testing.T) {
	fs := &Features{
		TLS:    tlsFeature(t),
		Gin:    Gin(WithGinPort(0)),
		HTTP:   HTTP(),
		Health: Health(WithHealthPath("/health")),
	}
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	client := fs.TLS.Client()
	for _, addr := range []net.Addr{fs.Gin.Addr(), fs.HTTP.Addr()} {
		code, body := get(t, client, fmt.Sprintf("https://localhost:%d/health", port(addr)))
		assert.Equal(t, http.StatusOK, code)
		assert.JSONEq(t, `{"status":"ok"}`, body)
	}

	code, _ := get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/health", port(fs.HTTP.Addr())))
	assert.Equal(t, http.StatusBadRequest, code, "plain HTTP against a TLS listener should be rejected")
}

func TestTLS_BytesTakePrecedenceOverFiles(t *testing.T) {
	caPath, certPath, keyPath := writeTLS(t)
	caBytes, _ := os.ReadFile(caPath)
	certBytes, _ := os.ReadFile(certPath)
	keyBytes, _ := os.ReadFile(keyPath)

	fs := &Features{TLS: TLS(
		WithCaBytes(caBytes),
		WithServerCertBytes(certBytes),
		WithServerKeyBytes(keyBytes),
		WithServerCert("/does/not/exist"),
		WithServerKey("/does/not/exist"),
	)}
	startFeatures(t, fs)
	assert.Len(t, fs.TLS.ServerConfig().Certificates, 1)
}

func TestTLS_FailsWithoutKey(t *testing.T) {
	_, certPath, _ := writeTLS(t)
	f := TLS(WithServerCert(certPath))
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorContains(t, err, "server key")
}

func TestGrpc_ServesOverTLS(t *testing.T) {
	srv := grpc.NewServer()
	healthpb.RegisterHealthServer(srv, health.NewServer())
	fs := &Features{TLS: tlsFeature(t), Grpc: GRPC(WithGrpcServer(srv), WithGrpcPort(0))}
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	creds := credentials.NewTLS(&tls.Config{RootCAs: fs.TLS.ServerConfig().RootCAs})
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port(fs.Grpc.Addr())), grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	defer conn.Close()

	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status)
}

func TestGrpc_ServesWithoutTLS(t *testing.T) {
	srv := grpc.NewServer()
	healthpb.RegisterHealthServer(srv, health.NewServer())
	fs := &Features{Grpc: GRPC(WithGrpcServer(srv), WithGrpcPort(0))}
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", port(fs.Grpc.Addr())), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	_, err = healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
	assert.NoError(t, err)
}

func TestHealth_FailingCheckReportsUnhealthy(t *testing.T) {
	fs := &Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"))}
	fs.Health.Check = func() bool { return false }
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	code, _ := get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/health", port(fs.HTTP.Addr())))
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.False(t, fs.Health.healthy())
}

func TestDocs_ServesContentOnBothRouters(t *testing.T) {
	fs := &Features{
		Gin:  Gin(WithGinPort(0)),
		HTTP: HTTP(),
		Docs: Docs(WithDocsApiPath("/api/docs"), WithDocsContent("<h1>docs</h1>")),
	}
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	for _, addr := range []net.Addr{fs.Gin.Addr(), fs.HTTP.Addr()} {
		code, body := get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/api/docs", port(addr)))
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "<h1>docs</h1>", body)
	}
}

func TestDocs_ServesDirectoryOnBothRouters(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte("openapi: 3.0.0"), 0644))
	fs := &Features{
		Gin:  Gin(WithGinPort(0)),
		HTTP: HTTP(),
		Docs: Docs(WithDocsApiPath("/api/docs"), WithDocsPath(dir)),
	}
	ctx := startFeatures(t, fs)
	serveFeatures(t, ctx, fs)

	for _, addr := range []net.Addr{fs.Gin.Addr(), fs.HTTP.Addr()} {
		code, body := get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/api/docs/openapi.yaml", port(addr)))
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "openapi: 3.0.0", body)
	}
}

func TestGin_CorsAppliesToRoutesAddedAfterStart(t *testing.T) {
	fs := &Features{Gin: Gin(WithGinPort(0), WithGinCors(cors.Config{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{"GET"},
	}))}
	ctx := startFeatures(t, fs)
	fs.Gin.Engine.GET("/hello", func(c *gin.Context) { c.String(http.StatusOK, "hi") })
	serveFeatures(t, ctx, fs)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/hello", port(fs.Gin.Addr())), nil)
	req.Header.Set("Origin", "https://example.com")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, "https://example.com", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestCache_DefaultsToMemory(t *testing.T) {
	fs := &Features{Cache: Cache()}
	ctx := startFeatures(t, fs)
	assert.NotNil(t, ctx.cacheFactory)
}

func TestCache_RedisNeedsRedisFeature(t *testing.T) {
	f := Cache(WithRedis())
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorContains(t, err, "Redis feature")
}

func TestCache_UnknownType(t *testing.T) {
	f := Cache(WithCacheType("memcached"))
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorContains(t, err, "memcached")
}

func TestRedis_FailsWithoutHost(t *testing.T) {
	registry.Set(registry.Registry{})
	f := Redis()
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorContains(t, err, "no redis host")
}

func TestRegistry_MissingFile(t *testing.T) {
	f := Registry(WithRegistryPath("/does/not/exist.yaml"))
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorIs(t, err, ErrRegistryFileNotFound)
}

func TestRegistry_LoadsFile(t *testing.T) {
	fs := &Features{Registry: Registry(WithRegistryPath(writeRegistry(t)))}
	startFeatures(t, fs)
	assert.Equal(t, "pg", registry.Get().Postgres.Name)
}

func TestRSA_GeneratesKeyWithoutPaths(t *testing.T) {
	fs := &Features{RSA: RSA()}
	startFeatures(t, fs)
	assert.NotNil(t, keys.RSA())
}

func TestJWT_LoadsTokenConfigurations(t *testing.T) {
	privPath, pubPath := writeRSA(t)
	fs := &Features{JWT: JWT(
		WithJWTPrivateKeyPath(privPath),
		WithJWTPublicKeyPath(pubPath),
		WithTokenConfigurationPaths([]string{writeToken(t)}),
	)}
	fs.JWT.TokenConfigurations = append(fs.JWT.TokenConfigurations, tokenConfig(AuthIssuer))
	ctx := startFeatures(t, fs)

	_, ok := ctx.issuerToTokenConfigs["issuer"]
	assert.True(t, ok, "issuer from file should be loaded")
	_, ok = ctx.AuthIssuerConfig()
	assert.True(t, ok, "inline issuer should be loaded")
}

func tokenConfig(issuer string) jwt.TokenConfiguration {
	return jwt.TokenConfiguration{
		Issuer:                  issuer,
		Audience:                []string{"audience"},
		IdGenType:               "uuid",
		ValidityDurationSeconds: 3600,
	}
}

func TestSQLite_OpensDatabaseAndAppliesSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	fs := &Features{SQLite: SQLite(WithSQLiteDatabase("feature-test", path, "CREATE TABLE things (id INTEGER)"))}
	startFeatures(t, fs)

	db, err := dbsqlite.Get("feature-test")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("INSERT INTO things (id) VALUES (1)")
	assert.NoError(t, err)
}

func TestSQL_UnknownPackage(t *testing.T) {
	f := SQLFeature{Enabled: true}
	err := f.start(NewAppContext(context.Background(), zap.NewNop()), &Features{})
	assert.ErrorContains(t, err, "unknown sql package")
}

func TestPGX_BuildsPGXFeature(t *testing.T) {
	assert.Equal(t, PGXPackage, PGX().SQLPackage)
	assert.Equal(t, SQLXPackage, SQLX().SQLPackage)
}

func TestWithConfig_BuildsFeaturesFromYAML(t *testing.T) {
	path := writeFile(t, `
gin:
  enabled: true
  port: 9000
  cors:
    enabled: true
    origins: ["https://example.com"]
redis:
  enabled: true
  host: redis.local
  port: 6380
cache:
  enabled: true
  cache_type: redis
sql:
  enabled: true
  sql_package: PGX
  sql_files_dirs: ["./sql"]
`)
	cfg, err := LoadConfig(path)
	require.NoError(t, err)
	fs := WithConfig(cfg)

	assert.Equal(t, 9000, fs.Gin.Port)
	assert.Equal(t, []string{"https://example.com"}, fs.Gin.Cors.AllowOrigins)
	assert.Equal(t, "redis.local", fs.Redis.Host)
	assert.Equal(t, 6380, fs.Redis.Port)
	assert.Equal(t, cacheTypeRedis, fs.Cache.CacheType)
	assert.Equal(t, PGXPackage, fs.SQL.SQLPackage)
	assert.Equal(t, []string{"./sql"}, fs.SQL.SQLDirs)
	assert.False(t, fs.HTTP.Enabled)
}

func TestApplyFlags_OptionsWinOverFlags(t *testing.T) {
	registryPathFlag, sqlFilesFlag = "/from/flag.yaml", "a.sql,b.sql"
	t.Cleanup(func() { registryPathFlag, sqlFilesFlag = "", "" })

	fs := &Features{Registry: Registry(WithRegistryPath("/from/option.yaml")), SQL: PGX()}
	applyFlags(fs)

	assert.Equal(t, "/from/option.yaml", fs.Registry.Path)
	assert.Equal(t, []string{"a.sql", "b.sql"}, fs.SQL.SQLFiles)
}
