# Development Guide

## Prerequisites

- Go version compatible with `apps/api/go.mod`
- Docker Desktop or Docker Engine with Compose v2
- PostgreSQL migration CLI (`migrate`)
- Git
- Optional: `golangci-lint`

## Important module note

The repository contains a root `go.mod` and `apps/api/go.mod`, and the API commands intentionally run from `apps/api`.

Use the root Makefile for common project commands or run Go commands directly from:

```text
apps/api
```

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

The local stack includes PostgreSQL, migrations, the API, Prometheus, Grafana, Loki, Tempo, and Alertmanager.

Compose startup dependencies use health/completion conditions where a service must be operational before its dependent starts. The API remains independent of telemetry backend availability.

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

## Makefile commands

The root Makefile runs Go commands from `apps/api` so they use the API module consistently.

Examples:

```bash
make test
make test-integration
make test-race
make vet
make lint
```

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
```

## Observability Verification

From the repository root:

Windows:

```powershell
.scriptserify-observability-stack.ps1
.scripts	est-alert-chain.ps1
```

Shell:

```bash
./scripts/verify-observability-stack.sh
./scripts/test-alert-chain.sh
```

The canonical operator guide is `docs/OBSERVABILITY.md`.

## Milestone verification

Run the Milestone 9 gate with:

Windows:

```powershell
.scriptserify-milestone9.ps1
```

Shell:

```bash
./scripts/verify-milestone9.sh
```

Run the opt-in live Reddit smoke test only when Reddit credentials are intentionally available in the environment:

```powershell
.scriptseddit-live-smoke.ps1 -Subreddit toronto
```
