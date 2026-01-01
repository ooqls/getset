package app

import (
	"context"

	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/crypto/jwt"
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
		issuerToTokenConfigs: make(map[string]jwt.TokenConfiguration),
	}
}

type AppContext struct {
	context.Context
	l                    *zap.Logger
	issuerToTokenConfigs map[string]jwt.TokenConfiguration
	cacheFactory         factory.CacheFactory
}

func (ctx *AppContext) L() *zap.Logger {
	return ctx.l
}

func (ctx *AppContext) AuthIssuerConfig() (*jwt.TokenConfiguration, bool) {
	config, ok := ctx.issuerToTokenConfigs[AuthIssuer]
	return &config, ok
}

func (ctx *AppContext) RefreshIssuerConfig() (*jwt.TokenConfiguration, bool) {
	config, ok := ctx.issuerToTokenConfigs[RefreshIssuer]
	return &config, ok
}

func (ctx *AppContext) CacheFactory() (factory.CacheFactory, bool) {
	return ctx.cacheFactory, ctx.cacheFactory != nil
}

func (ctx *AppContext) WithCacheFactory(factory factory.CacheFactory) *AppContext {
	ctx.cacheFactory = factory
	return ctx
}
