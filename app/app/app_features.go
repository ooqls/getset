package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type featureOpt struct {
	key   string
	value interface{}
}

func WithConfig(cfg *AppConfig) Features {
	return Features{
		LoggingAPI: LoggingApiFeature{
			Enabled: cfg.LoggingAPI.Enabled,
			Port:    cfg.LoggingAPI.Port,
		},
		HTTP: HTTPFeature{
			Enabled: cfg.HTTP.Enabled,
			Port:    cfg.HTTP.Port,
			Mux:     http.NewServeMux(),
		},
		Gin: GinFeature{
			Enabled: cfg.Gin.Enabled,
			Port:    cfg.Gin.Port,
			Engine:  gin.New(),
			Cors:    cfg.Gin.Cors.CorsConfig(),
		},
		Cache: CacheFeature{
			Enabled:   cfg.Cache.Enabled,
			CacheType: cacheType(cfg.Cache.CacheType),
		},
		Docs: DocsFeature{
			Enabled:     cfg.DocsConfig.Enabled,
			DocsPath:    cfg.DocsConfig.DocsDir,
			DocsApiPath: cfg.DocsConfig.DocsApiPath,
			DocsPort:    cfg.DocsConfig.DocsPort,
			DocsContent: cfg.DocsConfig.DocsContent,
		},
		TLS: TLSFeature{
			Enabled:        cfg.TLS.Enabled,
			ServerCertFile: cfg.TLS.CertFile,
			ServerKeyFile:  cfg.TLS.KeyFile,
			CAFile:         cfg.TLS.CaPath,
		},
		RSA: RSAFeature{
			Enabled:        cfg.RSA.Enabled,
			PrivateKeyPath: cfg.RSA.PrivateKeyPath,
			PublicKeyPath:  cfg.RSA.PublicKeyPath,
		},
		JWT: JWTFeature{
			Enabled:                 cfg.JWT.Enabled,
			tokenConfigurationPaths: cfg.JWT.TokenConfigurationPaths,
			PrivateKeyPath:          cfg.JWT.RSAKeyPath,
			PubKeyPath:              cfg.JWT.RSAPubKeyPath,
			tokenConfiguration:      cfg.JWT.TokenConfigurations,
		},
		Health: HealthFeature{
			Enabled:  cfg.Health.Enabled,
			Path:     cfg.Health.Path,
			Interval: cfg.Health.Interval,
		},
		SQL: SQLFeature{
			Enabled:               cfg.SQLFiles.Enabled,
			SQLPackage:            cfg.SQLFiles.SQLPackage,
			SQLFiles:              cfg.SQLFiles.SQLFiles,
			SQLDirs:               cfg.SQLFiles.SQLFilesDirs,
			CreateTableStatements: cfg.SQLFiles.CreateTableStmts,
			CreateIndexStatements: cfg.SQLFiles.CreateIndexStmts,
		},
		Registry: RegistryFeature{
			enabled:      cfg.Registry.Enabled,
			registryPath: &cfg.Registry.Path,
		},
		Grpc: GrpcFeature{
			Enabled: cfg.Grpc.Enabled,
			Port:    cfg.Grpc.Port,
			Server:  grpc.NewServer(),
		},
		SQLite: func() SQLiteFeature {
			f := SQLiteFeature{Enabled: cfg.SQLite.Enabled}
			for _, db := range cfg.SQLite.Databases {
				f.Databases = append(f.Databases, SQLiteDB{
					Name:   db.Name,
					Path:   db.Path,
					Schema: db.Schema,
				})
			}
			return f
		}(),
	}
}

type Features struct {
	LoggingAPI LoggingApiFeature
	Cache      CacheFeature
	RSA        RSAFeature
	JWT        JWTFeature
	SQL        SQLFeature
	Redis      RedisFeature
	Valkey     ValkeyFeature
	HTTP       HTTPFeature
	TLS        TLSFeature
	Registry   RegistryFeature
	Docs       DocsFeature
	Health     HealthFeature
	Gin        GinFeature
	Grpc       GrpcFeature
	SQLite     SQLiteFeature
}
