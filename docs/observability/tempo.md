# Tempo development configuration

The Village API exports traces over OTLP/HTTP to Tempo. The API does not depend on Tempo for startup, readiness, or serving traffic.

Tempo runs in monolithic mode with local filesystem storage:

- OTLP gRPC: `4317`
- OTLP HTTP: `4318`
- Query HTTP: `3200`
- WAL: `/data/tempo/wal`
- Trace blocks: `/data/tempo/blocks`
- Retention: `24h`
- Compacted block retention: `1h`
- Persistent volume: `tempo-data`

The development storage target is **2 GiB**. Docker named-volume size enforcement is intentionally not added here because it is platform/driver dependent; disk usage should be monitored and the volume cleared/recreated during local development when needed.

Tempo's local backend is appropriate for single-node development, while production should use object storage. citeturn892261search1turn892261search0

## Fail-open contract

When Tempo is unreachable:

1. the API continues serving traffic;
2. the BatchSpanProcessor keeps telemetry asynchronous;
3. export errors are logged as structured `slog` events;
4. application startup and request handling do not return Tempo errors.
