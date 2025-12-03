package elasticsearch

import (
	"context"
	"testing"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	ctx := context.Background()

	container := containers.StartElasticsearch(ctx)
	defer container.Terminate(ctx)

	port, err := container.MappedPort(ctx, "9200")
	if err != nil {
		t.Fatalf("failed to get mapped port for elasticsearch: %v", err)
	}
	t.Logf("Elasticsearch should be running at localhost:%s", port)

	registry.Set(registry.Registry{
		Elasticsearch: &registry.Database{
			Database: "elasticsearch",
			Server: registry.Server{
				Host: "localhost",
				Port: port.Int(),
				Auth: registry.Auth{
					Username: "elastic",
					Password: "changeme",
				},
				TLS: &registry.TLSConfig{
					Enabled:               true,
					InsecureSkipTLSVerify: true,
				},
			},
		},
	})

	// Get the mapped port
	err = InitDefault()
	assert.Nilf(t, err, "failed to initialize elasticsearch client: %v", err)

	l.Info("Elasticsearch client initialized successfully")
}
