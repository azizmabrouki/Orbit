# ADR-0008: CI and supply-chain security

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Every image deployed on Orbit is produced by CI. The pipeline must test,
build and publish images reproducibly, and it must stop known-vulnerable
images before they reach any environment. The pipeline itself is part of the
supply chain: third-party CI actions have been compromised in the past.

## Decision

- CI runs on **GitHub Actions**; images are published to **GHCR**.
- Each application has a pipeline triggered by **path filters** on its source
  directory:

  ```text
  test → build → scan (Trivy) → SBOM (Trivy) → push → update dev image tag
  ```

- **Trivy** scans images and fails the build on fixable critical
  vulnerabilities. The threshold is documented and adjustable.
- **Trivy** also generates an **SBOM** (CycloneDX), published with each build.
- Images are tagged with the commit SHA.
- Third-party actions are **pinned to full commit SHAs**, and every workflow
  declares minimal `permissions`.
- **Image signing with cosign** and signature verification at admission are
  deferred to the security phase, when a verifier exists
  (see [ADR-0010](0010-policy-enforcement-with-kyverno.md)).

## Consequences

- Positive: one tool covers both scanning and SBOM generation.
- Positive: pinned actions and minimal permissions reduce the impact of a
  compromised dependency or token.
- Negative: SHA-pinned actions must be updated deliberately (for example with
  Dependabot).
- Negative: until signing is introduced, the cluster trusts any image from the
  registry.

## Alternatives considered

- **Grype and Syft** — equally capable, but two tools instead of one.
- **Docker Scout** — tied to a Docker account and its usage limits.
- **Signing images immediately** — signatures provide no protection until
  something verifies them, so signing is introduced together with admission
  verification.
