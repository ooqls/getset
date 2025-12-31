package app

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

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
		HTTP:       HTTP(WithHttpPort(8082)),
		LoggingAPI: LoggingApi(WithLoggingApiPort(8081)),
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

func TestAppWithTestEnvironment(t *testing.T) {
	app := New("test", Features{})
	app.OnRunning(func(ctx *AppContext) error {
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
