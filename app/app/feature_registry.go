package app

import (
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

type RegistryConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type RegistryOption func(*RegistryFeature)

func WithRegistryPath(p string) RegistryOption {
	return func(f *RegistryFeature) { f.Path = p }
}

// RegistryFeature loads the Registry from Path, or a localhost Registry when
// Path is empty.
type RegistryFeature struct {
	Enabled bool
	Path    string
}

func Registry(opts ...RegistryOption) RegistryFeature {
	f := RegistryFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func registryFromConfig(c RegistryConfig) RegistryFeature {
	return RegistryFeature{Enabled: c.Enabled, Path: c.Path}
}

func (f *RegistryFeature) name() string  { return "registry" }
func (f *RegistryFeature) enabled() bool { return f.Enabled }

func (f *RegistryFeature) start(ctx *AppContext, fs *Features) error {
	l := ctx.L()

	if f.Path == "" {
		l.Debug("[Startup Registry] no registry path provided, using localhost")
		registry.InitLocalhost()
		return nil
	}

	if !fileExists(f.Path) {
		l.Info("[Startup Registry] registry file not found", zap.String("registry_path", f.Path))
		return ErrRegistryFileNotFound
	}

	l.Debug("[Startup Registry] initializing registry with path", zap.String("path", f.Path))
	return registry.Init(f.Path)
}
