# Extending Workload Support with Karta

Karta in NVIDIA Run:ai provides a declarative way to extend platform support for new workload types. It enables organizations to quickly incorporate emerging ML frameworks, tools, or Kubernetes resources without requiring platform updates or code changes. Administrators can introduce new workload types at any time through the [Workload Types](api:main-api:POST/api/v1/workload-types) API.

Once registered, these workloads become available across the organization, enabling teams to innovate and collaborate. Practitioners can then submit them using the [via YAML](../../submit-via-yaml.md) submission and manage them with the same orchestration, monitoring, and scheduling capabilities as native workloads. For details on feature support, see [Supported features](../supported-features.md).

<Note>

**Note**

NVIDIA Run:ai supports a broad range of workloads from the ML and Kubernetes ecosystem that are already registered as workload types in the platform and ready to use. See [Supported workload types](../supported-workload-types.md).

</Note>

## Karta <a href="#karta-overview" id="karta-overview"></a>

Karta provides a YAML-based interface for defining how NVIDIA Run:ai should interpret, optimize, and monitor new workload types, without requiring platform updates or code changes. For more details on the Karta structure, see [Defining a Karta](defining-a-karta.md#karta-structure).

### Core Functions <a href="#core-functions-of-karta" id="core-functions-of-karta"></a>

* **Structure Awareness** - Allow NVIDIA Run:ai to interpret the full composition of a workload, including component hierarchy, dependencies, and scaling logic. This ensures the platform can organize and optimize any supported workload, independent of framework or origin.
* **Monitoring and Status Mapping** - Defines how to track workload health by mapping specific status conditions from the framework to standard, abstract states (such as running, succeeded, failed). This mapping supports robust monitoring and lifecycle automation across diverse workload types.
* **Optimization Directives** - Encodes optimization strategies such as gang scheduling. These directives help ensure that workloads are utilized efficiently and reliably in any infrastructure environment.

### Structure Overview <a href="#karta-structure-overview" id="karta-structure-overview"></a>

A typical Karta manifest includes the following primary sections:

* **structureDefinition** - Specifies all components (root, children), their types, hierarchical relationships, and how to interpret their specs within the overall workload.
* **optimizationInstructions** - Encapsulates how workloads should be scheduled and optimized (e.g., using gang scheduling).
* **scaleDefinition** - Explains how workload components can be scaled (manual, auto-scaling boundaries).
* **statusDefinition** - Maps resource-specific conditions to standard running/completed/failed states for unified monitoring.

## Registering New Workload Types

The [Workload Types](api:main-api:POST/api/v1/workload-types) API allows administrators to register and manage new workload types by providing the required details, such as workload name, supported CRD versions, category, priority, and Kubernetes group. For more details on registering new workload types, see [Defining a Karta](defining-a-karta.md).

`POST /api/v1/workload-types` with the following:

* `categoryId` - The identifier of the workload category.
* `priorityId` - The identifier of the workload priority.
* `preemptibility` - The preemptibility value of the workload - `non-preemptible` / `preemptible`.
* `name` - The unique name of the workload type. This value must exactly match the Kubernetes Kind that represents the workload type.
* `group` - The Kubernetes group associated with the workload resource.
* `kartas` - Lists the versions of the custom resource definition (CRD) supported for this workload type, such as `v1`, `v1beta1`, or `v1alpha1`. This enables the platform to correctly parse, interpret, and manage manifests for this workload type according to the specific structure and schema associated with each listed version. On update, you may only add or remove supported versions, modifying existing version entries is not allowed.

Once a new workload type is registered, you can also apply workload policies to it by registering a spec mapping with the policy engine. See [Workload Type Policy Mapping](../../../platform-management/policies/workload-type-policy-mapping.md).

**Example POST request body:**

```bash
{
  "categoryId": "337f5e5d-288b-40d5-be14-901cc3acacc0",
  "priorityId": "a57eab25-838b-40cc-a576-57e4056f1d6c",
  "preemptibility": "non-preemptible"
  "name": "Deployment",
  "group": "apps",
  "kartas": [
    {
      "spec": {
        "structureDefinition": {
          "rootComponent": {
            "kind": {
              "group": "apps",
              "version": "v1",
              "kind": "Deployment"
             }
           }
        }
      }
    }
  ]
}
```
