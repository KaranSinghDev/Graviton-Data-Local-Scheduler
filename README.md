# data-gravity-operator

A Kubernetes Operator that implements **data-locality-aware workload scheduling**
for distributed physics data lakes modelled on the WLCG/Rucio storage topology.

When a physicist submits a `PhysicsJob` referencing a Rucio dataset
(`scope:name` format), the operator:

1. Resolves which RSE (Rucio Storage Element) holds the primary replica
2. Maps that RSE to a Kubernetes node via the `topology.cern.io/site` label
3. Creates an owned `batch/v1.Job` with `NodeAffinity` constraints injected —
   so compute runs *co-located with the data*, avoiding WAN transfers entirely

Standard `kube-scheduler` has no storage-topology awareness. This operator
provides that injection automatically, closing the gap between data placement
(Rucio) and compute placement (Kubernetes).

---

## Quickstart — local kind cluster

```bash
# 1. Install local toolchain (Go 1.24, kubebuilder, kind, kubectl, helm)
bash scripts/setup-env.sh
source scripts/env.sh

# 2. Spin up a 4-node kind cluster with WLCG site labels + deploy mock-rucio
bash scripts/setup-kind.sh

# 3. Run the end-to-end demo
bash scripts/demo.sh
```

The kind cluster comes with four worker nodes pre-labelled:

| Node | Label |
|------|-------|
| worker-0 | `topology.cern.io/site=cern-prod` |
| worker-1 | `topology.cern.io/site=bnl-osg2` |
| worker-2 | `topology.cern.io/site=in2p3-cc` |
| worker-3 | `topology.cern.io/site=triumf-lcg2` |

---

## Custom Resource: PhysicsJob

```yaml
apiVersion: hep.cern.local/v1alpha1
kind: PhysicsJob
metadata:
  name: atlas-daod-sample
spec:
  # Rucio DID — scope:name format
  dataset: "data23_13p6TeV:DAOD_PHYS.123456"
  # Container image for the compute workload
  image: "gitlab-registry.cern.ch/atlas/athena:24.0.12"
  command: ["Reco_tf.py", "--inputAODFile", "/data/input.AOD.pool.root"]
  # DataLocal | ClosestSite | AnyAvailable
  schedulingPolicy: DataLocal
  resources:
    requests:
      cpu: "2"
      memory: "4Gi"
```

Inspect the job as it progresses:

```bash
kubectl get pj   # shortName
```

```
NAME               DATASET                                  PHASE       RSE                   NODE
atlas-daod-sample  data23_13p6TeV:DAOD_PHYS.123456         Scheduled   CERN-PROD_DATADISK    worker-0
```

---

## Scheduling policies

| Policy | Behaviour |
|--------|-----------|
| `DataLocal` (default) | Hard-pins compute to the node whose `topology.cern.io/site` matches the primary RSE |
| `ClosestSite` | Same as DataLocal; extension point for geo-distance ranking across replicas |
| `AnyAvailable` | No affinity injected; scheduler places freely; RSE still recorded for observability |

---

## Development

```bash
# Regenerate deepcopy + CRD YAML after editing types
make generate manifests

# Run unit tests + envtest controller suite
make test

# Build Docker image (contains both manager and mock-rucio binaries)
make docker-build IMG=ghcr.io/karansinghdev/data-gravity-operator:dev

# Deploy via Helm (production)
helm install data-gravity helm/data-gravity-operator/ \
  --namespace data-gravity-system --create-namespace \
  --set rucioURL=https://rucio.cern.ch \
  --set mockRucio.enabled=false
```

---

## Prometheus metrics

| Metric | Type | Labels |
|--------|------|--------|
| `physjob_reconcile_total` | Counter | `result` |
| `physjob_reconcile_duration_seconds` | Histogram | — |
| `physjob_resolved_total` | Counter | `rse`, `policy` |
| `physjob_resolution_failures_total` | Counter | `reason` |
| `physjob_data_transfer_avoided_bytes` | Counter | `rse` |

The `physjob_data_transfer_avoided_bytes` counter accumulates estimated bytes
of WAN transfer eliminated by data-local scheduling. For a typical ATLAS DAOD
dataset (~2.5 TB), a single data-local job avoids 2.5 TB of inter-site traffic.

---

## Architecture

See [`docs/architecture.md`](docs/architecture.md) for the full component diagram,
reconcile loop pseudocode, and data-flow explanation.

---

## Repository layout

```
api/v1alpha1/               CRD types (PhysicsJobSpec, PhysicsJobStatus, Phase enum)
internal/controller/        Reconciler + Ginkgo/envtest suite (55 % coverage)
internal/storage/           StorageTopologyClient interface + Rucio HTTP client
internal/scheduling/        NodeAffinity builder (100 % coverage)
internal/metrics/           Prometheus registrations
internal/mockrucio/         Mock Rucio API — 9 ATLAS/CMS/LHCb datasets (80 % coverage)
cmd/main.go                 Manager entrypoint (--rucio-url flag)
cmd/mock-rucio/main.go      Standalone mock-rucio server
config/crd/bases/           Generated CRD YAML
config/rbac/                Generated ClusterRole
config/samples/             Example PhysicsJob CRs
deploy/                     kind cluster config + mock-rucio Kubernetes manifest
helm/data-gravity-operator/ Helm chart (CRD in crds/, RBAC, Deployment, optional mock)
scripts/                    setup-env.sh  setup-kind.sh  demo.sh
docs/                       Architecture doc + Mermaid diagram
```

---

## Tech stack

| Component | Version |
|-----------|---------|
| Go | 1.24 |
| controller-runtime | v0.21 |
| Kubernetes API | v0.33 (1.33) |
| Ginkgo / Gomega | v2 / v1 |
| Prometheus client | v1.22 |
| kubebuilder scaffold | v4.6 |
| kind | v0.26 |
| Helm | 3.17 |
