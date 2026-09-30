# ADR-0009: Observability

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Applications deployed through the Golden Path should be observable by
default, without teams configuring monitoring themselves. Observability
stacks are also the most memory-intensive part of a platform, and Orbit has a
strict resource budget (see [ADR-0001](0001-local-first-zero-cost-runtime.md)).

## Decision

- **Metrics:** Prometheus and Grafana, installed with kube-prometheus-stack in
  a trimmed configuration: short retention and Alertmanager disabled until
  alerting is demonstrated.
- **Logs:** Loki in single-binary mode, with **Grafana Alloy** as the
  collector.
- The Golden Path chart configures metrics scraping for every application
  that exposes `/metrics`, so applications are observable without additional
  configuration.
- Metrics are introduced first and logs second, each only when it supports a
  platform capability being demonstrated.
- Observability is an **optional capability** that can be disabled to save
  resources.
- **Tracing is deferred.** OpenTelemetry and Tempo are added only if a workload
  scenario requires distributed tracing.
- The portal links to, or displays, each application's signals
  (see [ADR-0007](0007-developer-portal.md)).

## Consequences

- Positive: widely used, industry-standard tools with a single UI (Grafana)
  for metrics and logs.
- Positive: applications get monitoring by default through the Golden Path.
- Negative: even trimmed, this is the largest optional capability in memory
  terms, and running it continuously on an 8 GB machine is not expected.

## Alternatives considered

- **VictoriaMetrics** — lighter than Prometheus, but less standard; Prometheus
  compatibility matters more than the memory saving.
- **Promtail** — reached end of life in March 2026; Alloy is its successor.
- **Elasticsearch or OpenSearch** for logs — far exceeds the resource budget.
