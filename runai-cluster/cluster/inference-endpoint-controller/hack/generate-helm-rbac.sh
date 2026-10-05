#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
{
  echo '# Generated from inference-endpoint-controller/config/rbac/role.yaml; run make copy-crds.'
  sed 's/name: manager-role/name: runai-inference-endpoint-controller/' config/rbac/role.yaml
  cat <<'YAML'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: runai-inference-endpoint-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: runai-inference-endpoint-controller
subjects:
  - kind: ServiceAccount
    name: inference-endpoint-controller
    namespace: {{ .Release.Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: inference-endpoint-controller-leader-election
  namespace: {{ .Release.Namespace }}
rules:
  - apiGroups: [coordination.k8s.io]
    resources: [leases]
    verbs: [get, list, watch, create, update, patch, delete]
  - apiGroups: [""]
    resources: [events]
    verbs: [create, patch]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: inference-endpoint-controller-leader-election
  namespace: {{ .Release.Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: inference-endpoint-controller-leader-election
subjects:
  - kind: ServiceAccount
    name: inference-endpoint-controller
    namespace: {{ .Release.Namespace }}
YAML
} > ../runai-chart/templates/clusterroles/inference-endpoint-controller.yaml
