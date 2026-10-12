package app

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func TestApp(t *testing.T) {
	app := New("test", Features{})

	app.OnStartup(func(ctx *AppContext) error {
		<-ctx.Done()
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	err := app.Run(ctx)
	assert.Nilf(t, err, "expected no error, got %v", err)
}

func TestAppWithError(t *testing.T) {
	app := New("test", Features{})
	app.OnStartup(func(ctx *AppContext) error {
		return fmt.Errorf("test error")
	})
	err := app.Run(context.Background())
	assert.NotNilf(t, err, "expected error, got %v", err)
}

func TestAppWithRegistry(t *testing.T) {
	privKeyPath, pubKeyPath := writeRSA(t)
	regiPath := writeRegistry(t)
	caPath, certPath, keyPath := writeTLS(t)
	tokenPath := writeToken(t)
	app := New("test", Features{
		Registry: Registry(WithRegistryPath(regiPath)),
		RSA:      RSA(WithPrivateKeyPath(privKeyPath), WithPublicKeyPath(pubKeyPath)),
		JWT: JWT(
			WithTokenConfigurationPaths([]string{tokenPath}),
			WithJWTPrivateKeyPath(privKeyPath),
			WithJWTPublicKeyPath(pubKeyPath),
		),
		TLS: TLS(
			WithServerCAFile(caPath),
			WithServerCert(certPath),
			WithServerKey(keyPath),
		),
		HTTP:       HTTP(),
		LoggingAPI: LoggingApi(),
		Health:     Health(WithHealthPath("/health"), WithHealthInterval(1)),
	})

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := app.Run(ctx)
		assert.Nilf(t, err, "expected no error, got %v", err)
		wg.Done()
	}()
	l := app.l
	assert.Eventually(t, func() bool {
		l.Info("checking if app is healthy", zap.Bool("healthy", app.IsHealthy()))
		return app.IsRunning()
	}, 5*time.Second, 1*time.Second, "expected app to be running")
	cancel()
	wg.Wait()

}

func TestApp_StartupErrorNamesFeature(t *testing.T) {
	setupRan := false
	app := New("test", Features{Registry: Registry(WithRegistryPath("/does/not/exist.yaml"))})
	app.OnStartup(func(ctx *AppContext) error {
		setupRan = true
		return nil
	})

	err := app.Run(context.Background())
	assert.ErrorIs(t, err, ErrRegistryFileNotFound)
	assert.ErrorContains(t, err, "registry")
	assert.False(t, setupRan, "OnStartup should not run after a feature fails")
}

func TestApp_ServesRoutesRegisteredInOnStartupThenStopsOnCancel(t *testing.T) {
	app := New("test", Features{HTTP: HTTP()})
	app.OnStartup(func(ctx *AppContext) error {
		app.Features().HTTP.Mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hi"))
		})
		return nil
	})

	var body string
	var stoppedCalled bool
	ctx, cancel := context.WithCancel(context.Background())
	app.OnRunning(func(appCtx *AppContext) error {
		defer cancel()
		_, body = get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/hello", port(app.Features().HTTP.Addr())))
		return nil
	})
	app.OnStopped(func(ctx *AppContext) error {
		stoppedCalled = true
		return nil
	})

	assert.NoError(t, app.Run(ctx))
	assert.Equal(t, "hi", body)
	assert.True(t, stoppedCalled)
	assert.False(t, app.IsRunning())

	_, err := net.Dial("tcp", app.Features().HTTP.Addr().String())
	assert.Error(t, err, "HTTP server should be closed after Run returns")
}

func TestApp_ServeFailureStopsServersAlreadyStarted(t *testing.T) {
	taken, err := net.Listen("tcp", "0.0.0.0:0")
	assert.NoError(t, err)
	defer taken.Close()

	app := New("test", Features{
		Gin:  Gin(WithGinPort(0)),
		HTTP: HTTP(WithHttpPort(port(taken.Addr()))),
	})

	done := make(chan error)
	go func() { done <- app.Run(context.Background()) }()

	select {
	case err := <-done:
		assert.ErrorContains(t, err, "failed to serve http")
	case <-time.After(5 * time.Second):
		t.Fatal("Run should return when a server fails to start, not hang")
	}

	_, err = net.Dial("tcp", app.Features().Gin.Addr().String())
	assert.Error(t, err, "Gin should be shut down when HTTP fails to serve")
}

func TestApp_IsHealthyFollowsHealthCheck(t *testing.T) {
	app := New("test", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"))})
	app.SetHealthCheck(func() bool { return false })

	ctx, cancel := context.WithCancel(context.Background())
	app.OnRunning(func(appCtx *AppContext) error {
		defer cancel()
		assert.True(t, app.IsHealthy(), "healthy until a check fails")
		get(t, http.DefaultClient, fmt.Sprintf("http://localhost:%d/health", port(app.Features().HTTP.Addr())))
		assert.False(t, app.IsHealthy())
		return nil
	})
	assert.NoError(t, app.Run(ctx))
}

func TestAppWithTestEnvironment(t *testing.T) {
	app := New("test", Features{
		Redis: Redis(),
		Cache: Cache(WithRedis()),
	})
	cacheFactory := make(chan factory.CacheFactory, 1)
	app.OnRunning(func(ctx *AppContext) error {
		cacheFactory <- ctx.CacheFactory()
		<-ctx.Done()
		return nil
	})
	app.WithTestEnvironment(TestEnvironment{
		Postgres: true,
		Redis:    true,
	})
	l := app.l

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := app.Run(ctx)
		assert.Nilf(t, err, "expected no error, got %v", err)
		l.Info("app run completed")
		wg.Done()
	}()
	assert.Eventuallyf(t, func() bool {
		l.Info("checking if app is healthy and running", zap.Bool("healthy", app.IsHealthy()), zap.Bool("running", app.IsRunning()))
		return app.IsHealthy() && app.IsRunning()
	}, 60*time.Second, 1*time.Second, "expected app to be healthy")

	assert.IsType(t, &factory.RedisCacheFactory{}, <-cacheFactory, "cache should be backed by the Redis feature")

	r := redis.GetConnection(ctx)
	defer r.Close()
	cmd := r.Ping(ctx)
	assert.Nilf(t, cmd.Err(), "expected no error, got %v", cmd.Err())

	p := pgx.GetPGX()
	assert.Nilf(t, p.Ping(ctx), "expected no error")

	cancel()
	wg.Wait()
}

func writeFile(t *testing.T, content string) string {
	path, err := os.CreateTemp("/tmp/", "test-*")
	assert.Nilf(t, err, "expected no error, got %v", err)

	err = os.WriteFile(path.Name(), []byte(content), 0644)
	assert.Nilf(t, err, "expected no error, got %v", err)
	return path.Name()
}

func writeRegistry(t *testing.T) string {
	reg := registry.Registry{
		Postgres: &registry.Database{
			Server: registry.Server{
				Name: "pg",
				Host: "localhost",
				Port: 5432,
			},
		},
		Nats: &registry.MessageBroker{
			Server: registry.Server{
				Name: "nats",
			},
			Topics: []string{"topic1", "topic2"},
		},
	}
	b, err := yaml.Marshal(reg)
	assert.Nilf(t, err, "should be able to marshal registry")
	return writeFile(t, string(b))
}

func writeRSA(t *testing.T) (privKeyPath, pubKeyPath string) {
	rsa, err := keys.NewRSA()
	assert.Nilf(t, err, "should be able to create RSA")

	privBytes, pubBytes := rsa.Pem()
	privKeyPath = writeFile(t, string(privBytes))
	pubKeyPath = writeFile(t, string(pubBytes))

	return
}

func writeTLS(t *testing.T) (caPath, certPath, keyPath string) {
	ca, err := keys.CreateX509CA()
	assert.Nilf(t, err, "should be able to create CA")

	cert, err := keys.CreateX509(*ca,
		keys.WithDNSNames([]string{"localhost"}),
		keys.WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}),
	)
	assert.Nilf(t, err, "should be able to create cert")
	_, caCertPem := ca.Pem()
	keyPem, certPem := cert.Pem()

	caPath = writeFile(t, string(caCertPem))
	certPath = writeFile(t, string(certPem))
	keyPath = writeFile(t, string(keyPem))

	return
}

func writeToken(t *testing.T) string {
	token := jwt.TokenConfiguration{
		Issuer:                  "issuer",
		Audience:                []string{"audience"},
		IdGenType:               "uuid",
		ValidityDurationSeconds: 3600,
	}
	b, err := yaml.Marshal(token)
	assert.Nilf(t, err, "should be able to marshal token")
	return writeFile(t, string(b))
}
