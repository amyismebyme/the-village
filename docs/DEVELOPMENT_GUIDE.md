# Development Guide

## Prerequisites

- Go version compatible with `apps/api/go.mod`
- Docker Desktop or Docker Engine with Compose v2
- PostgreSQL migration CLI (`migrate`)
- Git
- Optional: `golangci-lint`

## Important module note

The repository currently contains both a root `go.mod` and `apps/api/go.mod`, and both declare the same module path. This is confusing and can cause commands to operate against different dependency graphs depending on the current directory.

Until this is cleaned up, run Go commands from:

```text
apps/api
```

Recommended cleanup: remove the root module and treat `apps/api` as the single Go module, or intentionally create a `go.work` workspace if multiple Go modules are planned.

## Running the API locally

From `apps/api`:

```bash
go run ./cmd/api
```

The default database host is `postgres`, which is appropriate inside Docker but not from a host shell. For host execution against local PostgreSQL, set:

```text
DB_HOST=localhost
DB_PORT=5432
DB_USER=village
DB_PASSWORD=village
DB_NAME=village
DB_SSLMODE=disable
```

On PowerShell:

```powershell
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
go run ./cmd/api
```

## Local Docker stack

From the repository root:

```bash
docker compose up --build
```

Current caveats:

- The Compose file references `./docker/postgres/init.sql`, which is not present in the reviewed ZIP.
- Prometheus references `./infra/docker/prometheus/prometheus.yml`, which is not present in the reviewed ZIP.
- The API image requires PostgreSQL to be ready, but `depends_on` alone does not wait for readiness.

These paths and startup dependencies must be corrected before the full stack is reliable.

## Common Go commands

From `apps/api`:

```bash
go fmt ./...
go test ./...
go vet ./...
go test -race ./...
```

Lint:

```bash
golangci-lint run
```

## Makefile note

The root Makefile invokes Go commands relative to the repository root, but the active API module is under `apps/api`. As written, targets may use the duplicate root module rather than the API module.

Recommended change:

```make
API_DIR := apps/api

test:
	cd $(API_DIR) && go test ./...
```

Apply the same pattern to `run`, `build`, `vet`, `fmt`, and `lint`.

## Configuration

Application variables:

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `ENVIRONMENT` | `development` | runtime environment |
| `LOG_LEVEL` | `info` | structured log level |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `READ_TIMEOUT` | `10` | seconds |
| `WRITE_TIMEOUT` | `10` | seconds |
| `IDLE_TIMEOUT` | `60` | seconds |
| `SHUTDOWN_TIMEOUT` | `15` | seconds |

Database variables:

| Variable | Default |
|---|---|
| `DB_HOST` | `postgres` |
| `DB_PORT` | `5432` |
| `DB_USER` | `village` |
| `DB_PASSWORD` | `village` |
| `DB_NAME` | `village` |
| `DB_SSLMODE` | `disable` |
| `DB_MAX_CONNS` | `10` |
| `DB_MIN_CONNS` | `1` |
| `DB_MAX_CONN_LIFETIME` | `3600` seconds |
| `DB_MAX_CONN_IDLE_TIME` | `300` seconds |
| `DB_HEALTH_CHECK_PERIOD` | `60` seconds |

Invalid numeric or duration environment values currently fall back silently to defaults. Production configuration should fail fast instead so mistakes do not go unnoticed.

## Code organization rules

- Put HTTP decoding/encoding in handlers.
- Put business rules in services.
- Put persistence contracts in repository interfaces.
- Put SQL and pgx behavior in `repository/postgres`.
- Keep models independent from HTTP and PostgreSQL.
- Pass `context.Context` through handler → service → repository.
- Wrap errors with operation context and preserve `errors.Is` behavior.

## Adding a domain

For each new domain, use this sequence:

1. Model and validation
2. Repository interface
3. Service interface and implementation
4. Unit tests with a fake repository
5. PostgreSQL implementation
6. Repository integration tests
7. HTTP DTOs and handlers
8. Router registration
9. API integration tests
10. OpenAPI documentation
11. UI client and screens



## Router Verification

Before completing a router change, verify the complete assembled router rather than only individual route registration functions.

Run:

```powershell
go test ./internal/server/... -v
go test ./...
go test -race ./...
go vet ./...
golangci-lint run