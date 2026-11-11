package app

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
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

const defaultHealthInterval = 30

func WithHealthPath(path string) HealthOption {
	return func(f *HealthFeature) { f.Path = path }
}

func WithHealthInterval(seconds int) HealthOption {
	return func(f *HealthFeature) { f.Interval = seconds }
}

// HealthFeature evaluates Check directly every Interval seconds during the
// Serve phase, with or without routers. GET requests on enabled Gin/HTTP
// routers also evaluate Check and return its result. The initial status is
// healthy; each completed check replaces it.
type HealthFeature struct {
	Enabled  bool
	Path     string
	Interval int
	// Check is startup configuration. Use App.SetHealthCheck to replace it
	// while running. Evaluations are serialized; callbacks must return and
	// must not wait for another evaluation of this same Health Feature.
	Check func() bool
	state *healthState
}

// Runtime synchronization stays behind the Feature interface. HealthFeature
// remains copyable configuration; each App receives fresh runtime state.
type healthState struct {
	checkMu    sync.RWMutex
	evaluating chan struct{}
	status     atomic.Bool
	logger     *zap.Logger
	lifecycle  context.Context
}

func newHealthState() *healthState {
	s := &healthState{evaluating: make(chan struct{}, 1), logger: zap.NewNop(), lifecycle: context.Background()}
	s.status.Store(true)
	return s
}

func Health(opts ...HealthOption) HealthFeature {
	f := HealthFeature{Enabled: true, Interval: defaultHealthInterval}
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
	return !f.Enabled || f.state == nil || f.state.status.Load()
}

func (f *HealthFeature) interval() (time.Duration, error) {
	seconds := f.Interval
	if seconds == 0 {
		seconds = defaultHealthInterval
	}
	if seconds < 0 || int64(seconds) > math.MaxInt64/int64(time.Second) {
		return 0, fmt.Errorf("health interval must be positive and fit a time.Duration: %d seconds", f.Interval)
	}
	return time.Duration(seconds) * time.Second, nil
}

func (f *HealthFeature) setCheck(check func() bool) {
	// App.New initializes state before exposing the App to concurrent callers.
	f.state.checkMu.Lock()
	defer f.state.checkMu.Unlock()
	f.Check = check
}

func (f *HealthFeature) evaluate(ctx context.Context) (ok bool) {
	// A channel gate serializes checks without trapping canceled requests or
	// the scheduler behind an active callback that has not returned yet.
	select {
	case f.state.evaluating <- struct{}{}:
		defer func() { <-f.state.evaluating }()
	case <-ctx.Done():
		return f.state.status.Load()
	case <-f.state.lifecycle.Done():
		return f.state.status.Load()
	}
	// Cancellation may have happened while waiting for another evaluation.
	if ctx.Err() != nil || f.state.lifecycle.Err() != nil {
		return f.state.status.Load()
	}
	// Direct periodic evaluation must preserve the panic isolation formerly
	// provided by net/http. A broken callback is an unhealthy check.
	defer func() {
		if r := recover(); r != nil {
			ok = false
			f.state.status.Store(false)
			f.state.logger.Error("health check panicked", zap.Any("panic", r))
		}
	}()
	f.state.checkMu.RLock()
	check := f.Check
	f.state.checkMu.RUnlock()
	ok = check == nil || check()
	f.state.status.Store(ok)
	return ok
}

func (f *HealthFeature) start(ctx *AppContext, fs *Features) error {
	interval, err := f.interval()
	if err != nil {
		return err
	}
	f.Interval = int(interval / time.Second)
	if f.state == nil {
		f.state = newHealthState()
	}
	f.state.logger = ctx.L()
	f.state.lifecycle = ctx
	f.state.status.Store(true)
	ctx.L().Info("[Startup Health] registering health", zap.String("path", f.Path))

	fs.handle(f.Path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := f.evaluate(r.Context())
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

// serve starts direct evaluation after the first interval. The delay restarts
// after each evaluation completes, rather than catching up with missed ticks.
// The App owns this goroutine and waits for an active callback during shutdown.
func (f *HealthFeature) serve(ctx *AppContext, fs *Features) error {
	interval, err := f.interval()
	if err != nil {
		return err
	}

	ctx.spawn(func() {
		// Endpoint checks share the gate too. Shutdown must wait for one that
		// was already active, even if the periodic timer had not fired yet.
		defer func() {
			f.state.evaluating <- struct{}{}
			<-f.state.evaluating
		}()
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}
			f.evaluate(ctx)
			if ctx.Err() != nil {
				return
			}
			timer.Reset(interval)
		}
	})
	return nil
}
