# Production Configuration Policy

## Configuration classes

### Required in production

- `PORT`
- `ENVIRONMENT`
- `LOG_LEVEL`
- `LOG_FORMAT`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE`

Production also requires the timeout values to be positive and the relationship:

```text
DB_QUERY_TIMEOUT < REQUEST_TIMEOUT < WRITE_TIMEOUT
```

### Optional with safe defaults

The application provides development-friendly defaults for local use, including:

- `READ_TIMEOUT=10`
- `REQUEST_TIMEOUT=35`
- `WRITE_TIMEOUT=40`
- `IDLE_TIMEOUT=60`
- `SHUTDOWN_TIMEOUT=15`
- `DB_QUERY_TIMEOUT=30`
- connection-pool lifecycle values

These defaults are not a substitute for explicit production configuration.

### Development/test-only behavior

The repository integration environment may intentionally use local PostgreSQL credentials and `DB_SSLMODE=disable`. Those values must not be copied into production configuration.
