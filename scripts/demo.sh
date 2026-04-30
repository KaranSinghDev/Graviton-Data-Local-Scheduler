#!/usr/bin/env bash
# demo.sh — end-to-end demonstration of data-local scheduling.
# Prerequisite: bash scripts/setup-kind.sh has completed successfully.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

NAMESPACE="default"
OPERATOR_NS="data-gravity-system"
PJ_NAME="demo-atlas-daod"

cleanup() {
  echo ""
  echo "==> Cleaning up"
  kubectl delete physicsjob "${PJ_NAME}" -n "${NAMESPACE}" --ignore-not-found
}
trap cleanup EXIT

# ── 1. Start the operator in the background ───────────────────────────────────
echo "==> Starting operator (ctrl-C to stop)"
source scripts/env.sh
RUCIO_URL="http://$(kubectl -n ${OPERATOR_NS} get svc mock-rucio \
  -o jsonpath='{.spec.clusterIP}'):8080"
echo "    Rucio URL: ${RUCIO_URL}"

./bin/manager \
  --rucio-url="${RUCIO_URL}" \
  --metrics-bind-address=:8082 \
  --health-probe-bind-address=:8083 \
  --leader-elect=false &
OPERATOR_PID=$!
trap "kill ${OPERATOR_PID} 2>/dev/null; cleanup" EXIT
sleep 3

# ── 2. Submit a PhysicsJob ────────────────────────────────────────────────────
echo ""
echo "==> Submitting PhysicsJob '${PJ_NAME}'"
kubectl apply -f - <<EOF
apiVersion: hep.cern.local/v1alpha1
kind: PhysicsJob
metadata:
  name: ${PJ_NAME}
  namespace: ${NAMESPACE}
spec:
  dataset: "data23_13p6TeV:DAOD_PHYS.123456"
  image:   "busybox:1.36"
  command: ["sh", "-c", "echo running on data-local node; sleep 30"]
  schedulingPolicy: DataLocal
  resources:
    requests:
      cpu: 100m
      memory: 64Mi
EOF

# ── 3. Watch until Scheduled ──────────────────────────────────────────────────
echo ""
echo "==> Waiting for PhysicsJob to reach Scheduled phase..."
for i in $(seq 1 30); do
  PHASE=$(kubectl get physicsjob "${PJ_NAME}" -n "${NAMESPACE}" \
    -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
  if [[ "${PHASE}" == "Scheduled" || "${PHASE}" == "Running" ]]; then
    break
  fi
  sleep 2
done

# ── 4. Show results ───────────────────────────────────────────────────────────
echo ""
echo "==> PhysicsJob status"
kubectl get physicsjob "${PJ_NAME}" -n "${NAMESPACE}" \
  -o custom-columns=\
'NAME:.metadata.name,PHASE:.status.phase,RSE:.status.resolvedRSE,NODE:.status.scheduledNode,BYTES_AVOIDED:.status.bytesTransferAvoided'

echo ""
echo "==> Owned batch/v1.Job"
kubectl get jobs -n "${NAMESPACE}" -l app.kubernetes.io/managed-by=data-gravity-operator

RSE=$(kubectl get physicsjob "${PJ_NAME}" -n "${NAMESPACE}" \
  -o jsonpath='{.status.resolvedRSE}' 2>/dev/null || echo "")
echo ""
echo "==> NodeAffinity on the Job pod template"
kubectl get job "${PJ_NAME}" -n "${NAMESPACE}" \
  -o jsonpath='{.spec.template.spec.affinity}' | python3 -m json.tool 2>/dev/null || true

echo ""
echo "Dataset replica resolved to RSE: ${RSE}"
echo "Compute Job is pinned to the matching WLCG site node via NodeAffinity."
echo ""
echo "Press Ctrl-C to stop the operator and clean up."
wait ${OPERATOR_PID}
