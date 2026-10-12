package app

import (
	"fmt"

	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"go.uber.org/zap"
)

type JWTConfig struct {
	Enabled                 bool                     `yaml:"enabled"`
	RSAKeyPath              string                   `yaml:"rsa_key_path"`
	RSAPubKeyPath           string                   `yaml:"rsa_pub_key_path"`
	TokenConfigurationPaths []string                 `yaml:"token_configuration_paths"`
	TokenConfigurations     []jwt.TokenConfiguration `yaml:"token_configurations"`
}

type JWTOption func(*JWTFeature)

func WithTokenConfigurationPaths(p []string) JWTOption {
	return func(f *JWTFeature) { f.TokenConfigurationPaths = p }
}

func WithTokenConfigurations(cfg []jwt.TokenConfiguration) JWTOption {
	return func(f *JWTFeature) { f.TokenConfigurations = cfg }
}

func WithJWTPrivateKeyPath(p string) JWTOption {
	return func(f *JWTFeature) { f.PrivateKeyPath = p }
}

func WithJWTPublicKeyPath(p string) JWTOption {
	return func(f *JWTFeature) { f.PubKeyPath = p }
}

// JWTFeature loads the JWT signing key (or generates one when no paths are
// given) and puts token issuer configurations on the AppContext. If key
// paths are given but a file is missing, nothing is loaded.
type JWTFeature struct {
	Enabled                 bool
	PrivateKeyPath          string
	PubKeyPath              string
	TokenConfigurationPaths []string
	TokenConfigurations     []jwt.TokenConfiguration
}

func JWT(opts ...JWTOption) JWTFeature {
	f := JWTFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func jwtFromConfig(c JWTConfig) JWTFeature {
	return JWTFeature{
		Enabled:                 c.Enabled,
		PrivateKeyPath:          c.RSAKeyPath,
		PubKeyPath:              c.RSAPubKeyPath,
		TokenConfigurationPaths: c.TokenConfigurationPaths,
		TokenConfigurations:     c.TokenConfigurations,
	}
}

func (f *JWTFeature) name() string  { return "jwt" }
func (f *JWTFeature) enabled() bool { return f.Enabled }

func (f *JWTFeature) start(ctx *AppContext, fs *Features) error {
	l := ctx.L()

	if f.PrivateKeyPath != "" && f.PubKeyPath != "" {
		if !fileExists(f.PrivateKeyPath) {
			l.Info("[Startup JWT] JWT private key does not exist", zap.String("path", f.PrivateKeyPath))
			return nil
		}
		if !fileExists(f.PubKeyPath) {
			l.Info("[Startup JWT] JWT public key does not exist", zap.String("path", f.PubKeyPath))
			return nil
		}
		l.Debug("[Startup JWT] loading JWT keys",
			zap.String("jwt_private_key_path", f.PrivateKeyPath),
			zap.String("jwt_pub_key_path", f.PubKeyPath),
		)
		if err := keys.InitJwt(mustReadFile(f.PrivateKeyPath), mustReadFile(f.PubKeyPath)); err != nil {
			return err
		}
	} else {
		l.Debug("[Startup JWT] no JWT key paths provided, generating a key")
		rsaKey, err := keys.NewRSA()
		if err != nil {
			return err
		}
		keys.SetJwt(keys.NewJWTKey(*rsaKey))
	}

	for _, configPath := range f.TokenConfigurationPaths {
		if !fileExists(configPath) {
			return fmt.Errorf("JWT token config file not found: %s", configPath)
		}
		config, err := jwt.ParseTokenConfigFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to parse token config file %s: %v", configPath, err)
		}
		ctx.issuerToTokenConfigs[config.Issuer] = *config
	}

	for _, cfg := range f.TokenConfigurations {
		ctx.issuerToTokenConfigs[cfg.Issuer] = cfg
	}

	return nil
}
