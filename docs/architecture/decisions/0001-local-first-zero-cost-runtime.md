# ADR-0001: Local-first, zero-cost platform runtime

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Orbit must be reproducible by anyone who clones the repository. It must run
at no cost, without a cloud account, a paid service or a registered domain,
on a typical developer laptop (8 GB of RAM, Docker Desktop with the WSL2
backend, or Docker on Linux/macOS).

Memory is the binding constraint: on such a machine the container runtime
typically has 3.5–4 GiB available, shared by the cluster and every platform
component.

## Decision

- Orbit runs on a **single-node K3s cluster created with k3d**.
- The cluster is **disposable**: it can be deleted and recreated from the
  repository plus one out-of-band key (see [ADR-0006](0006-secrets-with-sops-and-age.md)).
- K3s built-ins are used where they serve the platform: **ServiceLB** and the
  **embedded NetworkPolicy controller**. The bundled Traefik is disabled so
  that ingress is managed through GitOps (see [ADR-0005](0005-gateway-api-with-traefik.md)).
- The source repository is **public on GitHub**. CI runs on **GitHub Actions**
  and images are stored in **GitHub Container Registry (GHCR)**, both free for
  public repositories and packages.
- Workloads are reached through hostnames under **`*.localhost`**, which
  browsers resolve to the loopback address without DNS configuration.
- Automation is written as **POSIX shell scripts** that run the same way
  locally (Git Bash, WSL, Linux, macOS) and in CI.
- Orbit has an explicit **resource budget**:
  - core platform: at most **1.5 GiB** of memory;
  - all optional capabilities enabled: at most **3 GiB**.

  The budget is measured and documented as capabilities are added.
- Capabilities that are expensive to run (observability, policy enforcement)
  are **optional** and can be enabled or disabled independently.
- **No infrastructure-as-code tool** (such as Terraform) is used in this
  version, because there is no cloud infrastructure to manage.

## Consequences

- Positive: zero cost; anyone can run the full platform; the cluster can be
  recreated in minutes, which makes failure and recovery scenarios cheap.
- Positive: the resource budget forces deliberate component choices and
  trimmed configurations.
- Negative: a single node cannot demonstrate high availability, node failure
  or multi-cluster topologies.
- Negative: K3s bundles components that differ from upstream Kubernetes
  defaults. If upstream parity becomes important for testing, kind can be
  used in CI without changing the local runtime.
- Negative: some non-browser tools may not resolve `*.localhost` and need an
  explicit hosts entry or resolve override.

## Alternatives considered

- **kind** — closer to upstream Kubernetes and common in CI, but no built-in
  load balancer and no NetworkPolicy enforcement with its default setup.
  Kept as an option for CI.
- **minikube** — VM-based by default and heavier for the same result.
- **Managed cloud Kubernetes** — realistic, but incurs cost.
- **Cloud free tiers** — some offer enough capacity for a small cluster, but
  require payment-card verification and carry account and capacity risk.
  A possible future target, and the point at which infrastructure-as-code
  would become justified.
