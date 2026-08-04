# Testing Strategy

## Testing pyramid

The repository currently uses two primary levels:

```text
Many fast unit tests
        ↓
Fewer PostgreSQL/HTTP integration tests
```

End-to-end browser tests should be added only after the frontend exists.

## Unit tests

Unit tests live beside the package under test and use the `_test.go` suffix.

Examples:

- configuration validation
- handler response behavior
- middleware request IDs, logging, recovery, and response recording
- health registry behavior
- metrics registration
- Community model validation
- Community service behavior using a fake repository

Run from `apps/api`:

```bash
go test ./...
```

Run one package:

```bash
go test ./internal/service -v
```

Run one test:

```bash
go test ./internal/service -run TestCreateCommunity -v
```

## Integration tests

Integration tests use the build tag:

```go
//go:build integration
```

They live under `apps/api/internal/integration` and use a real PostgreSQL container.

Current integration coverage:

- database connection
- expected migration tables
- health endpoint through the real router and database checker

The Community repository CRUD test exists but is skipped because repository methods are not complete.

Run manually from `apps/api` after starting the integration database and applying migrations:

```bash
go test -tags=integration ./internal/integration/... -v -count=1
```

Run from the repository root with helpers:

```powershell
.\scripts\integration-test.ps1
```

```bash
./scripts/integration-test.sh
```

## Test isolation

Integration tests should start from known database state. `DeleteAll` helpers currently truncate tables and reset identity sequences.

Be careful with foreign keys as relationships are introduced. `TRUNCATE` may need multiple tables or `CASCADE`, and cleanup order should be explicit.

Recommended future approach:

- one clean database per test run
- migrations applied once
- each test cleans only data it creates
- avoid parallel tests that mutate shared tables until isolation is designed

## Current test gaps

- Community repository CRUD is skipped.
- Resource repository has no real integration coverage.
- Community service tests do not currently test all error propagation paths.
- No router tests for method handling or unknown paths.
- No tests for configuration parse failures because invalid values silently default.
- No frontend tests.
- No CI workflow in the reviewed ZIP.

## Required Community vertical-slice tests

### Model

- missing name
- name shorter than 3 characters
- name longer than 100 characters
- missing slug
- invalid/lowercase slug rules
- description longer than 2,000 characters

### Service

- trims and normalizes input
- rejects invalid entity
- detects duplicate slug
- distinguishes “not found” from repository outage
- wraps Create/Get/List/Update/Delete failures
- handles nil Community input without panic
- prevents duplicate slug on update, excluding the current record

### Repository integration

- create returns generated ID and timestamps
- find by ID
- find by slug
- list ordering
- update and updated timestamp
- delete
- not-found translation
- unique constraint translation

### API integration

- create, read, list, update, delete flow
- malformed JSON
- validation errors
- duplicate slug conflict
- invalid ID
- not found
- method not allowed
- content type

## Race and static checks

Before merging:

```bash
go test ./...
go vet ./...
go test -race ./...
golangci-lint run
```

The current `go 1.26.5` directive may prevent these commands on environments that cannot obtain that toolchain. Pin to an intentionally supported Go version and document it.
