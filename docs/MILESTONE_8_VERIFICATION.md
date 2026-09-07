# Milestone 8 Verification

## Current verification scope

Milestone verification covers application configuration, error handling, context propagation, security-sensitive logging, dependency vulnerability scanning, Docker runtime health, Reddit smoke testing, failure/recovery behavior, performance sanity, documentation consistency, and clean builds.

## Current persistence state

Reddit ingestion persists normalized and deduplicated external items through the `ExternalItemRepository` using `(source, external_id)` uniqueness. Resource PostgreSQL CRUD is implemented; Resource service/HTTP endpoints remain future product work.

## Docker certificate note

The current Dockerfile still expects the local `zscaler.crt` certificate for developer builds. This is intentionally deferred and is not part of the Milestone 8 P1 cleanup.

## Live Reddit smoke

The live Reddit smoke test is opt-in and requires `REDDIT_CLIENT_ID`, `REDDIT_CLIENT_SECRET`, and `REDDIT_USER_AGENT`. Credentials must be supplied through the environment and must not be committed or printed.
