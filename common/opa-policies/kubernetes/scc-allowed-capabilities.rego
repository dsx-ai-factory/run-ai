# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

name = input.metadata.name

deny_allowed_scc[msg] {
    input.allowHostDirVolumePlugin != false
    msg = sprintf("%v: allowHostDirVolumePlugin must be set to false", [name])
}

deny_allowed_scc[msg] {
    input.allowHostNetwork != false
    msg = sprintf("%v: allowHostNetwork must be set to false", [name])
}

deny_allowed_scc[msg] {
    input.allowHostPorts != false
    msg = sprintf("%v: allowHostPorts must be set to false", [name])
}

deny_allowed_scc[msg] {
    input.allowPrivilegeEscalation != false
    msg = sprintf("%v: allowPrivilegeEscalation must be set to false", [name])
}

deny_allowed_scc[msg] {
    input.allowPrivilegedContainer != false
    msg = sprintf("%v: allowPrivilegedContainer must be set to false", [name])
}

deny_allowed_scc[msg] {
    input.allowedCapabilities != null
    msg = sprintf("%v: allowedCapabilities must be set to null", [name])
}

deny_allowed_scc[msg] {
    input.readOnlyRootFilesystem != false
    msg = sprintf("%v: readOnlyRootFilesystem must be set to false", [name])
}

exception[rules] {
    input.metadata.name == "runai-openshift-s3"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-backend-grafana"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-backend-keycloakx"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-backend-thanos-querier"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-backend-thanos-receive"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-user-job"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-mps-server"
    rules := ["allowed_scc"]
}

exception[rules] {
    input.metadata.name == "runai-jupyter-notebook"
    rules := ["allowed_scc"]
}