# Observability and Operations

## Logging

The API uses Go's structured `log/slog` package.

Request middleware currently logs:

- request ID
- HTTP method
- path
- status
- duration in milliseconds

Panic recovery logs the request ID, method, path, and recovered value.

Recommended additions:

- remote/client address only if operationally needed and privacy-reviewed
- user/member ID after authentication
- route template rather than raw path for cardinality control
- response size
- trace ID after tracing is introduced

Never log passwords, tokens, private messages, or sensitive mental-health information.

## Request IDs

Request IDs are placed in context and used by logging and recovery middleware. Preserve the ID in handler/service logs and return it in error responses so user reports can be correlated with logs.

## Metrics

The application registers HTTP, panic, error, database-query, build-info, and pool metrics.

### Current concern

Several collectors are registered but not actively updated in the reviewed code:

- HTTP counters/histograms/in-flight gauge
- panic counter
- application error counter
- database query counters/histograms

Registering a collector does not produce meaningful telemetry unless code increments or observes it.

The next cleanup should either wire these metrics into middleware/repositories or remove them until they are used.

### Cardinality guidance

Do not label metrics with raw IDs, slugs, URLs, request IDs, or unbounded error text. For HTTP metrics, use normalized route names such as `/communities/{id}` rather than `/communities/123`.

## Health, liveness, and readiness

Current behavior:

- `/health` checks PostgreSQL and returns 503 on dependency failure.
- `/ready` always reports ready.
- Docker `HEALTHCHECK` calls `/health`.

Recommended behavior:

- `/live`: only verifies that the process can respond.
- `/ready`: verifies required dependencies and migration compatibility.
- Docker/Kubernetes liveness probes use `/live`.
- Kubernetes readiness probes use `/ready`.

Using a database-dependent endpoint for liveness can cause restart loops during a database outage.

## Graceful shutdown

The application waits for `SIGINT` or `SIGTERM`, then calls `http.Server.Shutdown` with a configurable timeout and closes the database pool.

Future background workers must also receive cancellation and shut down within the same lifecycle.

## Initial service-level indicators

Once Community APIs exist, begin with:

- request success rate excluding expected 4xx responses
- p50/p95/p99 request latency by normalized route and method
- availability of read and write paths
- database pool saturation
- database query latency

## Initial SLO example

Do not finalize an SLO before usage data exists. A reasonable starting experiment could be:

- 99.5% successful Community API requests over 30 days
- 95% of read requests under 300 ms
- 95% of write requests under 500 ms

Review and adjust based on user impact and operational cost.

## Alerting principles

Alert on symptoms that affect users:

- sustained error-budget burn
- elevated latency
- readiness failure across enough replicas to reduce capacity
- connection-pool exhaustion
- migration or deployment failure

Avoid paging on every single error or transient dependency check.

## Runbook: API fails to start

1. Read startup error and structured logs.
2. Validate environment variables.
3. Confirm PostgreSQL host and port from the API's network namespace.
4. Confirm credentials and database name.
5. Check migration version.
6. Check pool configuration values.
7. Confirm the configured Go binary/container image exists.

## Runbook: `/health` returns 503

1. Inspect the response checks map.
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
