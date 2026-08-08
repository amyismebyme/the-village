# Repository and Service Boundary

## Purpose

This document defines the boundary between persistence and business logic.

Keeping this boundary clear is important as The Village grows into multiple domains.

## Repository

The repository answers:

> How do we store and retrieve this data?

For example:

```go
FindByID(ctx, id)
FindBySlug(ctx, slug)
Create(ctx, community)
Update(ctx, community)
Delete(ctx, id)
List(ctx)