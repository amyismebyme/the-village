# Architecture Review and Cleanup Plan

## Executive summary

The repository has a strong operational foundation for its age, but it now contains drift from several iterations of code generation. Before building substantial UI or adding more domains, complete a focused alignment pass. The highest-risk issue is not style; it is that the Community model, repository SQL, and database migration describe different entities.

## Priority 0 — blockers before Community CRUD/UI

### 1. Align Community model and database schema

**Current conflict**

- Model: `ID`, `Name`, `Slug`, `Category`, `Description`, `ExternalSource`, timestamps
- Repository SQL: expects `slug`, but not `category`
- Database table: has `url`, but no `slug` or `category`

**Impact**

Community queries will fail at runtime, and the UI cannot rely on a stable contract.

**Action**

Choose the product contract, add a migration, update seed data, then align model, SQL, tests, and API DTOs.

### 2. Complete Community repository methods

`Create`, `FindByID`, `Update`, and `Delete` are placeholders. The CRUD integration test is skipped.

**Action**

Implement all methods, translate database errors, add unique slug constraint, and unskip the test.

### 3. Fix duplicate Go module layout

Both root and `apps/api` contain a `go.mod` with the same module path but different dependencies.

**Impact**

Commands behave differently depending on working directory; Makefile and CI can test the wrong module.

**Action**

Keep `apps/api/go.mod` as the API module and remove the root module, unless a deliberate multi-module workspace is required. A frontend does not need a Go module.

### 4. Pin a real supported Go version

Both module files declare `go 1.26.5`. This is an unusual patch-level toolchain directive and may force automatic toolchain downloads or fail in CI/container environments.

**Action**

Choose the intentionally supported Go release, use the appropriate `go` and optional `toolchain` directives, and align Docker and CI.

## Priority 1 — complete before first production-like deployment

### 5. Clarify health semantics

`/health` checks the database, `/ready` always succeeds, and Docker uses `/health` as liveness.

**Risk**

A PostgreSQL outage could cause repeated API restarts.

**Action**

Introduce `/live` for process health and make `/ready` dependency-aware.

### 6. Remove laptop-specific certificate from production Docker build

The Dockerfile copies `zscaler.crt` into the image build.

**Risk**

- environment-specific build
- certificate material committed to repository
- portability and trust concerns

**Action**

Remove it from the standard Dockerfile. Handle corporate proxy certificates through local build configuration, BuildKit secrets, or a documented optional development Dockerfile. Review whether the certificate should remain in version control.

### 7. Repair Docker Compose paths

The local Compose file references files not present in the reviewed ZIP:

- `docker/postgres/init.sql`
- `infra/docker/prometheus/prometheus.yml`

**Action**

Create the files or remove the mounts. Prefer migrations over a separate init schema to avoid two schema sources of truth.

### 8. Wire or remove unused metrics

Many metrics are declared but not observed/incremented.

**Action**

Add an HTTP metrics middleware and repository instrumentation using bounded labels. Increment the panic metric in recovery middleware.

### 9. Standardize model versus DTO boundaries

`Resource` has JSON tags; `Community` does not.

**Action**

Keep domain models transport-neutral and create request/response DTOs in handlers or an API package. This prevents persistence/domain changes from accidentally changing the public API.

### 10. Standardize errors

`ErrNotFound` says “resource not found,” even though it is shared by multiple repositories. Service duplicate detection currently needs reliable not-found semantics.

**Action**

Use neutral messages and preserve sentinel errors through wrapping. Translate pgx errors at the repository boundary.

## Priority 2 — maintainability improvements

### 11. Resolve unused abstractions and empty files

Examples:

- `internal/app/dependencies.go` is unused.
- `internal/metrics/database.go` is empty.
- multiple `.gitkeep` files remain in populated areas.
- `internal/repository/postgres.go` and `repository/repository.go` should be reviewed for duplication or unclear purpose.

Delete unused scaffolding or explain it in code comments/issues.

### 12. Improve configuration parsing

Invalid integer/duration values silently fall back to defaults.

**Risk**

A production typo can go unnoticed.

**Action**

Return parsing errors from `Load` or use a strict loader in non-development environments.

### 13. Make handler logging consistent

Some handlers use the standard `log` package while others use injected `slog`.

**Action**

Use structured logger injection or a shared JSON response helper. Avoid calling `http.Error` after headers/body may already have been written.

### 14. Normalize API JSON field names

`StatusResponse.BuildTime` is tagged as `BuildTime`, while other fields use snake_case.

Fix before publishing the contract.

### 15. Improve service robustness

Community service improvements:

- reject nil Community inputs instead of panicking
- trim all relevant fields on update, not only slug
- distinguish `ErrNotFound` from infrastructure errors during duplicate checks
- check duplicate slug on update while excluding the current record
- decide who owns timestamp assignment
- avoid naming a field `repository` when `repo` is clearer and avoids package-name shadowing

### 16. Improve test fake behavior

The fake Community repository stores incoming pointers directly. A caller can mutate stored state outside the repository, unlike a real database.

Use copies when storing/returning to make unit tests more realistic.

### 17. Fix script robustness

The PowerShell script is syntactically balanced in this ZIP, but native command failures and `Push-Location` cleanup can be made safer with nested `try/finally` blocks. The Bash script depends on a locally installed `migrate` binary while CI may use a containerized migration tool.

Standardize one source of truth and test scripts in CI.

### 18. Add CI workflows

No workflow files are present in `.github/workflows` in the reviewed ZIP.

Add separate workflows for:

- formatting/vet/unit tests/lint
- PostgreSQL integration tests
- container build
- dependency/security scans later

## UI-readiness checklist

Before frontend implementation begins, stabilize:

- Community schema
- Community endpoint paths and methods
- request/response DTOs
- consistent JSON error envelope
- pagination contract
- slug behavior
- CORS policy
- API base URL configuration
- OpenAPI specification

Then build UI in this order:

1. App shell and navigation
2. Community list
3. Community detail
4. Community create/edit form
5. Empty/loading/error states
6. Accessible validation feedback
7. Authentication-aware actions later

## Recommended cleanup sprint

A short cleanup sprint should produce:

- one Go module
- one authoritative Community schema
- complete Community repository
- green `fmt`, `test`, `vet`, race, lint, and integration tests
- corrected local Docker stack
- liveness/readiness separation
- stable API error format
- OpenAPI skeleton
- CI workflows

This is not premature refactoring. These changes remove ambiguity at the boundaries the UI and future functionality will depend on.
