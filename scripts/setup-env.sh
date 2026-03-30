#!/usr/bin/env bash
# Installs all required development tools into ./tools/ (no root, no system changes).
# Run once from the project root: bash scripts/setup-env.sh

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLS_DIR="${PROJECT_ROOT}/tools"
mkdir -p "${TOOLS_DIR}"

OS="linux"
ARCH="amd64"

# ── versions ──────────────────────────────────────────────────────────────────
GO_VERSION="1.24.2"
KUBEBUILDER_VERSION="4.6.0"
KIND_VERSION="0.26.0"
KUBECTL_VERSION="1.33.0"
HELM_VERSION="3.17.2"
# ──────────────────────────────────────────────────────────────────────────────

echo "Installing tools to ${TOOLS_DIR}"

# ── Go ────────────────────────────────────────────────────────────────────────
if [ ! -f "${TOOLS_DIR}/go/bin/go" ]; then
  echo "→ Downloading Go ${GO_VERSION}..."
  TMP=$(mktemp -d)
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz" -o "${TMP}/go.tar.gz"
  tar -C "${TOOLS_DIR}" -xzf "${TMP}/go.tar.gz"
  rm -rf "${TMP}"
  echo "  Go installed at ${TOOLS_DIR}/go/bin/go"
else
  echo "  Go already present"
fi

# ── kubebuilder ───────────────────────────────────────────────────────────────
if [ ! -f "${TOOLS_DIR}/bin/kubebuilder" ]; then
  echo "→ Downloading kubebuilder ${KUBEBUILDER_VERSION}..."
  curl -fsSL "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${OS}_${ARCH}" \
    -o "${TOOLS_DIR}/bin/kubebuilder"
  chmod +x "${TOOLS_DIR}/bin/kubebuilder"
else
  echo "  kubebuilder already present"
fi
mkdir -p "${TOOLS_DIR}/bin"

# ── kind ──────────────────────────────────────────────────────────────────────
if [ ! -f "${TOOLS_DIR}/bin/kind" ]; then
  echo "→ Downloading kind ${KIND_VERSION}..."
  mkdir -p "${TOOLS_DIR}/bin"
  curl -fsSL "https://kind.sigs.k8s.io/dl/v${KIND_VERSION}/kind-${OS}-${ARCH}" \
    -o "${TOOLS_DIR}/bin/kind"
  chmod +x "${TOOLS_DIR}/bin/kind"
else
  echo "  kind already present"
fi

# ── kubectl ───────────────────────────────────────────────────────────────────
if [ ! -f "${TOOLS_DIR}/bin/kubectl" ]; then
  echo "→ Downloading kubectl ${KUBECTL_VERSION}..."
  curl -fsSL "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl" \
    -o "${TOOLS_DIR}/bin/kubectl"
  chmod +x "${TOOLS_DIR}/bin/kubectl"
else
  echo "  kubectl already present"
fi

# ── helm ──────────────────────────────────────────────────────────────────────
if [ ! -f "${TOOLS_DIR}/bin/helm" ]; then
  echo "→ Downloading helm ${HELM_VERSION}..."
  TMP=$(mktemp -d)
  curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-${OS}-${ARCH}.tar.gz" -o "${TMP}/helm.tar.gz"
  tar -C "${TMP}" -xzf "${TMP}/helm.tar.gz"
  mv "${TMP}/${OS}-${ARCH}/helm" "${TOOLS_DIR}/bin/helm"
  rm -rf "${TMP}"
  chmod +x "${TOOLS_DIR}/bin/helm"
else
  echo "  helm already present"
fi

# ── env.sh ────────────────────────────────────────────────────────────────────
cat > "${PROJECT_ROOT}/scripts/env.sh" <<EOF
export GOROOT="${TOOLS_DIR}/go"
export GOPATH="${TOOLS_DIR}/gopath"
export PATH="${TOOLS_DIR}/go/bin:${TOOLS_DIR}/bin:\${GOPATH}/bin:\${PATH}"
EOF

echo ""
echo "Done. Activate with:"
echo "  source scripts/env.sh"
echo ""
echo "Verify:"
echo "  go version"
echo "  kubebuilder version"
echo "  kind version"
echo "  kubectl version --client"
echo "  helm version"
