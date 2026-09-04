# Monitor Workloads by Category

A workload category represents the role or purpose of a workload such as training, building, or deploying models. Each workload type is automatically assigned a default category to ensure consistent classification across the platform.

Categories appear in the Overview dashboard, allowing administrators to filter, group, and monitor workloads based on their function. Administrators can modify the default category mapping for a workload type using the NVIDIA Run:ai API.

## Default Category Mapping

NVIDIA Run:ai defines the following default mappings of workload types to categories. To retrieve the default category per workload type, refer to the [List workload types](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#get-api-v1-workload-types) API.

<Info>

**Note**

* For more information on workload support, see [Introduction to workloads](../../workloads-in-nvidia-run-ai/introduction-to-workloads.md).
* To see the default priority assigned to each of the workload types listed below, refer to [Workload priority control](../runai-scheduler/scheduling/workload-priority-control.md).

</Info>

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

## Update the Default Category Mapping

Administrators can change the default category assigned to a workload type by updating the category mapping using the [NVIDIA Run:ai API](https://run-ai-docs.nvidia.com/api/2.26/). To update the category mapping:

1. Retrieve the list of workload types and their IDs using `GET /api/v1/workload-types`.
2. Identify the `workloadTypeId` of the workload type you want to modify.
3. Retrieve the list of available categories and their IDs using `GET /api/v1/workload-categories`.
4. Send a request to update the workload type with the new category using\
   `PUT /api/v1/workload-types/{workloadTypeId}` and include the `categoryId` in the request body.

## Using API

Go to the [Workload properties](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#get-api-v1-workload-categories) API reference to view the available actions.
