# Deploy NVIDIA Cloud Functions (NVCF) in NVIDIA Run:ai

NVIDIA Cloud Functions (NVCF) is a serverless API platform designed to deploy and manage AI workloads on GPUs. Through its integration with NVIDIA Run:ai, NVCF can be deployed directly onto NVIDIA Run:ai-managed GPU clusters. This allows users to take advantage of NVIDIA Run:ai's scheduling, quota management, and monitoring features. See [Supported features](nvcf.md#supported-features) for more details.

This guide provides the required steps for integrating NVIDIA Cloud Functions with the NVIDIA Run:ai platform.

## Workload Priority

By default, inference workloads in NVIDIA Run:ai are assigned a priority of `very-high`, which is non-preemptible. This behavior ensures that inference workloads, which often serve real-time or latency-sensitive traffic, are guaranteed the resources they need and will not be disrupted by other workloads. For more details, see [Workload priority control](../../platform-management/runai-scheduler/scheduling/workload-priority-control.md).

<Info>

**Note**

Changing the priority is not supported for NVCF workloads.

</Info>

## Setup

Follow the official instructions provided in the [NVIDIA Cloud Functions](https://docs.nvidia.com/cloud-functions/user-guide/latest/cloud-function/overview.html) documentation.

### Setting up a Cluster in NVCF

Cloud Functions administrators can install the NVIDIA Cluster Agent to enable existing GPU clusters as deployment targets for NVCF functions. Once installed, the cluster appears as a deployment option in the API and Cloud Functions menu, allowing authorized functions to deploy on it. See [Cluster Setup & Management](https://docs.nvidia.com/cloud-functions/user-guide/latest/cloud-function/cluster-management.html) to register, configure and verify the cluster.

### Setting up a Project in NVIDIA Run:ai

Once the cluster is registered to NVCF and appears as ready, create a project in the NVIDIA Run:ai UI with an **NVCF namespace**:

1. Follow the instructions detailed in [Projects](../../platform-management/aiinitiatives/organization/projects.md) to create a new project.
2. When setting the **Namespace**, choose **"Enter existing namespace from the cluster"** and enter **`nvcf-backend`**.
3. Assign the necessary resource quotas to the project.

<Info>

**Note**

* This is the designated NVIDIA Run:ai project for all NVCF functions; other projects will not be used.
* NVCF is assigned to a specific project and adheres to the project's resource quota. Ensure that your project has sufficient quota allocated to accommodate.

</Info>

## Deploying a Function

In NVCF, function creation defines the code and resources while deployment registers the function to a GPU cluster, making it available for execution:

1. Create the function as detailed in [Function Creation](https://docs.nvidia.com/cloud-functions/user-guide/latest/cloud-function/function-creation.html#function-creation).
2. Deploy the function as detailed in [Function Deployment](https://docs.nvidia.com/cloud-functions/user-guide/latest/cloud-function/function-deployment.html).

<Info>

**Note**

Using a custom Helm chart when creating a [Cloud Function](https://docs.nvidia.com/cloud-functions/user-guide/latest/cloud-function/function-creation.html) is not supported.

</Info>

## Managing and Monitoring

After the NVCF function is deployed, it is added to the [Workloads](../workloads.md) table, where it can be managed and monitored:

* Monitor resource usage, performance, and execution status.
* Manage workload lifecycle, including scaling, logging, and troubleshooting.

## Supported Features

| NVIDIA Run:ai Functionality                                                                                                                    | NVCF |
| ---------------------------------------------------------------------------------------------------------------------------------------------- | :--: |
| [Fairness](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#fairness-fair-resource-distribution)                |   v  |
| [Priority and preemption](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#priority-and-preemption)             |   v  |
| [Over quota](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#over-quota)                                       |   v  |
| [Node pools](../../platform-management/aiinitiatives/resources/node-pools.md)                                            |   v  |
| [Bin packing / Spread](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#placement-strategy-bin-pack-and-spread) |   v  |
| [Multi-GPU fractions](../../platform-management/runai-scheduler/resource-optimization/memory-fractions.md)                                            |      |
| [Multi-GPU dynamic fractions](../../platform-management/runai-scheduler/resource-optimization/dynamic-memory-fractions.md)                            |      |
| [Node level scheduler](../../platform-management/runai-scheduler/resource-optimization/node-level-scheduler.md)                                |   v  |
| [Multi-GPU memory swap](../../platform-management/runai-scheduler/resource-optimization/memory-swap.md)                                        |      |
| [Gang scheduling](../../platform-management/runai-scheduler/scheduling/concepts-and-principles.md#gang-scheduling)                             |   v  |
| [Monitoring](../../installation-and-configuration/operations/observability/system-monitoring.md)                                                                       |   v  |
| [RBAC](../../platform-management/authentication/overview.md#role-based-access-control-rbac-in-run-ai)                                         |   v  |
| Workload awareness                                                                                                                             |   v  |
| [Workload submission](../workloads.md)                                                                                                         |      |
| [Workload actions (stop/run)](../workloads.md)                                                                                                 |      |
| [Workload Policies](../../platform-management/policies/native-workload-policies.md)                                                                   |      |
| [Scheduling rules](../../platform-management/policies/scheduling-rules.md)                                                                     |      |
