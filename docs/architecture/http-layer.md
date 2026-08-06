# HTTP Layer

## Purpose

The HTTP layer is responsible for translating HTTP requests into service calls and converting service results into JSON responses.

The HTTP layer **must not** contain business logic.

---

## Request Flow

```text
Client
   │
   ▼
HTTP Handler
   │
   ▼
Service
   │
   ▼
Repository
   │
   ▼
PostgreSQL
```

---

## Responsibilities

Handlers are responsible for:

- Parsing request bodies.
- Reading route parameters.
- Calling the appropriate service.
- Returning HTTP status codes.
- Returning JSON responses.

Handlers are **not** responsible for:

- Validation rules.
- Duplicate detection.
- Database access.
- Business rules.

---

## API Helpers

All handlers use the shared `internal/api` package.

### Success

```go
api.WriteJSON(w, http.StatusOK, community)
```

### Bad Request

```go
api.BadRequest(w, "validation failed")
```

### Not Found

```go
api.NotFound(w, "community not found")
```

### Conflict

```go
api.Conflict(w, "community already exists")
```

### Internal Server Error

```go
api.InternalServerError(w)
```

This keeps responses consistent across every endpoint in the application.