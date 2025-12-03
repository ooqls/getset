package elasticsearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("elasticsearch")
var esClient *elasticsearch.TypedClient
var m sync.Mutex

type ElasticsearchOptions struct {
	Host      string
	Port      int
	User      string
	Pw        string
	DB        string
	TlsConfig *tls.Config
}

func Init(opts ElasticsearchOptions) error {
	m.Lock()
	defer m.Unlock()

	var err error
	// headers := http.Header{}
	// headers.Del("Accept")
	// headers.Set("Accept", "application/json")
	// headers.Set("Content-Type", "application/json")
	esClient, err = elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("https://%s:%d", opts.Host, opts.Port),
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
		return fmt.Errorf("error creating the client: %s", err)
	}
	ctx := context.Background()
	success := false
	for i := 0; i < 5; i++ {
		if i > 0 {
			l.Info("Retrying to connect to elasticsearch", zap.Int("attempt", i))
			time.Sleep(time.Second)
		}

		info, err := esClient.Info().Do(ctx)
		if err != nil {
			l.Error(fmt.Sprintf("failed to get elasticsearch info, attempt %d", i), zap.Error(err))
			continue
		}

		l.Debug("Connected to Elasticsearch", zap.String("version", info.Version.Int))

		l.Debug("Elasticsearch client initialized successfully")
		success = true
		break
	}

	if !success {
		return fmt.Errorf("failed to initialize elasticsearch client")
	}

	return nil
}

func InitDefault() error {
	reg := registry.Get()
	if reg.Elasticsearch == nil {
		return fmt.Errorf("elasticsearch not found in registry")
	}

	tlsConfig, err := reg.Elasticsearch.TLS.TLSConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS config: %s", err)
	}

	opts := ElasticsearchOptions{
		Host:      reg.Elasticsearch.Host,
		Port:      reg.Elasticsearch.Port,
		User:      reg.Elasticsearch.Auth.Username,
		Pw:        reg.Elasticsearch.Auth.Password,
		DB:        reg.Elasticsearch.Database,
		TlsConfig: tlsConfig,
	}

	return Init(opts)
}

func Get() *elasticsearch.TypedClient {
	m.Lock()
	defer m.Unlock()

	return esClient
}
