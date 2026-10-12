package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

func New(appName string, features Features) *App {
	return &App{
		appName:  appName,
		l:        log.NewLogger(appName),
		features: features,
	}
}

type App struct {
	appName         string
	setup           func(ctx *AppContext) error
	running         func(ctx *AppContext) error
	stopped         func(ctx *AppContext) error
	onPanic         func(err interface{})
	l               *zap.Logger
	isRunning       atomic.Bool
	features        Features
	testEnvironment *TestEnvironment
}

func (a *App) WithTestEnvironment(env TestEnvironment) {
	a.testEnvironment = &env
}

func (a *App) IsRunning() bool {
	return a.isRunning.Load()
}

func (a *App) OnStartup(f func(ctx *AppContext) error) *App {
	a.setup = f
	return a
}

func (a *App) OnRunning(f func(ctx *AppContext) error) *App {
	a.running = f
	return a
}

func (a *App) OnStopped(f func(ctx *AppContext) error) *App {
	a.stopped = f
	return a
}

// IsHealthy reports whether the App is running and, when the Health feature
// is enabled, its last health check passed.
func (a *App) IsHealthy() bool {
	return a.IsRunning() && a.features.Health.healthy()
}

func (a *App) SetHealthCheck(f func() bool) *App {
	a.features.Health.Check = f
	return a
}

func (a *App) Features() *Features {
	return &a.features
}

// Run starts every enabled feature, runs the lifecycle hooks, and blocks
// until ctx is cancelled or the process receives SIGINT/SIGTERM.
func (a *App) Run(ctx context.Context) error {
	flag.Parse()
	applyFlags(&a.features)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if a.testEnvironment != nil {
		cleanup, err := a.testEnvironment.Start(context.Background())
		if err != nil {
			return fmt.Errorf("failed to start test environment: %v", err)
		}
		defer cleanup()
	}

	err := a.run(ctx)
	if err != nil {
		a.l.Error("app stopped with an error", zap.Error(err))
	}
	return err
}

func (a *App) run(parent context.Context) (err error) {
	l := a.l
	if a.onPanic != nil {
		defer func() {
			if r := recover(); r != nil {
				l.Warn("recovered from panic", zap.Any("error", r))
				a.onPanic(r)
			}
		}()
	}

	ctx, cancel := context.WithCancel(parent)
	appCtx := NewAppContext(ctx, l)
	// On failure, stop whatever already started before returning.
	defer func() {
		if err != nil {
			cancel()
			appCtx.wg.Wait()
		}
	}()
	defer cancel()

	fs := &a.features
	for _, f := range fs.ordered() {
		if !f.enabled() {
			continue
		}
		l.Info("[Startup] starting feature", zap.String("feature", f.name()))
		if err := f.start(appCtx, fs); err != nil {
			return fmt.Errorf("failed to start %s: %w", f.name(), err)
		}
	}

	if a.setup != nil {
		l.Debug("[Startup] running OnStartup")
		if err := a.setup(appCtx); err != nil {
			return fmt.Errorf("OnStartup failed: %w", err)
		}
	}

	for _, f := range fs.ordered() {
		s, ok := f.(server)
		if !ok || !s.enabled() {
			continue
		}
		l.Info("[Serve] serving feature", zap.String("feature", s.name()))
		if err := s.serve(appCtx, fs); err != nil {
			return fmt.Errorf("failed to serve %s: %w", s.name(), err)
		}
	}

	a.isRunning.Store(true)
	if a.running != nil {
		appCtx.spawn(func() {
			l.Debug("[Running] running OnRunning")
			if err := a.running(appCtx); err != nil {
				l.Error("[Running] OnRunning returned an error", zap.Error(err))
			}
		})
	}

	appCtx.wg.Wait()
	a.isRunning.Store(false)
	l.Debug("[Stopped] app stopped")

	if a.stopped != nil {
		if err := a.stopped(appCtx); err != nil {
			return fmt.Errorf("OnStopped failed: %w", err)
		}
	}

	return nil
}
