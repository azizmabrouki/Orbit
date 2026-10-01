#!/bin/sh
# Create Orbit's local k3d cluster from infrastructure/k3d/orbit.yaml and wait
# until it is ready. Installs nothing into the cluster, so a bootstrap script
# can call it first. Exits 0 if the cluster already exists.
set -eu

CLUSTER=orbit
TIMEOUT=${ORBIT_CREATE_TIMEOUT:-180s}

script_dir=$(cd "$(dirname "$0")" && pwd)
config="$script_dir/../k3d/orbit.yaml"

"$script_dir/check-prereqs.sh"

if k3d cluster get "$CLUSTER" >/dev/null 2>&1; then
	echo "Cluster '$CLUSTER' already exists; nothing to do."
	exit 0
fi

echo "Creating cluster '$CLUSTER'..."
k3d cluster create --config "$config" --wait --timeout "$TIMEOUT"

context="k3d-$CLUSTER"

echo "Waiting for the node to be Ready..."
kubectl --context "$context" wait --for=condition=Ready node --all --timeout="$TIMEOUT"

# K3s applies each add-on as a separate manifest, a few seconds after the
# node is Ready, so wait for every expected deployment by name.
# Not "wait --for=condition=Available": K3s's deployments allow one
# unavailable replica, so they report Available before any pod is ready.
# "rollout status" waits until every replica is updated and available.
echo "Waiting for the kube-system deployments to be available..."
for deployment in coredns local-path-provisioner metrics-server; do
	kubectl --context "$context" -n kube-system wait --for=create "deployment/$deployment" --timeout="$TIMEOUT"
	kubectl --context "$context" -n kube-system rollout status "deployment/$deployment" --timeout="$TIMEOUT"
done

echo "Cluster '$CLUSTER' is ready (kubectl context: $context)."
