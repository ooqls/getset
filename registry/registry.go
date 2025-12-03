package registry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

type TLSConfig struct {
	Enabled               bool   `yaml:"enabled"`
	CertPath              string `yaml:"cert_path"`
	KeyPath               string `yaml:"key_path"`
	CaPath                string `yaml:"ca_path"`
	InsecureSkipTLSVerify bool   `yaml:"insecure_skip_tls_verify"`
}

func (cfg *TLSConfig) TLSConfig() (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	defaultConfig := &tls.Config{}
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		if transport.TLSClientConfig != nil {
			defaultConfig = transport.TLSClientConfig
		}
	}
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, err
		}
		if defaultConfig.Certificates == nil {
			defaultConfig.Certificates = make([]tls.Certificate, 0)
		}

		defaultConfig.Certificates = append(defaultConfig.Certificates, cert)
	}

	if cfg.CaPath != "" {
		caCert, err := os.ReadFile(cfg.CaPath)
		if err != nil {
			return nil, err
		}

		if defaultConfig.RootCAs == nil {
			defaultConfig.RootCAs = x509.NewCertPool()
		}
		defaultConfig.RootCAs.AppendCertsFromPEM(caCert)
	}

	defaultConfig.InsecureSkipVerify = cfg.InsecureSkipTLSVerify

	return defaultConfig, nil
}

type Auth struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Topics struct {
	Messages string `yaml:"messages"`
}

type Server struct {
	Name  string                 `yaml:"name"`
	Host  string                 `yaml:"host"`
	Port  int                    `yaml:"port"`
	TLS   *TLSConfig             `yaml:"tls,omitempty"`
	Auth  Auth                   `yaml:"auth"`
	Extra map[string]interface{} `yaml:"extra"`
}

type Database struct {
	Server   `yaml:",inline"`
	Database string `yaml:"database"`
}

type MessageBroker struct {
	Server `yaml:",inline"`
	Topics []string `yaml:"topics"`
}

func (s *Server) GetConnectionString() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Registry struct {
	Kafka *MessageBroker `yaml:"kafka,omitempty"`
	Nats  *MessageBroker `yaml:"nats,omitempty"`

	Redis         *Database `yaml:"redis,omitempty"`
	Valkey        *Database `yaml:"valkeys,omitempty"`
	Postgres      *Database `yaml:"postgres,omitempty"`
	Mongo         *Database `yaml:"mongo,omitempty"`
	Neo4j         *Database `yaml:"neo4j,omitempty"`
	Elasticsearch *Database `yaml:"elasticsearch,omitempty"`
}
