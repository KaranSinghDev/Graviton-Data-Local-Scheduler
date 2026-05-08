#!/usr/bin/env bash
# setup-kind.sh — bootstrap a kind cluster and load the operator image.
#
# After this completes, run one of:
#   bash scripts/demo.sh       # Helm-based in-cluster demo (recommended)
#   bash scripts/dev-run.sh    # Out-of-cluster manager for fast development
#
# Prerequisites: bash scripts/setup-env.sh && source scripts/env.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_TAG="data-gravity-operator:dev"
CLUSTER_NAME="data-gravity"

cd "${PROJECT_ROOT}"

echo "==> Creating kind cluster '${CLUSTER_NAME}'"
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "    cluster already exists, skipping"
else
  kind create cluster --config deploy/kind-config.yaml
fi

echo "==> Building operator image (${IMAGE_TAG})"
docker build -t "${IMAGE_TAG}" .

echo "==> Loading image into kind"
kind load docker-image "${IMAGE_TAG}" --name "${CLUSTER_NAME}"

echo ""
echo "Cluster is ready. Worker node site labels:"
kubectl get nodes -L topology.cern.io/site
echo ""
echo "Next steps:"
echo "  bash scripts/demo.sh       # full in-cluster demo via Helm"
echo "  bash scripts/dev-run.sh    # run operator on host for fast iteration"
