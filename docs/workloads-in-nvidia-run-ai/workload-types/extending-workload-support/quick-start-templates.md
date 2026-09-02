# Quick Start Templates

This section provides ready-to-use templates for defining new workload types. Each template represents a common Kubernetes pattern and can serve as a foundation when registering a new workload type through the [Workload Types](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#post-api-v1-workload-types) API.

Use these templates to accelerate development - simply start from the closest match and adapt it to your CRD’s structure.

## Before You Begin

When using these templates:

* Choose the template that most closely matches your workload’s structure
* Replace the **GVK** (`group`, `version`, `kind`) fields with those from your CRD
* Verify **status conditions** against the CRD source or official documentation
* Define **child** or **referenced components** only if they’re required for scheduling or optimization

## Generic Job

Represents a single-component workload where pods are defined directly within the CRD specification.

```json
rootComponent:
  name: "job"
  kind:
    group: "batch"
    version: "v1"
    kind: "Job"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Running"
          status: "True"
      completed:
        byConditions:
        - type: "Complete"
          status: "True"
      failed:
        byConditions:
        - type: "Failed"
          status: "True"
  
  specDefinition:
    podTemplateSpecPath: ".spec.template"
```

## Deployment

Represents a controlling resource that manages subordinate ReplicaSets.

```json
rootComponent:
  name: "deployment"
  kind:
    group: "apps"
    version: "v1"
    kind: "Deployment"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Available"
          status: "True"

childComponents:
- name: "replicaset"
  ownerName: "deployment"
  kind:
    group: "apps"
    version: "v1"
    kind: "ReplicaSet"
  specDefinition:
    podTemplateSpecPath: ".spec.template"
```

## Distributed Training (PyTorchJob)

Defines a multi-component workload with role-based pods such as **master** and **worker** roles.

```json
rootComponent:
  name: "pytorchjob"
  kind:
    group: "kubeflow.org"
    version: "v1"
    kind: "PyTorchJob"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Running"
          status: "True"
      completed:
        byConditions:
        - type: "Succeeded"
          status: "True"
      failed:
        byConditions:
        - type: "Failed"
          status: "True"

childComponents:
- name: "master"
  ownerName: "pytorchjob"
  kind:
    group: "apps"
    version: "v1"
    kind: "ReplicaSet"
  specDefinition:
    podTemplateSpecPath: ".spec.pytorchReplicaSpecs.Master.template"
  podSelector:
    componentTypeSelector:
      keyPath: '.metadata.labels["training.kubeflow.org/replica-type"]'
      value: "master"

- name: "worker"
  ownerName: "pytorchjob"
  kind:
    group: "apps"
    version: "v1"
    kind: "ReplicaSet"
  specDefinition:
    podTemplateSpecPath: ".spec.pytorchReplicaSpecs.Worker.template"
  podSelector:
    componentTypeSelector:
      keyPath: '.metadata.labels["training.kubeflow.org/replica-type"]'
      value: "worker"
```

## Inference Service (Knative)

Describes a workload that references an external component such as a Knative **Revision**.

```json
rootComponent:
  name: "service"
  kind:
    group: "serving.knative.dev"
    version: "v1"
    kind: "Service"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Ready"
          status: "True"

referencedComponents:
- name: "revision"
  kind:
    group: "serving.knative.dev"
    version: "v1"
    kind: "Revision"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Ready"
          status: "True"
```
