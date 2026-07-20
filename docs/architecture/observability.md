# Observability

## Purpose

The Village API is built with observability as a first-class concern.

Every incoming request passes through middleware that provides:

- Request IDs
- Request logging
- Panic recovery

This foundation enables reliable debugging, incident response, and future integration with distributed tracing and metrics.

## Middleware Order

Recovery
↓

Request ID
↓

Logging
↓

Router

## Future Work

- Structured JSON logging
- OpenTelemetry tracing
- Grafana dashboards
- Distributed tracing

## Started Adding Prometheus entries in Sprint 2
Added custom metric for version 