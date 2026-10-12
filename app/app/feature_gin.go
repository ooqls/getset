package app

import (
	"net"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type GinConfig struct {
	Enabled bool       `yaml:"enabled"`
	Port    int        `yaml:"port"`
	Cors    CorsConfig `yaml:"cors"`
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

type GinOption func(*GinFeature)

func WithGin(e *gin.Engine) GinOption {
	return func(f *GinFeature) { f.Engine = e }
}

func WithGinPort(port int) GinOption {
	return func(f *GinFeature) { f.Port = port }
}

func WithGinCors(cors cors.Config) GinOption {
	return func(f *GinFeature) { f.Cors = &cors }
}

type GinFeature struct {
	Enabled bool
	Port    int
	Cors    *cors.Config
	Engine  *gin.Engine
	addr    net.Addr
}

func Gin(opts ...GinOption) GinFeature {
	f := GinFeature{
		Enabled: true,
		Port:    8080,
		Engine:  gin.New(),
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func ginFromConfig(c GinConfig) GinFeature {
	return GinFeature{
		Enabled: c.Enabled,
		Port:    c.Port,
		Engine:  gin.New(),
		Cors:    c.Cors.CorsConfig(),
	}
}

func (f *GinFeature) name() string  { return "gin" }
func (f *GinFeature) enabled() bool { return f.Enabled }

// Addr is the address Gin is bound to, or nil before it serves.
func (f *GinFeature) Addr() net.Addr { return f.addr }

func (f *GinFeature) start(ctx *AppContext, fs *Features) error {
	if f.Engine == nil {
		f.Engine = gin.New()
	}
	if f.Cors != nil {
		f.Engine.Use(cors.New(*f.Cors))
	}
	return nil
}

func (f *GinFeature) serve(ctx *AppContext, fs *Features) error {
	addr, err := serveHTTP(ctx, fs, f.name(), f.Port, f.Engine)
	f.addr = addr
	return err
}
