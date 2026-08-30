package main

import data.kubernetes

name = input.metadata.name

violation[{"msg": msg, "details": {}}] {
    kubernetes.is_deployment
    c := input_containers[_]
    c.securityContext.privileged
    msg := sprintf("Privileged container is not allowed: %v, securityContext: %v", [c.name, c.securityContext])
}

input_containers[c] {
    c := input.spec.template.spec.containers[_]
}

input_containers[c] {
    c := input.spec.template.spec.initContainers[_]
}
