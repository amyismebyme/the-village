# External Integration Architecture

External integrations are provider implementations behind shared,
provider-neutral contracts.

## Architecture

External provider
↓
provider client
↓
external integration DTO
↓
normalization
↓
external.Item
↓
domain/service
↓
repository

## Provider identity

Every externally sourced item is identified by:

- source
- external_id

The combination is the idempotency identity.

## Context

All outbound requests receive the caller's context.

External clients must not create `context.Background()` for normal
request processing.

## Errors

External errors are categorized as:

- unauthorized
- forbidden
- not found
- rate limited
- upstream
- timeout
- invalid payload
- invalid configuration

Errors must preserve `errors.Is` / `errors.As` semantics.

## Observability

Metrics use bounded labels only:

- source
- operation
- status
- error type

Structured logs may include external IDs, but must never expose:

- secrets
- authorization headers
- request payloads
- access tokens