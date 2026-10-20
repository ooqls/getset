package app

import (
	"context"
	"sync"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/email"
	"go.uber.org/zap"
)

const (
	AuthIssuer    = "auth"
	RefreshIssuer = "refresh"
)

func NewAppContext(ctx context.Context, l *zap.Logger) *AppContext {
	return &AppContext{
		l:                    l,
		Context:              ctx,
		wg:                   &sync.WaitGroup{},
		issuerToTokenConfigs: make(map[string]jwt.TokenConfiguration),
	}
}

type AppContext struct {
	context.Context
	l                    *zap.Logger
	wg                   *sync.WaitGroup
	issuerToTokenConfigs map[string]jwt.TokenConfiguration
	cacheStore           cache.Store
	emailClient          email.EmailClient
}

func (ctx *AppContext) L() *zap.Logger {
	return ctx.l
}

// spawn runs f in a goroutine the App waits for before it stops.
func (ctx *AppContext) spawn(f func()) {
	ctx.wg.Go(f)
}

func (ctx *AppContext) AuthIssuerConfig() (*jwt.TokenConfiguration, bool) {
	config, ok := ctx.issuerToTokenConfigs[AuthIssuer]
	return &config, ok
}

func (ctx *AppContext) RefreshIssuerConfig() (*jwt.TokenConfiguration, bool) {
	config, ok := ctx.issuerToTokenConfigs[RefreshIssuer]
	return &config, ok
}

// CacheStore returns the store the Cache feature chose, or an in-memory
// store when the Cache feature is off. Use cache.New to get a namespace in it.
func (ctx *AppContext) CacheStore() cache.Store {
	if ctx.cacheStore == nil {
		ctx.L().Warn("cache store not set, using memory caching")
		ctx.cacheStore = cache.NewMemStore()
	}

	return ctx.cacheStore
}

func (ctx *AppContext) EmailClient() (email.EmailClient, bool) {
	return ctx.emailClient, ctx.emailClient != nil
}

func (ctx *AppContext) WithCacheStore(store cache.Store) *AppContext {
	ctx.cacheStore = store
	return ctx
}
