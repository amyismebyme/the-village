# Milestone 7 — Community Domain

## Sprint A — Foundation

Status: Complete

## Objective

Establish a production-quality Community domain with:

- Domain model.
- Validation.
- Service layer.
- PostgreSQL repository.
- Error translation.
- Repository helpers.
- Integration tests.
- Documentation.

## Completed Tasks

### Task 1 — Community Domain

Completed.

The Community model is frozen for the current milestone.

Implemented:

- Identity.
- Name.
- Slug.
- Description.
- External source.
- Timestamps.
- Validation.

### Task 2 — Repository Error Translation

Completed.

Repository errors are translated into application-level repository errors.

Examples:

- Not found.
- Already exists.
- Database failures.

### Task 3 — PostgreSQL Error Translator

Completed.

PostgreSQL-specific errors are translated before leaving the repository layer.

This prevents infrastructure-specific errors from leaking into higher layers.

### Task 4 — Repository Methods

Completed.

Community persistence operations are implemented:

- Create.
- FindByID.
- FindBySlug.
- List.
- Update.
- Delete.

### Task 5 — Repository Helpers

Completed.

Repository helpers were introduced to reduce duplicated PostgreSQL logic.

Examples include:

- Community row scanning.
- Common execution behavior.
- Existence checks.
- Integration-test cleanup.

### Task 6 — Service Cleanup

Completed.

The Community service now owns:

- Input normalization.
- Validation.
- Duplicate detection.
- Error wrapping.
- ID validation.

### Task 7 — Integration Tests

Completed.

Community service behavior has been tested against a real PostgreSQL instance.

Covered operations:

- Create.
- Get.
- List.
- Update.
- Delete.
- Duplicate detection.
- Validation.
- Not found.

### Task 8 — Documentation

Completed.

Architecture, testing, and repository/service boundaries are documented.

## Current Architecture

```text
HTTP Handler
     |
     v
Community Service
     |
     v
Community Repository
     |
     v
PostgreSQL



## One reconciliation I strongly recommend

Before moving to the next sprint, run this **final Sprint A gate**:

```bash
go fmt ./...
go vet ./...
go test ./... -v
go test -tags=integration ./internal/integration/... -v