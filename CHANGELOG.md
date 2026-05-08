# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-05-09

Initial release.

### Added
- `PhysicsJob` custom resource (`hep.cern.local/v1alpha1`) with `dataset`, `image`,
  `command`, `resources`, and `schedulingPolicy` fields.
- Reconciler implementing data-locality-aware scheduling:
  resolves a Rucio DID to its Storage Element, maps the RSE to a Kubernetes node via
  the `topology.cern.io/site` label, and creates an owned `batch/v1.Job` with
  `NodeAffinity` constraints injected.
- Three scheduling policies: `DataLocal` (default), `ClosestSite`, `AnyAvailable`.
- `StorageTopologyClient` interface with a Rucio HTTP client implementation.
- Mock Rucio HTTP server (`cmd/mock-rucio`) with seed data covering nine ATLAS,
  CMS, and LHCb dataset classes — sizes range 80 GB to 3.2 TB to give realistic
  WAN-bytes-avoided figures.
- Five Prometheus metrics:
  `physjob_reconcile_total`, `physjob_reconcile_duration_seconds`,
  `physjob_resolved_total`, `physjob_resolution_failures_total`, and
  `physjob_data_transfer_avoided_bytes`.
- Ginkgo controller test suite using envtest (DataLocal, AnyAvailable,
  RSENotFound paths).
- Helm chart with CRD bundled in `crds/`, ClusterRole/Binding, ServiceAccount,
  manager Deployment, and an optional in-chart mock-rucio for demo purposes.
- kind cluster configuration with four worker nodes labelled
  `topology.cern.io/site={cern-prod, bnl-osg2, in2p3-cc, triumf-lcg2}`.
- End-to-end demo script that installs the chart and submits a sample `PhysicsJob`.
- Out-of-cluster development runner (`scripts/dev-run.sh`) with port-forwarded
  mock-rucio.
- Architecture document (`docs/architecture.md`) with a Mermaid component
  diagram and reconcile-loop pseudocode.

[Unreleased]: https://github.com/KaranSinghDev/data-gravity-operator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/KaranSinghDev/data-gravity-operator/releases/tag/v0.1.0
