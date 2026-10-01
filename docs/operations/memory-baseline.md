# Memory baseline

Orbit has an explicit resource budget
([ADR-0001](../architecture/decisions/0001-local-first-zero-cost-runtime.md)):

- core platform: at most **1.5 GiB**;
- all optional capabilities enabled: at most **3 GiB**.

The budget targets an 8 GB laptop, including one running Docker Desktop, not
the development machine
([ADR-0012](../architecture/decisions/0012-native-linux-development-environment.md)).
Measurements are compared with the budget, not with the memory that happens to
be free on the machine that took them.

## Budget table

| Date | State | k3d containers (`docker stats`) | Node (`kubectl top nodes`) | Core budget | Share of core budget |
|---|---|---|---|---|---|
| 2026-10-01 | Empty cluster, idle | **494 MiB** (server 480 MiB + load balancer 14 MiB) | 490 MiB | 1536 MiB | 32 % |

Rows are added as components are installed.

## Method

1. Record host `free -m` before creating the cluster.
2. Create the cluster with `infrastructure/scripts/cluster-create.sh`.
3. Let it idle for 5 minutes.
4. Record `docker stats --no-stream` for every k3d container (server and load
   balancer), `kubectl top nodes` and `kubectl top pods -A`.

The budget figure is the sum of the k3d containers' memory usage from
`docker stats`, since that is what the cluster costs the host.

## 2026-10-01 — empty cluster, idle

**Environment**

- Machine: laptop, Intel i5-13420H (12 threads), 7.7 GB RAM (6.9 GiB usable),
  Ubuntu 26.04.1 LTS, kernel 7.0.0-38-generic, Docker Engine 29.8.2 (no VM).
- Tools: k3d v5.9.0, K3s v1.36.5+k3s1 (`rancher/k3s:v1.36.5-k3s1`),
  kubectl 1.35.9.
- Cluster: as defined in `infrastructure/k3d/orbit.yaml` (1 server, bundled
  Traefik disabled, nothing installed on top).
- Also running on the host: the desktop session, an IDE and other desktop
  applications; no other containers.

**Host before creating the cluster** (`free -m`, 22:29)

```
               total        used        free      shared  buff/cache   available
Mem:            7091        5203         222         601        2494        1887
Swap:           4095        2914        1181
```

**After 5 minutes idle** (22:40)

`docker stats --no-stream`:

| Container | Memory | CPU |
|---|---|---|
| k3d-orbit-server-0 | 480 MiB | 10.4 % |
| k3d-orbit-serverlb | 14.3 MiB | 0.0 % |
| **Total** | **494 MiB** | |

`kubectl top nodes`:

| Node | CPU | Memory |
|---|---|---|
| k3d-orbit-server-0 | 94m | 490 Mi |

`kubectl top pods -A`:

| Namespace | Pod | CPU | Memory |
|---|---|---|---|
| kube-system | coredns | 4m | 16 Mi |
| kube-system | local-path-provisioner | 1m | 15 Mi |
| kube-system | metrics-server | 11m | 27 Mi |

Host `free -m` afterwards:

```
               total        used        free      shared  buff/cache   available
Mem:            7091        5557         218         914        2522        1534
Swap:           4095        4095           0
```

**Reading**

- The empty cluster uses about **494 MiB, roughly a third of the 1.5 GiB core
  budget**, leaving about 1 GiB for the core platform components (Argo CD,
  Traefik, the secrets operator and the workloads).
- Almost all of it is the K3s server process itself (API server, controller
  manager, scheduler, kubelet, containerd and the embedded datastore). The
  three system pods together use under 60 Mi.
- Host memory is not representative: with the desktop applications open, the
  host was already using swap before the cluster existed and swap was full
  afterwards. This does not affect the comparison with the budget, but
  closing other applications before later measurements gives more stable
  numbers.
