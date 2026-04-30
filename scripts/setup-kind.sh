#!/usr/bin/env bash
# setup-kind.sh — spin up a 4-worker kind cluster wired for data-gravity-operator.
# Run after: bash scripts/setup-env.sh && source scripts/env.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_TAG="data-gravity-operator:dev"
CLUSTER_NAME="data-gravity"

cd "${PROJECT_ROOT}"

echo "==> Creating kind cluster '${CLUSTER_NAME}'"
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "    cluster already exists, skipping creation"
else
  kind create cluster --config deploy/kind-config.yaml
fi

echo "==> Building operator image (${IMAGE_TAG})"
docker build -t "${IMAGE_TAG}" .

echo "==> Loading image into kind"
kind load docker-image "${IMAGE_TAG}" --name "${CLUSTER_NAME}"

echo "==> Installing CRD"
kubectl apply -f config/crd/bases/hep.cern.local_physicsjobs.yaml

echo "==> Deploying mock-rucio"
kubectl apply -f deploy/mock-rucio.yaml
kubectl -n data-gravity-system rollout status deployment/mock-rucio --timeout=60s

echo ""
echo "Cluster is ready. Node labels:"
kubectl get nodes -L topology.cern.io/site
echo ""
echo "Run 'bash scripts/demo.sh' to submit a sample PhysicsJob."
