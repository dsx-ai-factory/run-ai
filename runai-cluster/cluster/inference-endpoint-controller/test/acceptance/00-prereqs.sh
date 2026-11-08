#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
# Track A step 0 · admin prerequisites. ✅ upstream, none of it is ours.
# Idempotent: safe to re-run.
set -euo pipefail

GATEWAY_API_VERSION="${GATEWAY_API_VERSION:-v1.6.0}"
GIE_VERSION="${GIE_VERSION:-v1.5.0}"            # last release that ships the reference EPP + x-k8s.io Objective/Rewrite CRDs
AGW_VERSION="${AGW_VERSION:-v1.5.0}"
AGW_NS="${AGW_NS:-agentgateway-system}"

step() { printf '\n\033[1;32m== %s\033[0m\n' "$*"; }

step "1/4 Gateway API ${GATEWAY_API_VERSION} standard CRDs"
kubectl apply --server-side --force-conflicts -f \
  "https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/standard-install.yaml"

step "2/4 GIE ${GIE_VERSION} CRDs (InferencePool + x-k8s.io InferenceObjective / InferenceModelRewrite)"
kubectl apply --server-side --force-conflicts -f \
  "https://github.com/kubernetes-sigs/gateway-api-inference-extension/releases/download/${GIE_VERSION}/manifests.yaml"
kubectl get crd | grep -E 'inference\.networking' || { echo "GIE CRDs missing"; exit 1; }

step "3/4 agentgateway ${AGW_VERSION} CRDs + controller, inference extension on"
helm upgrade -i agentgateway-crds oci://cr.agentgateway.dev/charts/agentgateway-crds \
  --version "${AGW_VERSION}" -n "${AGW_NS}" --create-namespace
helm upgrade -i agentgateway oci://cr.agentgateway.dev/charts/agentgateway \
  --version "${AGW_VERSION}" -n "${AGW_NS}" \
  --set inferenceExtension.enabled=true \
  --wait --timeout 5m

step "4/4 GatewayClass accepted?"
for i in $(seq 1 30); do
  if kubectl get gatewayclass agentgateway -o jsonpath='{.status.conditions[?(@.type=="Accepted")].status}' 2>/dev/null | grep -q True; then
    kubectl get gatewayclass agentgateway
    echo "OK: prerequisites installed"
    exit 0
  fi
  sleep 5
done
echo "GatewayClass agentgateway not Accepted after 150s"; kubectl get gatewayclass; exit 1
