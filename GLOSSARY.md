# getset

An application framework that gives a Go service its infrastructure (databases, caches, keys, servers) as switchable pieces, started and stopped together.

## Language

**App**:
A single running service built on getset: a name, a set of Features, and the user's lifecycle hooks.
_Avoid_: server, application instance

**Feature**:
One switchable piece of infrastructure an App can turn on, such as SQL, Gin, TLS or Cache, owning its own configuration and how it comes up.
_Avoid_: plugin, module, component

**Server Feature**:
A Feature that accepts network traffic (Gin, HTTP, gRPC, Logging API) and so also takes part in the Serve phase.

**Startup phase**:
The part of an App's life where enabled Features come up in a fixed order, before the user's startup hook runs.

**Serve phase**:
The part of an App's life, after the user's startup hook, where Server Features begin accepting traffic.

**App Context**:
What an App hands to the user's hooks: cancellation, logging, and the shared results Features produced (cache factory, token issuer configurations).
_Avoid_: request context

**Registry**:
The description of where each backing system (Postgres, Redis, Valkey, Elasticsearch) lives and how to authenticate to it.
_Avoid_: service discovery, config
