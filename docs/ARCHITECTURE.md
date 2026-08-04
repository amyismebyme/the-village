# Architecture

## Architectural style

The API is evolving toward a layered monolith. This is appropriate for the current product stage: it keeps deployment simple while preserving clear boundaries that can be tested independently.

```text
Client / UI
    ↓ HTTP
Router and middleware
    ↓
Handlers
    ↓
Services
    ↓
Repository interfaces
    ↓
PostgreSQL implementations
    ↓
PostgreSQL
```

## Startup flow

The executable entry point is `apps/api/cmd/api/main.go`.

```text
main
  └─ app.Run
      ├─ load and validate configuration
      ├─ create structured logger
      ├─ open and ping PostgreSQL pool
      ├─ create health registry
      ├─ register database health checker
      ├─ register Prometheus collectors
      ├─ build HTTP server and router
      ├─ start ListenAndServe in a goroutine
      ├─ wait for signal or server failure
      └─ gracefully shut down and close database pool
```

## Package responsibilities

### `internal/app`

Composition root and lifecycle management. It should be the only package that knows how all major dependencies are assembled.

Current note: `dependencies.go` defines a `Dependencies` struct but the application does not use it. Keep either direct construction in `Run` or adopt the struct deliberately; avoid maintaining both patterns.

### `internal/config`

Loads environment variables into a typed configuration object and validates application-level settings.

The database configuration is currently validated separately in `app.Run`. Consolidating this into `config.Validate` would provide a single startup validation path.

### `internal/database`

Owns PostgreSQL configuration, pool creation, pinging, health checks, and pool statistics.

### `internal/health`

Provides a generic checker interface and registry. Dependencies implement `Name()` and `Check(context.Context)` and are aggregated into one health response.

### `internal/server`

Constructs the HTTP server and router. The router currently receives a logger and health registry.

As domain handlers are added, prefer a dependency struct or explicit handler bundle rather than continually adding constructor parameters.

### `internal/middleware`

Current chain:

```text
Recovery(RequestID(Logging(mux)))
```

Execution order for an incoming request:

1. Recovery establishes panic protection.
2. RequestID attaches a request identifier.
3. Logging records method, path, status, and duration.
4. Router dispatches to the endpoint.

### `internal/handlers`

Contains transport concerns: request decoding, response encoding, HTTP status codes, and handler-specific logging. Business rules should not live here.

Current handlers are function-based. The health handler is constructor-based because it requires dependencies.

### `internal/model`

Contains domain entities. Community validation is currently implemented as a method on `model.Community` using reusable validation helpers.

There is a consistency issue: `Resource` has JSON tags while `Community` does not. Decide whether models are pure domain objects or API representations. The recommended approach is pure domain models plus handler DTOs.

### `internal/service`

Contains business logic and coordinates repository calls. The Community service normalizes input, validates the entity, detects duplicate slugs, and wraps repository errors.

### `internal/repository`

Defines persistence interfaces and shared repository errors.

### `internal/repository/postgres`

Contains concrete PostgreSQL implementations and a shared embedded `Repository` that exposes the pool.

The Community and Resource implementations are currently incomplete.

### `internal/metrics`

Defines Prometheus collectors for HTTP requests, latency, in-flight requests, panics, application errors, database operations, build metadata, and connection-pool state.

Some metrics are defined but not yet incremented by middleware or repository code. Documentation should distinguish “registered” from “actively emitted.”

## Request flow

For a current health request:

```text
GET /health
  → Recovery middleware
  → Request ID middleware
  → Logging middleware
  → health handler
  → health registry
  → database health checker
  → database.Health / Ping
  → JSON response
```

For the planned Community API:

```text
POST /communities
  → middleware
  → Community handler
  → Community service
  → Community.Validate
  → Community repository interface
  → PostgreSQL Community repository
  → communities table
```

## Dependency rules

Recommended dependency direction:

```text
handlers  → service interfaces / domain types
services  → repository interfaces / domain types
postgres  → repository interfaces / domain types / pgx
models    → validation helpers only
```

Avoid dependencies in the reverse direction. In particular:

- Models must not import handlers, services, or database packages.
- Services must not import PostgreSQL implementations.
- Repository interfaces must not import handlers.
- HTTP status codes must not leak into service or repository packages.

## Error strategy

Current repository errors are generic sentinel errors. Before CRUD handlers are added, standardize error semantics:

- `repository.ErrNotFound`
- `repository.ErrAlreadyExists`
- `repository.ErrInvalidID`

PostgreSQL implementations should translate `pgx.ErrNoRows` and unique-constraint violations into repository errors. Services may wrap those errors while preserving `errors.Is`. Handlers should map them to HTTP responses.

Recommended HTTP mapping:

| Domain/repository error | HTTP status |
|---|---:|
| Validation failure | 400 |
| Not found | 404 |
| Already exists/conflict | 409 |
| Unexpected dependency failure | 500 |

## Scaling approach

Keep the application as a modular monolith until product usage proves a need to split services. PostgreSQL, one API deployment, and one frontend deployment are sufficient for the foreseeable product stage.

Scale first through:

- stateless API instances
- connection-pool tuning
- caching only where measured
- pagination
- database indexes
- background jobs for expensive or asynchronous work

Do not introduce microservices solely for portfolio optics; reliability and operability are easier to demonstrate with a well-built monolith.
