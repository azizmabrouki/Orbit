# ADR-0010: Policy enforcement with Kyverno

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

The Golden Path makes the safe configuration the default, but it does not
prevent workloads from bypassing it. The platform needs guardrails that
reject unsafe workloads regardless of how they were produced, and that can be
tested before changes reach the cluster.

Policy enforcement is only meaningful once the delivery path exists, so it is
introduced in the security phase rather than at the start.

## Decision

- **Kyverno** is the policy engine.
- It runs in a **minimal configuration**, focused on the admission
  controller, with auxiliary controllers reduced or disabled to fit the
  resource budget.
- Initial policies:
  - disallow the `latest` image tag;
  - require resource requests and limits;
  - require liveness and readiness probes;
  - disallow privileged containers and require a non-root user;
  - allow images only from the Orbit registry.
- Later: verify cosign image signatures at admission
  (see [ADR-0008](0008-ci-and-supply-chain-security.md)).
- Policies start in **Audit** mode and move to **Enforce** once the Golden
  Path chart is compliant.
- The same policies run in **CI with the Kyverno CLI** against rendered
  manifests, so violations are caught before they reach the cluster.

## Consequences

- Positive: policies are written in YAML and can be tested in CI.
- Positive: supports image signature verification, which completes the
  supply-chain story.
- Negative: adds an admission webhook; if Kyverno is unavailable, admission
  behaviour depends on its failure policy, which must be chosen deliberately.

## Alternatives considered

- **ValidatingAdmissionPolicy** (built into Kubernetes, CEL) — no additional
  component, but no image signature verification and no equivalent CLI
  workflow for testing policies against manifests in CI.
- **OPA Gatekeeper** — powerful, but Rego adds a language to learn and
  maintain with no benefit for Orbit's policies.
