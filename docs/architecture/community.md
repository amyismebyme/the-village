# Community API

## Overview

The Community API provides CRUD operations for managing communities in The Village API.

### Base URL

```text
/api/v1
```

### Resource

```text
/api/v1/communities
```

The API uses JSON for request and response bodies.

---

# 1. POST `/api/v1/communities`

Create a new Community.

## Request

```http
POST /api/v1/communities
Content-Type: application/json
```

### Request body

```json
{
  "name": "Toronto Men",
  "slug": "toronto-men",
  "description": "A community for men in Toronto.",
  "external_source": "manual"
}
```

### Fields

| Field             | Type   | Required | Description                            |
| ----------------- | ------ | -------: | -------------------------------------- |
| `name`            | string |      Yes | Community display name                 |
| `slug`            | string |      Yes | URL-safe unique community identifier   |
| `description`     | string |       No | Community description                  |
| `external_source` | string |       No | Identifies the source of the community |

Unknown JSON fields are rejected.

### Success

```http
201 Created
Content-Type: application/json
```

Example:

```json
{
  "id": 1,
  "name": "Toronto Men",
  "slug": "toronto-men",
  "description": "A community for men in Toronto.",
  "external_source": "manual",
  "created_at": "2026-08-11T10:00:00Z",
  "updated_at": "2026-08-11T10:00:00Z"
}
```

### Status codes

| Status | Meaning                                                |
| -----: | ------------------------------------------------------ |
|  `201` | Community created                                      |
|  `400` | Invalid request, malformed JSON, or validation failure |
|  `409` | Community with the slug already exists                 |
|  `500` | Internal/database failure                              |

---

# 2. GET `/api/v1/communities`

Return all Communities.

## Request

```http
GET /api/v1/communities
```

### Success

```http
200 OK
Content-Type: application/json
```

Example:

```json
{
  "communities": [
    {
      "id": 1,
      "name": "Toronto Men",
      "slug": "toronto-men",
      "description": "A community for men in Toronto.",
      "external_source": "manual",
      "created_at": "2026-08-11T10:00:00Z",
      "updated_at": "2026-08-11T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Mississauga Men",
      "slug": "mississauga-men",
      "description": "A community for men in Mississauga.",
      "external_source": "manual",
      "created_at": "2026-08-11T10:05:00Z",
      "updated_at": "2026-08-11T10:05:00Z"
    }
  ]
}
```

### Empty result

The API always returns an array rather than `null`.

```json
{
  "communities": []
}
```

### Status codes

| Status | Meaning                           |
| -----: | --------------------------------- |
|  `200` | Communities returned successfully |
|  `500` | Internal/database failure         |

---

# 3. GET `/api/v1/communities/{id}`

Retrieve a single Community.

## Request

```http
GET /api/v1/communities/1
```

### Path parameters

| Parameter | Type    | Required | Description          |
| --------- | ------- | -------: | -------------------- |
| `id`      | integer |      Yes | Community identifier |

### Success

```http
200 OK
Content-Type: application/json
```

Example:

```json
{
  "id": 1,
  "name": "Toronto Men",
  "slug": "toronto-men",
  "description": "A community for men in Toronto.",
  "external_source": "manual",
  "created_at": "2026-08-11T10:00:00Z",
  "updated_at": "2026-08-11T10:00:00Z"
}
```

### Invalid ID

Examples:

```text
/api/v1/communities/abc
/api/v1/communities/0
/api/v1/communities/-1
```

Return:

```http
400 Bad Request
```

### Status codes

| Status | Meaning                   |
| -----: | ------------------------- |
|  `200` | Community returned        |
|  `400` | Invalid Community ID      |
|  `404` | Community does not exist  |
|  `500` | Internal/database failure |

---

# 4. PUT `/api/v1/communities/{id}`

Update an existing Community.

## Request

```http
PUT /api/v1/communities/1
Content-Type: application/json
```

### Request body

```json
{
  "name": "Toronto Men's Community",
  "slug": "toronto-men",
  "description": "Updated description.",
  "external_source": "manual"
}
```

The Community ID is determined by the URL.

For example:

```text
PUT /api/v1/communities/1
```

updates Community `1`.

The client does not control the resource ID through the JSON body.

### Success

```http
200 OK
Content-Type: application/json
```

Example:

```json
{
  "id": 1,
  "name": "Toronto Men's Community",
  "slug": "toronto-men",
  "description": "Updated description.",
  "external_source": "manual",
  "created_at": "2026-08-11T10:00:00Z",
  "updated_at": "2026-08-11T10:15:00Z"
}
```

### Status codes

| Status | Meaning                                             |
| -----: | --------------------------------------------------- |
|  `200` | Community updated                                   |
|  `400` | Invalid ID, malformed JSON, or validation failure   |
|  `404` | Community does not exist                            |
|  `409` | Requested slug already belongs to another Community |
|  `500` | Internal/database failure                           |

Duplicate-slug detection is handled by the service layer, not by the HTTP handler.

---

# 5. DELETE `/api/v1/communities/{id}`

Delete a Community.

## Request

```http
DELETE /api/v1/communities/1
```

### Success

```http
204 No Content
```

The response contains no body.

### Invalid ID

Examples:

```text
/api/v1/communities/abc
/api/v1/communities/0
/api/v1/communities/-1
```

Return:

```http
400 Bad Request
```

### Status codes

| Status | Meaning                   |
| -----: | ------------------------- |
|  `204` | Community deleted         |
|  `400` | Invalid Community ID      |
|  `404` | Community does not exist  |
|  `500` | Internal/database failure |

---

# 6. HTTP Method Enforcement

Each endpoint supports one HTTP method.

| Endpoint                   | Method   |
| -------------------------- | -------- |
| `/api/v1/communities`      | `POST`   |
| `/api/v1/communities`      | `GET`    |
| `/api/v1/communities/{id}` | `GET`    |
| `/api/v1/communities/{id}` | `PUT`    |
| `/api/v1/communities/{id}` | `DELETE` |

Handlers reject unsupported methods with:

```http
405 Method Not Allowed
```

The response includes an `Allow` header describing the supported method.

Example:

```http
HTTP/1.1 405 Method Not Allowed
Allow: GET
```

---

# 7. Request Validation

The HTTP handler is responsible for HTTP-specific validation.

This includes:

* HTTP method
* JSON syntax
* request body presence
* `Content-Type`
* unknown JSON fields
* path parameter format
* request size
* multiple JSON values in a single request body

Unknown JSON fields are rejected.

For example:

```json
{
  "name": "Toronto Men",
  "slug": "toronto-men",
  "banana": "hello"
}
```

is invalid.

The API uses strict JSON decoding:

```go
decoder.DisallowUnknownFields()
```

Multiple JSON objects in a single request body are also rejected.

---

# 8. Domain Validation

Domain validation remains inside the model/service layers.

The handler does not duplicate business rules such as:

```text
name required
name length
slug required
slug format
description length
```

The general request flow is:

```text
HTTP request
     │
     ▼
HTTP handler
     │
     ├── method validation
     ├── JSON validation
     ├── request validation
     │
     ▼
Community Service
     │
     ├── normalization
     ├── domain validation
     ├── duplicate detection
     │
     ▼
Repository
     │
     ▼
PostgreSQL
```

---

# 9. Error Contract

All API errors use a consistent structure.

```json
{
  "error": {
    "code": "community_not_found",
    "message": "community 42 not found"
  }
}
```

Clients should use:

```text
error.code
error.message
```

rather than parsing arbitrary server messages.

## Error codes

| Code                       | Meaning                                   |
| -------------------------- | ----------------------------------------- |
| `invalid_request`          | Generic malformed or invalid HTTP request |
| `invalid_community`        | Community failed domain validation        |
| `invalid_id`               | Invalid Community ID                      |
| `community_not_found`      | Requested Community does not exist        |
| `community_already_exists` | Community slug already exists             |
| `internal_error`           | Unexpected internal/database failure      |

### Example: invalid ID

```http
400 Bad Request
```

```json
{
  "error": {
    "code": "invalid_id",
    "message": "invalid community id"
  }
}
```

### Example: not found

```http
404 Not Found
```

```json
{
  "error": {
    "code": "community_not_found",
    "message": "community 42 not found"
  }
}
```

### Example: duplicate

```http
409 Conflict
```

```json
{
  "error": {
    "code": "community_already_exists",
    "message": "community with slug \"toronto-men\" already exists"
  }
}
```

### Example: internal error

```http
500 Internal Server Error
```

```json
{
  "error": {
    "code": "internal_error",
    "message": "internal server error"
  }
}
```

---

# 10. API Contract Summary

| Method   | Endpoint                   | Success | Client Errors       | Server Error |
| -------- | -------------------------- | ------: | ------------------- | -----------: |
| `POST`   | `/api/v1/communities`      |   `201` | `400`, `409`        |        `500` |
| `GET`    | `/api/v1/communities`      |   `200` | —                   |        `500` |
| `GET`    | `/api/v1/communities/{id}` |   `200` | `400`, `404`        |        `500` |
| `PUT`    | `/api/v1/communities/{id}` |   `200` | `400`, `404`, `409` |        `500` |
| `DELETE` | `/api/v1/communities/{id}` |   `204` | `400`, `404`        |        `500` |

---

# 11. Example CRUD Flow

### Create

```http
POST /api/v1/communities
```

```json
{
  "name": "Toronto Men",
  "slug": "toronto-men",
  "description": "A community for men in Toronto.",
  "external_source": "manual"
}
```

Returns:

```http
201 Created
```

### Retrieve

```http
GET /api/v1/communities/1
```

Returns:

```http
200 OK
```

### Update

```http
PUT /api/v1/communities/1
```

Returns:

```http
200 OK
```

### Delete

```http
DELETE /api/v1/communities/1
```

Returns:

```http
204 No Content
```

### Verify deletion

```http
GET /api/v1/communities/1
```

Returns:

```http
404 Not Found
```

---

# 12. Versioning

The current API is versioned under:

```text
/api/v1
```

Community endpoints therefore use:

```text
/api/v1/communities
```

Future incompatible API changes should use a new API version rather than silently changing the existing contract.

For example:

```text
/api/v2/communities
```

This allows existing clients, including the future Village UI, to continue using the existing API contract while newer clients migrate to the next version.
