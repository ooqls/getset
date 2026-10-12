package app

import (
	"net"

	v1 "github.com/ooqls/getset/log/api/v1"
)

type LoggingAPIConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type LoggingApiOption func(*LoggingApiFeature)

func WithLoggingApiPort(port int) LoggingApiOption {
	return func(f *LoggingApiFeature) { f.Port = port }
}

// LoggingApiFeature serves the log-level API on its own port.
type LoggingApiFeature struct {
	Enabled bool
	Port    int
	addr    net.Addr
}

func LoggingApi(opts ...LoggingApiOption) LoggingApiFeature {
	f := LoggingApiFeature{Enabled: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func loggingApiFromConfig(c LoggingAPIConfig) LoggingApiFeature {
	return LoggingApiFeature{Enabled: c.Enabled, Port: c.Port}
}

func (f *LoggingApiFeature) name() string  { return "logging-api" }
func (f *LoggingApiFeature) enabled() bool { return f.Enabled }

// Addr is the address the logging API is bound to, or nil before it serves.
func (f *LoggingApiFeature) Addr() net.Addr { return f.addr }

func (f *LoggingApiFeature) start(ctx *AppContext, fs *Features) error {
	return nil
}

func (f *LoggingApiFeature) serve(ctx *AppContext, fs *Features) error {
	addr, err := serveHTTP(ctx, fs, f.name(), f.Port, v1.Std())
	f.addr = addr
	return err
}
