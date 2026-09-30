# ADR-0002: Developer contract — the Orbit application manifest

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

The core of an Internal Developer Platform is the contract between the teams
that build applications and the platform that runs them. Developers should be
able to deploy an application without writing Kubernetes manifests, Argo CD
Applications, Gateway routes, monitoring configuration or security policies.

The contract must be small, explicit, versioned and machine-validated, and it
must be usable both by hand and by the developer portal
(see [ADR-0007](0007-developer-portal.md)).

## Decision

- Each application is described by an **Orbit application manifest**,
  `orbit.yaml`, plus optional **per-environment files** (`env/dev.yaml`,
  `env/prod.yaml`) that carry environment-specific values such as the image
  tag and replica count.
- The manifest **is the values input** of the Orbit Golden Path Helm chart
  (`helm/orbit-app`). There is no separate translation layer.
- The chart ships a **`values.schema.json`**. Manifests are validated in CI
  and by Helm at render time. The schema is the versioned public contract.
- **Responsibilities** are split as follows:

  | Platform (Golden Path chart) | Application team |
  |---|---|
  | Deployment, Service, ServiceAccount | Application code |
  | HTTPRoute on the shared Gateway | Dockerfile |
  | Probes, resource defaults | `/healthz`, `/readyz` endpoints |
  | Security context defaults | `/metrics` endpoint |
  | Metrics scraping configuration | JSON logs to stdout |
  | Network and admission policies | `orbit.yaml` |

- A workload on the Golden Path must: listen on the configured port, expose
  `/healthz` and `/readyz`, write structured JSON logs to stdout, and run as a
  non-root user.

Illustrative example — the schema is authoritative:

```yaml
# gitops/apps/orbit-demo/orbit.yaml
name: orbit-demo
team: platform
image:
  repository: ghcr.io/<owner>/orbit-demo
service:
  port: 8080
route:
  enabled: true
resources:
  size: small
```

```yaml
# gitops/apps/orbit-demo/env/dev.yaml
image:
  tag: 3f2c1a9
replicas: 1
```

## Consequences

- Positive: developers write one short YAML file; no custom controller has to
  be built or operated; standard Helm tooling applies.
- Positive: the portal generates and edits the same files a developer would
  write by hand, so the portal is never required.
- Negative: Helm values are not a true API. Contract stability depends on
  schema discipline and on versioning the chart for breaking changes.
- Negative: the abstraction deliberately limits what applications can
  configure. Escape hatches are added only for demonstrated needs.

## Alternatives considered

- **Custom resource definition with a controller** — a stronger API boundary,
  but requires writing and operating a controller. Revisit if the Helm-based
  contract proves insufficient.
- **Score specification** — adds a translation tool and a smaller ecosystem
  without solving a problem Orbit currently has.
- **Backstage software templates** — see [ADR-0007](0007-developer-portal.md).
- **Raw Helm or Kustomize per application** — no abstraction; every team
  re-implements the same Kubernetes configuration.
