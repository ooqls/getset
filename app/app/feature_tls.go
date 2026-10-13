package app

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CaPath   string `yaml:"ca_path"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type TLSOption func(*TLSFeature)

func WithServerCAFile(ca string) TLSOption {
	return func(f *TLSFeature) { f.CAFile = ca }
}

func WithCaBytes(ca []byte) TLSOption {
	return func(f *TLSFeature) { f.CABytes = ca }
}

func WithServerCert(certPath string) TLSOption {
	return func(f *TLSFeature) { f.ServerCertFile = certPath }
}

func WithServerCertBytes(cert []byte) TLSOption {
	return func(f *TLSFeature) { f.ServerCertBytes = cert }
}

func WithServerKey(keyPath string) TLSOption {
	return func(f *TLSFeature) { f.ServerKeyFile = keyPath }
}

func WithServerKeyBytes(key []byte) TLSOption {
	return func(f *TLSFeature) { f.ServerKeyBytes = key }
}

// TLSFeature turns on TLS for every server the App runs and for the App's
// own client. Bytes take precedence over files. The CA, when given, is
// trusted on top of the system pool.
type TLSFeature struct {
	Enabled         bool
	CAFile          string
	CABytes         []byte
	ServerCertFile  string
	ServerCertBytes []byte
	ServerKeyFile   string
	ServerKeyBytes  []byte
	config          *tls.Config
}

func TLS(opts ...TLSOption) TLSFeature {
	f := TLSFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func tlsFromConfig(c TLSConfig) TLSFeature {
	return TLSFeature{
		Enabled:        c.Enabled,
		CAFile:         c.CaPath,
		ServerCertFile: c.CertFile,
		ServerKeyFile:  c.KeyFile,
	}
}

func (f *TLSFeature) name() string  { return "tls" }
func (f *TLSFeature) enabled() bool { return f.Enabled }

// ServerConfig returns a copy of the resolved TLS config. Only valid after
// the feature has started.
func (f *TLSFeature) ServerConfig() *tls.Config {
	return f.config.Clone()
}

// Client returns an HTTP client that trusts the configured CA and presents
// the configured certificate. Without TLS it returns http.DefaultClient.
func (f *TLSFeature) Client() *http.Client {
	if !f.Enabled || f.config == nil {
		return http.DefaultClient
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: f.config.Clone()}}
}

func (f *TLSFeature) start(ctx *AppContext, fs *Features) error {
	l := ctx.L()

	pool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to get system cert pool: %v", err)
	}

	caBytes, err := bytesOrFile(f.CABytes, f.CAFile, "ca")
	if err != nil {
		return err
	}
	if caBytes != nil {
		if !pool.AppendCertsFromPEM(caBytes) {
			return fmt.Errorf("failed to add CA certificates")
		}
	} else {
		l.Info("[Startup TLS] no CA given, using the system pool")
	}

	certBytes, err := bytesOrFile(f.ServerCertBytes, f.ServerCertFile, "server cert")
	if err != nil {
		return err
	}
	keyBytes, err := bytesOrFile(f.ServerKeyBytes, f.ServerKeyFile, "server key")
	if err != nil {
		return err
	}
	if certBytes == nil || keyBytes == nil {
		return fmt.Errorf("TLS needs both a server cert and a server key")
	}

	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return fmt.Errorf("failed to load key pair: %v", err)
	}

	f.config = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientCAs:    pool,
	}
	l.Debug("[Startup TLS] TLS material loaded", zap.String("cert_file", f.ServerCertFile))
	return nil
}

// bytesOrFile returns b if set, otherwise the contents of path, otherwise nil.
func bytesOrFile(b []byte, path, what string) ([]byte, error) {
	if len(b) > 0 {
		return b, nil
	}
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file %s: %v", what, path, err)
	}
	return b, nil
}
