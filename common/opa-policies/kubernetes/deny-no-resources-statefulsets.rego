# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

import data.kubernetes

name = input.metadata.name

deny_no_resources[msg]{ 
    kubernetes.is_statefulset
    c := input_containers[_]
    
    not c.resources.limits.memory
    not c.resources.limits.cpu
    not c.resources.requests.memory
    not c.resources.requests.cpu

    msg := sprintf("resources for the %v container,missing memory/cpu requests/limits", [name])
}

input_containers[c] {
    c := input.spec.template.spec.containers[_]
}

input_containers[c] {
    c := input.spec.template.spec.initContainers[_]
}