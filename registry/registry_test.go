package registry

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ooqls/getset/crypto/keys"
	"github.com/stretchr/testify/assert"
)

var cfg string = `
postgres:
  name: pg
	host: pg
	port: 10
	tls:
    enabled: false
  auth:
    enabled: false
kafka:
  "group.id": aaa
token_config:
	audience: test
	issuer: test
	validity_duration_seconds: 100
`

// writes out the temp certs and returns the path to the key, cert and ca cert
func writeTempCerts(t *testing.T) (string, string, string) {
	ca, err := keys.CreateX509CA()
	assert.NoError(t, err)

	cert, err := keys.CreateX509(*ca)
	assert.NoError(t, err)

	keyB, certB := cert.Pem()
	_, caCertB := ca.Pem()
	// save the keyB and certB in tmp files
	keyF, err := os.CreateTemp(os.TempDir(), "key-*.pem")
	assert.Nil(t, err)

	certF, err := os.CreateTemp(os.TempDir(), "cert-*.pem")
	assert.Nilf(t, err, "should not fail to create a cert file")

	caCertF, err := os.CreateTemp(os.TempDir(), "ca-cert-*.pem")
	assert.Nilf(t, err, "should not fail to create ca file")

	_, err = keyF.Write(keyB)
	assert.NoError(t, err)

	_, err = certF.Write(certB)
	assert.Nil(t, err)

	_, err = caCertF.Write(caCertB)
	assert.Nil(t, err)

	return keyF.Name(), certF.Name(), caCertF.Name()
}

func TestInit(t *testing.T) {
	f, err := os.CreateTemp("/tmp/", "registry-test")
	if err != nil {
		log.Print(err.Error())
		t.FailNow()
	}

	_, err = f.Write([]byte(strings.ReplaceAll(cfg, "\t", "  ")))
	if err != nil {
		log.Print(err.Error())
		t.FailNow()
	}

	err = Init(f.Name())
	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, Get().Kafka)
}

func TestTLSConfig(t *testing.T) {
	keyPath, certPath, caPath := writeTempCerts(t)

	reg := Registry{
		Redis: &Database{
			Server: Server{
				TLS: &TLSConfig{
					CaPath:                caPath,
					CertPath:              certPath,
					KeyPath:               keyPath,
					Enabled:               true,
					InsecureSkipTLSVerify: true,
				},
			},
		},
	}

	cfg, err := reg.Redis.TLS.TLSConfig()
	assert.Nilf(t, err, "should not fail to get tls config")
	assert.NotNil(t, cfg)
}

func TestTlsConnect(t *testing.T) {
	keyPath, certPath, caPath := writeTempCerts(t)

	go http.ListenAndServeTLS("0.0.0.0:8443", certPath, keyPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))

	reg := Registry{
		Redis: &Database{
			Server: Server{
				TLS: &TLSConfig{
					CaPath:                caPath,
					CertPath:              certPath,
					KeyPath:               keyPath,
					Enabled:               true,
					InsecureSkipTLSVerify: true,
				},
			},
		},
	}

	cfg, err := reg.Redis.TLS.TLSConfig()
	assert.Nilf(t, err, "should not fail to get tls config")

	// create a new client
	transport := &http.Transport{
		TLSClientConfig: cfg,
	}
	client := &http.Client{
		Transport: transport,
	}
	// make a request

	resp, err := client.Get("https://localhost:8443") // Use the client to make the request
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(body))
}
