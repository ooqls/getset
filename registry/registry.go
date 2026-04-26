package registry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/crypto/keys"
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
	Enabled      bool   `yaml:"enabled"`
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"password_file"`
}

func (a *Auth) ResolvePassword(algo crypto.Algorithm) (string, error) {
	if a.PasswordFile == "" {
		return "", nil
	}
	b, err := os.ReadFile(a.PasswordFile)
	if err != nil {
		return "", fmt.Errorf("failed to read password file %s: %v", a.PasswordFile, err)
	}
	decrypted, err := algo.Decrypt(b)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %v", err)
	}

	return string(decrypted), nil
}

// NewAuthWithPassword writes the given password to a temp file and returns
// an Auth with PasswordFile set. Avoids storing the password in the struct.
func NewAuthWithPassword(username, password string) (Auth, error) {
	f, err := os.CreateTemp("", "getset-auth-*")
	if err != nil {
		return Auth{}, fmt.Errorf("failed to create temp password file: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(password); err != nil {
		return Auth{}, fmt.Errorf("failed to write password file: %v", err)
	}
	return Auth{
		Enabled:      true,
		Username:     username,
		PasswordFile: f.Name(),
	}, nil
}

type Topics struct {
	Messages string `yaml:"messages"`
}

type Server struct {
	Name       string                 `yaml:"name"`
	Protocol   string                 `yaml:"protocol"`
	Host       string                 `yaml:"host"`
	Port       int                    `yaml:"port"`
	TLS        *TLSConfig             `yaml:"tls,omitempty"`
	Auth       Auth                   `yaml:"auth"`
	Extra      map[string]interface{} `yaml:"extra"`
	Encryption *Encryption            `yaml:"encryption,omitempty"`
}

func (a *Server) ResolvePassword() (string, error) {
	if a.Auth.PasswordFile == "" {
		return "", nil
	}

	b, err := os.ReadFile(a.Auth.PasswordFile)
	if err != nil {
		return "", fmt.Errorf("failed to read password file %s: %v", a.Auth.PasswordFile, err)
	}

	if a.Encryption == nil {
		return string(b), nil
	}
	
	algo, err := a.Encryption.Algorithm()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption algorithm: %v", err)
	}

	decrypted, err := algo.Decrypt(b)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %v", err)
	}

	return string(decrypted), nil
}

type Encryption struct {
	RSA *struct {
		PrivateKeyPath string `yaml:"private_key_path"`
		PublicKeyPath  string `yaml:"public_key_path"`
	} `yaml:"rsa,omitempty"`
	AESGCM *struct {
		Salt    string `yaml:"salt"`
		KeyPath string `yaml:"key_path"`
	} `yaml:"aes_gcm,omitempty"`
	X509 *struct {
		CertPath string `yaml:"cert_path"`
	} `yaml:"x509,omitempty"`
}

func (e *Encryption) Algorithm() (crypto.Algorithm, error) {

	switch true {
	case e.RSA != nil:
		return crypto.NewRsaAlgorithm(), nil
	case e.AESGCM != nil:
		salt := []byte(e.AESGCM.Salt)
		b, err := os.ReadFile(e.AESGCM.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read AES-GCM key file %s: %v", e.AESGCM.KeyPath, err)
		}

		key, err := crypto.DeriveAESGCMKey(string(b), [crypto.SALT_SIZE]byte(salt))
		if err != nil {
			return nil, fmt.Errorf("failed to derive AES-GCM key: %v", err)
		}

		return crypto.NewAESGCMAlgorithmWithKey(key, [crypto.SALT_SIZE]byte(salt)), nil

	case e.X509 != nil:
		x509, err := keys.ParseX509File(e.X509.CertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x509 certificate for encryption: %v", err)
		}
		return crypto.NewX509Algorithm(x509), nil

	default:
		return crypto.NewNoopAlgorithm(), nil
	}
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
	return fmt.Sprintf("%s://%s:%d", s.Protocol, s.Host, s.Port)
}

type Registry struct {
	Kafka *MessageBroker `yaml:"kafka,omitempty"`
	Nats  *MessageBroker `yaml:"nats,omitempty"`

	Redis         *Database `yaml:"redis,omitempty"`
	Valkey        *Database `yaml:"valkey,omitempty"`
	Postgres      *Database `yaml:"postgres,omitempty"`
	Mongo         *Database `yaml:"mongo,omitempty"`
	Neo4j         *Database `yaml:"neo4j,omitempty"`
	Elasticsearch *Database `yaml:"elasticsearch,omitempty"`
}
