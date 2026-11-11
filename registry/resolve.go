package registry

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// Connection is a resolved snapshot. Adapters translate it into driver settings
// without interpreting auth enablement, secrets or TLS files again.
type Connection struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      *tls.Config
}

// Resolve applies Registry auth and TLS policy. Disabled auth does not resolve
// secrets or forward configured credentials. A nonnil passwordOverride opts in
// to authentication and bypasses the Registry password, even if it is empty.
// Each call owns its TLS configuration; nothing inherits HTTP global defaults.
func (s *Server) Resolve(passwordOverride *string) (Connection, error) {
	if s == nil {
		return Connection{}, fmt.Errorf("missing Registry server")
	}
	c := Connection{Host: s.Host, Port: s.Port}
	var err error
	c.TLS, err = s.TLS.TLSConfig()
	if err != nil {
		return Connection{}, fmt.Errorf("resolve TLS: %w", err)
	}
	if c.TLS != nil {
		c.TLS.ServerName = s.Host
	}
	if passwordOverride != nil {
		c.Username, c.Password = s.Auth.Username, *passwordOverride
	} else if s.Auth.Enabled {
		c.Username = s.Auth.Username
		c.Password, err = s.ResolvePassword()
		if err != nil {
			return Connection{}, fmt.Errorf("resolve password: %w", err)
		}
	}
	return c, nil
}

// TLSConfig resolves an isolated client configuration. Nil and disabled
// settings mean plaintext; enabled settings require valid TLS material.
func (cfg *TLSConfig) TLSConfig() (*tls.Config, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}
	if (cfg.CertPath == "") != (cfg.KeyPath == "") {
		return nil, fmt.Errorf("TLS client certificate and key must be configured together")
	}
	c := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.InsecureSkipTLSVerify}
	if cfg.CertPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("load TLS client certificate: %w", err)
		}
		c.Certificates = []tls.Certificate{cert}
	}
	if cfg.CaPath != "" {
		b, err := os.ReadFile(cfg.CaPath)
		if err != nil {
			return nil, fmt.Errorf("read TLS CA file: %w", err)
		}
		// AppendCertsFromPEM silently ignores malformed blocks. Validate the
		// entire bundle first so mixed valid/invalid material cannot pass.
		rest := bytes.TrimSpace(b)
		count := 0
		for len(rest) > 0 {
			if !bytes.HasPrefix(rest, []byte("-----BEGIN CERTIFICATE-----")) {
				return nil, fmt.Errorf("TLS CA file contains invalid certificate PEM")
			}
			block, tail := pem.Decode(rest)
			if block == nil || block.Type != "CERTIFICATE" || bytes.Count(rest[:len(rest)-len(tail)], []byte("-----BEGIN ")) != 1 {
				return nil, fmt.Errorf("TLS CA file contains invalid certificate PEM")
			}
			if _, err := x509.ParseCertificate(block.Bytes); err != nil {
				return nil, fmt.Errorf("parse TLS CA certificate: %w", err)
			}
			count++
			rest = bytes.TrimSpace(tail)
		}
		if count == 0 {
			return nil, fmt.Errorf("TLS CA file contains no certificates")
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system TLS roots: %w", err)
		}
		if !roots.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("TLS CA file contains no certificates")
		}
		c.RootCAs = roots
	}
	return c, nil
}
