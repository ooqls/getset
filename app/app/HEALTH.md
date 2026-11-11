# Health evaluation

The Health Feature owns evaluation and the latest result. HTTP/Gin expose that
behavior; they are not prerequisites for it. This is not a loopback reachability
or TLS probe.

```go
a := app.New("worker", app.Features{
    Health: app.Health(app.WithHealthInterval(30)),
})
a.SetHealthCheck(func() bool {
    return dependenciesReady()
})
```

- Status starts healthy. Startup does not invoke the callback.
- Periodic checks begin in the Serve phase, even with no routers. The first runs
  after the interval; each subsequent interval starts after that check finishes.
- Zero or omitted intervals mean 30 seconds. Negative or overflowing intervals
  return an error before startup completes. Intervals are whole seconds.
- Every Health GET also evaluates the check: HTTP 200 and `{"status":"ok"}` on
  success, HTTP 500 and `{"status":"error"}` on failure. Configure a Health path
  when using HTTP/Gin. A nil callback succeeds.
- A later success replaces a failure; failures are not sticky.
- A panicking callback is logged and records an unhealthy result. Periodic work
  continues, and an endpoint evaluation returns the normal HTTP 500 response.
- `App.IsHealthy()` reads the last result without running a check. It remains
  false when the App is not running; disabled Health adds no health restriction.
- Endpoint and periodic checks never overlap within one App. Each App owns its
  own status and synchronization, even when configuration values are reused.
- `App.SetHealthCheck` may safely replace the callback while running, including
  with nil. An active evaluation finishes with its captured callback; later
  evaluations use the replacement. Replacing a callback does not evaluate it.
- Other Feature fields, including direct assignment to `Health.Check`, are
  startup configuration and must not be changed concurrently with evaluation.
- Cancellation stops future periodic work and cancels queued evaluations. An
  already-running `func() bool` cannot be forcibly canceled: shutdown waits for
  it to return. Keep callbacks bounded and do not wait for another evaluation
  of the same Health Feature from inside a callback.

The fixed Feature startup order and existing endpoint response format are
unchanged. Periodic evaluation no longer exercises listeners or TLS; use
external monitoring if endpoint reachability is required.

```sh
go test -race ./app/app -run 'TestHealth_|TestApp_IsHealthy|TestTLS_ServesGinAndHTTPOverTLS'
```

Scheduling, serialization and shutdown tests use `testing/synctest`: no actual
interval waits, polling sleeps or listening sockets are needed for evaluation.
