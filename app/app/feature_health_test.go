package app

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHealth_IntervalDefaultsAndValidation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		feature   HealthFeature
		want      int
		wantError bool
	}{
		{name: "constructor default", feature: Health(), want: 30},
		{name: "zero option", feature: Health(WithHealthInterval(0)), want: 30},
		{name: "omitted config interval", feature: healthFromConfig(HealthConfig{Enabled: true}), want: 30},
		{name: "configured interval", feature: Health(WithHealthInterval(7)), want: 7},
		{name: "negative", feature: Health(WithHealthInterval(-1)), wantError: true},
	}
	if strconv.IntSize == 64 {
		seconds := int64(math.MaxInt64)/int64(time.Second) + 1
		cases = append(cases, struct {
			name      string
			feature   HealthFeature
			want      int
			wantError bool
		}{name: "overflow", feature: Health(WithHealthInterval(int(seconds))), wantError: true})
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := NewAppContext(t.Context(), zap.NewNop())
			err := tt.feature.start(ctx, &Features{})
			if tt.wantError {
				require.ErrorContains(t, err, "health interval")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.feature.Interval)
			assert.True(t, tt.feature.healthy())
		})
	}
}

func TestHealth_InvalidIntervalStopsStartup(t *testing.T) {
	a := New("invalid-health", Features{Health: Health(WithHealthInterval(-1))})
	started := false
	a.OnStartup(func(*AppContext) error { started = true; return context.Canceled })
	require.ErrorContains(t, a.run(t.Context()), "failed to start health")
	assert.False(t, started)
}

func TestHealth_DisabledDoesNotEvaluate(t *testing.T) {
	var calls atomic.Int32
	f := Health(WithHealthInterval(-1))
	f.Enabled = false
	f.Check = func() bool { calls.Add(1); return false }
	a := New("disabled-health", Features{Health: f})
	a.OnRunning(func(*AppContext) error {
		assert.True(t, a.IsHealthy(), "disabled Health does not restrict a running App")
		return nil
	})
	require.NoError(t, a.run(t.Context()))
	assert.Zero(t, calls.Load())
	assert.False(t, a.IsHealthy())
}

func TestHealth_EndpointsEvaluateAndRecover(t *testing.T) {
	a := New("health", Features{HTTP: HTTP(), Gin: Gin(), Health: Health(WithHealthPath("/health"))})
	fs := a.Features()
	startFeatures(t, fs)
	for _, tt := range []struct {
		name    string
		handler http.Handler
	}{
		{name: "HTTP", handler: fs.HTTP.Mux},
		{name: "Gin", handler: fs.Gin.Engine},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			a.SetHealthCheck(func() bool { calls++; return false })
			response := httptest.NewRecorder()
			tt.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.JSONEq(t, `{"status":"error"}`, response.Body.String())
			assert.Equal(t, 1, calls)
			assert.False(t, fs.Health.healthy())
			a.SetHealthCheck(func() bool { calls++; return true })
			assert.False(t, fs.Health.healthy(), "replacing a callback does not evaluate it")
			response = httptest.NewRecorder()
			tt.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
			assert.Equal(t, http.StatusOK, response.Code)
			assert.JSONEq(t, `{"status":"ok"}`, response.Body.String())
			assert.Equal(t, 2, calls)
			assert.True(t, fs.Health.healthy())
			a.SetHealthCheck(nil)
			response = httptest.NewRecorder()
			tt.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
			assert.Equal(t, http.StatusOK, response.Code)
		})
	}
}

func TestHealth_HeadlessEvaluation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("headless-health", Features{Health: Health(WithHealthInterval(1))})
		var failed, recovered atomic.Int32
		a.SetHealthCheck(func() bool { failed.Add(1); return false })
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- a.run(ctx) }()
		synctest.Wait()
		require.True(t, a.IsRunning())
		assert.True(t, a.IsHealthy(), "healthy until the first evaluation")
		assert.Zero(t, failed.Load())
		time.Sleep(time.Second - time.Nanosecond)
		synctest.Wait()
		assert.Zero(t, failed.Load())
		time.Sleep(time.Nanosecond)
		synctest.Wait()
		assert.Equal(t, int32(1), failed.Load())
		assert.False(t, a.IsHealthy())
		a.SetHealthCheck(func() bool { recovered.Add(1); return true })
		assert.False(t, a.IsHealthy(), "IsHealthy only reads cached status")
		time.Sleep(time.Second)
		synctest.Wait()
		assert.Equal(t, int32(1), recovered.Load())
		assert.True(t, a.IsHealthy())
		cancel()
		synctest.Wait()
		require.NoError(t, <-done)
		assert.False(t, a.IsHealthy(), "a stopped App is not healthy")
		time.Sleep(time.Hour)
		synctest.Wait()
		assert.Equal(t, int32(1), failed.Load())
		assert.Equal(t, int32(1), recovered.Load())
	})
}

func TestHealth_SerializesEndpointAndPeriodicChecks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"), WithHealthInterval(1))})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		appCtx := NewAppContext(ctx, zap.NewNop())
		fs := a.Features()
		require.NoError(t, fs.HTTP.start(appCtx, fs))
		require.NoError(t, fs.Health.start(appCtx, fs))
		entered, release := make(chan struct{}, 3), make(chan struct{})
		var releaseOnce sync.Once
		releaseCheck := func() { releaseOnce.Do(func() { close(release) }) }
		t.Cleanup(releaseCheck)
		var active, maximum, replacements atomic.Int32
		a.SetHealthCheck(func() bool {
			n := active.Add(1)
			maximum.Store(n)
			entered <- struct{}{}
			<-release
			active.Add(-1)
			return false
		})
		require.NoError(t, fs.Health.serve(appCtx, fs))
		time.Sleep(time.Second)
		synctest.Wait()
		require.Len(t, entered, 1)
		responses := make(chan int, 2)
		for range 2 {
			go func() {
				response := httptest.NewRecorder()
				fs.HTTP.Mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
				responses <- response.Code
			}()
		}
		synctest.Wait()
		assert.Equal(t, int32(1), active.Load(), "queued endpoint checks must not overlap the periodic check")
		a.SetHealthCheck(func() bool {
			n := active.Add(1)
			if n > maximum.Load() {
				maximum.Store(n)
			}
			replacements.Add(1)
			active.Add(-1)
			return true
		})
		releaseCheck()
		synctest.Wait()
		assert.Equal(t, http.StatusOK, <-responses)
		assert.Equal(t, http.StatusOK, <-responses)
		assert.Equal(t, int32(1), maximum.Load())
		assert.Equal(t, int32(2), replacements.Load())
		assert.True(t, fs.Health.healthy())
		cancel()
		appCtx.wg.Wait()
	})
}

func TestHealth_ShutdownWaitsForActiveCheck(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health-shutdown", Features{Health: Health(WithHealthInterval(1))})
		entered, release := make(chan struct{}), make(chan struct{})
		var releaseOnce sync.Once
		releaseCheck := func() { releaseOnce.Do(func() { close(release) }) }
		t.Cleanup(releaseCheck)
		var calls atomic.Int32
		a.SetHealthCheck(func() bool { calls.Add(1); close(entered); <-release; return true })
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- a.run(ctx) }()
		synctest.Wait()
		time.Sleep(time.Second)
		synctest.Wait()
		select {
		case <-entered:
		default:
			t.Fatal("expected the periodic check to begin")
		}
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("shutdown returned before the active check finished")
		default:
		}
		releaseCheck()
		synctest.Wait()
		require.NoError(t, <-done)
		time.Sleep(time.Hour)
		synctest.Wait()
		assert.Equal(t, int32(1), calls.Load())
	})
}

func TestHealth_ConcurrentReplacement(t *testing.T) {
	a := New("health-replacement", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"))})
	fs := a.Features()
	startFeatures(t, fs)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 1000 {
			if i%2 == 0 {
				a.SetHealthCheck(nil)
			} else {
				a.SetHealthCheck(func() bool { return false })
			}
		}
	})
	for range 4 {
		wg.Go(func() {
			for range 100 {
				response := httptest.NewRecorder()
				fs.HTTP.Mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
				assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, response.Code)
			}
		})
	}
	wg.Wait()
	a.SetHealthCheck(func() bool { return false })
	response := httptest.NewRecorder()
	fs.HTTP.Mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestHealth_CanceledEndpointDoesNotWaitForActiveCheck(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health-cancellation", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"), WithHealthInterval(1))})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		appCtx := NewAppContext(ctx, zap.NewNop())
		fs := a.Features()
		require.NoError(t, fs.HTTP.start(appCtx, fs))
		require.NoError(t, fs.Health.start(appCtx, fs))
		release := make(chan struct{})
		var releaseOnce sync.Once
		releaseCheck := func() { releaseOnce.Do(func() { close(release) }) }
		t.Cleanup(releaseCheck)
		var calls atomic.Int32
		a.SetHealthCheck(func() bool { calls.Add(1); <-release; return false })
		require.NoError(t, fs.Health.serve(appCtx, fs))
		time.Sleep(time.Second)
		synctest.Wait()
		require.Equal(t, int32(1), calls.Load())
		requestCtx, cancelRequest := context.WithCancel(t.Context())
		defer cancelRequest()
		done := make(chan struct{})
		go func() {
			defer close(done)
			fs.HTTP.Mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(requestCtx, http.MethodGet, "/health", nil))
		}()
		synctest.Wait()
		cancelRequest()
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Error("canceled endpoint should not wait for the active callback")
		}
		assert.Equal(t, int32(1), calls.Load())
		cancel()
		releaseCheck()
		appCtx.wg.Wait()
	})
}

func TestHealth_IntervalRestartsAfterCompletion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health-delay", Features{Health: Health(WithHealthInterval(2))})
		var calls atomic.Int32
		a.SetHealthCheck(func() bool { calls.Add(1); time.Sleep(3 * time.Second); return true })
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- a.run(ctx) }()
		synctest.Wait()
		time.Sleep(2 * time.Second)
		synctest.Wait()
		assert.Equal(t, int32(1), calls.Load())
		time.Sleep(3 * time.Second)
		synctest.Wait()
		time.Sleep(2*time.Second - time.Nanosecond)
		synctest.Wait()
		assert.Equal(t, int32(1), calls.Load(), "the next interval starts after completion")
		time.Sleep(time.Nanosecond)
		synctest.Wait()
		assert.Equal(t, int32(2), calls.Load())
		cancel()
		time.Sleep(3 * time.Second)
		synctest.Wait()
		require.NoError(t, <-done)
	})
}

func TestHealth_ReusedConfigurationHasIndependentState(t *testing.T) {
	health := Health(WithHealthPath("/health"))
	a := New("first", Features{HTTP: HTTP(), Health: health})
	b := New("second", Features{HTTP: HTTP(), Health: health})
	a.SetHealthCheck(func() bool { return false })
	b.SetHealthCheck(func() bool { return true })
	startFeatures(t, a.Features())
	startFeatures(t, b.Features())
	first, second := httptest.NewRecorder(), httptest.NewRecorder()
	a.Features().HTTP.Mux.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/health", nil))
	b.Features().HTTP.Mux.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusInternalServerError, first.Code)
	assert.Equal(t, http.StatusOK, second.Code)
	assert.False(t, a.Features().Health.healthy())
	assert.True(t, b.Features().Health.healthy())
}

func TestHealth_PanickingCheckIsUnhealthyAndCanRecover(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health-panic", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"), WithHealthInterval(1))})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		appCtx := NewAppContext(ctx, zap.NewNop())
		fs := a.Features()
		require.NoError(t, fs.HTTP.start(appCtx, fs))
		require.NoError(t, fs.Health.start(appCtx, fs))
		a.SetHealthCheck(func() bool { panic("check failed") })
		require.NoError(t, fs.Health.serve(appCtx, fs))
		time.Sleep(time.Second)
		synctest.Wait()
		assert.False(t, fs.Health.healthy())
		response := httptest.NewRecorder()
		fs.HTTP.Mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.JSONEq(t, `{"status":"error"}`, response.Body.String())
		a.SetHealthCheck(func() bool { return true })
		time.Sleep(time.Second)
		synctest.Wait()
		assert.True(t, fs.Health.healthy(), "a panic must release the evaluation gate")
		cancel()
		appCtx.wg.Wait()
	})
}

func TestHealth_ShutdownDrainsEndpointCheckAndCancelsQueuedChecks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := New("health-endpoint-shutdown", Features{HTTP: HTTP(), Health: Health(WithHealthPath("/health"))})
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		appCtx := NewAppContext(ctx, zap.NewNop())
		fs := a.Features()
		require.NoError(t, fs.HTTP.start(appCtx, fs))
		require.NoError(t, fs.Health.start(appCtx, fs))
		require.NoError(t, fs.Health.serve(appCtx, fs))
		release := make(chan struct{})
		var releaseOnce sync.Once
		releaseCheck := func() { releaseOnce.Do(func() { close(release) }) }
		t.Cleanup(releaseCheck)
		var calls atomic.Int32
		a.SetHealthCheck(func() bool { calls.Add(1); <-release; return false })
		activeDone := make(chan struct{})
		go func() {
			defer close(activeDone)
			fs.HTTP.Mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
		}()
		synctest.Wait()
		require.Equal(t, int32(1), calls.Load())
		queuedDone := make(chan struct{})
		go func() {
			defer close(queuedDone)
			fs.HTTP.Mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
		}()
		synctest.Wait()
		cancel()
		drained := make(chan struct{})
		go func() { appCtx.wg.Wait(); close(drained) }()
		synctest.Wait()
		select {
		case <-queuedDone:
		default:
			t.Error("App cancellation must release queued endpoint checks")
		}
		select {
		case <-drained:
			t.Error("Health returned before the endpoint check finished")
		default:
		}
		releaseCheck()
		synctest.Wait()
		<-drained
		<-activeDone
		assert.Equal(t, int32(1), calls.Load())
	})
}
