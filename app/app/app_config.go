package app

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/ooqls/getset/crypto/jwt"
	"gopkg.in/yaml.v2"
)

type DocsConfig struct {
	Enabled     bool   `yaml:"enabled"`
	DocsApiPath string `yaml:"docs_api_path"`
	DocsDir     string `yaml:"docs_dir"`
	DocsPort    int    `yaml:"docs_port"`
}

type ServerConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CaPath   string `yaml:"ca_path"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type JWTConfig struct {
	Enabled                 bool                     `yaml:"enabled"`
	RSAKeyPath              string                   `yaml:"rsa_key_path"`
	RSAPubKeyPath           string                   `yaml:"rsa_pub_key_path"`
	TokenConfigurationPaths []string                 `yaml:"token_configuration_paths"`
	TokenConfigurations     []jwt.TokenConfiguration `yaml:"token_configurations"`
}

type SQLFilesConfig struct {
	Enabled          bool       `yaml:"enabled"`
	SQLPackage       sqlPackage `yaml:"sql_package"`
	SQLFilesDirs     []string   `yaml:"sql_files_dirs"`
	SQLFiles         []string   `yaml:"sql_files"`
	CreateTableStmts []string   `yaml:"create_table_stmts"`
	CreateIndexStmts []string   `yaml:"create_index_stmts"`
}

type RegistryConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type LoggingAPIConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type HealthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

type CorsConfig struct {
	Enabled                bool     `yaml:"enabled"`
	AllowAllOrigins        bool     `yaml:"allow_all_origins"`
	AllowOrigins           []string `yaml:"origins"`
	AllowMethods           []string `yaml:"methods"`
	Headers                []string `yaml:"headers"`
	ExposeHeaders          []string `yaml:"expose_headers"`
	AllowCredentials       bool     `yaml:"allow_credentials"`
	AllowWildcard          bool     `yaml:"allow_wildcard"`
	AllowBrowserExtensions bool     `yaml:"allow_browser_extensions"`
	AllowWebSockets        bool     `yaml:"allow_web_sockets"`
	AllowFiles             bool     `yaml:"allow_files"`
	AllowPrivateNetwork    bool     `yaml:"allow_private_network"`
	MaxAge                 int      `yaml:"max_age"`
}

func (c *CorsConfig) CorsConfig() *cors.Config {
	if !c.Enabled {
		return nil
	}

	return &cors.Config{
		AllowAllOrigins:        c.AllowAllOrigins,
		AllowOrigins:           c.AllowOrigins,
		AllowMethods:           c.AllowMethods,
		AllowHeaders:           c.Headers,
		ExposeHeaders:          c.ExposeHeaders,
		AllowCredentials:       c.AllowCredentials,
		AllowWildcard:          c.AllowWildcard,
		AllowBrowserExtensions: c.AllowBrowserExtensions,
		AllowWebSockets:        c.AllowWebSockets,
		AllowFiles:             c.AllowFiles,
		AllowPrivateNetwork:    c.AllowPrivateNetwork,
		MaxAge:                 time.Duration(c.MaxAge) * time.Second,
	}
}

type GinConfig struct {
	Enabled bool       `yaml:"enabled"`
	Port    int        `yaml:"port"`
	Cors    CorsConfig `yaml:"cors"`
}

type HTTPConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type RSAConfig struct {
	Enabled        bool   `yaml:"enabled"`
	PrivateKeyPath string `yaml:"private_key_path"`
	PublicKeyPath  string `yaml:"public_key_path"`
}

type AppConfig struct {
	LoggingAPI   LoggingAPIConfig `yaml:"logging_api"`
	Gin          GinConfig        `yaml:"gin"`
	DocsConfig   DocsConfig       `yaml:"docs"`
	ServerConfig ServerConfig     `yaml:"server"`
	TLS          TLSConfig        `yaml:"tls"`
	JWT          JWTConfig        `yaml:"jwt"`
	SQLFiles     SQLFilesConfig   `yaml:"sql"`
	Registry     RegistryConfig   `yaml:"registry"`
	Health       HealthConfig     `yaml:"health"`
	HTTP         HTTPConfig       `yaml:"http"`
	RSA          RSAConfig        `yaml:"rsa"`
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
