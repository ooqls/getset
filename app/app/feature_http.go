package app

import (
	"net"
	"net/http"
)

type HTTPConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type HTTPOption func(*HTTPFeature)

func WithHttpPort(port int) HTTPOption {
	return func(f *HTTPFeature) { f.Port = port }
}

func WithHttpMux(mux *http.ServeMux) HTTPOption {
	return func(f *HTTPFeature) { f.Mux = mux }
}

type HTTPFeature struct {
	Enabled bool
	Port    int
	Mux     *http.ServeMux
	addr    net.Addr
}

func HTTP(opts ...HTTPOption) HTTPFeature {
	f := HTTPFeature{
		Enabled: true,
		Mux:     http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func httpFromConfig(c HTTPConfig) HTTPFeature {
	return HTTPFeature{
		Enabled: c.Enabled,
		Port:    c.Port,
		Mux:     http.NewServeMux(),
	}
}

func (f *HTTPFeature) name() string  { return "http" }
func (f *HTTPFeature) enabled() bool { return f.Enabled }

// Addr is the address the HTTP server is bound to, or nil before it serves.
func (f *HTTPFeature) Addr() net.Addr { return f.addr }

func (f *HTTPFeature) start(ctx *AppContext, fs *Features) error {
	if f.Mux == nil {
		f.Mux = http.NewServeMux()
	}
	return nil
}

func (f *HTTPFeature) serve(ctx *AppContext, fs *Features) error {
	addr, err := serveHTTP(ctx, fs, f.name(), f.Port, f.Mux)
	f.addr = addr
	return err
}
