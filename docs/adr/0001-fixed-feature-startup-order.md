# Feature startup order is a fixed list in the App

Features start in one hand-ordered list (`Features.ordered`), with a comment on each constraint (Gin before routes, Registry before backends, TLS before servers, Redis/Valkey before Cache, routers before Health). We considered letting each Feature declare what it requires and topologically sorting them, but with ~15 Features and a handful of constraints that is machinery for a problem we don't have; one list keeps all ordering knowledge in one place. Revisit if Features become pluggable from outside the `app` package.
