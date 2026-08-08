# Community Domain

## Purpose

The Community domain represents a group or community within The Village platform.

A Community is currently the first domain implemented end-to-end and establishes the architectural pattern that future domains should follow.

## Domain Model

The Community entity contains:

- `ID`
- `Name`
- `Slug`
- `Description`
- `ExternalSource`
- `CreatedAt`
- `UpdatedAt`

## Layering

The Community domain follows this dependency direction:

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

The domain model does not depend on HTTP, PostgreSQL, or infrastructure concerns.

## Model Responsibilities

The model is responsible for representing Community data and enforcing domain-level validation.

Examples:

- Name is required.
- Name must be between 3 and 100 characters.
- Slug is required.
- Slug must be lowercase and URL-safe.
- Description is optional.
- Description has a maximum length of 2000 characters.

## Service Responsibilities

The service layer owns business rules.

Responsibilities include:

- Normalizing input.
- Validating Communities.
- Detecting duplicate slugs.
- Coordinating repository operations.
- Translating repository errors into service-level errors.
- Wrapping errors with useful context.

Business rules should not be moved into HTTP handlers.

## Repository Responsibilities

The repository owns PostgreSQL persistence.

Responsibilities include:

- Creating Communities.
- Finding Communities by ID.
- Finding Communities by slug.
- Listing Communities.
- Updating Communities.
- Deleting Communities.
- Providing test cleanup helpers.

The repository should not contain HTTP concerns or API response formatting.

## Handler Responsibilities

HTTP handlers should remain thin.

Handlers are responsible for:

1. Parsing HTTP requests.
2. Calling the service layer.
3. Translating service errors into HTTP responses.
4. Encoding JSON responses.
5. Returning appropriate HTTP status codes.

Handlers should not contain business rules or SQL.

## Architectural Rule

Dependencies must flow inward:

    HTTP
      ↓
    Service
      ↓
    Repository
      ↓
    Database

The database must never become a dependency of the service or handler layer directly.