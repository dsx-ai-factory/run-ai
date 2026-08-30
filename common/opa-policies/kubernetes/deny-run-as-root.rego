package main 

import data.kubernetes

name = input.metadata.name

# Deny Pods that allow running as root
deny_run_as_root[msg] {
    kubernetes.is_pod
    c := input.spec
    input_validate_security_context_exist(c)

    msg = sprintf("Pods must not run containers as root: %s", [name])
}

# Deny Deployments that allow running as root
deny_run_as_root[msg] {
    kubernetes.is_deployment
    c := input.spec.template.spec
    input_validate_security_context_exist(c)

    # Check all containers in the deployment
    some i
    container := c.containers[i]
    not input_validate_security_context_exist(container)

    msg = sprintf("Containers must not run as root in Deployment %s", [name])
}

# Ensure securityContext exists
input_validate_security_context_exist(c) {
    has_field(c, "securityContext")
}

# Ensure securityContext contains runAsNonRoot
input_validate_security_context_exist(c) {
    has_field(c.securityContext, "runAsNonRoot")
}

# Ensure runAsNonRoot is set to true
input_validate_security_context_exist(c) {
    c.securityContext.runAsNonRoot == true
}

# Check if an object has a field
has_field(object, field) = true {
    object[field]
}

exception[rules] {
  kubernetes.is_deployment
  input.metadata.name == "runai-operator-kube-promet-operator"

  rules := ["run_as_root"]
}

exception[rules] {
  kubernetes.is_deployment
  input.metadata.name == "runai-operator-kube-state-metrics"

  rules := ["run_as_root"]
}