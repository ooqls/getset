package app

import (
	"fmt"

	"github.com/ooqls/getset/db/valkey"
	"github.com/ooqls/getset/registry"
	"go.uber.org/zap"
)

type ValkeyConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"password_file"`
}

type ValkeyOption func(*ValkeyFeature)

func WithValkeyHost(host string) ValkeyOption {
	return func(f *ValkeyFeature) { f.Host = host }
}

func WithValkeyPort(port int) ValkeyOption {
	return func(f *ValkeyFeature) { f.Port = port }
}

func WithValkeyUser(username string) ValkeyOption {
	return func(f *ValkeyFeature) { f.Username = username }
}

func WithValkeyPasswordFile(path string) ValkeyOption {
	return func(f *ValkeyFeature) { f.PasswordFile = path }
}

// ValkeyFeature connects to Valkey. It starts from the Registry's Valkey
// entry when the Registry feature is enabled; any field set here overrides it.
type ValkeyFeature struct {
	Enabled      bool
	Host         string
	Port         int
	Username     string
	PasswordFile string
}

func Valkey(opts ...ValkeyOption) ValkeyFeature {
	f := ValkeyFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func valkeyFromConfig(c ValkeyConfig) ValkeyFeature {
	return ValkeyFeature(c)
}

func (f *ValkeyFeature) name() string  { return "valkey" }
func (f *ValkeyFeature) enabled() bool { return f.Enabled }

func (f *ValkeyFeature) start(ctx *AppContext, fs *Features) error {
	db := registry.Database{}
	if registry.IsSet() && registry.Get().Valkey != nil {
		db = *registry.Get().Valkey
	}
	if f.Host != "" {
		db.Host = f.Host
	}
	if f.Port != 0 {
		db.Port = f.Port
	}
	if f.Username != "" {
		db.Auth.Enabled = true
		db.Auth.Username = f.Username
	}
	if f.PasswordFile != "" {
		db.Auth.Enabled = true
		db.Auth.PasswordFile = f.PasswordFile
	}
	if db.Host == "" {
		return fmt.Errorf("no valkey host: set one on the Valkey feature or in the registry")
	}

	ctx.L().Debug("[Startup Valkey] connecting", zap.String("host", db.Host), zap.Int("port", db.Port))
	return valkey.Init(&db)
}
