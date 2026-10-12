package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ooqls/getset/db/conn"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("elasticsearch")

var c = conn.New("elasticsearch", openFromRegistry, func(*elasticsearch.TypedClient) error {
	return nil
})

type ElasticsearchOptions struct {
	Protocol  string
	Host      string
	Port      int
	User      string
	Pw        string
	DB        string
	TlsConfig *tls.Config
}

// open creates a client and waits up to five seconds for the cluster to
// answer.
func open(opts ElasticsearchOptions) (*elasticsearch.TypedClient, error) {
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("%s://%s:%d", opts.Protocol, opts.Host, opts.Port),
		},
		Username:          opts.User,
		Password:          opts.Pw,
		EnableDebugLogger: true,
		Transport: &http.Transport{
			TLSClientConfig: opts.TlsConfig,
		},
		Logger: &elastictransport.TextLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating the client: %s", err)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if i > 0 {
			l.Info("Retrying to connect to elasticsearch", zap.Int("attempt", i))
			time.Sleep(time.Second)
		}

		info, err := client.Info().Do(ctx)
		if err != nil {
			l.Error(fmt.Sprintf("failed to get elasticsearch info, attempt %d", i), zap.Error(err))
			continue
		}

		l.Debug("Connected to Elasticsearch", zap.String("version", info.Version.Int))
		return client, nil
	}

	return nil, fmt.Errorf("failed to initialize elasticsearch client")
}

func openFromRegistry() (*elasticsearch.TypedClient, error) {
	reg := registry.Get()
	if reg.Elasticsearch == nil {
		return nil, fmt.Errorf("elasticsearch not found in registry")
	}

	tlsConfig, err := reg.Elasticsearch.TLS.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS config: %s", err)
	}

	pw, err := reg.Elasticsearch.Server.ResolvePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve elasticsearch password: %v", err)
	}

	return open(ElasticsearchOptions{
		Protocol:  reg.Elasticsearch.Protocol,
		Host:      reg.Elasticsearch.Host,
		Port:      reg.Elasticsearch.Port,
		User:      reg.Elasticsearch.Auth.Username,
		Pw:        pw,
		DB:        reg.Elasticsearch.Database,
		TlsConfig: tlsConfig,
	})
}

// Init opens a client with opts and makes it the one Get returns.
func Init(opts ElasticsearchOptions) error {
	return c.Init(func() (*elasticsearch.TypedClient, error) { return open(opts) })
}

// InitDefault opens the client from the registry unless one is already open.
func InitDefault() error {
	return c.InitDefault()
}

// Get returns the client, opening it from the registry on first use. It
// panics if the client can't be opened.
func Get() *elasticsearch.TypedClient {
	return c.MustGet()
}
