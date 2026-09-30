# ADR-0011: Go for platform-owned services

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Orbit owns two services: a sample workload that exercises the platform, and
the backend of the developer portal. Both should be small and cheap to run,
and they should model the practices the Golden Path expects of applications.

## Decision

- Platform-owned services are written in **Go**.
- The sample workload, **`orbit-demo`**, is intentionally minimal:
  - endpoints: `/`, `/healthz`, `/readyz`, `/metrics`;
  - structured JSON logs to stdout;
  - graceful shutdown;
  - configuration through environment variables.
- Images are built from **minimal base images** and run as a **non-root**
  user.

## Consequences

- Positive: static binaries, small images and a low memory footprint.
- Positive: the sample workload doubles as the reference implementation of
  the Golden Path's workload requirements
  (see [ADR-0002](0002-developer-contract.md)).
- Negative: the portal uses two languages (TypeScript and Go). This is an
  accepted trade-off for keeping credentials and API access on the server.

## Alternatives considered

- **Python** — larger images and a higher memory footprint for the same
  functionality.
- **Node.js** — would share a language with the portal frontend, but gives
  larger images and higher memory use than Go.
