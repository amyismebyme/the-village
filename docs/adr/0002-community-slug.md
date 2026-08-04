# ADR 0002: Use Unique Community Slugs for Public URLs

- Status: Proposed
- Date: 2026-08-03

## Context

The Community model contains a slug, while the current database migration contains a URL column. The frontend will need stable, readable Community routes.

## Decision

Store a unique lowercase URL-safe slug for each Community and use it in public frontend routes. Store any external destination separately as an optional external URL.

## Consequences

- database requires a unique slug constraint
- service normalizes and validates slugs
- rename behavior must be considered because changing a slug changes a public URL
- redirects or immutable slugs may be needed later
