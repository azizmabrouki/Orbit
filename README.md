# Orbit

> A production-inspired Internal Developer Platform for self-service
> application delivery.

## Overview

Orbit aims to provide developers with a standardized Golden Path
for building and deploying applications while abstracting away
infrastructure and operational complexity.

Developers describe an application in a short `orbit.yaml` manifest.
Orbit builds, scans, deploys, routes and observes it, using Git as the
single source of truth. It runs entirely on a laptop, at no cost.

## Status

🚧 Architecture defined — implementation in progress.

## Architecture

See the [architecture overview](docs/architecture/overview.md) and the
[architecture decision records](docs/architecture/decisions/README.md).

## Technology

| Capability | Technology |
|---|---|
| Local Kubernetes | k3d (K3s) |
| CI / registry | GitHub Actions, GHCR |
| Packaging | Helm — Orbit Golden Path chart |
| GitOps | Argo CD + ApplicationSets |
| Routing | Gateway API + Traefik |
| Secrets | SOPS + age |
| Developer portal | React, TypeScript, Vite, Go |
| Supply-chain security | Trivy (scanning, SBOM), cosign |
| Observability | Prometheus, Grafana, Loki, Alloy |
| Policy | Kyverno |

## Documentation

- [Architecture](docs/architecture/overview.md)
- Deployment
- Operations
