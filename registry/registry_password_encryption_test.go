package registry

import (
	"os"
	"testing"

	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp("", "registry-enc-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })
	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestServerResolvePasswordRSA(t *testing.T) {
	rsaKey, err := keys.NewRSA()
	require.NoError(t, err)
	keys.SetRSA(rsaKey)

	plaintext := "rsa-secret"
	encrypted, err := crypto.RSAEncrypt([]byte(plaintext))
	require.NoError(t, err)

	tests := []struct {
		description    string
		server         func() Server
		expectedResult string
		expectError    bool
	}{
		{
			description: "RSA encrypted password file decrypts correctly",
			server: func() Server {
				pwFile := writeTempFile(t, encrypted)
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: pwFile,
					},
					Encryption: &Encryption{
						RSA: &struct {
							PrivateKeyPath string `yaml:"private_key_path"`
							PublicKeyPath  string `yaml:"public_key_path"`
						}{},
					},
				}
			},
			expectedResult: plaintext,
			expectError:    false,
		},
		{
			description: "RSA encrypted password file missing returns error",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: "/tmp/does-not-exist-rsa-test",
					},
					Encryption: &Encryption{
						RSA: &struct {
							PrivateKeyPath string `yaml:"private_key_path"`
							PublicKeyPath  string `yaml:"public_key_path"`
						}{},
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			s := tc.server()
			result, err := s.ResolvePassword()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestServerResolvePasswordX509(t *testing.T) {
	ca, err := keys.CreateX509CA()
	require.NoError(t, err)

	cert, err := keys.CreateX509(*ca)
	require.NoError(t, err)

	keyB, certB := cert.Pem()
	certFile := writeTempFile(t, append(keyB, certB...))

	plaintext := "x509-secret"
	encrypted, err := cert.Encrypt([]byte(plaintext))
	require.NoError(t, err)

	tests := []struct {
		description    string
		server         func() Server
		expectedResult string
		expectError    bool
	}{
		{
			description: "X509 encrypted password file decrypts correctly",
			server: func() Server {
				pwFile := writeTempFile(t, encrypted)
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: pwFile,
					},
					Encryption: &Encryption{
						X509: &struct {
							CertPath string `yaml:"cert_path"`
						}{
							CertPath: certFile,
						},
					},
				}
			},
			expectedResult: plaintext,
			expectError:    false,
		},
		{
			description: "X509 with missing cert file returns error",
			server: func() Server {
				pwFile := writeTempFile(t, encrypted)
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: pwFile,
					},
					Encryption: &Encryption{
						X509: &struct {
							CertPath string `yaml:"cert_path"`
						}{
							CertPath: "/tmp/does-not-exist-x509-cert",
						},
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			description: "X509 with missing password file returns error",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: "/tmp/does-not-exist-x509-pw",
					},
					Encryption: &Encryption{
						X509: &struct {
							CertPath string `yaml:"cert_path"`
						}{
							CertPath: certFile,
						},
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			s := tc.server()
			result, err := s.ResolvePassword()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestServerResolvePasswordAESGCM(t *testing.T) {
	salt := [crypto.SALT_SIZE]byte{}
	copy(salt[:], []byte("testsalt12345678"))

	passphrase := "aes-gcm-key-passphrase"
	derivedKey, err := crypto.DeriveAESGCMKey(passphrase, salt)
	require.NoError(t, err)

	plaintext := "aesgcm-secret"
	encrypted, err := crypto.AESGCMEncryptWithKey(derivedKey, salt, []byte(plaintext))
	require.NoError(t, err)

	keyFile := writeTempFile(t, []byte(passphrase))

	tests := []struct {
		description    string
		server         func() Server
		expectedResult string
		expectError    bool
	}{
		{
			description: "AES-GCM encrypted password file decrypts correctly",
			server: func() Server {
				pwFile := writeTempFile(t, encrypted)
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: pwFile,
					},
					Encryption: &Encryption{
						AESGCM: &struct {
							Salt    string `yaml:"salt"`
							KeyPath string `yaml:"key_path"`
						}{
							Salt:    string(salt[:]),
							KeyPath: keyFile,
						},
					},
				}
			},
			expectedResult: plaintext,
			expectError:    false,
		},
		{
			description: "AES-GCM with missing key file returns error",
			server: func() Server {
				pwFile := writeTempFile(t, encrypted)
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: pwFile,
					},
					Encryption: &Encryption{
						AESGCM: &struct {
							Salt    string `yaml:"salt"`
							KeyPath string `yaml:"key_path"`
						}{
							Salt:    string(salt[:]),
							KeyPath: "/tmp/does-not-exist-aes-key",
						},
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			description: "AES-GCM with missing password file returns error",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: "/tmp/does-not-exist-aes-pw",
					},
					Encryption: &Encryption{
						AESGCM: &struct {
							Salt    string `yaml:"salt"`
							KeyPath string `yaml:"key_path"`
						}{
							Salt:    string(salt[:]),
							KeyPath: keyFile,
						},
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			s := tc.server()
			result, err := s.ResolvePassword()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
