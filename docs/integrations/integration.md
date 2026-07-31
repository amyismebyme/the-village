# Integration Tests

These tests verify the application against real infrastructure.

They require Docker and PostgreSQL.

Run:

```bash
go test -tags=integration ./internal/integration/...

