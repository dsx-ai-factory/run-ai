# Supported Workload Types

NVIDIA Run:ai supports a broad range of workloads from the ML and Kubernetes ecosystem that are already registered as workload types in the platform and ready to use. These workloads are managed by [Karta](extending-workload-support/#karta-overview), which ensures they receive the same advanced scheduling, orchestration, and monitoring capabilities as NVIDIA Run:ai native workloads. For more details on feature support, see [Supported features](supported-features.md).

All supported workload types are pre-registered using the [Workload Types](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#post-api-v1-workload-types) API. You can retrieve the full list of these workloads and their current configuration defaults.

<Info>

**Note**

To define and register workload types beyond those provided by NVIDIA Run:ai, see [Extending workload support with Karta](extending-workload-support/).

</Info>

## List of Supported Workload Types

The following workload types are already registered in the platform and ready to use:

* **NVIDIA** - [NIM Services](https://docs.nvidia.com/nim-operator/latest/service.html), [DynamoGraphDeployment](https://docs.nvidia.com/dynamo/latest/index.html)
* **Kubernetes** - LeaderWorkerSet (LWS), Deployment, StatefulSet, ReplicaSet, Pod, CronJob, Job ([kubernetes.io](https://kubernetes.io/))
* **Kubeflow** - TFJob, PyTorchJob, MPIJob, XGBoostJob, JAXJob, Notebook, ScheduledWorkflow ([kubeflow.org](http://kubeflow.org))
* **Ray** - RayService, RayCluster, RayJob ([ray.io](http://ray.io))
* **Tekton** - PipelineRun, TaskRun ([tekton.dev](http://tekton.dev))
* **Slinky** - [NodeSet](https://github.com/SlinkyProject/slurm-operator#nodesets)
* **Additional frameworks** - [SeldonDeployment](http://machinelearning.seldon.io), [AMLJob](https://learn.microsoft.com/en-us/cli/azure/ml/job?view=azure-cli-latest\&utm_source=chatgpt.com), [Workflow](http://argoproj.io), [DevWorkspace](http://workspace.devfile.io), [Service](http://serving.knative.dev), [VirtualMachineInstance](http://kubevirt.io/), InferenceService ([KServe](https://kserve.github.io/website/))

<Info>

**Scope of Slinky NodeSet Support**

NVIDIA Run:ai recognizes a [Slinky](https://github.com/SlinkyProject/slurm-operator) `NodeSet` - the group of Slurm compute node (`slurmd`) pods that the Slinky operator runs on Kubernetes - and manages it as a standard workload. The NodeSet receives the same scheduling, prioritization, preemption, and monitoring as any other supported workload type, and appears in the [Workloads](../workloads.md) table for management.

This support is scoped to the `NodeSet` workload as it runs on Kubernetes. It is not an integration with the Slurm control plane: NVIDIA Run:ai does not submit, schedule, or manage Slurm jobs, partitions, accounts, or the Slurm controller (`slurmctld`).

</Info>

## Submitting Supported Workload Types

Supported workload types are submitted as standard Kubernetes YAML manifest&#x73;**.** Once submitted, the workload is created and appears in the [Workloads](../workloads.md) table for monitoring and management. See [Submit supported workload types via YAML](../submit-via-yaml.md).

## Workload Types and Defaults

Each workload type is assigned a default category, which determines its default priority and preemptibility. These defaults influence how workloads are scheduled and prioritized within a project, as well as how they are grouped for monitoring and reporting.

<Info>

**Note**

To retrieve the default category, priority and preemptibility per workload type, refer to the [List workload types](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-properties#get-api-v1-workload-types) API.

</Info>

| Category | Default Priority | Default Preemptibility |
| -------- | ---------------- | ---------------------- |
| Build    | High             | Non-preemptible        |
| Train    | Low              | Preemptible            |
| Deploy   | Very high        | Non-preemptible        |

### Workload Types by Default Category

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

## Updating Default Category and Priority Mapping

Administrators can change the default priority and category assigned to a workload type by updating the mapping using the [NVIDIA Run:ai API](https://run-ai-docs.nvidia.com/api/2.26/):

* To update the priority mapping, see [Workload priority and preemption](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md#updating-the-default-mapping)
* To update the category mapping, see [Monitor workloads by category](../../platform-management/monitor-performance/workload-categories.md#update-the-default-category-mapping)
