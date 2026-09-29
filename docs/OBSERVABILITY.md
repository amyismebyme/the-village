# Observability Guide

The Village API uses Prometheus, Grafana, Loki, Tempo, Alertmanager, and OpenTelemetry to provide metrics, logs, traces, dashboards, and alert routing.

## Stack

| Component | Purpose | Local endpoint |
|---|---|---|
| Prometheus | Scrapes API metrics, evaluates recording/alerting rules | `http://localhost:9090` |
| Alertmanager | Receives and retains alerts from Prometheus | `http://localhost:9093` |
| Grafana | Dashboards and investigation UI | `http://localhost:3000` |
| Loki | Structured log storage/query backend | `http://localhost:3100` |
| Tempo | Distributed trace storage/query backend | `http://localhost:3200` |
| OpenTelemetry | API trace instrumentation and OTLP export | API → Tempo |

The API keeps telemetry backends out of the serving path. A telemetry backend outage must not make the API unavailable.

## Health endpoints

The health contract is intentionally split:

- `GET /health` is liveness. It verifies that the API process can serve a request and does not require PostgreSQL.
- `GET /ready` is readiness. It evaluates required dependency checks, including PostgreSQL, and returns `503` when the instance should not receive traffic.
- `GET /metrics` exposes Prometheus metrics.

For orchestration:

- startup/liveness probes should use `/health`
- readiness probes should use `/ready`

## Metrics

Prometheus scrapes `/metrics` every 15 seconds in the local Compose configuration.

Key application metrics include:

- `village_http_requests_total`
- `village_http_request_duration_seconds`
- `village_http_requests_in_flight`
- `village_panics_total`
- `village_errors_total`
- `village_db_queries_total`
- `village_db_query_duration_seconds`
- `village_external_requests_total`
- `village_external_retry_exhausted_total`
- `village_worker_failures_total`
- `village_rate_limiter_waits_total`
- `village_cache_hits_total`
- `village_cache_misses_total`
- `village_build_info`
- `village_db_pool_*`

Metric labels must remain bounded. Do not introduce raw IDs, request IDs, full URLs, arbitrary error text, or user-generated values as labels.

Recording rules in `infra/docker/prometheus/recording_rules.yml` derive higher-level signals such as HTTP 5xx ratio, request latency, external failure rate, worker failure rate, rate-limiter pressure, and cache hit ratio.

## Logs

The API emits structured `log/slog` records.

Operationally useful correlation fields include:

- request ID
- trace ID
- span ID
- HTTP method
- normalized route
- status
- duration

Use request ID for API request correlation. Use trace ID/span ID to move from logs into distributed traces.

Never log passwords, bearer tokens, secrets, private messages, or other sensitive application data.

## Traces

OpenTelemetry traces follow request and service boundaries through the API.

Representative spans include:

- HTTP server request spans
- community service spans
- `reddit.ingest`
- `reddit.authenticate`
- `reddit.fetch_listing`
- `reddit.normalize`
- `reddit.deduplicate`
- `external_item.upsert_batch`
- bounded PostgreSQL operation spans

The local API exports OTLP HTTP traces to Tempo.

## Grafana investigation workflow

Grafana is the primary investigation surface:

1. Start with a symptom or alert in Prometheus.
2. Use Prometheus to inspect the metric, labels, alert state, and evaluation window.
3. Open Grafana Explore with Loki and filter by service/route/time window.
4. Use the trace ID from structured logs to find the corresponding trace in Tempo.
5. Check `/ready` and dependency health when the symptom suggests a downstream outage.
6. Confirm whether the issue is serving-path, dependency, worker, external integration, or telemetry-only.

## Alert flow

The operational alert path is:

`application behavior → Prometheus metric → recording rule → alerting rule → Prometheus firing state → Alertmanager`

The live alert-chain test exercises this path for API unavailability by stopping the API container, allowing Prometheus to observe `up == 0`, waiting for the configured alert window, and then confirming the alert is visible in both Prometheus and Alertmanager.

Deterministic rule tests remain separate from the live chain test:

`scripts/test-alerts.ps1`
`scripts/test-alerts.sh`

Live chain tests:

`scripts/test-alert-chain.ps1`
`scripts/test-alert-chain.sh`

## Failure interpretation

### API returns 503 on /ready

Check the response checks and PostgreSQL health first. A readiness failure is expected during a dependency outage and should prevent new traffic from being routed to that instance.

### API returns 200 on /health but /ready is 503

This means the process is alive but a required dependency check is failing. Treat this as a dependency/readiness problem, not an automatic process-crash condition.

### Prometheus cannot scrape the API

Check:

`docker compose ps`
`http://localhost:8080/metrics`
Prometheus target state at `http://localhost:9090/targets`

Then inspect API logs and container networking.

### Alerts are firing but Alertmanager is empty

Check:

`http://localhost:9090/api/v1/alerts`
`http://localhost:9093/-/ready`

Then inspect the Prometheus Alertmanager target configuration and Alertmanager logs.

### Logs exist but traces are missing

Check API OpenTelemetry configuration and Tempo health. A Tempo outage should not make the API unavailable; the serving path remains independent of trace storage.

## Verification

From the repository root:

Windows:

`.scriptserify-observability-stack.ps1`
`.scripts	est-alert-chain.ps1`

Shell:

`./scripts/verify-observability-stack.sh`
`./scripts/test-alert-chain.sh`

Milestone 9 verification remains the combined gate:

`.scriptserify-milestone9.ps1`
`./scripts/verify-milestone9.sh`

## Operational principle

Use metrics to detect the symptom, logs to explain the request context, traces to follow the distributed path, and dependency/readiness checks to isolate the failing component.
