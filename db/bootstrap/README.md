# SQL bootstrap

`bootstrap.Run` prepares a database from a plan using a SQLX or PGX adapter:

```go
err := bootstrap.Run(ctx, bootstrap.Plan{
    Files:   []string{"init.sql"},
    Dirs:    []string{"./sql"},
    Tables:  []string{"CREATE TABLE IF NOT EXISTS users (id integer)"},
    Indexes: []string{"CREATE INDEX IF NOT EXISTS users_id ON users (id)"},
}, bootstrap.SQLX("")) // Or bootstrap.PGX(passwordFile).
```

The module owns discovery, ordering, cancellation, and failure policy:

- Explicit files run first, followed by sorted `*.sql` matches within each directory, table statements, then index statements. Supplied order and duplicates are preserved.
- Each file is executed whole. SQL is not split on semicolons.
- An empty plan, including directories with no matches, does not select a connection or read a password file.
- Nonempty plans select one connection. A password file overrides the Registry password for files and statements alike, without resolving the original password. Its contents are used verbatim.
- Without an override, an already initialized connection is preserved; otherwise it is opened from the Registry.
- Discovery, file, connection, and execution failures return immediately. Errors identify the file or statement position and wrap the cause. Driver error details may contain driver-provided context.
- The caller's context controls execution. No transaction is added: completed SQL may remain after a later failure or cancellation.

The adapter borrows the connection managed by `db/conn`. Run never closes it or changes connection lifetime. This is bootstrap execution, not versioned migrations, rollback, or an exactly-once guarantee.

## Migrating callers

The old `sqlx.SeedSQLX`, `sqlx.SeedSQLXFile`, `pgx.SeedPGX`, and `pgx.SeedPGXFile` helpers have been removed. Use a plan and the corresponding adapter, and handle the returned error.

`eventsource/initv1.Init` now returns an error. The SQL Feature returns bootstrap errors through `App.Run` instead of logging file failures or panicking on statement failures.

`postgres.GetRegistryOptions(passwordOverride *string)` now returns `(Options, error)`. Pass `nil` to resolve the Registry password, or a nonnil pointer to bypass it. SQLX and PGX also expose error-returning `Get` functions; existing panic-based getters remain for other callers.

## Tests

```sh
go test -race ./db/bootstrap ./db/conn ./db/postgres ./eventsource/initv1
go test -race -tags=integration ./db/bootstrap
```

The tagged tests require Docker and exercise both real Postgres adapters, including password overrides, retained connections, partial progress after failure, and cancellation.
