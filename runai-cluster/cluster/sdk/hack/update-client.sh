#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

go mod tidy
go mod vendor

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

THIS_PKG="github.com/run-ai/runai/runai-cluster/cluster/sdk"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/apis"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/client" \
    --output-pkg "${THIS_PKG}/client" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/apis"
