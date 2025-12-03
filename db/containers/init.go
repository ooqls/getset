package containers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/ooqls/getset/registry"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var reg registry.Registry = registry.Registry{}

const (
	opt_logging = "logging"
	opt_env     = "env"
	opt_tag     = "tags"
)

type Options struct {
	key   string
	value interface{}
}

func WithLogging() Options {
	return Options{
		key:   opt_logging,
		value: true,
	}
}

func WithEnv(key string, value map[string]string) Options {
	return Options{
		key:   key,
		value: value,
	}
}

func WithTags(tag string) Options {
	return Options{
		key:   opt_tag,
		value: tag,
	}
}

func isArm64() bool {
	arch := runtime.GOARCH
	return arch == "arm64"
}

func applyOptions(c *testcontainers.ContainerRequest, opts ...Options) {
	for _, opt := range opts {
		switch opt.key {
		case opt_logging:
			w := c.BuildLogWriter()
			io.Copy(w, os.Stdout)
		case opt_env:
			envMap := opt.value.(map[string]string)
			maps.Copy(c.Env, envMap)
		case opt_tag:
			baseImage := strings.Split(c.Image, ":")[0]
			c.Image = fmt.Sprintf("%s:%s", baseImage, opt.value)
		}
	}
}

func requestContainer(ctx context.Context, req testcontainers.GenericContainerRequest, mappedPort string) (nat.Port, *testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nat.Port("0"), nil, fmt.Errorf("failed to request container: %v", err)
	}

	time.Sleep(time.Second * 10)

	port, err := container.MappedPort(ctx, nat.Port(mappedPort))
	if err != nil {
		return nat.Port("0"), nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	return port, &container, nil
}

func StartValkey(ctx context.Context, opts ...Options) testcontainers.Container {
	image := "valkey/valkey:latest"
	c := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"6379"},
		WaitingFor:   &wait.LogStrategy{Log: "Ready to accept connections"},
		Env:          map[string]string{},
	}

	applyOptions(&c, opts...)

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}

	port, container, err := requestContainer(ctx, gc, "6379")
	if err != nil {
		panic(fmt.Errorf("failed to request container: %v", err))
	}
	time.Sleep(3 * time.Second)
	zap.L().Debug("valkey should be running", zap.Int("port", port.Int()))

	reg.Valkey = &registry.Database{
		Server: registry.Server{
			Host: "localhost",
			Port: port.Int(),
		},
	}
	registry.Set(reg)

	return *container
}

func StartRedis(ctx context.Context, opts ...Options) testcontainers.Container {
	c := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379"},
		WaitingFor:   &wait.LogStrategy{Log: "Ready to accept connections"},
		Env: map[string]string{
			"REDIS_PASSWORD": "password",
		},
	}

	applyOptions(&c, opts...)

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}

	port, container, err := requestContainer(ctx, gc, "6379")
	if err != nil {
		panic(fmt.Errorf("failed to request redis container: %v", err))
	}
	time.Sleep(3 * time.Second)
	zap.L().Debug("redis should be running", zap.Int("port", port.Int()))

	reg.Redis = &registry.Database{
		Database: "0",
		Server: registry.Server{
			Host: "localhost",
			Port: port.Int(),

			Auth: registry.Auth{
				Enabled:  true,
				Password: "password",
			},
		},
	}
	registry.Set(reg)

	return *container
}

func StartPostgres(ctx context.Context, opts ...Options) testcontainers.Container {
	image := "postgres:latest"
	if isArm64() {
		image = "arm64v8/postgres:latest"
	}

	c := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"5432"},
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "user100",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: &wait.LogStrategy{Log: "database system is ready to accept connections"},
	}

	applyOptions(&c, opts...)

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}

	port, container, err := requestContainer(ctx, gc, "5432")
	if err != nil {
		panic(fmt.Errorf("failed to request postgres container: %v", err))
	}
	time.Sleep(3 * time.Second)
	zap.L().Debug("postgres should be running", zap.Int("port", port.Int()))

	reg.Postgres = &registry.Database{
		Database: "test",
		Server: registry.Server{
			Host: "localhost",
			Port: port.Int(),
			Auth: registry.Auth{
				Enabled:  true,
				Username: "user",
				Password: "user100",
			},
		},
	}
	registry.Set(reg)

	return *container
}

func StartElasticsearch(ctx context.Context, opts ...Options) testcontainers.Container {
	image := "elasticsearch:8.18.0"
	if isArm64() {
		image = "arm64v8/elasticsearch:8.18.0"
	}

	c := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"9200"},
		Env: map[string]string{
			"ELASTIC_PASSWORD": "changeme",
			"discovery.type":   "single-node",
			"ES_JAVA_OPTS":     "-Xms512m -Xmx512m",
		},
		WaitingFor: wait.ForHTTP("/_cluster/health").
			WithBasicAuth("elastic", "changeme").
			WithAllowInsecure(true).
			WithMethod("GET").
			WithTLS(true, &tls.Config{InsecureSkipVerify: true}).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
	}

	applyOptions(&c, opts...)

	gc := testcontainers.GenericContainerRequest{
		ContainerRequest: c,
		Started:          true,
	}

	port, container, err := requestContainer(ctx, gc, "9200")
	if err != nil {
		panic(fmt.Errorf("failed to request an elasticsearch container: %v", err))
	}
	time.Sleep(3 * time.Second)
	zap.L().Debug("elasticsearch should be running", zap.Int("port", port.Int()))

	reg.Elasticsearch = &registry.Database{
		Database: "elasticsearch",
		Server: registry.Server{
			Host: "localhost",
			Port: port.Int(),
			Auth: registry.Auth{
				Enabled:  true,
				Password: "changeme",
				Username: "elastic",
			},
			TLS: &registry.TLSConfig{
				Enabled:               true,
				InsecureSkipTLSVerify: true,
			},
		},
	}
	registry.Set(reg)

	return *container
}
