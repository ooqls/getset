package app

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

var (
	tlsKeyPathFlag  string = "tls-key-path"
	tlsCertPathFlag string = "tls-cert-path"
	tlsCaPathFlag   string = "tls-ca-path"
)

const (
	tls_caFileOpt   string = "opt-server-ca-pool"
	tls_crtOpt      string = "opt-server-cert"
	tls_crtBytesOpt string = "opt-server-cert-bytes"
	tls_caBytesOpt  string = "opt-server-ca-bytes"
	tls_keyFile     string = "opt-server-key-file"
	tls_keyBytes    string = "opt-server-key-bytes"
)

type tlsOpt struct {
	featureOpt
}

func WithServerCAFile(ca string) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_caFileOpt,
			value: ca,
		},
	}
}

func WithServerCert(cert string) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_crtOpt,
			value: cert,
		},
	}
}

func WithServerCertBytes(cert []byte) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_crtBytesOpt,
			value: cert,
		},
	}
}

func WithServerKey(keyPath string) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_keyFile,
			value: keyPath,
		},
	}
}

func WithServerKeyBytes(b []byte) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_keyBytes,
			value: b,
		},
	}
}

func WithCaBytes(ca []byte) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_caBytesOpt,
			value: ca,
		},
	}
}

func WithKeyFile(p string) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_keyFile,
			value: p,
		},
	}
}

func WithKeyBytes(b []byte) tlsOpt {
	return tlsOpt{
		featureOpt: featureOpt{
			key:   tls_keyBytes,
			value: b,
		},
	}
}

type TLSFeature struct {
	Enabled         bool
	CAFile          string
	CABytes         []byte
	ServerCertFile  string
	ServerCertBytes []byte
	ServerKeyBytes  []byte
	ServerKeyFile   string
}

func (f *TLSFeature) TLSConfig() (*tls.Config, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system cert pool: %v", err)
	}

	cfg := &tls.Config{
		RootCAs: pool,
	}

	if f.CAFile != "" {
		b, err := os.ReadFile(f.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read ca file %s: %v", f.CAFile, err)
		}

		ok := pool.AppendCertsFromPEM(b)
		if !ok {
			return nil, fmt.Errorf("failed to add pem bytes from ca file %s", f.CAFile)
		}
	}

	if len(f.ServerCertFile) > 0 && len(f.ServerKeyFile) > 0 {
		cert, err := tls.LoadX509KeyPair(f.ServerCertFile, f.ServerKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load x509 key pair: %v", err)
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	} else if len(f.ServerCertBytes) > 0 && len(f.ServerKeyBytes) > 0 {
		cert, err := tls.X509KeyPair(f.ServerCertBytes, f.ServerKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to load x509 key pair: %v", err)
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	return cfg, nil
}
func TLS(opts ...tlsOpt) TLSFeature {
	f := TLSFeature{
		Enabled: true,
	}

	for _, opt := range opts {
		switch opt.key {
		case tls_crtOpt:
			f.ServerCertFile = opt.value.(string)
		case tls_crtBytesOpt:
			f.ServerCertBytes = opt.value.([]byte)
		case tls_caFileOpt:
			f.CAFile = opt.value.(string)
		case tls_caBytesOpt:
			f.CABytes = opt.value.([]byte)
		case tls_keyFile:
			f.ServerKeyFile = opt.value.(string)
		case tls_keyBytes:
			f.ServerKeyBytes = opt.value.([]byte)
		}
	}

	return f

}
