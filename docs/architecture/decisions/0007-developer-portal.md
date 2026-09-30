# ADR-0007: Developer portal

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Developers should be able to discover applications, see their status and use
the Golden Path without learning the platform's internals. A portal provides
this, but it must not become a second source of truth alongside Git, and it
must fit the platform's resource budget
(see [ADR-0001](0001-local-first-zero-cost-runtime.md)).

## Decision

- Orbit has a **custom developer portal**:
  - frontend: **React, TypeScript and Vite**;
  - backend: a **Go backend-for-frontend** that also serves the built
    frontend, packaged as **one container**.
- **Git remains the source of truth.** Every change the portal makes is a
  **pull request** opened through the GitHub API: creating an application
  adds its manifest; promoting a release updates `env/prod.yaml`
  (see [ADR-0004](0004-environments-and-promotion.md)).
- The portal has **no database**:
  - the application catalog is read from the manifests in Git;
  - live status comes from the Argo CD API (sync and health) and the
    Kubernetes API;
  - metrics come from Prometheus when observability is enabled.
- The backend holds all credentials (GitHub, Argo CD); the browser never sees
  them. Credentials are delivered as SOPS-encrypted secrets
  (see [ADR-0006](0006-secrets-with-sops-and-age.md)). The GitHub credential
  is scoped to the Orbit repository only.
- The portal is **deployed as an Orbit application** through the Golden Path
  chart.
- The portal is **optional**: everything it does can also be done directly
  through Git.
- Authentication is single-user for local use at first; GitHub OAuth sign-in
  is a later addition.
- The portal is delivered in two phases:
  1. **read-only**: application catalog and status;
  2. **self-service**: create applications and promote releases.

## Consequences

- Positive: a small footprint and no state to back up or migrate.
- Positive: every portal action is reviewable and auditable in Git.
- Positive: the portal uses the platform in the same way as any other
  application, which continuously validates the Golden Path.
- Negative: features that Backstage provides out of the box must be built.
  Scope is controlled by exposing only Golden Path actions.
- Negative: portal writes are asynchronous: a change takes effect only after
  its pull request is merged and Argo CD has synced.

## Alternatives considered

- **Backstage** — the de facto open-source portal, but it needs a database and
  roughly a gigabyte of memory or more, and its plugin ecosystem would
  dominate the project's complexity.
- **Commercial portals (SaaS)** — not self-hosted and not free.
- **No portal** — Git-only self-service remains supported, but a portal
  significantly improves discoverability and the developer experience.
