# Production Hardening Audit

This document records the cleanup applied before Milestone 8.

## Route normalization

Route normalization is implemented once in `internal/httputil.RouteLabel` and reused by request logging and HTTP metrics. Middleware no longer mutates `http.Request.Pattern`.

## Pagination validation

The handler parses query parameters and rejects syntactically invalid or explicitly invalid HTTP values. The service owns the pagination policy (default and maximum). The PostgreSQL repository assumes validated pagination inputs and owns SQL `LIMIT/OFFSET`.

## Nil ownership

Community service methods reject nil domain objects. PostgreSQL Create/Update no longer duplicate the same nil guard.

## Error metrics

`village_errors_total` is incremented by the HTTP metrics middleware for 4xx (`client`) and 5xx (`server`) responses.

`village_panics_total` is incremented by recovery middleware when a panic is recovered.

## Request timeout

`RequestTimeout` is wired into the production router using `Config.RequestTimeout`. The intended relationship is:

`DB_QUERY_TIMEOUT < REQUEST_TIMEOUT < WRITE_TIMEOUT`.

## Build metadata

Docker builds inject version, git commit, build timestamp, and environment using linker flags. Local builds retain safe defaults.

## Corporate CA

`zscaler.crt` is no longer required for Docker builds. When needed, provide it as a BuildKit secret. It is not copied into the runtime image.
