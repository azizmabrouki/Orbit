# ADR-0012: Native Linux development environment

- **Status:** Accepted
- **Date:** 2026-09-30

## Context

Orbit was first set up on Windows 11, where Docker Desktop runs containers in
a WSL2 virtual machine. On an 8 GB laptop that virtual machine is capped at
about 3.7 GiB, which is the whole budget for the cluster and every platform
component. CI runs on Ubuntu, and shell scripts run from Windows are exposed
to path and line-ending differences that CI never sees.

The first plan was to run the Milestone 1 tooling from Git Bash on Windows.
It was replaced the same day, before anything was built on it.

## Decision

- Orbit is developed **on native Ubuntu LTS**, installed as a dual boot next
  to Windows.
- Containers run on **Docker Engine**, not Docker Desktop.
- The resource budget in [ADR-0001](0001-local-first-zero-cost-runtime.md) is
  **unchanged**. It targets any 8 GB laptop, including one running Docker
  Desktop, not the development machine. Measurements are compared with the
  budget, not with the memory that happens to be free on the development
  machine.
- The architecture is unaffected: automation was already POSIX shell
  (see [ADR-0001](0001-local-first-zero-cost-runtime.md)).

## Consequences

- Positive: containers get roughly 5.5–6 GB instead of about 3.7 GiB (an
  estimate, to be measured), which leaves headroom for development tools.
- Positive: local scripts run in the same environment as CI.
- Negative: switching between Orbit and Windows requires a reboot.
- Negative: anything not committed to Git has to be moved between the two
  systems by hand.

## Alternatives considered

- **Git Bash on Windows** — nothing to install, but keeps the virtual
  machine's memory cap and the Windows script differences.
- **An Ubuntu WSL distribution** — closer to CI, but still inside the same
  virtual machine memory limit.
- **Docker Desktop on Linux** — also runs a virtual machine, which loses the
  memory gain.
