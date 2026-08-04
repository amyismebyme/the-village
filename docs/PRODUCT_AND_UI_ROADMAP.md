# Product and UI Roadmap

## Product principle

The Village should prioritize a small number of trustworthy, useful workflows rather than immediately becoming a broad social network. The first product should help someone discover a relevant community and understand how to participate safely.

## First usable product slice

### Community discovery

Users can:

- browse available communities
- filter or search by category/location later
- open a community detail page
- understand the community's purpose, source, and participation path

### Community management

Authorized operators can:

- create a community
- edit its name, slug, category, description, and external link
- archive or delete it

Authentication and role enforcement may follow after the API contract exists, but write endpoints should not be publicly exposed in a real environment without protection.

## Recommended Community shape

The final shape should be chosen deliberately. A practical starting point:

```text
id
name
slug
description
category
external_url (optional)
external_source
created_at
updated_at
```

Future additions:

- location/region
- online versus in-person
- moderation status
- image/avatar
- membership count
- owner/creator
- archived_at

Do not add all future fields now. Add them when a user workflow requires them.

## Frontend foundation

The `apps/frontend` directory is currently empty. Recommended stack from the project vision:

- Next.js
- React
- TypeScript
- accessible component system
- generated or typed API client from OpenAPI

## Initial routes

```text
/                         landing page
/communities              community list
/communities/[slug]       community detail
/communities/new          create form, protected later
/communities/[slug]/edit  edit form, protected later
```

## Initial UI components

- application header/navigation
- community card
- community list/grid
- category badge
- empty state
- loading skeleton
- API error message
- community form
- confirmation dialog for destructive actions
- health/status indicator only for internal/admin use

## UX requirements

Because the product relates to loneliness and potentially sensitive topics:

- use supportive, non-judgmental language
- avoid manipulative engagement patterns
- provide clear reporting/moderation paths when social features arrive
- design for accessibility from the start
- clearly distinguish peer community resources from professional crisis support
- avoid presenting the application as medical treatment

## API contract needed by the UI

Before implementation, define:

### List response

- array shape
- ordering
- pagination metadata
- filtering parameters

### Detail response

- lookup by numeric ID or slug; slug is preferable for public URLs

### Error envelope

Example:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "The request contains invalid fields.",
    "fields": {
      "name": "must contain between 3 and 100 characters"
    },
    "request_id": "..."
  }
}
```

### Concurrency

For early CRUD, last-write-wins may be acceptable. Later, add optimistic concurrency using `updated_at`, ETags, or a version field if concurrent editing becomes realistic.

## Delivery sequence

### Phase 1 — API completion

- align schema
- complete repository and service
- add HTTP handlers
- add API integration tests
- publish OpenAPI

### Phase 2 — read-only UI

- app shell
- list and detail pages
- loading/error/empty states
- basic responsive layout

### Phase 3 — management UI

- create and edit forms
- client and server validation
- delete/archive confirmation
- authentication and authorization

### Phase 4 — product expansion

- resources associated with communities
- events or meetups
- membership and profiles
- moderation and reporting
- search and recommendations

## SRE integration with UI delivery

Each UI/API feature should ship with:

- unit tests
- API integration tests
- frontend component/page tests
- normalized route metrics
- structured logs
- documented failure states
- deployment rollback plan

The goal is not to bolt reliability onto the finished product. Reliability should be part of the definition of done for each vertical slice.
