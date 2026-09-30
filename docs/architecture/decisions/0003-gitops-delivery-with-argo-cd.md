# ADR-0003: GitOps delivery with Argo CD

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Orbit uses Git as the source of truth for everything running on the platform.
Deployments must be auditable, reversible and free of manual cluster access,
and onboarding a new application must not require hand-written deployment
configuration.

## Decision

- **Argo CD** is the only mechanism that applies changes to the cluster after
  bootstrap. Managed resources are never changed with `kubectl apply` or
  `helm install`.
- **Bootstrap** is a single script that creates the k3d cluster, installs
  Argo CD, injects the secrets decryption key
  (see [ADR-0006](0006-secrets-with-sops-and-age.md)) and applies one root
  Application. From then on Argo CD manages all platform components,
  including itself (**app-of-apps** pattern).
- An **ApplicationSet** generates one Argo CD Application per application and
  environment by discovering application manifests under `gitops/apps/`.
  Onboarding an application means adding its directory; nobody writes an
  Argo CD Application by hand.
- Argo CD runs in a **trimmed configuration**: single replicas, with Dex and
  the notifications controller disabled.
- **CI never has cluster credentials.** It changes the cluster only by
  committing to Git.
- Orbit uses a **single repository**:

  ```text
  apps/              application source code (sample workload, portal)
  helm/orbit-app/    Golden Path chart
  gitops/
    bootstrap/       root Application
    platform/        platform components
    apps/<name>/     orbit.yaml, env/dev.yaml, env/prod.yaml
  infrastructure/    local cluster configuration and bootstrap scripts
  ```

## Consequences

- Positive: full audit trail in Git history; Argo CD corrects drift; the
  entire platform can be rebuilt from the repository.
- Positive: application onboarding is self-service and declarative.
- Positive: Argo CD's API exposes sync and health status, which the portal
  uses (see [ADR-0007](0007-developer-portal.md)).
- Negative: CI commits image tag updates to the same repository that holds
  application code. Workflows use path filters so that desired-state changes
  do not trigger builds.
- Negative: application code and desired state share one repository. This is
  acceptable at Orbit's scale; a separate configuration repository can be
  introduced later without changing the model.
- Negative: Argo CD is the largest component of the core platform's memory
  footprint.

## Alternatives considered

- **Flux** — lighter and supports SOPS natively, but has no built-in UI and no
  equivalent to ApplicationSets for generating applications from a directory
  structure.
- **Push-based deployment from CI** — requires cluster credentials in CI and
  provides no drift detection.
- **Separate configuration repository** — cleaner separation of concerns, but
  extra coordination overhead with no benefit at the current scale.
