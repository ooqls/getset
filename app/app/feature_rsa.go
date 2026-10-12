package app

import (
	"github.com/ooqls/getset/crypto/keys"
	"go.uber.org/zap"
)

type RSAConfig struct {
	Enabled        bool   `yaml:"enabled"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
}

type RSAOption func(*RSAFeature)

func WithPrivateKeyPath(p string) RSAOption {
	return func(f *RSAFeature) { f.PrivateKeyPath = p }
}

func WithPublicKeyPath(p string) RSAOption {
	return func(f *RSAFeature) { f.PublicKeyPath = p }
}

// RSAFeature loads the process RSA key pair from the given paths, or
// generates one when no paths are given. If paths are given but a file is
// missing, the key is left unset.
type RSAFeature struct {
	Enabled        bool
	PrivateKeyPath string
	PublicKeyPath  string
}

func RSA(opts ...RSAOption) RSAFeature {
	f := RSAFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func rsaFromConfig(c RSAConfig) RSAFeature {
	return RSAFeature(c)
}

func (f *RSAFeature) name() string  { return "rsa" }
func (f *RSAFeature) enabled() bool { return f.Enabled }

func (f *RSAFeature) start(ctx *AppContext, fs *Features) error {
	l := ctx.L()

	if f.PrivateKeyPath == "" || f.PublicKeyPath == "" {
		l.Debug("[Startup RSA] no RSA key paths provided, generating a key")
		rsa, err := keys.NewRSA()
		if err != nil {
			return err
		}
		keys.SetRSA(rsa)
		return nil
	}

	if !fileExists(f.PrivateKeyPath) {
		l.Info("[Startup RSA] RSA private key not found", zap.String("path", f.PrivateKeyPath))
		return nil
	}
	if !fileExists(f.PublicKeyPath) {
		l.Info("[Startup RSA] RSA public key not found", zap.String("path", f.PublicKeyPath))
		return nil
	}

	l.Debug("[Startup RSA] loading RSA keys",
		zap.String("rsa_private_key_path", f.PrivateKeyPath),
		zap.String("rsa_pub_key_path", f.PublicKeyPath),
	)
	return keys.InitRSA(mustReadFile(f.PrivateKeyPath), mustReadFile(f.PublicKeyPath))
}
