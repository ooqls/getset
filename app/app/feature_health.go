package app

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type HealthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	Interval int    `yaml:"interval"`
}

type HealthOption func(*HealthFeature)

func WithHealthPath(path string) HealthOption {
	return func(f *HealthFeature) { f.Path = path }
}

func WithHealthInterval(seconds int) HealthOption {
	return func(f *HealthFeature) { f.Interval = seconds }
}

// HealthFeature serves a health endpoint at Path on the Gin and HTTP routers
// and polls it every Interval seconds. Check, when set, decides the result.
type HealthFeature struct {
	Enabled  bool
	Path     string
	Interval int
	Check    func() bool
	status   *atomic.Bool
}

func Health(opts ...HealthOption) HealthFeature {
	f := HealthFeature{Enabled: true, Interval: 30}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func healthFromConfig(c HealthConfig) HealthFeature {
	return HealthFeature{Enabled: c.Enabled, Path: c.Path, Interval: c.Interval}
}

func (f *HealthFeature) name() string  { return "health" }
func (f *HealthFeature) enabled() bool { return f.Enabled }

// healthy is the result of the last check; true until a check fails, and
// always true when the feature is off.
func (f *HealthFeature) healthy() bool {
	return f.status == nil || f.status.Load()
}

func (f *HealthFeature) start(ctx *AppContext, fs *Features) error {
	ctx.L().Info("[Startup Health] serving health", zap.String("path", f.Path))
	f.status = &atomic.Bool{}
	f.status.Store(true)

	fs.handle(f.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := f.Check == nil || f.Check()
		f.status.Store(ok)
		w.Header().Set("Content-Type", "application/json")
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"error"}`))
		}
	}))
	return nil
}

// serve starts polling the health endpoint on the HTTP router, or Gin's when
// HTTP is off.
func (f *HealthFeature) serve(ctx *AppContext, fs *Features) error {
	var addr net.Addr
	switch {
	case fs.HTTP.Enabled:
		addr = fs.HTTP.Addr()
	case fs.Gin.Enabled:
		addr = fs.Gin.Addr()
	default:
		ctx.L().Warn("[Serve Health] no Gin or HTTP router to serve health on")
		return nil
	}

	scheme := "http"
	if fs.TLS.Enabled {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://localhost:%d%s", scheme, addr.(*net.TCPAddr).Port, f.Path)
	client := fs.TLS.Client()
	interval := time.Duration(f.Interval) * time.Second

	ctx.spawn(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				resp, err := client.Get(url)
				if err != nil {
					ctx.L().Error("[Health] health check request failed", zap.Error(err))
					continue
				}
				resp.Body.Close()
			}
		}
	})
	return nil
}
