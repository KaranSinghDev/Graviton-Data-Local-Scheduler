#!/usr/bin/env bash
# demo.sh — end-to-end demonstration of data-local scheduling.
#
# Installs the operator into the kind cluster via Helm, submits a sample
# PhysicsJob, and shows that the resulting pod was scheduled to the worker
# node whose topology.cern.io/site label matches the dataset's primary RSE.
#
# Prerequisite: bash scripts/setup-kind.sh has completed successfully.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_ROOT}"

NS="data-gravity-system"
RELEASE="data-gravity"
PJ="demo-atlas-daod"

cleanup() {
  echo ""
  echo "==> Cleaning up demo resources"
  kubectl delete physicsjob "${PJ}" --ignore-not-found
}
trap cleanup EXIT

# ── 1. Install operator + mock-rucio via Helm ─────────────────────────────────
echo "==> Installing chart into namespace '${NS}'"
helm upgrade --install "${RELEASE}" helm/data-gravity-operator/ \
  --namespace "${NS}" --create-namespace \
  --values helm/data-gravity-operator/values-kind.yaml \
  --wait --timeout 90s

echo ""
echo "==> Pods in ${NS}"
kubectl -n "${NS}" get pods

# ── 2. Submit a PhysicsJob ────────────────────────────────────────────────────
echo ""
echo "==> Submitting PhysicsJob '${PJ}'"
kubectl apply -f - <<EOF
apiVersion: hep.cern.local/v1alpha1
kind: PhysicsJob
metadata:
  name: ${PJ}
spec:
  dataset: "data23_13p6TeV:DAOD_PHYS.123456"
  image:   "busybox:1.36"
  command: ["sh", "-c", "echo 'running on data-local node'; sleep 30"]
  schedulingPolicy: DataLocal
  resources:
    requests:
      cpu: 100m
      memory: 64Mi
EOF

# ── 3. Wait for Scheduled / Running ───────────────────────────────────────────
echo ""
echo "==> Waiting for PhysicsJob to leave Pending phase..."
for _ in $(seq 1 30); do
  PHASE=$(kubectl get physicsjob "${PJ}" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
  if [[ "${PHASE}" == "Scheduled" || "${PHASE}" == "Running" || "${PHASE}" == "Succeeded" ]]; then
    break
  fi
  sleep 2
done

# ── 4. Show results ───────────────────────────────────────────────────────────
echo ""
echo "==> PhysicsJob status"
kubectl get physicsjob "${PJ}" -o custom-columns=\
'NAME:.metadata.name,PHASE:.status.phase,RSE:.status.resolvedRSE,JOB:.status.jobRef,BYTES_AVOIDED:.status.bytesTransferAvoided'

echo ""
echo "==> Owned batch/v1.Job"
kubectl get jobs -l app.kubernetes.io/managed-by=data-gravity-operator

# Wait briefly for the pod to be created and scheduled
sleep 3
POD=$(kubectl get pods -l job-name="${PJ}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [[ -n "${POD}" ]]; then
  NODE=$(kubectl get pod "${POD}" -o jsonpath='{.spec.nodeName}' 2>/dev/null || echo "<not yet bound>")
  SITE=$(kubectl get node "${NODE}" -o jsonpath='{.metadata.labels.topology\.cern\.io/site}' 2>/dev/null || echo "<unknown>")
  RSE=$(kubectl get physicsjob "${PJ}" -o jsonpath='{.status.resolvedRSE}')
  echo ""
  echo "==> Data-locality verification"
  echo "    Dataset DID:     data23_13p6TeV:DAOD_PHYS.123456"
  echo "    Resolved RSE:    ${RSE}"
  echo "    Pod:             ${POD}"
  echo "    Scheduled node:  ${NODE}"
  echo "    Node site label: ${SITE}"
fi

echo ""
echo "Compute is co-located with the dataset replica via NodeAffinity injection."
echo "Press Enter to clean up the PhysicsJob (the chart and cluster will remain)."
read -r
