package app

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

func init() {
	flag.StringVar(&registryPathFlag, "registry", "", "Path to the registry path")
	flag.StringVar(&sqlFilesFlag, "sql-files", "", "Comma separated list of files")
	flag.StringVar(&rsaPrivKeyPathFlag, "rsa-private-key", "", "Path to an RSA private key")
	flag.StringVar(&rsaPubKeyPathFlag, "rsa-public-key", "", "Path to the RSA public key")
	flag.StringVar(&jwtPrivKeyPathFlag, "jwt-private-key", "", "Path to a JWT private key")
	flag.StringVar(&jwtPubKeyPathFlag, "jwt-public-key", "", "Path to a jwt public key")
	flag.StringVar(&tlsKeyPathFlag, "tls-key-path", "", "Path to the TLS key file")
	flag.StringVar(&tlsCertPathFlag, "tls-cert-path", "", "Path to the TLS cert file")
	flag.StringVar(&tlsCaPathFlag, "tls-ca-path", "", "Path to the TLS CA file")
	flag.StringVar(&healthPathFlag, "health-path", "", "Path to the health path")
	flag.IntVar(&docsPortFlag, "docs-port", 8080, "Port to serve docs on")
	flag.StringVar(&docsPathFlag, "docs-path", "/docs/", "Path to the docs directory")
	flag.StringVar(&docsApiPathFlag, "docs-api-path", "/api/docs", "Path to the docs API")
}

func New(appName string, features Features) *app {
	return &app{
		appName:    appName,
		l:          log.NewLogger(appName),
		features:   features,
		threadWg:   &sync.WaitGroup{},
		httpClient: http.DefaultClient,
	}
}

type app struct {
	appName         string
	setup           func(ctx *AppContext) error
	running         func(ctx *AppContext) error
	stopped         func(ctx *AppContext) error
	healthCheck     func() bool
	onPanic         func(err interface{})
	l               *zap.Logger
	state           AppState
	features        Features
	testEnvironment *TestEnvironment
	httpClient      *http.Client
	stopServers     []func() (string, error)
	threadWg        *sync.WaitGroup
}

func (a *app) WithTestEnvironment(env TestEnvironment) {
	a.testEnvironment = &env
}

func (a *app) IsRunning() bool {
	return a.state.Running
}

func (a *app) OnStartup(f func(ctx *AppContext) error) *app {
	a.setup = func(ctx *AppContext) error {
		if a.features.Gin.Enabled {
			a.features.Gin.Engine.Use(cors.New(*a.features.Gin.Cors))
		}

		return f(ctx)
	}
	return a
}

func (a *app) OnRunning(f func(ctx *AppContext) error) *app {
	a.running = f
	return a
}

func (a *app) OnStopped(f func(ctx *AppContext) error) *app {
	a.stopped = f
	return a
}

func (a *app) IsHealthy() bool {
	return a.state.Healthy
}

func (a *app) SetHealthCheck(f func() bool) *app {
	a.healthCheck = f
	return a
}

func (a *app) Run(ctx context.Context) error {
	flag.Parse()
	if a.testEnvironment != nil {
		cleanup, err := a.testEnvironment.Start(context.Background())
		if err != nil {
			return fmt.Errorf("failed to start test environment: %v", err)
		}
		defer cleanup()
	}

	if err := a._startup(ctx); err != nil {
		return err
	}

	return nil
}

func (a *app) Features() *Features {
	return &a.features
}
