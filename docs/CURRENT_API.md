# Current HTTP API

This document describes routes that are currently registered in `internal/server/router.go`.

## `GET /`

Basic root response indicating that the API is running.

## `GET /health`

Process liveness endpoint.

Healthy response:

```json
{
  "status": "healthy"
}
```

- Returns `200 OK` when the API process can serve the request.
- Does not depend on PostgreSQL or other runtime dependencies.

This endpoint is suitable for container and Kubernetes startup/liveness probes.

## `GET /ready`

Dependency-aware readiness endpoint.

Ready response:

```json
{
  "status": "ready",
  "checks": [
    {
      "name": "database"
    }
  ]
}
```

- Returns `200 OK` when all registered required checks pass.
- Returns `503 Service Unavailable` when a required dependency is unhealthy or no registry is configured.

This endpoint is suitable for container and Kubernetes readiness probes.

## `GET /version`

Returns build-version information. See the runtime package for defaults and linker-injected variables.

## `GET /status`

Returns runtime metadata, including:

- status
- version
- build time
- Go version
- uptime
- process start time
- Git commit

One response field currently serializes as `BuildTime` rather than snake_case because its JSON tag is `json:"BuildTime"`. Standardize this before treating the contract as stable.

## `GET /metrics`

Prometheus exposition endpoint using the default registry.

Metrics registered by the application include:

- `village_http_requests_total`
- `village_http_request_duration_seconds`
- `village_http_requests_in_flight`
- `village_panics_total`
- `village_errors_total`
- `village_db_queries_total`
- `village_db_query_duration_seconds`
- `village_build_info`
- PostgreSQL pool metrics under `village_db_pool_*`

Additional operational metrics cover external requests/retries, workers, rate limiting, and cache behavior.

## Planned Community API

The intended Community API is documented in `docs/openapi.yaml` and currently includes:

```text
POST   /api/v1/communities
GET    /api/v1/communities
GET    /api/v1/communities/{id}
PUT    /api/v1/communities/{id}
DELETE /api/v1/communities/{id}
```

Keep OpenAPI and integration tests aligned with any future route changes.
