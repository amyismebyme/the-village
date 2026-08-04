# Database and Migrations

## Database technology

The API uses PostgreSQL through `github.com/jackc/pgx/v5/pgxpool`.

The application opens a pool during startup, applies pool limits from environment configuration, pings the database, and fails startup if the connection cannot be established.

## Migration tooling

Migrations use `golang-migrate` naming conventions:

```text
000001_initial.up.sql
000001_initial.down.sql
000002_seed.up.sql
000002_seed.down.sql
```

Apply migrations from the repository root:

```bash
migrate \
  -path ./migrations \
  -database "postgres://village:village@localhost:5433/village?sslmode=disable" \
  up
```

## Current schema

Migration `000001_initial.up.sql` creates:

### `communities`

| Column | Type | Notes |
|---|---|---|
| `id` | `BIGSERIAL` | primary key |
| `name` | `TEXT` | required |
| `description` | `TEXT` | required |
| `url` | `TEXT` | required |
| `external_source` | `TEXT` | defaults to `internal` |
| `created_at` | `TIMESTAMPTZ` | defaults to current time |
| `updated_at` | `TIMESTAMPTZ` | defaults to current time |

Index: `idx_communities_name`.

### `resources`

| Column | Type |
|---|---|
| `id` | `BIGSERIAL` |
| `title` | `TEXT` |
| `description` | `TEXT` |
| `url` | `TEXT` |
| `category` | `TEXT` |
| `created_at` | `TIMESTAMPTZ` |
| `updated_at` | `TIMESTAMPTZ` |

Index: `idx_resources_category`.

## Critical schema drift

The current Community model and repository SQL expect:

```text
id, name, slug, description, external_source, created_at, updated_at
```

The migration actually creates:

```text
id, name, description, url, external_source, created_at, updated_at
```

The model also includes `Category`, which is not in the table.

This mismatch means the current Community `List` and `FindBySlug` SQL cannot work against the checked-in migration. It must be resolved before Community CRUD development continues.

## Recommended Community schema

For a product-owned Community entity, use a stable slug and optional external URL:

```sql
CREATE TABLE communities (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    external_url TEXT,
    external_source TEXT NOT NULL DEFAULT 'internal',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_communities_name ON communities(name);
CREATE INDEX idx_communities_category ON communities(category);
```

Do not edit an already-applied migration in shared environments. Add a new forward migration that introduces/renames columns and updates seed data. During early local-only development, resetting migrations is possible, but record the decision.

## Timestamp behavior

`updated_at DEFAULT NOW()` only sets the value on insert. Repository update statements must explicitly set:

```sql
updated_at = NOW()
```

Alternatively, add a database trigger, but explicit SQL is easier to understand at this stage.

## Error translation

PostgreSQL repository methods should translate:

- `pgx.ErrNoRows` → `repository.ErrNotFound`
- unique violation on `communities.slug` → `repository.ErrAlreadyExists`

Keep raw database details out of HTTP responses.

## Seed data

Migration `000002_seed.up.sql` seeds two communities and two resources. Seed data currently uses the old `url` column and must be updated if the Community schema changes.

## Integration environment

`testdata/docker-compose.integration.yml` exposes PostgreSQL on host port `5433`, uses database/user/password `village`, and defines a health check.

The integration scripts remove the volume before each run, ensuring migrations are applied to a clean database.
