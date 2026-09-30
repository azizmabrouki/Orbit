# Architecture Decision Records

This directory records the significant architectural decisions behind Orbit.

Each record describes the context, the decision, its consequences and the
alternatives that were considered. Records are immutable once accepted: a
decision that changes is superseded by a new record rather than edited.

## Status values

- **Proposed** — under discussion
- **Accepted** — part of Orbit's design
- **Superseded** — replaced by a later record (linked)

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-local-first-zero-cost-runtime.md) | Local-first, zero-cost platform runtime | Accepted |
| [0002](0002-developer-contract.md) | Developer contract: the Orbit application manifest | Accepted |
| [0003](0003-gitops-delivery-with-argo-cd.md) | GitOps delivery with Argo CD | Accepted |
| [0004](0004-environments-and-promotion.md) | Environments and promotion | Accepted |
| [0005](0005-gateway-api-with-traefik.md) | Traffic management with Gateway API and Traefik | Accepted |
| [0006](0006-secrets-with-sops-and-age.md) | Secrets management with SOPS and age | Accepted |
| [0007](0007-developer-portal.md) | Developer portal | Accepted |
| [0008](0008-ci-and-supply-chain-security.md) | CI and supply-chain security | Accepted |
| [0009](0009-observability.md) | Observability | Accepted |
| [0010](0010-policy-enforcement-with-kyverno.md) | Policy enforcement with Kyverno | Accepted |
| [0011](0011-go-for-platform-services.md) | Go for platform-owned services | Accepted |

## Template

```markdown
# ADR-NNNN: Title

- **Status:** Proposed | Accepted | Superseded by ADR-XXXX
- **Date:** YYYY-MM-DD

## Context

## Decision

## Consequences

## Alternatives considered
```
