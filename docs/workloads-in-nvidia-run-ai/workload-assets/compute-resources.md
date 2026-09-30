# Compute Resources

Compute resources are one type of [workload assets](overview.md). A compute resource is a template that simplifies how workloads are submitted and can be used by AI practitioners when they submit their workloads.

A compute resource asset is a preconfigured building block that encapsulates all the specifications of compute requirements for the workload including:

* GPU devices and GPU memory
* CPU memory and CPU compute

## Compute Resource Table

The Compute resource table can be found under **Workload manager** in the NVIDIA Run:ai UI.

The Compute resource table provides a list of all the compute resources defined in the platform and allows you to manage them.

<figure><img src="../../assets/compute-new.png" alt=""/><figcaption></figcaption></figure>

The Compute resource table consists of the following columns:

| Column                        | Description                                                                                                                                                       |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Compute resource              | The name of the compute resource                                                                                                                                  |
| Description                   | A description of the essence of the compute resource                                                                                                              |
| GPU devices request per pod   | The number of requested physical devices per pod of the workload that uses this compute resource                                                                  |
| GPU memory request per device | The amount of GPU memory per requested device that is granted to each pod of the workload that uses this compute resource                                         |
| CPU memory request            | The minimum amount of CPU memory per pod of the workload that uses this compute resource                                                                          |
| CPU memory limit              | The maximum amount of CPU memory per pod of the workload that uses this compute resource                                                                          |
| CPU compute request           | The minimum number of CPU cores per pod of the workload that uses this compute resource                                                                           |
| CPU compute limit             | The maximum number of CPU cores per pod of the workload that uses this compute resource                                                                           |
| Scope                         | The [scope](overview.md#asset-scope) of this compute resource within the organizational tree. Click the name of the scope to view the organizational tree diagram |
| Workload(s)                   | The list of workloads associated with the compute resource                                                                                                        |
| Template(s)                   | The list of workload templates that use this compute resource                                                                                                     |
| Created by                    | The name of the user who created the compute resource                                                                                                             |
| Creation time                 | The timestamp of when the compute resource was created                                                                                                            |
| Last updated                  | The timestamp of when the compute resource was last updated                                                                                                       |
| Cluster                       | The cluster that the compute resource is associated with                                                                                                          |

### Workloads Associated with the Compute Resource

Click one of the values in the Workload(s) column to view the list of workloads and their parameters.

| Column   | Description                                                                                                 |
| -------- | ----------------------------------------------------------------------------------------------------------- |
| Workload | The workload that uses the compute resource                                                                 |
| Type     | Workspace/Training/Inference                                                                                |
| Status   | Represents the workload lifecycle. See the full list of [workload status](../workloads.md#workload-status). |

### Customizing the Table View

* Filter - Click ADD FILTER, select the column to filter by, and enter the filter values
* Search - Click SEARCH and type the value to search by
* Sort - Click each column header to sort by
* Column selection - Click COLUMNS and select the columns to display in the table

## Adding a New Compute Resource

To add a new compute resource:

1. Go to the Compute resource table
2. Click **+NEW COMPUTE RESOURCE**
3. Select under which **cluster** to create the compute resource
4. Select a [**scope**](overview.md#asset-scope)
5. Enter a **name** for the compute resource. The name must be unique.
6. Optional: Provide a **description** of the essence of the compute resource
7. Set the resource types needed within a single node. The NVIDIA Run:ai Scheduler tries to match a single node that complies with the compute resource for each of the workload’s pods.
   *   **GPU devices** - The number of devices (physical GPUs) per pod. For example, if you requested 3 devices per pod and the running workload using this compute resource consists of 3 pods, there are 9 physical GPU devices used in total.

       <Note><p><strong>Note</strong></p><ul><li>When setting it to zero, the workload using this computer resource neither requests nor uses GPU resources while running.</li><li>You can set any number of GPU devices and specify the memory requirement to any portion size (1..100), or memory size value using GB or MB units per device.</li></ul></Note>
   * Enable **GPU fractioning** to allocate a portion of the GPU memory to each device
     * Select the memory request format
       * **% (of device)** - Fraction of a GPU device’s memory
       * **MB (memory size)** - An explicit GPU memory unit
       * **GB (memory size)** - An explicit GPU memory unit
     * Set the memory **Request** - The minimum amount of GPU memory that is provisioned per device. This means that any pod of a running workload that uses this compute resource, receives this amount of GPU memory for each device(s) the pod utilizes.
     *   Optional: Set the memory **Limit** - The maximum amount of GPU memory that is provisioned per device. This means that any pod of a running workload that uses this compute resource, receives **at most** this amount of GPU memory for each device(s) the pod utilizes. The limit value must be equal to or higher than the request.

         <Note><p><strong>Note</strong></p><ul><li>GPU memory limit is disabled by default. If you cannot see <strong>Limit</strong> in the compute resource form, then it must be enabled by your Administrator, under <strong>General settings</strong> → Resources → GPU resource optimization.</li><li>When a <strong>Limit</strong> is set and is bigger than the <strong>Request</strong>, the Scheduler allows each pod to reach the maximum amount of GPU memory in an opportunistic manner (only upon availability).</li><li>If the GPU memory <strong>Limit</strong> is bigger that the <strong>Request</strong>, the pod is prone to be killed by the NVIDIA Run:ai toolkit (out of memory signal). The greater the difference between the GPU memory used and the request, the higher the risk of being killed.</li><li>If GPU resource optimization is turned off, the minimum and maximum are in fact equal.</li></ul></Note>
   * **CPU resources**
     * Select the units for the CPU compute (Cores / Millicores)
       * Set the CPU compute **Request** - The minimum amount of CPU compute that is provisioned per pod. This means that any pod of a running workload that uses this compute resource, receives this amount of CPU compute for each pod.
       * Optional: Set the CPU compute **Limit** - The maximum amount of CPU compute that is provisioned per pod. This means that any pod of a running workload that uses this compute resource, receives **at most** this amount of CPU compute. The limit value must be equal to or higher than the request. By default, the limit is set to “Auto” which means that the pod may consume up to the node's maximum available CPU compute resources.
     * Select the units for the CPU memory (MB / GB)
       * Set the CPU memory **Request** - The minimum amount of CPU memory that is provisioned per pod. This means that any pod of a running workload that uses this compute resource, receives this amount of CPU memory for each pod.
       *   Optional: Set the CPU memory **Limit** - The maximum amount of CPU memory that is provisioned per pod. This means that any pod of a running workload that uses this compute resource, receives **at most** this amount of CPU memory. The limit value must be equal to or higher than the request. By default, the limit is set to “Auto”, which means that the pod may consume up to the node's maximum available CPU memory resources.

           <Note><p><strong>Note</strong></p><p>If the CPU Memory <strong>Limit</strong> is bigger that the <strong>Request</strong>, the pod is prone to be killed by the operating system (out of memory signal). The greater the difference between the CPU memory used and the request, the higher the risk of being killed.</p></Note>
8. Optional: More settings
   * **Increase shared memory size**\
     When enabled, the shared memory size available to the pod is increased from the default 64MB to the node's total available memory or the CPU memory limit, if set above.
   * **Set extended resource(s)**\
     Click **+EXTENDED RESOURCES** to add resource/quantity pairs. For more information on how to set extended resources, see the [Extended resources](https://kubernetes.io/docs/tasks/configure-pod-container/extended-resource/) and [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) guides
9. Click **CREATE COMPUTE RESOURCE**

<Note>

**Note**

It is also possible to add compute resources directly when creating a specific [workspace](../using-workspaces/running-workspace.md), [training](../using-training/train-models.md) or [inference](../using-inference/nvidia-run-ai-inference-overview.md) workload.

</Note>

## Editing a Compute Resource

To edit a compute resource:

1. Select the compute resource you want to edit
2. Click **Edit**
3. Update the compute resource and click **SAVE COMPUTE RESOURCE**

<Note>

**Note**

The already bound workload that is using this asset will not be affected.

</Note>

## Copying a Compute Resource

To copy an existing compute resource:

1. Select the compute resource you want to copy
2. Click **MAKE A COPY**
3. Enter a **name** for the compute resource. The name must be unique.
4. Update the compute resource and click **CREATE COMPUTE RESOURCE**

## Deleting a Compute Resource

1. Select the compute resource you want to delete
2. Click **DELETE**
3. On the dialog, click **DELETE** to confirm

<Note>

**Note**

The already bound workload that is using this asset will not be affected.

</Note>

## Using CLI

To view the available actions on compute resources, see the compute resource [CLI v2 reference](../../reference/cli/runai/runai_compute.md).

## Using API

Go to the [Compute resources](api:main-api:GET/api/v1/asset/compute) API reference to view the available actions
