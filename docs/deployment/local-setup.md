# Local setup

Orbit runs on a disposable, single-node [K3s](https://k3s.io) cluster created
with [k3d](https://k3d.io) (see
[ADR-0001](../architecture/decisions/0001-local-first-zero-cost-runtime.md)).
This page covers the host tools, and how to create and delete the cluster.

## Prerequisites

| Tool | Minimum | Tested with | Purpose |
|---|---|---|---|
| Docker Engine | daemon reachable | 29.8.2 | Runs the k3d node containers |
| k3d | 5.9.0 | v5.9.0 | Creates the K3s cluster in Docker |
| kubectl | 1.35 | 1.35.9 | Talks to the cluster (within one minor version of K3s 1.36) |
| Go | 1.27 | 1.27.1 | Builds Orbit's Go services (later milestones) |

The K3s version is pinned in the cluster config: `rancher/k3s:v1.36.5-k3s1`
(K3s `v1.36.5+k3s1`, stable channel). k3d pulls it on first create.

Orbit is developed on Ubuntu with Docker Engine
([ADR-0012](../architecture/decisions/0012-native-linux-development-environment.md)).
Your user must be in the `docker` group so the scripts can run without `sudo`.

The repository checks the tools; it never installs them. To check them:

```sh
infrastructure/scripts/check-prereqs.sh
```

It prints each tool's version with `[ok]` or `[FAIL]` and exits non-zero,
naming every missing or too-old tool.

## Installing k3d and Go on Ubuntu

Install pinned versions from the official releases, not from apt or snap.
The steps below install system-wide into `/usr/local`. To install without
`sudo`, use `~/.local/bin` and `~/.local/go` instead (and make sure
`~/.local/bin` is on your `PATH`).

### k3d v5.9.0

From the GitHub release, verified against the published checksums:

```sh
cd "$(mktemp -d)"
curl -fsSLO https://github.com/k3d-io/k3d/releases/download/v5.9.0/k3d-linux-amd64
curl -fsSLO https://github.com/k3d-io/k3d/releases/download/v5.9.0/checksums.txt
# Paths in checksums.txt start with "_dist/".
grep '_dist/k3d-linux-amd64$' checksums.txt | sed 's|_dist/||' | sha256sum -c -
sudo install -m 0755 k3d-linux-amd64 /usr/local/bin/k3d
k3d version
```

### Go 1.27.1

From the go.dev tarball, verified against the SHA256 published on
<https://go.dev/dl/>:

```sh
cd "$(mktemp -d)"
curl -fsSLO https://go.dev/dl/go1.27.1.linux-amd64.tar.gz
echo '63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445  go1.27.1.linux-amd64.tar.gz' | sha256sum -c -
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.27.1.linux-amd64.tar.gz
# Add to your shell profile:
export PATH="$PATH:/usr/local/go/bin"
go version
```

## Creating and deleting the cluster

All scripts are POSIX `sh` and work from any directory.

```sh
infrastructure/scripts/cluster-create.sh   # check prerequisites, create, wait until ready
infrastructure/scripts/cluster-delete.sh   # delete the cluster and its containers
```

- `cluster-create.sh` runs the prerequisite check, creates the cluster from
  `infrastructure/k3d/orbit.yaml`, and waits until the node is Ready and every
  `kube-system` deployment has rolled out. If the cluster already exists, it
  says so and exits 0. Set `ORBIT_CREATE_TIMEOUT` (default `180s`) to change
  the wait timeout.
- `cluster-delete.sh` deletes the cluster. If it does not exist, it says so
  and exits 0.

The first create pulls the K3s and k3d images (about 1.5 minutes); later
creates take under a minute. The kubeconfig context is `k3d-orbit` and
becomes the current context.

`cluster-create.sh` installs nothing into the cluster, so that a later
bootstrap script can call it before installing Argo CD
([ADR-0003](../architecture/decisions/0003-gitops-delivery-with-argo-cd.md)).

## What the cluster contains

The whole shape is in [`infrastructure/k3d/orbit.yaml`](../../infrastructure/k3d/orbit.yaml):

| Item | Setting | Why |
|---|---|---|
| Nodes | 1 server, 0 agents | Memory budget; single node is enough locally ([ADR-0001](../architecture/decisions/0001-local-first-zero-cost-runtime.md)) |
| K3s image | `rancher/k3s:v1.36.5-k3s1` | Pinned for reproducibility |
| Host port 80 | Mapped to the k3d load balancer | Traffic enters through the load balancer on host ports ([ADR-0005](../architecture/decisions/0005-gateway-api-with-traefik.md)) |
| Bundled Traefik | Disabled | Traefik is installed and configured through GitOps instead ([ADR-0005](../architecture/decisions/0005-gateway-api-with-traefik.md)) |
| ServiceLB | On (K3s default) | Gives `LoadBalancer` services an address behind the k3d load balancer |
| NetworkPolicy controller | On (K3s default) | NetworkPolicies are enforced |
| CoreDNS, local-path storage, metrics-server | On (K3s default) | Cluster DNS, a default `StorageClass`, `kubectl top` |

The cluster is disposable: delete and recreate it at any time.

Host port 80 must be free when the cluster is created.

## Hostnames under `*.localhost`

Workloads will be reached at `<app>.<env>.localhost`, for example
`http://orbit-demo.dev.localhost`. Browsers resolve every `*.localhost` name
to the loopback address without any DNS setup. Some non-browser tools do not;
for those, add a hosts entry or pass the address explicitly, for example:

```sh
echo '127.0.0.1 orbit-demo.dev.localhost' | sudo tee -a /etc/hosts
# or, without changing /etc/hosts:
curl --resolve orbit-demo.dev.localhost:80:127.0.0.1 http://orbit-demo.dev.localhost/
```
