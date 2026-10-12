package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// feature is one piece of infrastructure an App can turn on. start runs
// during the startup phase, before the user's OnStartup hook.
type feature interface {
	name() string
	enabled() bool
	start(ctx *AppContext, fs *Features) error
}

// server is a feature that accepts traffic. serve runs after the user's
// OnStartup hook, so every route is registered before traffic arrives. It
// must not block: long-running work goes through ctx.spawn and stops when
// ctx is cancelled.
type server interface {
	feature
	serve(ctx *AppContext, fs *Features) error
}

type Features struct {
	LoggingAPI LoggingApiFeature
	Cache      CacheFeature
	RSA        RSAFeature
	JWT        JWTFeature
	SQL        SQLFeature
	HTTP       HTTPFeature
	TLS        TLSFeature
	Registry   RegistryFeature
	Redis      RedisFeature
	Valkey     ValkeyFeature
	Docs       DocsFeature
	Health     HealthFeature
	Gin        GinFeature
	Grpc       GrpcFeature
	SQLite     SQLiteFeature
}

// ordered returns every feature in startup order. Servers serve in the same
// relative order.
func (fs *Features) ordered() []feature {
	return []feature{
		// Gin CORS middleware only applies to routes added after it, so Gin
		// goes before anything that registers routes.
		&fs.Gin,
		&fs.HTTP,
		&fs.Grpc,
		// Registry before anything that looks up a backend in it.
		&fs.Registry,
		// TLS before anything that serves or dials.
		&fs.TLS,
		&fs.Redis,
		&fs.Valkey,
		// Cache uses the Redis or Valkey connection.
		&fs.Cache,
		&fs.RSA,
		&fs.JWT,
		&fs.SQL,
		&fs.SQLite,
		&fs.Docs,
		&fs.LoggingAPI,
		// Health polls Gin or HTTP, so it serves after them.
		&fs.Health,
	}
}

// handle registers h for GET requests at path on every enabled router.
func (fs *Features) handle(path string, h http.Handler) {
	if fs.HTTP.Enabled {
		fs.HTTP.Mux.Handle("GET "+path, h)
	}
	if fs.Gin.Enabled {
		fs.Gin.Engine.GET(path, gin.WrapH(h))
	}
}

// handlePrefix registers h for GET requests under prefix on every enabled
// router, with prefix stripped from the request path.
func (fs *Features) handlePrefix(prefix string, h http.Handler) {
	prefix = strings.TrimSuffix(prefix, "/")
	h = http.StripPrefix(prefix, h)
	if fs.HTTP.Enabled {
		fs.HTTP.Mux.Handle("GET "+prefix+"/", h)
	}
	if fs.Gin.Enabled {
		fs.Gin.Engine.GET(prefix+"/*filepath", gin.WrapH(h))
	}
}
