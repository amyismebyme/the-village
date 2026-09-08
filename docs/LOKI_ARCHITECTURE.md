# Loki Logging Architecture

The Village API keeps application logging independent of Loki.

```text
slog
  ↓
JSON structured stdout
  ↓
Promtail / Grafana Alloy / other collector
  ↓
Loki
  ↓
Grafana
```

## Application boundary

Application packages use `log/slog`. They do not import Loki clients or Loki HTTP APIs.

The logger writes JSON to standard output when `LOG_FORMAT=json` (the project default). This keeps logs useful during local development, Docker execution, CI, and environments where Loki is unavailable.

## Failure behavior

Loki availability is not an application dependency. Collector buffering/retry behavior belongs to the deployment layer.

```text
Loki unavailable
    ↓
collector retries / buffers according to deployment policy
    ↓
API continues serving requests and workers
```

## Trace correlation

When an active OpenTelemetry span exists, structured logs should carry `trace_id` and `span_id` through the logging integration layer. These fields are correlation data and should not be promoted to Loki labels.

Avoid high-cardinality Loki labels such as `request_id`, `external_id`, full URLs, user-generated titles, or credentials.

## Deployment responsibility

Promtail, Grafana Alloy, or another supported collector reads the API's structured stdout and forwards events to Loki. The application remains unaware of Loki's transport and storage APIs.
