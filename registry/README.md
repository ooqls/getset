# Registry connection policy

`Server.Resolve(passwordOverride)` is the shared interpretation seam for SQLX,
PGX, Redis, Valkey and Elasticsearch. It returns a resolved connection snapshot;
adapters retain only driver-specific translation. Database selection remains
specific to each backing system.

## Authentication

- `auth.enabled: false` (including an omitted flag) skips password files and
  decryption, and does not forward the configured username or password.
- `auth.enabled: true` resolves the configured credentials. Password bytes are
  not trimmed.
- A nonnil explicit password override enables authentication and bypasses the
  Registry password, even if the override is empty. SQL bootstrap and explicit
  Redis/Valkey Feature credentials retain this opt-in behavior.
- An explicit username override on Redis/Valkey also enables authentication;
  any Registry password remains in effect unless explicitly overridden.
- Postgres requires a role name in its wire protocol. Enable authentication to
  supply that name, even for a server that accepts an empty password.

`Server.ResolvePassword` remains a lower-level password/decryption operation;
adapters should use `Resolve` rather than independently applying enablement.

## TLS

- Nil or disabled TLS means plaintext. Enabled TLS requires encryption with no
  plaintext fallback, and verifies the peer certificate and hostname by default.
- `insecure_skip_tls_verify: true` explicitly disables peer verification; it
  does not disable encryption. Prefer a configured CA outside local tests.
- Custom CA certificates extend system roots. Invalid CA bundles, invalid
  client certificates, and incomplete certificate/key pairs return errors.
- Each resolution owns its TLS configuration. It neither inherits nor mutates
  `http.DefaultTransport` or the App's separate server TLS configuration.
- Elasticsearch derives HTTP/HTTPS from TLS. An explicitly supplied protocol
  must agree with that choice; contradictory settings return an error.

## Postgres adapters and compatibility

Both adapters use `postgres.Options.Config` to apply the resolved credentials
and TLS settings. PGX has no plaintext fallback. SQLX still returns `*sqlx.DB`,
but uses pgx's `database/sql` adapter internally (`DriverName()` is now `pgx`).
Postgres errors are now pgx/pgconn errors rather than lib/pq errors. The existing
`postgres` driver registration remains available for direct `sql.Open` callers.

Ambient password/passfile and TLS settings do not override this configuration.
`PGSERVICE` is rejected instead of reading an ambient service file. Explicit
`postgres.Options` initialization still works, including its `Tls` field.

This intentionally changes older configurations that populated credentials
without enabling auth, relied on opportunistic PGX TLS, or assumed SQLX ignored
TLS. Update those configurations explicitly. Connection ownership, reuse and
replacement remain with `db/conn`; startup order is unchanged.

## Verification

```sh
go test -race ./registry ./db/postgres
go test -race ./db/elasticsearch -run TestRegistryPolicy
go test -race -tags=integration ./db/integrationTest/registry ./db/bootstrap
```

The tagged tests require Docker and verify actual Postgres and Redis-protocol
connections through both adapters, including no downgrade, credential overrides,
and effective encryption. Elasticsearch uses a local HTTP/TLS endpoint.
