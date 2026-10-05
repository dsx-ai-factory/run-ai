#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
export KIND_CLUSTER="${KIND_CLUSTER:-inference-endpoint-controller-test-e2e}"
test_dir=$(mktemp -d)
export KUBECONFIG="$test_dir/kubeconfig"
trap 'rm -rf "$test_dir"' EXIT
# Fail without deleting anything if this cluster name is already in use.
kind create cluster --name "$KIND_CLUSTER"
trap 'kind delete cluster --name "$KIND_CLUSTER"; rm -rf "$test_dir"' EXIT
go test -tags=e2e ./test/e2e/ -v -timeout 15m
