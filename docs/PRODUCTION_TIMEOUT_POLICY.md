# Production Timeout Policy

The API uses separate timeout boundaries for the HTTP server, request context, PostgreSQL queries, idle connections, and graceful shutdown.

| Boundary | Default | Purpose |
|---|---:|---|
| Read timeout | 10s | Maximum time to receive an HTTP request |
| Request timeout | 35s | Maximum lifetime of an application request context |
| DB query timeout | 30s | Maximum duration of an individual PostgreSQL repository operation |
| Write timeout | 40s | Maximum time allowed for the server to write the response |
| Idle timeout | 60s | Maximum idle keep-alive connection duration |
| Shutdown timeout | 15s | Graceful shutdown deadline |

The intended relationship is:

```text
DB query timeout (30s)
        <
request timeout (35s)
        <
write timeout (40s)
```

This ensures a PostgreSQL query normally reaches its own deadline before the request context expires, while the HTTP server retains enough time to return the resulting error response.

## Configuration

```text
READ_TIMEOUT=10
REQUEST_TIMEOUT=35
WRITE_TIMEOUT=40
IDLE_TIMEOUT=60
SHUTDOWN_TIMEOUT=15
DB_QUERY_TIMEOUT=30
```

Values are expressed in seconds.

## Context propagation

The request context flows through:

```text
HTTP request
  ↓
handler
  ↓
service
  ↓
repository
  ↓
pgx
```

Production request paths must not replace the incoming context with `context.Background()`.

`context.Background()` is appropriate for application startup/shutdown and test harness setup, but not for normal request processing.
