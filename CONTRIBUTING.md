# Contributing

Thanks for your interest in `data-gravity-operator`. This document covers
the development workflow.

## Reporting bugs and requesting features

Open a GitHub issue with:
- The expected vs. observed behaviour
- A minimal reproduction (a `PhysicsJob` YAML, the operator log line, and
  the observed `kubectl get pj` output is usually enough)
- Operator version (`kubectl get deploy -n data-gravity-system data-gravity-operator-controller-manager -o jsonpath='{.spec.template.spec.containers[0].image}'`)
- Kubernetes version (`kubectl version --short`)

## Development setup

```bash
bash scripts/setup-env.sh         # installs Go 1.24, kubebuilder, kind, kubectl, helm into ./tools/
source scripts/env.sh             # adds ./tools/bin to PATH for this shell
make generate manifests           # regenerates DeepCopy code + CRD YAML
make test                         # runs unit tests + Ginkgo/envtest controller suite
```

For an interactive dev loop against a running kind cluster, see
`scripts/dev-run.sh` — it port-forwards the in-cluster mock-rucio service
and runs the manager binary on your host.

## Code style

- Run `make fmt vet` before pushing. CI also runs `golangci-lint`.
- One logical change per commit; follow the conventional commit format used
  in `git log` (`type(scope): description` — `feat`, `fix`, `chore`, `test`,
  `docs`, etc.).
- Keep generated files (`zz_generated.deepcopy.go`, CRD YAML, Helm chart
  CRDs) in sync with their source: run `make generate manifests` after
  editing types in `api/v1alpha1/`.

## Tests

- Unit tests live next to the code (`*_test.go`). Add table-driven tests
  for storage-client logic and affinity-builder logic.
- Controller behaviour is tested via the Ginkgo suite at
  `internal/controller/`. New scheduling policies or status transitions
  should add a new `Context` block.
- `make test` is required to be green for any PR.

## Adding a new dataset to mock-rucio

Edit `internal/mockrucio/data.go`. Each entry needs an RSE list and a
`DatasetSizeBytes` estimate (used for the
`physjob_data_transfer_avoided_bytes` metric). After editing, the unit
test `TestHandler_AllSeedDIDs` will assert the new entry is reachable.

## Release process

See `RELEASING.md` (added alongside the release workflow).
