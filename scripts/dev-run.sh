#!/usr/bin/env bash
# dev-run.sh — Run the operator manager on your host, talking to a kind cluster.
#
# Useful for fast iteration: edit code → `go build` → restart the manager,
# without rebuilding/reloading the Docker image.
#
# Prerequisites:
#   1. bash scripts/setup-kind.sh         (kind cluster + image loaded)
#   2. source scripts/env.sh              (PATH includes ./tools/bin)
#
# This script will:
#   - install the CRD into the cluster (if missing)
#   - deploy mock-rucio (if missing)
#   - port-forward mock-rucio to localhost
#   - build the manager binary into ./bin/manager
#   - run the manager against the kind cluster's kubeconfig
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

NS="data-gravity-system"
LOCAL_RUCIO_PORT="${LOCAL_RUCIO_PORT:-18080}"

# ── 1. Ensure CRD is installed ────────────────────────────────────────────────
if ! kubectl get crd physicsjobs.hep.cern.local >/dev/null 2>&1; then
  echo "==> Installing CRD"
  kubectl apply -f config/crd/bases/hep.cern.local_physicsjobs.yaml
fi

# ── 2. Ensure mock-rucio is deployed ──────────────────────────────────────────
if ! kubectl -n "${NS}" get deploy mock-rucio >/dev/null 2>&1; then
  echo "==> Deploying mock-rucio into ${NS}"
  kubectl apply -f deploy/mock-rucio.yaml
fi
kubectl -n "${NS}" rollout status deploy/mock-rucio --timeout=60s

# ── 3. Build the manager binary on the host ───────────────────────────────────
echo "==> Building ./bin/manager"
mkdir -p bin
go build -o bin/manager ./cmd/main.go

# ── 4. Port-forward mock-rucio to the host ────────────────────────────────────
echo "==> Port-forwarding mock-rucio  localhost:${LOCAL_RUCIO_PORT} → ${NS}/mock-rucio:8080"
kubectl -n "${NS}" port-forward svc/mock-rucio "${LOCAL_RUCIO_PORT}:8080" >/dev/null 2>&1 &
PF_PID=$!
trap "kill ${PF_PID} 2>/dev/null || true" EXIT
sleep 2

# Sanity check: hit the mock once to confirm the tunnel is up
if ! curl -sf "http://localhost:${LOCAL_RUCIO_PORT}/dids/data23_13p6TeV/DAOD_PHYS.123456/replicas" >/dev/null; then
  echo "ERROR: cannot reach mock-rucio via localhost:${LOCAL_RUCIO_PORT}"
  exit 1
fi
echo "    mock-rucio reachable"

# ── 5. Run the manager ────────────────────────────────────────────────────────
echo ""
echo "==> Starting manager (Ctrl-C to stop)"
exec ./bin/manager \
  --rucio-url="http://localhost:${LOCAL_RUCIO_PORT}" \
  --metrics-bind-address=0 \
  --health-probe-bind-address=:8081 \
  --leader-elect=false
