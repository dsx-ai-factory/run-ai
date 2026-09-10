# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

package main

import data.kubernetes

name = input.metadata.name

violation[{"msg": msg}] {
  kubernetes.is_deployment
  container := input.spec.template.spec.containers[_]
  has_disallowed_capabilities(container)
  msg := sprintf("container <%v> has a disallowed capability. Allowed capabilities are %v", [container.name, get_default(input.parameters, "allowedCapabilities", "NONE")])
}

violation[{"msg": msg}] {
  kubernetes.is_deployment
  container := input.spec.template.spec.containers[_]
  missing_drop_capabilities(container)
  msg := sprintf("container <%v> is not dropping all required capabilities. Container must drop all of %v", [container.name, input.parameters.requiredDropCapabilities])
}

violation[{"msg": msg}] {
  kubernetes.is_deployment
  container := input.spec.template.spec.initContainers[_]
  has_disallowed_capabilities(container)
  msg := sprintf("init container <%v> has a disallowed capability. Allowed capabilities are %v", [container.name, get_default(input.parameters, "allowedCapabilities", "NONE")])
}

violation[{"msg": msg}] {
  kubernetes.is_deployment
  container := input.spec.template.spec.initContainers[_]
  missing_drop_capabilities(container)
  msg := sprintf("init container <%v> is not dropping all required capabilities. Container must drop all of %v", [container.name, input.parameters.requiredDropCapabilities])
}

has_disallowed_capabilities(container) {
  allowed := {c | c := input.parameters.allowedCapabilities[_]}
  not allowed["*"]
  capabilities := {c | c := container.securityContext.capabilities.add[_]}
  count(capabilities - allowed) > 0
}

missing_drop_capabilities(container) {
  must_drop := {c | c := input.parameters.requiredDropCapabilities[_]}
  dropped := {c | c := container.securityContext.capabilities.drop[_]}
  count(must_drop - dropped) > 0
}

get_default(obj, param, _default) = out {
  out = obj[param]
}

get_default(obj, param, _default) = out {
  not obj[param]
  not obj[param] == false
  out = _default
}
