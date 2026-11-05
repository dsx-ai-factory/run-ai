#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."

# config/rbac/role.yaml stays the single description of what the controller needs,
# but the chart must not bind child-object writes cluster-wide: a cluster-wide
# "create deployments" plus "create roles/rolebindings" would let the controller's
# service account escalate anywhere. The rules are therefore split by verb:
# read verbs land in the cluster-wide ClusterRole (the controller watches children
# everywhere) and write verbs in the "-project" ClusterRole, which KRM binds in
# every project namespace. The InferenceEndpoint rules stay fully cluster-wide
# because the CRD is Run:ai-owned and must be reconcilable wherever it appears.
split_rules() {
  awk -v role="$1" '
    function flush(   i, keep, n, out) {
      if (!in_rule) return
      n = 0
      for (i = 1; i <= nverbs; i++) {
        if (resources[1] ~ /^inferenceendpoints/) keep = (role == "cluster")
        else if (verbs[i] ~ /^(get|list|watch)$/) keep = (role == "cluster")
        else keep = (role == "project")
        if (keep) out[++n] = verbs[i]
      }
      if (n > 0) {
        print "- apiGroups:"
        for (i = 1; i <= ngroups; i++) print "  - " groups[i]
        print "  resources:"
        for (i = 1; i <= nresources; i++) print "  - " resources[i]
        print "  verbs:"
        for (i = 1; i <= n; i++) print "  - " out[i]
      }
      in_rule = 0
    }
    /^rules:/ { in_rules = 1; next }
    !in_rules { next }
    /^- apiGroups:/ { flush(); in_rule = 1; ngroups = nresources = nverbs = 0; section = "groups"; next }
    /^  resources:/ { section = "resources"; next }
    /^  verbs:/ { section = "verbs"; next }
    /^  - / {
      item = substr($0, 5)
      if (section == "groups") groups[++ngroups] = item
      else if (section == "resources") resources[++nresources] = item
      else verbs[++nverbs] = item
      next
    }
    END { flush() }
  ' config/rbac/role.yaml
}

{
  echo '# Generated from inference-endpoint-controller/config/rbac/role.yaml; run make copy-crds.'
  cat <<'YAML'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-inference-endpoint-controller
rules:
YAML
  split_rules cluster
  cat <<'YAML'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: runai-inference-endpoint-controller-project
rules:
YAML
  split_rules project
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
