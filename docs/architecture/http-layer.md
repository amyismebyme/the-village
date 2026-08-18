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


```

This keeps responses consistent across every endpoint in the application.

## Route Registration Architecture

HTTP route registration is intentionally separated into three levels.

### Application router

`internal/server/router.go` constructs the `http.ServeMux` and applies the HTTP middleware stack.

### Route composition

`internal/server/routes.go` is the central composition point:

```text
registerRoutes()
├── registerSystemRoutes()
└── registerAPIV1Routes()