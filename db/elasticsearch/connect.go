package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
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
	protocol := "http"
	if opts.TlsConfig != nil {
		protocol = "https"
	}
	if opts.Protocol != "" && opts.Protocol != protocol {
		return nil, fmt.Errorf("elasticsearch protocol %q contradicts TLS settings (expected %s)", opts.Protocol, protocol)
	}
	var tlsConfig *tls.Config
	if opts.TlsConfig != nil {
		tlsConfig = opts.TlsConfig.Clone()
		if tlsConfig.ServerName == "" {
			tlsConfig.ServerName = opts.Host
		}
	}
	client, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port))),
		},
		Username:          opts.User,
		Password:          opts.Pw,
		EnableDebugLogger: true,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
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
	var lastErr error
	for i := 0; i < 5; i++ {
		if i > 0 {
			l.Info("Retrying to connect to elasticsearch", zap.Int("attempt", i))
			time.Sleep(time.Second)
		}

		info, err := client.Info().Do(ctx)
		if err != nil {
			lastErr = err
			l.Error(fmt.Sprintf("failed to get elasticsearch info, attempt %d", i), zap.Error(err))
			continue
		}

		l.Debug("Connected to Elasticsearch", zap.String("version", info.Version.Int))
		return client, nil
	}

	return nil, fmt.Errorf("failed to initialize elasticsearch client: %w", lastErr)
}

func openFromRegistry() (*elasticsearch.TypedClient, error) {
	if !registry.IsSet() {
		return nil, fmt.Errorf("elasticsearch Registry is not initialized")
	}
	reg := registry.Get()
	if reg.Elasticsearch == nil {
		return nil, fmt.Errorf("elasticsearch not found in registry")
	}

	resolved, err := reg.Elasticsearch.Server.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve elasticsearch connection: %w", err)
	}

	return open(ElasticsearchOptions{
		Protocol:  reg.Elasticsearch.Protocol,
		Host:      resolved.Host,
		Port:      resolved.Port,
		User:      resolved.Username,
		Pw:        resolved.Password,
		DB:        reg.Elasticsearch.Database,
		TlsConfig: resolved.TLS,
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
