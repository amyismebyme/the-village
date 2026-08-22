# Current HTTP API

This document describes routes that are currently registered in `internal/server/router.go`. Community CRUD routes are not yet available.

## `GET /`

Basic root response indicating that the API is running.

## `GET /health`

Dependency-aware health endpoint.

Healthy response:

```json
{
  "status": "healthy",
  "checks": {
    "database": "healthy"
  }
}
```

- Returns `200 OK` when all registered checks pass.
- Returns `503 Service Unavailable` when a dependency is unhealthy or no registry is configured.

This endpoint currently includes database health, so it is closer to a readiness/dependency-health endpoint than a pure liveness endpoint.

## `GET /ready`

Current response:

```json
{
  "status": "ready"
}
```

The current implementation always returns `200` and does not inspect dependencies. This overlaps semantically with `/health` and should be clarified before deployment.

Recommended semantics:

- `/health` or `/live`: process is alive and event loop responds.
- `/ready`: required dependencies are available and instance can receive traffic.

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

Some metrics are defined and registered but are not yet updated in the request/repository code.

## Planned Community API

The intended first product API is:

```text
POST   /communities
GET    /communities
GET    /communities/{id}
PUT    /communities/{id}
DELETE /communities/{id}
```

Do not implement UI against these routes until the schema and request/response contract are aligned and documented in OpenAPI.


## Community list pagination

`GET /api/v1/communities` supports offset pagination:

```text
?limit=20&offset=0
```

Rules:
- `limit` defaults to `20`.
- `limit` must be greater than `0` and cannot exceed `100`.
- `offset` defaults to `0` and cannot be negative.
- invalid or excessive pagination parameters return `400`.
- results are ordered deterministically by `name, id`.
- an empty page returns `"communities": []` with pagination metadata.

Response shape:

```json
{
  "communities": [],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 123
  }
}
```
