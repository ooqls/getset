package elasticsearch

import (
	"encoding/base64"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ooqls/getset/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryPolicy(t *testing.T) {
	if registry.IsSet() {
		old := *registry.Get()
		t.Cleanup(func() { registry.Set(old) })
	}
	for _, tt := range []struct {
		name      string
		serverTLS bool
		trusted   bool
		insecure  bool
		protocol  string
		auth      bool
		wantError string
	}{
		{name: "nil TLS is HTTP and disabled auth ignores missing file"},
		{name: "verified HTTPS and auth", serverTLS: true, trusted: true, auth: true},
		{name: "untrusted HTTPS fails closed", serverTLS: true, wantError: "failed to initialize"},
		{name: "explicit insecure HTTPS", serverTLS: true, insecure: true},
		{name: "HTTPS contradicts disabled TLS", protocol: "https", wantError: "contradicts TLS"},
		{name: "HTTP contradicts enabled TLS", serverTLS: true, trusted: true, protocol: "http", wantError: "contradicts TLS"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, c.Close()) // Isolate the process-global connection between cases.
			t.Cleanup(func() { require.NoError(t, c.Close()) })
			headers := make(chan string, 64)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				headers <- r.Header.Get("Authorization")
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"name":"test","cluster_name":"test","cluster_uuid":"test","tagline":"You Know, for Search","version":{"number":"8.19.6","build_flavor":"default","build_type":"docker","build_hash":"test","build_date":"2025-01-01T00:00:00Z","build_snapshot":false,"lucene_version":"9.12.0","minimum_wire_compatibility_version":"7.17.0","minimum_index_compatibility_version":"7.0.0"}}`))
			})
			server := httptest.NewUnstartedServer(handler)
			if tt.serverTLS {
				server.StartTLS()
			} else {
				server.Start()
			}
			t.Cleanup(server.Close)
			u, err := url.Parse(server.URL)
			require.NoError(t, err)
			host, portString, err := net.SplitHostPort(u.Host)
			require.NoError(t, err)
			port, err := strconv.Atoi(portString)
			require.NoError(t, err)
			var tlsConfig *registry.TLSConfig
			if tt.serverTLS {
				tlsConfig = &registry.TLSConfig{Enabled: true, InsecureSkipTLSVerify: tt.insecure}
				if tt.trusted {
					ca := filepath.Join(t.TempDir(), "ca.pem")
					require.NoError(t, os.WriteFile(ca, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600))
					tlsConfig.CaPath = ca
				}
			}
			password := filepath.Join(t.TempDir(), "missing-password")
			if tt.auth {
				require.NoError(t, os.WriteFile(password, []byte("test-password"), 0600))
			}
			registry.Set(registry.Registry{Elasticsearch: &registry.Database{Server: registry.Server{
				Host: host, Port: port, Protocol: tt.protocol, TLS: tlsConfig,
				Auth: registry.Auth{Enabled: tt.auth, Username: "test-user", PasswordFile: password},
			}}})
			err = InitDefault()
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				assert.Empty(t, headers, "invalid security settings must not reach the handler")
				return
			}
			require.NoError(t, err)
			select {
			case header := <-headers:
				if tt.auth {
					assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("test-user:test-password")), header)
				} else {
					assert.Empty(t, header)
				}
			default:
				t.Fatal("expected the adapter to contact the test endpoint")
			}
		})
	}
}
