package registry

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempPasswordFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "registry-pw-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(f.Name()) })
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestServerResolvePassword(t *testing.T) {
	plainFile := writeTempPasswordFile(t, "secret123")

	tests := []struct {
		description    string
		server         func() Server
		expectedResult string
		expectError    bool
	}{
		{
			description: "no password file set returns empty string",
			server: func() Server {
				return Server{}
			},
			expectedResult: "",
			expectError:    false,
		},
		{
			description: "password file set with no encryption returns raw file content",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						Username:     "user",
						PasswordFile: plainFile,
					},
				}
			},
			expectedResult: "secret123",
			expectError:    false,
		},
		{
			description: "password file does not exist returns error",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: "/tmp/does-not-exist-getset-test",
					},
				}
			},
			expectedResult: "",
			expectError:    true,
		},
		{
			description: "password file set with noop encryption (all nil fields) returns raw file content",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: plainFile,
					},
					Encryption: &Encryption{},
				}
			},
			expectedResult: "secret123",
			expectError:    false,
		},
		{
			description: "password file missing with encryption set returns error",
			server: func() Server {
				return Server{
					Auth: Auth{
						Enabled:      true,
						PasswordFile: "/tmp/does-not-exist-getset-test",
					},
					Encryption: &Encryption{},
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
