# ADR-0004: Environments and promotion

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

A platform must let teams deploy continuously to a non-production
environment and promote tested releases to production in a controlled,
reviewable way. Orbit's resource budget
(see [ADR-0001](0001-local-first-zero-cost-runtime.md)) limits how many
environments can run at the same time.

## Decision

- Orbit has **two environments: `dev` and `prod`**.
- Each application gets **one namespace per environment**, named
  `<app>-<env>` (for example `orbit-demo-dev`), which scopes RBAC, network
  policies and resource usage per application.
- Images are tagged with the **Git commit SHA** and are immutable. The
  `latest` tag is never deployed.
- **Continuous delivery to dev:** when a change to an application is merged
  to `main`, CI builds, scans and pushes the image, then updates the image
  tag in `gitops/apps/<app>/env/dev.yaml`. Argo CD deploys it.
- **Promotion to prod** is a pull request that copies the image tag from
  `env/dev.yaml` to `env/prod.yaml`. Merging the pull request is the approval
  and triggers the deployment.
- **Rollback** is a revert of the promotion commit.
- The developer portal automates creating promotion pull requests but does
  not change the model (see [ADR-0007](0007-developer-portal.md)).

## Consequences

- Positive: the same artifact is built once and promoted; every production
  change is reviewable, auditable and reversible through Git.
- Positive: the promotion model is explicit and requires no additional
  component.
- Negative: environments share a single node, so they are isolated logically,
  not physically.
- Negative: CI must be allowed to push dev tag updates directly to `main`.
  Production changes always go through a pull request.

## Alternatives considered

- **A third `staging` environment** — more realistic, but exceeds the memory
  budget and demonstrates nothing that two environments do not.
- **Argo CD Image Updater** — automates tag updates but adds a component and
  hides the promotion logic.
- **Kargo** — a capable promotion engine, but adds significant complexity and
  memory for a two-environment workflow.
