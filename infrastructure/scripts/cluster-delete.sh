#!/bin/sh
# Delete Orbit's local k3d cluster and its containers.
# Exits 0 if the cluster does not exist.
set -eu

CLUSTER=orbit

if ! command -v k3d >/dev/null 2>&1; then
	echo "k3d not found. See docs/deployment/local-setup.md." >&2
	exit 1
fi

# Without the daemon, "k3d cluster get" fails as if the cluster did not exist.
if ! docker version >/dev/null 2>&1; then
	echo "Docker daemon not reachable; cannot check for cluster '$CLUSTER'." >&2
	exit 1
fi

if ! k3d cluster get "$CLUSTER" >/dev/null 2>&1; then
	echo "Cluster '$CLUSTER' does not exist; nothing to do."
	exit 0
fi

echo "Deleting cluster '$CLUSTER'..."
k3d cluster delete "$CLUSTER"
echo "Cluster '$CLUSTER' deleted."
