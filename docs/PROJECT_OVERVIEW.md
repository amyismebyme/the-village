# The Village — Project Overview

## Mission

The Village is a community platform intended to help men form meaningful, healthy, real-world connections. The product goal is paired with an engineering goal: build the system as a production-minded Site Reliability Engineering portfolio project rather than as a disposable CRUD demo.

The repository therefore emphasizes reliability foundations early: structured logging, health checks, graceful shutdown, metrics, database pooling, migrations, Docker-based integration testing, and clear layering.

## Current state

The project currently contains a working Go API foundation and a partially implemented Community domain. The frontend directory exists but does not yet contain an application.

### Working platform capabilities

- Environment-based configuration
- Configuration validation
- Structured `slog` logging
- HTTP server with read, write, idle, and shutdown timeouts
- Graceful shutdown on `SIGINT` and `SIGTERM`
- Request ID middleware
- Request logging middleware
- Panic recovery middleware
- Liveness/readiness/status/version endpoints
- PostgreSQL connection pooling through `pgxpool`
- Database health checks
- Prometheus application and database-pool metrics
- SQL migrations and seed data
- Docker Compose definitions for local and integration databases
- Unit tests across configuration, middleware, handlers, runtime, metrics, health, validation, and database helpers
- Integration tests for database connectivity, migrations, and health

### In-progress product capabilities

- Community domain model
- Community validation
- Community service interface and initial unit tests
- PostgreSQL Community repository skeleton

### Not yet implemented

- Community HTTP CRUD endpoints
- Complete Community PostgreSQL CRUD implementation
- Resource repository CRUD implementation
- Authentication and authorization
- User/member domain
- Frontend/UI application
- OpenAPI specification
- CI workflows in the checked ZIP
- Kubernetes, Terraform, Grafana dashboards, Loki, and OpenTelemetry

## Repository layout

```text
.
├── apps/
│   ├── api/                    Go API module
│   │   ├── cmd/api/            executable entry point
│   │   └── internal/           application packages
│   └── frontend/               frontend placeholder
├── docs/                       project documentation
├── migrations/                 database schema and seed migrations
├── scripts/                    integration-test helpers
├── testdata/                   integration Docker Compose environment
├── Dockerfile                  API container image
├── docker-compose.yml          local multi-service stack
├── Makefile                    developer commands
└── golangci.yml                lint configuration
```

## Product direction

The first usable product slice should be the Community domain:

1. Create and manage communities.
2. Browse communities.
3. View a community detail page.
4. Associate useful resources and future events with a community.
5. Add membership, moderation, and discovery later.

The immediate engineering priority is to finish one vertical slice end to end before adding broad infrastructure. That means aligning the database schema, completing repository methods, exposing CRUD APIs, documenting the API, and then implementing the first UI screens.

## Reliability direction

The SRE learning track should continue alongside product delivery:

- Define service-level indicators for request success and latency.
- Establish service-level objectives once real traffic exists.
- Keep health and readiness semantics distinct.
- Add tracing only after stable endpoint boundaries exist.
- Add dashboards and alerts based on user-impacting signals.
- Practice failure injection through integration tests and controlled local experiments.

## Recommended next milestone sequence

1. **Architecture alignment:** resolve schema/model/repository mismatches and module layout.
2. **Community vertical slice:** repository, service, handlers, routes, validation, integration tests.
3. **API contract:** OpenAPI document and stable error envelope.
4. **Frontend foundation:** Next.js/TypeScript app, API client, Community list/detail/create flows.
5. **Operational maturity:** CI, dashboards, SLOs, alert rules, deployment environments.
