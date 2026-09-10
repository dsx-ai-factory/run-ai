# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

import data.kubernetes

name = input.metadata.name

deny_privilege_escalation[{"msg": msg, "details": {}}] {
    kubernetes.is_deployment
    c := input_containers[_]
    input_allow_privilege_escalation(c)
    msg := sprintf("Privilege escalation container is not allowed: %v", [name])
}

input_allow_privilege_escalation(c) {
   c.securityContext.allowPrivilegeEscalation == true
}

input_containers[c] {
    c := input.spec.template.spec.containers[_]
}

input_containers[c] {
    c := input.spec.template.spec.initContainers[_]
}

has_field(object, field) = true {
    object[field]
}