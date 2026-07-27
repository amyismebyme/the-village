# Testing Strategy

## Overview

The Village API follows a layered testing strategy to ensure reliability, maintainability, and confidence in production deployments.

Testing is an essential part of the software development lifecycle and supports the project's goal of building production-quality backend services using Site Reliability Engineering (SRE) principles.

---

# Testing Pyramid

```
                End-to-End
             Integration Tests
               Unit Tests
```

The majority of tests are unit tests because they provide fast feedback and isolate failures.

---

# Unit Tests

Unit tests validate individual packages independently.

Current coverage includes:

- Handlers
- Middleware
- Runtime
- Metrics

All unit tests run automatically during Continuous Integration.

Execute locally:

```bash
go test ./...
```

---

# Benchmarks

Benchmarks monitor performance of frequently executed code.

Current benchmarks include:

- Health Handler
- Logging Middleware

Execute locally:

```bash
go test -bench=. ./...
```

or

```bash
go test -bench=. -benchmem ./...
```

---

# Race Detection

The Go race detector is executed in CI.

It identifies concurrent memory access bugs.

Run locally:

```bash
go test -race ./...
```

---

# Coverage

Coverage reports are automatically generated during CI.

Generate locally:

```bash
go test ./... \
-covermode=atomic \
-coverprofile=coverage.out

go tool cover -func=coverage.out
```

HTML report:

```bash
go tool cover -html=coverage.out
```

---

# Continuous Integration

Every Pull Request automatically performs:

- Dependency download
- Formatting verification
- go vet
- Unit tests
- Race detection
- Coverage generation

This prevents regressions before merging code.

---

# Future Improvements

Planned enhancements include:

- Integration tests
- End-to-end tests
- Docker Compose test environment
- Load testing
- Chaos testing
- Performance regression detection
- Security scanning