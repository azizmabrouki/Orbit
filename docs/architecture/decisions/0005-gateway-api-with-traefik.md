# ADR-0005: Traffic management with Gateway API and Traefik

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Applications on Orbit need HTTP routing from outside the cluster. The
Kubernetes project retired ingress-nginx in March 2026 and recommends Gateway
API for new platforms. Gateway API also separates responsibilities: the
platform owns the Gateway, and application teams own their routes.

## Decision

- **Gateway API** is the routing interface.
- The platform owns a **shared Gateway**. The Golden Path chart generates an
  **HTTPRoute** for each application that enables routing in its manifest.
- **Traefik** is the Gateway API implementation. It is installed and
  configured through GitOps like every other platform component; the Traefik
  bundled with K3s is disabled.
- Traffic reaches the Gateway through the k3d load balancer mapped to host
  ports.
- Hostnames follow `<app>.<env>.localhost`, for example
  `orbit-demo.dev.localhost`.
- TLS is deferred until there is a need to demonstrate it.
- Only standard Gateway API resources are used; Traefik-specific custom
  resources are avoided so that the implementation remains replaceable.

## Consequences

- Positive: Orbit uses the current Kubernetes routing standard, with a clear
  split between platform-owned and application-owned configuration.
- Positive: Traefik is lightweight and its version and configuration are
  managed in Git.
- Negative: Gateway API CRDs must be installed and kept in step with the
  Traefik version.

## Alternatives considered

- **ingress-nginx** — retired; no further releases or security fixes.
- **Envoy Gateway** — a strong Gateway API implementation, but heavier and
  with more configuration surface than Orbit needs.
- **K3s-bundled Traefik** — works out of the box, but its version and
  configuration live outside GitOps and Gateway API support is not enabled by
  default.
