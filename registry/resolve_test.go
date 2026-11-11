package registry

import (
	"crypto/tls"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Resolve(t *testing.T) {
	t.Parallel()
	password := writeTempPasswordFile(t, "password\n")
	empty, override := "", "override\n"
	for _, tt := range []struct {
		name         string
		auth         Auth
		override     *string
		wantUser     string
		wantPassword string
		wantError    bool
	}{
		{name: "disabled ignores configured credentials", auth: Auth{Username: "ignored", PasswordFile: filepath.Join(t.TempDir(), "missing")}},
		{name: "enabled reads password", auth: Auth{Enabled: true, Username: "user", PasswordFile: password}, wantUser: "user", wantPassword: "password\n"},
		{name: "enabled reports missing password", auth: Auth{Enabled: true, PasswordFile: filepath.Join(t.TempDir(), "missing")}, wantError: true},
		{name: "override enables auth and bypasses password", auth: Auth{Username: "user", PasswordFile: filepath.Join(t.TempDir(), "missing")}, override: &override, wantUser: "user", wantPassword: override},
		{name: "empty override still explicit", auth: Auth{Username: "user", PasswordFile: password}, override: &empty, wantUser: "user"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Server{Host: "localhost", Port: 1234, Auth: tt.auth, Encryption: &Encryption{}}
			got, err := s.Resolve(tt.override)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, s.Host, got.Host)
			assert.Equal(t, s.Port, got.Port)
			assert.Equal(t, tt.wantUser, got.Username)
			assert.Equal(t, tt.wantPassword, got.Password)
			assert.Nil(t, got.TLS)
		})
	}
}

func TestTLSConfig_Validation(t *testing.T) {
	t.Parallel()
	invalidCA := writeTempPasswordFile(t, "not a certificate")
	for _, tt := range []struct {
		name      string
		config    *TLSConfig
		wantError bool
		wantTLS   bool
	}{
		{name: "absent"},
		{name: "disabled skips files", config: &TLSConfig{CertPath: "missing", CaPath: "missing"}},
		{name: "enabled", config: &TLSConfig{Enabled: true}, wantTLS: true},
		{name: "missing key", config: &TLSConfig{Enabled: true, CertPath: "missing"}, wantError: true},
		{name: "missing certificate", config: &TLSConfig{Enabled: true, KeyPath: "missing"}, wantError: true},
		{name: "missing CA", config: &TLSConfig{Enabled: true, CaPath: filepath.Join(t.TempDir(), "missing")}, wantError: true},
		{name: "malformed CA", config: &TLSConfig{Enabled: true, CaPath: invalidCA}, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.config.TLSConfig()
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantTLS {
				require.NotNil(t, got)
				assert.False(t, got.InsecureSkipVerify)
				assert.Equal(t, uint16(tls.VersionTLS12), got.MinVersion)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestServer_ResolveDoesNotInheritHTTPDefaults(t *testing.T) {
	old := http.DefaultTransport
	globalTLS := &tls.Config{ServerName: "unrelated", InsecureSkipVerify: true}
	http.DefaultTransport = &http.Transport{TLSClientConfig: globalTLS}
	t.Cleanup(func() { http.DefaultTransport = old })
	s := Server{Host: "db.example", TLS: &TLSConfig{Enabled: true}}
	first, err := s.Resolve(nil)
	require.NoError(t, err)
	second, err := s.Resolve(nil)
	require.NoError(t, err)
	assert.False(t, first.TLS.InsecureSkipVerify)
	assert.Equal(t, "db.example", first.TLS.ServerName)
	first.TLS.ServerName = "modified"
	assert.Equal(t, "db.example", second.TLS.ServerName)
	assert.Equal(t, "unrelated", globalTLS.ServerName)
}

func TestTLSConfig_TrustedAndUntrustedConnections(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	t.Cleanup(server.Close)
	ca := filepath.Join(t.TempDir(), "ca.pem")
	validPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	require.NoError(t, os.WriteFile(ca, validPEM, 0600))
	for _, tt := range []struct {
		name string
		data []byte
	}{
		{name: "valid then invalid", data: append(append([]byte(nil), validPEM...), []byte("invalid trailing material")...)},
		{name: "invalid then valid", data: append([]byte("-----BEGIN CERTIFICATE-----\ninvalid\n-----END CERTIFICATE-----\n"), validPEM...)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "invalid-bundle.pem")
			require.NoError(t, os.WriteFile(file, tt.data, 0600))
			_, err := (&TLSConfig{Enabled: true, CaPath: file}).TLSConfig()
			require.Error(t, err)
		})
	}
	for _, tt := range []struct {
		name      string
		config    TLSConfig
		wantError bool
	}{
		{name: "trusted", config: TLSConfig{Enabled: true, CaPath: ca}},
		{name: "untrusted", config: TLSConfig{Enabled: true}, wantError: true},
		{name: "explicit bypass", config: TLSConfig{Enabled: true, InsecureSkipTLSVerify: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := tt.config.TLSConfig()
			require.NoError(t, err)
			transport := &http.Transport{TLSClientConfig: cfg}
			t.Cleanup(transport.CloseIdleConnections)
			client := &http.Client{Transport: transport}
			resp, err := client.Get(server.URL)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
	}
}
