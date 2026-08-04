# Decisions and Conventions

This file records current recommended conventions. Convert important, long-lived decisions into individual ADRs under `docs/adr` as they are accepted.

## Architecture

- Use a modular monolith.
- Keep one deployable Go API until scale or ownership demands separation.
- Use handler → service → repository layering.
- Use dependency injection at the composition root.

## Domain models

- Domain models should not contain SQL-specific behavior.
- Prefer separate API DTOs rather than JSON tags on domain models.
- Domain validation covers intrinsic rules; services cover cross-entity and persistence-dependent rules.

## IDs and URLs

- Use PostgreSQL `BIGSERIAL`/`int64` internally for the current phase.
- Use a unique slug for human-readable Community URLs.
- Reconsider externally exposed opaque IDs before public launch if enumeration/privacy concerns arise.

## Errors

- Repositories translate database-specific errors into repository errors.
- Services wrap errors with operation context.
- Handlers map known errors to a stable JSON error envelope.
- Never match errors by string.

## Logging

- Use `slog` consistently.
- Include request ID in request-scoped logs.
- Do not log sensitive content.
- Prefer bounded, searchable attributes.

## Metrics

- Use normalized route labels.
- Avoid high-cardinality IDs and raw paths.
- Measure user-impacting success and latency.

## Database

- Migrations are the only source of schema truth.
- Do not duplicate schema creation in Docker init scripts.
- Every schema change receives an up and down migration.
- Repository update statements set `updated_at` explicitly.

## Testing

- Unit-test business rules with fakes.
- Integration-test SQL and full HTTP flows against PostgreSQL.
- Keep integration workflows separate from unit-test workflows.
- Do not leave important integration tests permanently skipped.

## API

- Use consistent snake_case JSON.
- Document contracts in OpenAPI before frontend coupling grows.
- Add pagination before collection size becomes unbounded.
- Keep health/operational endpoints separate from product APIs.

## Security and privacy

- Do not commit private certificates, secrets, or local credentials.
- Protect all write operations before internet exposure.
- Treat community participation and mental-health-adjacent data as sensitive.
- Add rate limiting, abuse controls, and auditability before public social interaction features.
