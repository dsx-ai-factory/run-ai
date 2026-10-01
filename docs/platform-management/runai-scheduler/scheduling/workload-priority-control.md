# Workload Priority and Preemption

NVIDIA Run:ai defines workload priority and preemptibility to determine how workloads are scheduled within a project. These mechanisms influence scheduling order, resource allocation, and whether running workloads may be interrupted when higher-priority workloads require resources.

## How Priority and Preemptibility Work

* **Workload priority** - Determines the workload's position in the project scheduling queue managed by the NVIDIA Run:ai Scheduler. By adjusting the priority, you can increase the likelihood that a workload will be scheduled and preferred over others within the same project, ensuring that critical tasks are given higher priority and resources are allocated efficiently.
* **Workload preemptibility** - Determines the workload's resource usage policy and its guarantee against interruption:
  * Preemptible workloads can use opportunistic resources beyond the project’s quota and may be interrupted at any time by higher priority workload, even if running within the project's quota.
  * Semi-preemptible workloads guarantee their minimum required replicas run within the project's deserved quota and are not interrupted. Any replicas beyond that minimum behave as preemptible replicas - they can use opportunistic resources beyond the project's quota, and may be interrupted at any time by a higher priority workload.
  * Non-preemptible workloads must run within the project’s deserved quota, cannot use over-quota resources, and will not be interrupted once scheduled.

Refer to the [NVIDIA Run:ai Scheduler: Concepts and Principles](concepts-and-principles.md#priority-and-preemption) doc for more information.

<Note>

**Note**

This applies only within a single project. It does not impact the scheduling queues or workloads of other projects.

</Note>

## Priority Dictionary

Workload priority is defined by selecting a priority from a predefined list in the NVIDIA Run:ai priority dictionary. Each string corresponds to a specific Kubernetes [PriorityClass](concepts-and-principles.md#priority-and-preemption), which in turn determines scheduling behavior.

| Priority Class Name | Kubernetes Priority Value |
| --- | --- |
| `very-low` | 25 |
| `low` | 40 |
| `medium-low` | 65 |
| `medium` | 80 |
| `medium-high` | 90 |
| `high` | 125 |
| `very-high` | 150 |

## Default Priority and Preemptibility per Workload

NVIDIA Run:ai defines the following default mappings of workload types to priorities and preemptibility. Each workload type comes with a default category that determines it default priority and preemptibility value. To retrieve the default priority and preemptibility per workload type, refer to the [List workload types](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#get-api-v1-workload-types) API.

<Note>

**Note**

* For more information on workload support, see [Introduction to workloads](../../../workloads-in-nvidia-run-ai/introduction-to-workloads.md).
* Changing the priority is not supported for NVCF workloads.

</Note>

| Category | Default Priority | Default Preemptibility |
| -------- | ---------------- | ---------------------- |
| Build    | High             | Non-preemptible        |
| Train    | Low              | Preemptible            |
| Deploy   | Very high        | Non-preemptible        |

### NVIDIA Run:ai Native Workloads

| Workload Type | Build | Train | Deploy |
| --- | --- | --- | --- |
| Workspaces | ✓ | ✗ | ✗ |
| Standard training | ✗ | ✓ | ✗ |
| Distributed training | ✗ | ✓ | ✗ |
| Custom inference | ✗ | ✗ | ✓ |
| NVIDIA NIM inference | ✗ | ✗ | ✓ |
| Hugging Face inference | ✗ | ✗ | ✓ |
| Distributed inference | ✗ | ✗ | ✓ |

### Supported Workload Types

| Workload Type | Build | Train | Deploy |
| --- | --- | --- | --- |
| AMLJob | ✗ | ✓ | ✗ |
| CronJob | ✗ | ✗ | ✓ |
| Deployment | ✗ | ✗ | ✓ |
| DevWorkspace | ✓ | ✗ | ✗ |
| DynamoGraphDeployment | ✗ | ✗ | ✓ |
| InferenceService (KServe) | ✗ | ✗ | ✓ |
| JAXJob | ✗ | ✓ | ✗ |
| Job | ✓ | ✗ | ✗ |
| JobSet | ✗ | ✓ | ✗ |
| LeaderWorkerSet (LWS) | ✗ | ✗ | ✓ |
| MPIJob | ✗ | ✓ | ✗ |
| NIMCache | ✗ | ✗ | ✓ |
| NIMServices | ✗ | ✗ | ✓ |
| NodeSet (Slinky) | ✗ | ✗ | ✓ |
| Notebook | ✓ | ✗ | ✗ |
| PipelineRun | ✗ | ✗ | ✓ |
| Pod | ✗ | ✗ | ✓ |
| PyTorchJob | ✗ | ✓ | ✗ |
| RayCluster | ✗ | ✓ | ✗ |
| RayJob | ✗ | ✓ | ✗ |
| RayService | ✗ | ✗ | ✓ |
| ReplicaSet | ✗ | ✗ | ✓ |
| ScheduledWorkflow | ✗ | ✗ | ✓ |
| SeldonDeployment | ✗ | ✗ | ✓ |
| Service | ✗ | ✗ | ✓ |
| SPOTRequest | ✗ | ✗ | ✗ |
| StatefulSet | ✗ | ✗ | ✓ |
| TaskRun | ✓ | ✗ | ✗ |
| TFJob | ✗ | ✓ | ✗ |
| VirtualMachineInstance | ✗ | ✓ | ✗ |
| Workflow | ✗ | ✗ | ✓ |
| XGBoostJob | ✗ | ✓ | ✗ |

## Setting Priority and Preemptibility During Workload Submission

<Note>

**Note**

* If preemptibility is not explicitly configured, the system uses the default preemptibility behavior associated with the selected workload priority.
* Changing a workload’s priority and preemptibility may impact its ability to be scheduled. For example, switching a workload from a low priority, preemptible value (which allows over-quota usage) to high priority, non-preemptible value (which requires in-quota resources) may reduce its chances of being scheduled in cases where the required quota is unavailable.

</Note>

### NVIDIA Run:ai Native Workloads

For native NVIDIA Run:ai workloads, priority and preemptibility can be set during workload submission using one of the following methods:

* **UI** - Set workload priority and preemptibility under **General** settings
* **API** - Set using the `priorityClass` and `preemptibility` field
* **CLI** - Set using the `--priority` and `--preemptibility` flag

### Supported Workload Types

<Note>

**Note**

If priority or preemptibility is set through the UI, API, or CLI, those values override any values defined in the YAML manifest.

</Note>

For [supported workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md) submitted with a YAML manifest, priority and preemptibility can be set as follows:

* **UI** - Set workload priority and preemptibility under **General** settings
* **API** - Set using the `priority` and `preemptibility` fields
* **CLI** - Set using the `--priority` and `--preemptibility` flags
* **via YAML manifest** - Set by adding the following labels to your YAML manifest under the `metadata.labels` section of your workload definition.
  * Use the following values for priority - `very-low`, `low`, `medium-low`, `medium`, `medium-high`, `high`, `very-high` :
  *   Use the following values for preemptibility - `preemptible`, `semi-preemptible`, or `non-preemptible`

      ```yaml
      metadata:
        labels:
          priorityClassName: <priority>
          kai.scheduler/preemptibility: <preemptibility_value>
      ```

## Updating the Default Mapping

Administrators can change the default priority and preemptibility assigned to a workload type by updating the mapping using the [NVIDIA Run:ai API](https://run-ai-docs.nvidia.com/api/2.26/). To update the priority mapping:

1. Retrieve the list of workload types and their IDs using `GET /api/v1/workload-types`.
2. Identify the `workloadTypeId` of the workload type you want to modify.
3. Retrieve the list of available priorities and their IDs using `GET /api/v1/workload-priorities`.
4. Send a request to update the workload type with the new priority using\
   `PUT /api/v1/workload-types/{workloadTypeId}` and include the `priorityId` in the request body.

## Using API

Go to the [Workload priorities](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#get-api-v1-workload-priorities) API reference to view the available actions.
