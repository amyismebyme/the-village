# ADR 0001: Use a Modular Monolith

- Status: Proposed
- Date: 2026-08-03

## Context

The Village needs to deliver product functionality while also demonstrating production-minded engineering. The current team/project size does not justify independent services, distributed transactions, or multiple deployment pipelines.

## Decision

Build the backend as one deployable Go application with internal package boundaries for handlers, services, repositories, models, and operational concerns.

## Consequences

### Positive

- simpler deployment and local development
- easier end-to-end testing
- lower operational burden
- straightforward transactions and consistency
- package boundaries still support later extraction

### Negative

- boundaries rely on code discipline
- one deployment contains multiple domains
- failures can affect the whole API unless isolated internally

## Revisit when

- domains require independent scaling
- team ownership separates
- deployment cadence conflicts
- reliability isolation provides measurable value
