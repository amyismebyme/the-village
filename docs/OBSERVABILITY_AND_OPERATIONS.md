# Observability and Operations

## Logging

The API uses Go's structured `log/slog` package.

Request middleware currently logs:

- request ID
- HTTP method
- normalized route
- status
- duration in milliseconds
- trace ID/span ID when a valid span is present

Panic recovery logs the request ID, method, path, and recovered value.

Never log passwords, tokens, private messages, or sensitive application data.

## Request IDs

Request IDs are placed in context and used by logging and recovery middleware. Preserve the ID in handler/service logs and return it in error responses so user reports can be correlated with logs.

Trace IDs and span IDs provide the cross-service correlation path into OpenTelemetry/Tempo.

## Metrics

The application registers HTTP, panic, error, database-query, external integration, worker, cache, rate-limiter, build-info, and pool metrics.

Prometheus recording rules provide higher-level operational signals such as:

- HTTP 5xx ratio
- HTTP request p95
- external request rate and 5xx ratio
- worker failure rate
- rate-limiter pressure
- cache hit ratio

Do not label metrics with raw IDs, slugs, URLs, request IDs, or unbounded error text. For HTTP metrics, use normalized route names such as `/communities/{id}` rather than concrete IDs.

## Health, liveness, and readiness

Current behavior:

- `/health` is a process liveness endpoint and does not depend on PostgreSQL.
- `/ready` is a dependency-aware readiness endpoint and evaluates registered checks, including PostgreSQL.
- `/health` returns `200` when the process can serve requests.
- `/ready` returns `200` when required checks pass and `503` when a required dependency is unhealthy.

Operational probe semantics:

- startup/liveness → `/health`
- readiness → `/ready`

Using a dependency-dependent endpoint for liveness can cause unnecessary process restarts during dependency outages, so keep the two responsibilities separate.

## Graceful shutdown

The application waits for `SIGINT` or `SIGTERM`, then calls `http.Server.Shutdown` with a configurable timeout and closes the database pool.

Future background workers must also receive cancellation and shut down within the same lifecycle.

## Initial service-level indicators

Use the following as the initial operational measurement set:

- request success rate excluding expected 4xx responses
- p50/p95/p99 request latency by normalized route and method
- availability of read and write paths
- database pool saturation
- database query latency
- external integration error/retry rates
- worker failure rate

Do not finalize an SLO target before production usage data exists.

## Alerting principles

Alert on symptoms that affect users:

- sustained error-budget burn
- elevated latency
- readiness failure across enough replicas to reduce capacity
- connection-pool exhaustion
- migration or deployment failure
- external dependency failure when user impact is sustained

Avoid paging on every single error or transient dependency check.

## Observability backends

The local stack uses:

- Prometheus for metrics and rule evaluation
- Alertmanager for alert routing and retention
- Grafana for dashboards and investigation
- Loki for structured logs
- Tempo for distributed traces
- OpenTelemetry for trace instrumentation/export

See `docs/OBSERVABILITY.md` for the operator-facing workflow.

## Runbook: API fails to start

1. Read startup error and structured logs.
2. Validate environment variables.
3. Confirm PostgreSQL host and port from the API's network namespace.
4. Confirm credentials and database name.
5. Check migration version.
6. Check pool configuration values.
7. Confirm the configured Go binary/container image exists.

## Runbook: `/ready` returns 503

1. Inspect the response checks map/array.
2. If `database` is unhealthy, test PostgreSQL connectivity.
3. Check container status and database logs.
4. Check connection-pool metrics.
5. Verify credentials and DNS/service name.
6. Determine whether this is a dependency outage or an application regression.

## Runbook: integration tests fail

1. Run the helper with `-KeepRunning` on Windows.
2. Inspect `docker ps` and PostgreSQL logs.
3. Run `migrate version` against port 5433.
4. Query `schema_migrations`.
5. Run one failing test with `-run` and `-count=1`.
6. Tear down with volumes after diagnosis.
