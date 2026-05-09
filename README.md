# data-gravity-operator

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Tests](https://github.com/KaranSinghDev/data-gravity-operator/actions/workflows/test.yml/badge.svg)](https://github.com/KaranSinghDev/data-gravity-operator/actions/workflows/test.yml)
[![Lint](https://github.com/KaranSinghDev/data-gravity-operator/actions/workflows/lint.yml/badge.svg)](https://github.com/KaranSinghDev/data-gravity-operator/actions/workflows/lint.yml)
[![Container](https://img.shields.io/badge/container-ghcr.io-blue.svg)](https://github.com/KaranSinghDev/data-gravity-operator/pkgs/container/data-gravity-operator)
[![Helm Chart](https://img.shields.io/badge/helm-chart-0F1689.svg?logo=helm)](https://karansinghdev.github.io/data-gravity-operator)

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

## Installation

### Production — Helm chart

```bash
helm repo add data-gravity https://karansinghdev.github.io/data-gravity-operator
helm repo update

helm install data-gravity data-gravity/data-gravity-operator \
  --namespace data-gravity-system --create-namespace \
  --set rucioURL=https://rucio.cern.ch
```

The container image is hosted at
`ghcr.io/karansinghdev/data-gravity-operator` and supports `linux/amd64`
and `linux/arm64`.

### Local quickstart — kind cluster

```bash
# 1. Install the local toolchain (Go 1.24, kubebuilder, kind, kubectl, helm)
bash scripts/setup-env.sh
source scripts/env.sh

# 2. Spin up a 4-node kind cluster + build/load operator image
bash scripts/setup-kind.sh

# 3. End-to-end demo: helm install + submit PhysicsJob + verify NodeAffinity routing
bash scripts/demo.sh
```

For development against a running kind cluster (skip the docker rebuild
loop), use `bash scripts/dev-run.sh` to run the manager binary on your
host with a port-forwarded mock-rucio.

The kind cluster has four worker nodes labelled with realistic WLCG sites:

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

Inspect:

```bash
kubectl get pj   # shortName for physicsjob
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
| `ClosestSite` | Same as DataLocal; extension point for a geo-distance ranking across replicas |
| `AnyAvailable` | No affinity injected; scheduler places freely; RSE still recorded for observability |

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

See [`docs/architecture.md`](docs/architecture.md) for the full component
diagram, reconcile-loop pseudocode, and data-flow explanation.

---

## Repository layout

```
api/v1alpha1/               CRD types (PhysicsJobSpec, PhysicsJobStatus, Phase enum)
internal/controller/        Reconciler + Ginkgo/envtest suite
internal/storage/           StorageTopologyClient interface + Rucio HTTP client
internal/scheduling/        NodeAffinity builder
internal/metrics/           Prometheus registrations
internal/mockrucio/         Mock Rucio API — 9 ATLAS/CMS/LHCb datasets
cmd/main.go                 Manager entrypoint (--rucio-url flag)
cmd/mock-rucio/main.go      Standalone mock-rucio server
config/                     Generated CRD + RBAC + sample CR
deploy/                     kind cluster config + mock-rucio Kubernetes manifest
helm/data-gravity-operator/ Helm chart (CRD in crds/, RBAC, Deployment, optional mock)
scripts/                    setup-env.sh  setup-kind.sh  demo.sh  dev-run.sh
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

---

## Citing this work

If you use `data-gravity-operator` in academic work, please cite it via the
metadata in [`CITATION.cff`](CITATION.cff). Each tagged GitHub release is
archived on Zenodo with a DOI; replace the DOI below with the version-specific
one for the release you used.

```bibtex
@software{singh_data_gravity_operator_2026,
  author       = {Singh, Karan},
  title        = {data-gravity-operator: Data-Locality-Aware Workload
                  Scheduling for Kubernetes on WLCG Data Lakes},
  year         = 2026,
  publisher    = {Zenodo},
  version      = {0.1.0},
  url          = {https://github.com/KaranSinghDev/data-gravity-operator}
}
```

---

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development setup and
contribution guidelines. Maintainers cutting a release should follow
[`RELEASING.md`](RELEASING.md).

---

## License

Licensed under the Apache License, Version 2.0. See [`LICENSE`](LICENSE) for
the full text. All third-party dependencies are also Apache 2.0 or
compatible permissive licenses.
