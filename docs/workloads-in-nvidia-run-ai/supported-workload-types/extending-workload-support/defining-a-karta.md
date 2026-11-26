# Defining a Karta

This section describes how to extend NVIDIA Run:ai to support ML frameworks, tools, or Kubernetes resources by defining and registering them through the Workload Types API and a corresponding Karta definition. See the [Quick start templates](quick-start-templates.md) for ready-to-use examples of Karta configurations.

A Karta is a declarative workload contract that describes how NVIDIA Run:ai should interpret a custom resource. It defines how the platform identifies the workload’s structure, locates its pods, tracks its state, and applies scheduling, monitoring, and optimization logic.

Without a Karta, NVIDIA Run:ai cannot reliably schedule, monitor, or optimize a custom resource, because it lacks the information required to traverse the resource’s schema and interpret its semantics.

## Workload Types API

Use this endpoint `POST /api/v1/workload-types` to register a new workload type. Each request must include the workload’s metadata and one or more Karta definitions, one per supported CRD version (for example, `v1`, `v1beta1`). See [Workload Types](api:main-api:POST/api/v1/workload-types) API for more details.

| Field                | Description                                                                                                                                                                                                                                                                                                                                                                                 | Example                                                                           |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `categoryId`         | The unique identifier of the workload category. See [List workload categories](api:main-api:GET/api/v1/workload-categories) API.                                                                                                                                                                                            | `046b6c7f-...`                                                                    |
| `priorityId`         | The unique identifier of the workload priority. See [Get workload priorities](api:main-api:GET/api/v1/workload-priorities) API.                                                                                                                                                                                             | `046b6c7f-...`                                                                    |
| `preemptibility`     | The preemptibility value of the workload - `non-preemptible` / `preemptible`.                                                                                                                                                                                                                                                                                                               | `non-preemptible`                                                                 |
| `name`               | The unique name of the workload type. This value must match the Kubernetes Kind that represents the workload type.                                                                                                                                                                                                                                                                          | `Deployment`                                                                      |
| `group`              | The Kubernetes API group associated with the workload resource.                                                                                                                                                                                                                                                                                                                             | `apps`                                                                            |
| `kartas`             | One or more Karta definitions (one per CRD version, e.g. v1, v1beta1). Lists the versions of the custom resource definition (CRD) supported for this workload type, such as v1, v1beta1, or v1alpha1. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. | See [Karta structure](defining-a-karta.md#karta-structure)                        |
| `resourceInterfaces` | Deprecated. Use `kartas` instead. Kept for backward compatibility.                                                                                                                                                                                                                                                                                                                          |                                                                                   |

## What is a Karta?

In Kubernetes, a workload is not a standalone execution unit (pod). It comprises various components, including an ingress entry point, a collection of pods, and storage.

A Karta enables NVIDIA Run:ai to interpret, schedule, monitor, and optimize custom Kubernetes workloads by describing how the CRD’s structure maps to workload behavior. A Karta tells the platform:

* What resources represent the workload (CRD group, version, kind)
* How to find pods and their spec fields
* How to map status conditions to unified states
* How to interpret component hierarchy and scheduling rules

<Note>

**Note**

The Karta defines the contract between a CRD and the NVIDIA Run:ai platform. It does not create the CRD; it explains how NVIDIA Run:ai should treat it.

</Note>

## Karta Structure

A Karta is a structured YAML or JSON object that describes how a workload is represented in Kubernetes. The following sections describe each required component of the Karta definition.

### Minimum Requirements

A minimal Karta must define a `rootComponent` with a name, full GVK (`group`, `version`, `kind`) and a `statusDefinition` that describes its runtime state. All other fields are optional but unlock additional platform capabilities.

```json
rootComponent:
  name: "minimal"
  kind:
    group: "minimal.org"
    version: "v1"
    kind: "minimalRI"
  statusDefinition:
    statusMappings:
      running:
        byConditions:
        - type: "Running"
          status: "True"
```

#### Contract Summary: Required vs Optional

| Section                    | Required                                                  | Purpose                                                                                                |
| -------------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `rootComponent`            | Yes                                                       | Defines the primary resource (CRD) that represents the workload                                        |
| `statusDefinition`         | Yes (under `rootComponent`)                               | Maps the CRD’s status fields to NVIDIA Run:ai canonical workload states (for example, running, failed) |
| `specDefinition`           | Yes (where pods exist)                                    | Specifies where NVIDIA Run:ai can locate the pod specifications associated with the workload           |
| `childComponents`          | Optional (only if workload has owned resources)           | Describes subordinate Kubernetes resources that are created and managed by the root CRD                |
| `podSelector`              | Optional (required in multi-component workloads)          | Associates pods with their component or instance within the workload                                   |
| `instanceIdPath`           | Optional (required if a component has multiple instances) | Uniquely identifies each instance when a component produces multiple specs                             |
| `additionalChildKinds`     | Optional                                                  | Declares additional resource kinds (GVKs) that the workload creates or controls.                       |
| `optimizationInstructions` | Optional                                                  | Instructs scheduler behaviors (e.g., gang scheduling).                                                 |
| `scaleDefinition`          | Optional                                                  | Defines how the workload or its components can be scaled                                               |
| `suspendDefinition`        | Optional                                                  | Enables suspend and resume lifecycle operations for the workload type                                  |

### Root Component

Every Karta begins with a `rootComponent` definition:

* Must specify the full Kubernetes GVK (`group`, `version`, `kind`)
* Must include a `statusDefinition` to describe the workload’s runtime state

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
```

### Child Components

`childComponents` represent resources owned by the root component.

* Must include `ownerRef` (points to parent component)
* Usually includes a `specDefinition`
* All paths must be **absolute**, starting from the CRD root

```json
childComponents:
- name: "worker"
  ownerName: "pytorchjob"
  kind:
    group: "apps"
    version: "v1"
    kind: "StatefulSet"
  specDefinition:
    podTemplateSpecPath: ".spec.pytorchReplicaSpecs.Worker.template"
```

### Paths

All paths within a Karta are written in [jq](https://jqlang.org/) syntax. The jq query language provides both path navigation and various query capabilities, and is widely used in the Kubernetes ecosystem. When defining paths:

* Use the correct jq type for each property (path/query)
* Provide default values where necessary

### Spec Definitions

`specDefinition` determines how NVIDIA Run:ai locates pod specifications within the CRD. Several mutually exclusive definition types are supported, such as `podTemplateSpecPath`, `podSpecPath`, `metaDataPath` and `fragmentedPodDefinition`.

| Pattern | Use Case | Example |
| --- | --- | --- |
| `podTemplateSpecPath` | CRD embeds a full pod template | `.spec.pytorchReplicaSpecs.Master.template` |
| `podSpecPath` and `metadataPath` | CRD directly embeds a podSpec and/or an objectMeta | `.spec.jobTemplate.spec`<br/>`.spec.jobTemplate.metadata` |
| `fragmentedPodSpecDefinition` | Pod fields scattered across CRD | `specDefinition:<br/> fragmentedPodDefinition:<br/> labelsPath: ".spec.labels"<br/> annotationsPath: ".spec.annotations"<br/> resourcesPath: ".spec.resources"<br/> schedulerNamePath: ".spec.schedulerName"` |

### Spec Definition Requirements for MNNVL Support

To support Multi-Node NVLink (MNNVL) workloads, a Karta must define one or more spec definitions. Each spec definition must provide access to the following semantic elements:

* Containers
* Pod affinity and anti-affinity
* Resource claims
* Node affinity

A spec definition may be expressed using one of the following mechanisms:

* `podTemplateSpecPath`
* `podSpecPath`
* `fragmentedPodSpecDefinition`

When `podTemplateSpecPath` or `podSpecPath` is used, coverage of all required semantic elements listed above is **implicit**. When `fragmentedPodSpecDefinition` is used, the definition **must explicitly specify paths** for each required element, including:

* `containersPath` or `containerPath`
* `podAffinityPath`
* `resourceClaimsPath`
* `nodeAffinityPath`

### Component Instances

A component’s spec definition might point to multiple specs (in map/array format). In those cases, it’s crucial to be able to distinguish between each instance of that component.

If a component produces multiple instances (arrays or maps), define `instanceIdPath` to identify each instance uniquely.

```json
# Array of specs
instanceIdPath: ".spec.jobs[].name"

# Map of specs (use map keys as instance IDs)
instanceIdPath: ".spec.jobs | to_entries[] | .key"
```

### Pod Selectors

Use `podSelector` to associate pods with components or instances. This is required when multiple components define pods or when a component manages multiple instances. All selectors of each kind (component, instance) must be **mutually exclusive** within their scope. Paths in pod selectors refer to paths in the pod JSON/YAML.

*   `componentTypeSelector` - A key and value selector that associates pod to the current component. If the value is not provided, only key existence is checked:

    ```json
    podSelector:
      componentTypeSelector:
        keyPath: '.metadata.labels["training.kubeflow.org/replica-type"]'
        value: "master"
    ```
*   `componentInstanceSelector` - A path on the pod that holds its matching instance id:

    ```json
    podSelector:
      componentInstanceSelector:
        idPath: '.metadata.labels["jobset.sigs.k8s.io/replicatedjob-name"]'
    ```
*   `replicaSelector` - Identifies which replica group a pod belongs to, for workloads where each replica group contains an identical sub-structure hierarchy. Use this when multiple replica groups exist and each group contains its own child components (for example, a leader and its workers per replica group):

    ```json
    podSelector:
      replicaSelector:
        idPath: '.metadata.labels["<replica-group-label>"]'
    ```

**DynamoGraphDeployment Example**

The following example shows how the selectors apply to a `DynamoGraphDeployment` workload with three services: Frontend, Middle, and Backend. All three services share the same structural definition - they are the same component type - and are differentiated only by their instance identity (the service name).

The Dynamo operator labels each pod with:

| Label                       | Example Value                   | Purpose                                     |
| --------------------------- | ------------------------------- | ------------------------------------------- |
| `nvidia.com/dynamo-graph`   | `dynamo-disagg-app`             | Identifies the parent DynamoGraphDeployment |
| `nvidia.com/dynamo-service` | `Frontend`, `Middle`, `Backend` | Identifies which service the pod belongs to |

The resulting component hierarchy is:

```
DynamoGraphDeployment (dynamo-disagg-app)
└── service                          ← single component definition
    ├── [instance: Frontend]         ← 1 replica pod
    ├── [instance: Middle]           ← 1 replica pod
    └── [instance: Backend]          ← 1 replica pod
```

The `podSelector` for this workload uses `componentTypeSelector` and `componentInstanceSelector` together:

```json
podSelector:
  componentTypeSelector:
    keyPath: '.metadata.labels["nvidia.com/dynamo-service"]'
  componentInstanceSelector:
    idPath: '.metadata.labels["nvidia.com/dynamo-service"]'
```

* `componentTypeSelector` - Confirms pod membership by checking for the existence of the `nvidia.com/dynamo-service` label. Since all Dynamo service pods carry this label, its existence is sufficient to confirm the pod belongs to the `service` component type.
* `componentInstanceSelector` - Extracts the label value (`Frontend`, `Middle`, or `Backend`) to create a separate component instance for each service.
* `replicaSelector` - Not used. Each service has standard Deployment replicas with no sub-structure beneath them. Their scale is fully captured by `scaleDefinition.replicasPath`.

### Status Definitions

The `statusDefinition` maps CRD conditions or phases to the canonical Karta statuses. A status definition is required for the `rootComponent`. For each generic status, you can define how it is evaluated using one or more of the following mechanisms:

#### Conditions and Phases

* Statuses can be derived from CRD conditions, phases, or both.
* If both conditions and phases are defined, both are evaluated when determining the status.
* When using conditions or phases, you must first define a `conditionsDefinition` or `phaseDefinition`.
* Multiple, independent definitions can be provided for the same generic status.
* When defining a status using `byConditions`, all specified conditions must be met (AND logic).

```json
statusDefinition:
        conditionsDefinition:
          path: ".status.conditions"
          typeFieldName: "type"
          statusFieldName: "status"
        statusMappings:
          initializing:
          - byConditions:
            - type: "Created"
              status: "True"
            - type: "Running"
              status: "False"
          running:
          - byConditions:
            - type: "Running"
              status: "True"
            - type: "Succeeded"
              status: "False"
            - type: "Failed"
              status: "False"
          completed:
          - byConditions:
            - type: "Succeeded"
              status: "True"
          failed:
          - byConditions:
            - type: "Failed"
              status: "True"
```

#### Matched Expressions

Matched expressions consist of:

* An expression
* An expected result

If the evaluated expression output matches the expected result, the condition is considered met and the corresponding status is applied.

Examples:

```json
Running: []boltv1alpha1.StatusMatcher{
{
	ByExpression: &boltv1alpha1.ExpressionMatcher{
		Expression:     ".status.phase",
		ExpectedResult: "running",
	},
},
```

```json
Running: []boltv1alpha1.StatusMatcher{
{
	ByExpression: &boltv1alpha1.ExpressionMatcher{
		Expression:     ".status.readyReplicas == .status.readyReplicas",
		ExpectedResult: "true",
	},
},
```

### Suspend Definition

Define `suspendDefinition` on the `rootComponent` to enable suspend and resume lifecycle operations for the workload type. When defined, NVIDIA Run:ai can suspend a running workload and resume it without deleting the resource.

`suspendDefinition` contains two arrays:

- `suspendActions` - the patches applied to the CRD to suspend the workload.
- `resumeActions` - the patches applied to the CRD to resume the workload.

Each action specifies a `path` (jq path to the field to patch) and a `value` (the string value to set at that path).

```json
suspendDefinition:
  suspendActions:
  - path: ".spec.runPolicy.suspend"
    value: "true"
  resumeActions:
  - path: ".spec.runPolicy.suspend"
    value: "false"
```

To expose the suspend lifecycle phases in NVIDIA Run:ai, also add the corresponding mappings to `statusDefinition.statusMappings`:

- `suspended` - the workload is fully suspended (all pods stopped).
- `suspending` - the workload is in the process of suspending (pods are still draining).
- `resuming` - the workload is resuming (new pods are starting).

These can be expressed using `byConditions` or `byExpression`, depending on how the CRD surfaces this state. For example:

```json
statusMappings:
  suspended:
  - byConditions:
    - type: "Suspended"
      status: "True"
  suspending:
  - byExpression:
      expression: ".spec.runPolicy.suspend == true and (.status.replicaStatuses // {} | to_entries | map(.value.active // 0) | add // 0) > 0"
      expectedResult: "true"
  resuming:
  - byConditions:
    - type: "Resuming"
      status: "True"
```

### Additional Child Kinds

List any additional GVKs created or managed by the CRD but not defined explicitly under `childComponents`. This is essential for permission management so that your CRD can be managed correctly.

```json
  additionalChildKinds:
   - group: apps
     version: v1
     kind: Deployment
   - group: leaderworkerset.x-k8s.io
     version: v1
     kind: LeaderWorkerSet
```

### Optimization Instructions

Optimization instructions define how NVIDIA Run:ai schedules or groups pods. All paths in the configuration refer to pod YAML or JSON fields.

#### Supported Instruction Types

The supported instruction type is `gangScheduling`, which instructs the Scheduler on how to group related pods:

* Each pod-group definition contains a list of the included members
* Each defined member can provide a list of distinct keys to group pods by and a list of filters to determine which pods should be included

```json
optimizationInstructions:
  gangScheduling:
    podGroups:
    - name: "job"
        members:
        - componentName: "master"
          groupByKeyPaths: 
          - '.metadata.labels["training.kubeflow.org/job-name"]'
        - componentName: "worker"
          groupByKeyPaths:
          - '.metadata.labels["training.kubeflow.org/job-name"]'
```

**Grouping examples**

The following definitions are equivalent when the `master` and `worker` share the same job-name label:

```json
optimizationInstructions:
  gangScheduling:
    podGroups:
    - name: "job"
        members:
        - componentName: "master"
          groupByKeyPaths: 
          - '.metadata.labels["training.kubeflow.org/job-name"]'
        - componentName: "worker"
          groupByKeyPaths:
          - '.metadata.labels["training.kubeflow.org/job-name"]'
```

```json
optimizationInstructions:
  gangScheduling:
    podGroups:
    - name: "job"
        members:
        - componentName: "job"
          groupByKeyPaths: 
          - '.metadata.labels["training.kubeflow.org/job-name"]'
```

**Using default values**

When multiple groups are possible, use a default value to cover cases where a single group is used (the used pattern: -{name}-{index}):

```json
optimizationInstructions:
   gangScheduling:
     podGroups:
     - name: "group"
       members:
       - componentName: "group"
         groupByKeyPaths:
         - '.metadata.labels["leaderworkerset.sigs.k8s.io/name"]'
         - '.metadata.labels["leaderworkerset.sigs.k8s.io/group-index"] // "0"'
```

**Filters**

When different components are not named components (array/map of specs in the same component) you can use filters to form different pod groups. In this example, for a CRD that defines multiple jobs but all under the `job` component, we use jq query as filter to identify pods that use NVIDIA GPUs, and queries with hard-coded values as grouping keys.

```json
optimizationInstructions:
  gangScheduling:
    podGroups:
    - name: "gpu-jobs"
        members:
        - componentName: "job"
          filters:
          - 'any(.spec.jobs[].spec.containers[]; (.resources.limits["nvidia.com/gpu"] // 0) > 0)'
          groupByKeyPaths: 
          - 'gpu'
    - name: "no-gpu-jobs"
        members:
        - componentName: "job"
          filters:
          - 'any(.spec.jobs[].spec.containers[]; (.resources.limits["nvidia.com/gpu"] // 0) == 0)'
          groupByKeyPaths: 
          - 'no-gpu'
```

## Best Practices

Before submitting a new Karta, confirm:

* All kinds use full GVK (`group`, `version`, `kind`)
* The `rootComponent` includes a valid `statusDefinition`
* Absolute paths only in `specDefinitions`
* No duplicate child kinds; don’t list explicitly defined components in `additionalChildKinds`
* Mutually exclusive `podSelectors` in multi-component workloads
* Null-safe jq expressions (e.g. // 0 defaults)
* Target actual components in `optimizationInstructions`, not just the root CRD
* Status conditions match real framework APIs
* All `childComponents` has `ownerRef` directed to existing components and there are no ownership cycles

## Examples

These examples illustrate how common workload types are modeled using `karta` definitions, including their component structure, status mapping, and scheduling configuration.

<Accordion title="DynamoGraphDeployment">

**DynamoGraphDeployment**

```json
"kartas": [
  {
    "spec": {
      "structureDefinition": {
        "rootComponent": {
          "name": "dynamographdeployment",
          "kind": {
            "group": "nvidia.com",
            "version": "v1alpha1",
            "kind": "DynamoGraphDeployment"
          },
          "statusDefinition": {
            "phaseDefinition": {
              "path": ".status.state"
            },
            "statusMappings": {
              "failed": [
                { "byPhase": "failed" }
              ],
              "initializing": [
                { "byPhase": "pending" }
              ],
              "running": [
                { "byPhase": "successful" }
              ]
            }
          }
        },
        "childComponents": [
          {
            "name": "service",
            "ownerRef": "dynamographdeployment",
            "instanceIdPath": ".spec.services | to_entries[] | .key",
            "specDefinition": {
              "fragmentedPodSpecDefinition": {
                "annotationsPath": ".spec.services | .[] | .annotations",
                "containerPath": ".spec.services | .[] | .extraPodSpec.mainContainer",
                "imagePath": ".spec.services | .[] | .extraPodSpec.mainContainer.image",
                "labelsPath": ".spec.services | .[] | .labels",
                "nodeAffinityPath": ".spec.services | .[] | .extraPodSpec.affinity.nodeAffinity",
                "podAffinityPath": ".spec.services | .[] | .extraPodSpec.affinity.podAffinity",
                "priorityClassNamePath": ".spec.services | .[] | .extraPodSpec.priorityClassName",
                "resourceClaimsPath": ".spec.services | .[] | .extraPodSpec.resourceClaims",
                "resourcesPath": ".spec.services | .[] | .resources",
                "schedulerNamePath": ".spec.services | .[] | .extraPodSpec.schedulerName"
              }
            },
            "scaleDefinition": {
              "replicasPath": ".spec.services[] | (.replicas // 1) * (.multinode.nodeCount // 1)",
              "minReplicasPath": ".spec.services | .[] | .autoscaling.minReplicas",
              "maxReplicasPath": ".spec.services | .[] | .autoscaling.maxReplicas"
            },
            "podSelector": {
              "componentInstanceSelector": {
                "idPath": ".metadata.labels[\"nvidia.com/dynamo-component\"]"
              },
              "replicaSelector": {
                "keyPath": ".metadata.labels[\"grove.io/podcliquescalinggroup-replica-index\"] // .metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]"
              }
            }
          }
        ],
        "additionalChildKinds": [
          { "group": "nvidia.com", "version": "v1alpha1", "kind": "DynamoComponentDeployment" },
          { "group": "leaderworkerset.x-k8s.io", "version": "v1", "kind": "LeaderWorkerSet" },
          { "group": "scheduler.grove.io", "version": "v1alpha1", "kind": "PodGang" },
          { "group": "grove.io", "version": "v1alpha1", "kind": "PodClique" },
          { "group": "grove.io", "version": "v1alpha1", "kind": "PodCliqueSet" },
          { "group": "grove.io", "version": "v1alpha1", "kind": "PodCliqueScalingGroup" }
        ]
      },
      "optimizationInstructions": {
        "gangScheduling": {
          "podGroups": [
            {
              "name": "service",
              "members": [
                {
                  "componentName": "service",
                  "groupByKeyPaths": [
                    ".metadata.labels[\"nvidia.com/dynamo-component\"]"
                  ]
                }
              ]
            }
          ]
        }
      }
    }
  }
]

```

</Accordion>

<Accordion title="LeaderWorkerSet (LWS)">

**LeaderWorkerSet (LWS)**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "kind": {
                "group": "leaderworkerset.x-k8s.io",
                "kind": "LeaderWorkerSet",
                "version": "v1"
              },
              "name": "leaderworkerset",
              "statusDefinition": {
                "conditionsDefinition": {
                  "messageFieldName": "message",
                  "path": ".status.conditions",
                  "statusFieldName": "status",
                  "typeFieldName": "type"
                },
                "statusMappings": {
                  "failed": [
                    {
                      "byConditions": [
                        { "status": "False", "type": "Available" },
                        { "status": "False", "type": "Progressing" },
                        { "status": "False", "type": "UpdateInProgress" }
                      ]
                    }
                  ],
                  "initializing": [
                    {
                      "byConditions": [
                        { "status": "True", "type": "Progressing" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "status": "True", "type": "Available" },
                        { "status": "False", "type": "Progressing" }
                      ]
                    }
                  ]
                }
              }
            },
            "childComponents": [
              {
                "kind": {
                  "group": "apps",
                  "kind": "StatefulSet",
                  "version": "v1"
                },
                "name": "leader",
                "ownerRef": "leaderworkerset",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"leaderworkerset.sigs.k8s.io/worker-index\"]",
                    "value": "0"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.leaderWorkerTemplate.leaderTemplate"
                }
              },
              {
                "kind": {
                  "group": "apps",
                  "kind": "StatefulSet",
                  "version": "v1"
                },
                "name": "worker",
                "ownerRef": "leaderworkerset",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.annotations[\"leaderworkerset.sigs.k8s.io/leader-name\"]"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": "(.spec.replicas // 1) * ((.spec.leaderWorkerTemplate.size // 1) - 1)"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.leaderWorkerTemplate.workerTemplate"
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "group",
                  "members": [
                    {
                      "componentName": "leader",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"leaderworkerset.sigs.k8s.io/name\"]",
                        ".metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]"
                      ]
                    },
                    {
                      "componentName": "worker",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"leaderworkerset.sigs.k8s.io/name\"]",
                        ".metadata.labels[\"leaderworkerset.sigs.k8s.io/group-index\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="Inference Service (KServe)">

**Inference Service (KServe)**

```json
   "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "inferenceservice",
              "kind": {
                "group": "serving.kserve.io",
                "version": "v1beta1",
                "kind": "InferenceService"
              },
              "specDefinition": {
                "fragmentedPodSpecDefinition": {
                  "resourcesPath": ".spec.domain.resources",
                  "priorityClassNamePath": ".spec.priorityClassName",
                  "nodeAffinityPath": ".spec.affinity.nodeAffinity"
                }
              },
              "statusDefinition": {
                "conditionsDefinition": {
                  "path": ".status.conditions",
                  "typeFieldName": "type",
                  "statusFieldName": "status",
                  "messageFieldName": "message"
                },
                "statusMappings": {
                  "failed": [
                    {
                      "byConditions": [
                        { "type": "PredictorReady", "status": "False" },
                        { "type": "PredictorConfigurationReady", "status": "False" },
                        { "type": "RoutesReady", "status": "False" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "type": "PredictorReady", "status": "True" },
                        { "type": "RoutesReady", "status": "True" },
                        { "type": "LatestDeploymentReady", "status": "True" }
                      ]
                    }
                  ]
                }
              }
            },
            "childComponents": [
              {
                "name": "predictor",
                "kind": {
                  "group": "apps",
                  "version": "v1",
                  "kind": "Deployment"
                },
                "ownerRef": "inferenceservice",
                "specDefinition": {
                  "podSpecPath": ".spec.predictor",
                  "metadataPath": ".spec.predictor",
                  "fragmentedPodSpecDefinition": {
                    "containerPath": ".spec.predictor.model"
                  }
                },
                "scaleDefinition": {
                  "minReplicasPath": ".spec.predictor.minReplicas",
                  "maxReplicasPath": ".spec.predictor.maxReplicas"
                },
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"component\"]",
                    "value": "predictor"
                  }
                }
              },
              {
                "name": "transformer",
                "kind": {
                  "group": "apps",
                  "version": "v1",
                  "kind": "Deployment"
                },
                "ownerRef": "inferenceservice",
                "specDefinition": {
                  "podSpecPath": ".spec.transformer",
                  "metadataPath": ".spec.transformer"
                },
                "scaleDefinition": {
                  "minReplicasPath": ".spec.transformer.minReplicas",
                  "maxReplicasPath": ".spec.transformer.maxReplicas"
                },
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"component\"]",
                    "value": "transformer"
                  }
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "service",
                  "members": [
                    {
                      "componentName": "predictor",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"serving.kserve.io/inferenceservice\"]"
                      ]
                    },
                    {
                      "componentName": "transformer",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"serving.kserve.io/inferenceservice\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="PyTorchJob">

**PyTorchJob**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "pytorchjob",
              "kind": {
                "group": "kubeflow.org",
                "version": "v1",
                "kind": "PyTorchJob"
              },
              "statusDefinition": {
                "conditionsDefinition": {
                  "path": ".status.conditions",
                  "typeFieldName": "type",
                  "statusFieldName": "status",
                  "messageFieldName": "message"
                },
                "statusMappings": {
                  "completed": [
                    {
                      "byConditions": [
                        { "type": "Succeeded", "status": "True" }
                      ]
                    }
                  ],
                  "failed": [
                    {
                      "byConditions": [
                        { "type": "Failed", "status": "True" }
                      ]
                    }
                  ],
                  "initializing": [
                    {
                      "byConditions": [
                        { "type": "Created", "status": "True" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "type": "Running", "status": "True" }
                      ]
                    }
                  ],
                  "suspended": [
                    {
                      "byConditions": [
                        { "type": "Suspended", "status": "True" }
                      ]
                    }
                  ],
                  "suspending": [
                    {
                      "byExpression": {
                        "expression": ".spec.runPolicy.suspend == true and (.status.replicaStatuses // {} | to_entries | map(.value.active // 0) | add // 0) > 0",
                        "expectedResult": "true"
                      }
                    }
                  ]
                }
              },
              "suspendDefinition": {
                "suspendActions": [
                  { "path": ".spec.runPolicy.suspend", "value": "true" }
                ],
                "resumeActions": [
                  { "path": ".spec.runPolicy.suspend", "value": "false" }
                ]
              }
            },
            "childComponents": [
              {
                "name": "Master",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "pytorchjob",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
                    "value": "Master"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.pytorchReplicaSpecs.Master.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.pytorchReplicaSpecs.Master.template"
                }
              },
              {
                "name": "Worker",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "pytorchjob",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
                    "value": "Worker"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.pytorchReplicaSpecs.Worker.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.pytorchReplicaSpecs.Worker.template"
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "job",
                  "members": [
                    {
                      "componentName": "Master",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"training.kubeflow.org/job-name\"]"
                      ]
                    },
                    {
                      "componentName": "Worker",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"training.kubeflow.org/job-name\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="JaxJob">

**JaxJob**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "jaxjob",
              "kind": {
                "group": "kubeflow.org",
                "version": "v1",
                "kind": "JAXJob"
              },
              "statusDefinition": {
                "conditionsDefinition": {
                  "path": ".status.conditions",
                  "typeFieldName": "type",
                  "statusFieldName": "status",
                  "messageFieldName": "message"
                },
                "statusMappings": {
                  "completed": [
                    {
                      "byConditions": [
                        { "type": "Succeeded", "status": "True" }
                      ]
                    }
                  ],
                  "failed": [
                    {
                      "byConditions": [
                        { "type": "Failed", "status": "True" }
                      ]
                    }
                  ],
                  "initializing": [
                    {
                      "byConditions": [
                        { "type": "Created", "status": "True" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "type": "Running", "status": "True" }
                      ]
                    }
                  ]
                }
              }
            },
            "childComponents": [
              {
                "name": "Worker",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "jaxjob",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
                    "value": "Worker"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.jaxReplicaSpecs.Worker.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.jaxReplicaSpecs.Worker.template"
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "job",
                  "members": [
                    {
                      "componentName": "Worker",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"training.kubeflow.org/job-name\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="MPIJob">

**MPIJob**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "mpijob",
              "kind": {
                "group": "kubeflow.org",
                "version": "v1",
                "kind": "MPIJob"
              },
              "statusDefinition": {
                "conditionsDefinition": {
                  "path": ".status.conditions",
                  "typeFieldName": "type",
                  "statusFieldName": "status",
                  "messageFieldName": "message"
                },
                "statusMappings": {
                  "completed": [
                    {
                      "byConditions": [
                        { "type": "Succeeded", "status": "True" }
                      ]
                    }
                  ],
                  "failed": [
                    {
                      "byConditions": [
                        { "type": "Failed", "status": "True" }
                      ]
                    }
                  ],
                  "initializing": [
                    {
                      "byConditions": [
                        { "type": "Created", "status": "True" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "type": "Running", "status": "True" }
                      ]
                    }
                  ]
                }
              }
            },
            "childComponents": [
              {
                "name": "launcher",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "mpijob",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
                    "value": "launcher"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.mpiReplicaSpecs.Launcher.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.mpiReplicaSpecs.Launcher.template"
                }
              },
              {
                "name": "worker",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "mpijob",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"training.kubeflow.org/replica-type\"]",
                    "value": "worker"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.mpiReplicaSpecs.Worker.replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.mpiReplicaSpecs.Worker.template"
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "job",
                  "members": [
                    {
                      "componentName": "launcher",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"training.kubeflow.org/job-name\"]"
                      ]
                    },
                    {
                      "componentName": "worker",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"training.kubeflow.org/job-name\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="RayService">

**RayService**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "rayservice",
              "kind": {
                "group": "ray.io",
                "version": "v1",
                "kind": "RayService"
              },
              "statusDefinition": {
                "phaseDefinition": {
                  "path": ".status.state"
                },
                "statusMappings": {
                  "failed": [
                    { "byPhase": "failed" }
                  ],
                  "running": [
                    { "byPhase": "ready" }
                  ]
                }
              }
            },
            "childComponents": [
              {
                "name": "head",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "rayservice",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"ray.io/node-type\"]",
                    "value": "head"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": "1"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.rayClusterSpec.headGroupSpec.template"
                }
              },
              {
                "name": "worker",
                "instanceIdPath": ".spec.rayClusterSpec.workerGroupSpecs[].groupName",
                "kind": {
                  "group": "",
                  "version": "v1",
                  "kind": "Pod"
                },
                "ownerRef": "rayservice",
                "podSelector": {
                  "componentTypeSelector": {
                    "keyPath": ".metadata.labels[\"ray.io/node-type\"]",
                    "value": "worker"
                  },
                  "componentInstanceSelector": {
                    "idPath": ".metadata.labels[\"ray.io/group\"]"
                  }
                },
                "scaleDefinition": {
                  "replicasPath": ".spec.rayClusterSpec.workerGroupSpecs[].replicas"
                },
                "specDefinition": {
                  "podTemplateSpecPath": ".spec.rayClusterSpec.workerGroupSpecs[].template"
                }
              }
            ]
          },
          "optimizationInstructions": {
            "gangScheduling": {
              "podGroups": [
                {
                  "name": "service",
                  "members": [
                    {
                      "componentName": "head",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"ray.io/cluster\"]"
                      ]
                    },
                    {
                      "componentName": "worker",
                      "groupByKeyPaths": [
                        ".metadata.labels[\"ray.io/cluster\"]"
                      ]
                    }
                  ]
                }
              ]
            }
          }
        }
      }
    ]
  }'

```

</Accordion>

<Accordion title="Notebook">

**Notebook**

```json
    "kartas": [
      {
        "spec": {
          "structureDefinition": {
            "rootComponent": {
              "name": "notebook",
              "kind": {
                "group": "kubeflow.org",
                "version": "v1beta1",
                "kind": "Notebook"
              },
              "specDefinition": {
                "podTemplateSpecPath": ".spec.template"
              },
              "statusDefinition": {
                "conditionsDefinition": {
                  "path": ".status.conditions",
                  "typeFieldName": "type",
                  "statusFieldName": "status",
                  "messageFieldName": "message"
                },
                "statusMappings": {
                  "completed": [
                    {
                      "byConditions": [
                        { "type": "Succeeded", "status": "True" }
                      ]
                    }
                  ],
                  "failed": [
                    {
                      "byConditions": [
                        { "type": "Failed", "status": "True" }
                      ]
                    }
                  ],
                  "initializing": [
                    {
                      "byConditions": [
                        { "type": "Created", "status": "True" }
                      ]
                    }
                  ],
                  "running": [
                    {
                      "byConditions": [
                        { "type": "Running", "status": "True" }
                      ]
                    }
                  ]
                }
              }
            }
          },
          "optimizationInstructions": {}
        }
      }
    ]
  }'

```

</Accordion>
