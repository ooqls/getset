package cache

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

type TLS struct {
	KeyPath  string
	CertPath string
	CaPath   string
}

func (t *TLS) GetTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(t.CertPath, t.KeyPath)
	if err != nil {
		return nil, err
	}

	caCert, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(t.CaPath)
	if err == nil {
		caB, err := os.ReadFile(t.CaPath)
		if err != nil {
			return nil, err
		}

		caCert.AppendCertsFromPEM(caB)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCert,
	}, nil
}
