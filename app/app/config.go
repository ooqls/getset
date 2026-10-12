package app

import (
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	LoggingAPI LoggingAPIConfig `yaml:"logging_api"`
	Gin        GinConfig        `yaml:"gin"`
	Docs       DocsConfig       `yaml:"docs"`
	TLS        TLSConfig        `yaml:"tls"`
	JWT        JWTConfig        `yaml:"jwt"`
	SQL        SQLConfig        `yaml:"sql"`
	Registry   RegistryConfig   `yaml:"registry"`
	Health     HealthConfig     `yaml:"health"`
	HTTP       HTTPConfig       `yaml:"http"`
	RSA        RSAConfig        `yaml:"rsa"`
	Cache      CacheConfig      `yaml:"cache"`
	Redis      RedisConfig      `yaml:"redis"`
	Valkey     ValkeyConfig     `yaml:"valkey"`
	Grpc       GrpcConfig       `yaml:"grpc"`
	SQLite     SQLiteConfig     `yaml:"sqlite"`
}

func LoadConfig(path string) (*AppConfig, error) {
	cfg := &AppConfig{}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func WithConfig(cfg *AppConfig) Features {
	return Features{
		LoggingAPI: loggingApiFromConfig(cfg.LoggingAPI),
		Cache:      cacheFromConfig(cfg.Cache),
		RSA:        rsaFromConfig(cfg.RSA),
		JWT:        jwtFromConfig(cfg.JWT),
		SQL:        sqlFromConfig(cfg.SQL),
		HTTP:       httpFromConfig(cfg.HTTP),
		TLS:        tlsFromConfig(cfg.TLS),
		Registry:   registryFromConfig(cfg.Registry),
		Redis:      redisFromConfig(cfg.Redis),
		Valkey:     valkeyFromConfig(cfg.Valkey),
		Docs:       docsFromConfig(cfg.Docs),
		Health:     healthFromConfig(cfg.Health),
		Gin:        ginFromConfig(cfg.Gin),
		Grpc:       grpcFromConfig(cfg.Grpc),
		SQLite:     sqliteFromConfig(cfg.SQLite),
	}
}
