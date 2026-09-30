# ADR-0006: Secrets management with SOPS and age

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

With GitOps, all desired state lives in Git, including secrets, which must
therefore be encrypted before they are committed. The platform itself needs
secrets (for example the portal's GitHub and Argo CD credentials), and so do
applications.

Orbit is a developer platform, so encrypting a secret should be a developer
workflow that does not depend on access to a running cluster.

Argo CD's documentation advises against decrypting or injecting secrets while
manifests are generated, because Argo CD caches generated manifests and shows
them in diffs. It recommends that secrets be materialised inside the
destination cluster instead.

## Decision

- **SOPS with age** is Orbit's secret encryption standard.
- Developers encrypt secrets locally with the `sops` CLI. A `.sops.yaml` file
  in the repository defines the creation rules and the **public** age
  recipient.
- **Decryption happens inside the cluster**, performed by a SOPS operator that
  turns an encrypted custom resource into a Kubernetes Secret. Argo CD only
  ever handles ciphertext.
- The primary candidate is **sops-secrets-operator**; the choice is validated
  when GitOps is implemented. The fallback is the KSOPS plugin for Argo CD,
  with the caching trade-off explicitly documented.
- The **age private key is never committed**. It is kept outside the
  repository by the platform operator and injected into the cluster by the
  bootstrap script.

## Consequences

- Positive: encryption does not depend on a running cluster; any file can be
  encrypted, not only Kubernetes Secrets.
- Positive: the key is independent of the cluster lifecycle, so a recreated
  cluster decrypts existing secrets as soon as the key is injected.
- Positive: consistent with Argo CD's recommended pattern of in-cluster
  secret materialisation.
- Negative: adds an operator to the core platform.
- Negative: the age private key is a single root secret. Losing it means
  re-encrypting every secret, so it must be backed up.
- Negative: rotating the key requires re-encrypting all secrets.

## Alternatives considered

- **Sealed Secrets** — simple and well suited to Argo CD, but the encryption
  key is created and held by the in-cluster controller, which ties secrets to
  the cluster's lifecycle, and it only handles Kubernetes Secrets.
- **External Secrets Operator** — requires an external secret store; running
  one locally (for example OpenBao) exceeds the resource budget.
- **KSOPS or helm-secrets in the Argo CD repo server** — decrypts during
  manifest generation, so plaintext enters Argo CD's cache.
