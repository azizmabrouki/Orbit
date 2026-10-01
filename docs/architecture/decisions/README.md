# Architecture Decision Records

This directory is the authoritative record of the architecture and technology
decisions behind Orbit. Each record describes the context, the decision, its
consequences and the alternatives that were considered.

## Rules

- A decision is recorded when it is made, in the same change that relies on
  it.
- Accepted records are immutable. A decision that changes is superseded by a
  new record, and the old record is marked **Superseded by ADR-NNNN**.
- Anything undecided or contradictory goes under
  [Open questions](#open-questions), never into a record.

## Status values

- **Proposed**: under discussion
- **Accepted**: part of Orbit's design
- **Superseded**: replaced by a later record (linked)

## Index

| ADR | Title | Decision in brief | Status |
|---|---|---|---|
| [0001](0001-local-first-zero-cost-runtime.md) | Local-first, zero-cost platform runtime | Single-node k3d/K3s on a laptop; public GitHub repository, Actions and GHCR; no Terraform; memory budget ≤ 1.5 GiB core, ≤ 3 GiB in total | Accepted |
| [0002](0002-developer-contract.md) | Developer contract: the Orbit application manifest | `orbit.yaml` plus `env/<env>.yaml` is the values input of the `helm/orbit-app` Golden Path chart, validated by `values.schema.json` | Accepted |
| [0003](0003-gitops-delivery-with-argo-cd.md) | GitOps delivery with Argo CD | Argo CD is the only applier after bootstrap: app-of-apps, one ApplicationSet-generated Application per application and environment, single repository, no cluster credentials in CI | Accepted |
| [0004](0004-environments-and-promotion.md) | Environments and promotion | Two environments (`dev`, `prod`); immutable SHA tags; CI updates dev; prod promotion by pull request; rollback by revert | Accepted |
| [0005](0005-gateway-api-with-traefik.md) | Traffic management with Gateway API and Traefik | Gateway API implemented by Traefik, installed through GitOps; `<app>.<env>.localhost`; TLS deferred | Accepted |
| [0006](0006-secrets-with-sops-and-age.md) | Secrets management with SOPS and age | SOPS + age, decrypted inside the cluster by an operator; age key kept outside Git | Accepted |
| [0007](0007-developer-portal.md) | Developer portal | Custom portal (React, TypeScript, Vite + Go backend, one container) that writes only through pull requests and has no database | Accepted |
| [0008](0008-ci-and-supply-chain-security.md) | CI and supply-chain security | GitHub Actions: test → build → Trivy scan + SBOM → GHCR → dev tag update; SHA-pinned actions; cosign deferred | Accepted |
| [0009](0009-observability.md) | Observability | Prometheus + Grafana (trimmed), Loki + Alloy; optional; tracing deferred | Accepted |
| [0010](0010-policy-enforcement-with-kyverno.md) | Policy enforcement with Kyverno | Kyverno in the security phase: Audit then Enforce, same policies checked in CI | Accepted |
| [0011](0011-go-for-platform-services.md) | Go for platform-owned services | Go for platform-owned services, starting with `orbit-demo` | Accepted |
| [0012](0012-native-linux-development-environment.md) | Native Linux development environment | Develop on native Ubuntu with Docker Engine; the memory budget still targets an 8 GB laptop | Accepted |

## Deferred by decision

These items are decided: they are introduced when there is a reason for them.

| Item | Introduced when | ADR |
|---|---|---|
| Image signing (cosign) and signature verification at admission | Security phase | 0008, 0010 |
| Distributed tracing (OpenTelemetry, Tempo) | A workload scenario needs it | 0009 |
| Alertmanager | Alerting is demonstrated | 0009 |
| TLS | There is a need to demonstrate it | 0005 |
| Portal sign-in with GitHub OAuth | After single-user local use | 0007 |
| Terraform | There is cloud infrastructure to manage | 0001 |

## Open questions

1. **SOPS operator.** sops-secrets-operator is the primary candidate and the
   KSOPS plugin the fallback, to be validated when GitOps is implemented.
   [ADR-0006](0006-secrets-with-sops-and-age.md) lists repo-server
   decryption (KSOPS) as a rejected alternative because plaintext enters
   Argo CD's cache, yet keeps it as the fallback. Using the fallback needs
   its own record that accepts that trade-off.
2. **Branch protection on `main`.**
   [ADR-0004](0004-environments-and-promotion.md) requires CI to push dev
   image tag updates directly to `main`. How that works with branch
   protection is undecided (CI and GitOps phases).
3. **Portal GitHub credential.** Fine-grained personal access token or
   GitHub App. Only the scope is decided: the Orbit repository alone
   (portal self-service phase).
4. **Kyverno failure policy.**
   [ADR-0010](0010-policy-enforcement-with-kyverno.md) says the admission
   webhook's failure policy must be chosen deliberately. It has not been
   chosen yet (security phase).
5. **Chart and image sourcing.** Preferring upstream charts and images over
   Bitnami, and using third-party ones only for a clear reason, was proposed
   during the architecture review but never recorded as a decision. Confirm
   it as a record or drop it.

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
