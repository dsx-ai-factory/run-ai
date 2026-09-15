# Launching Workloads with GPU Compute Modes

This quick start provides a step-by-step walkthrough for running two workspaces that share a single GPU using SM-sharing, and for verifying the compute mode after submission.

A GPU compute mode always accompanies a GPU memory fraction — the memory fraction determines how much GPU memory the workload receives, and the compute mode determines how the workloads sharing that GPU divide the GPU compute. For background on both modes, refer to [GPU compute modes](../compute-modes.md).

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 1 GPU.
* A GPU is available that is not currently running other fractional workloads.

<Info>

**Note**

* The compute mode is exclusive per GPU. The first workload placed on a GPU fixes the compute mode of that GPU, and every workload that joins afterwards must request the same mode. Starting this walkthrough on a GPU that is already running time-slicing workloads prevents the SM-sharing workspaces from being placed on it.
* Flexible workload submission is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload submission.

</Info>

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="CLI v2">

Run the below --help command to obtain the login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [API authentication](https://run-ai-docs.nvidia.com/api/2.26/getting-started/how-to-authenticate-to-the-api).

</Tab>

</Tabs>

## Step 2: Submitting the First Workspace

<Tabs>

<Tab title="UI">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Workspace**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workspace will run
5. Select **Start from scratch** to launch a new workspace quickly
6. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the **‘jupyter-lab’** environment for your workspace (Image URL: `jupyter/scipy-notebook`)
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. To add a new compute resource:
   * Click the **+** icon to create a new compute resource
   * Enter sm-sharing-fraction as the **name** for the compute resource. The name must be unique.
   * Set **GPU devices** per pod - 1
   * Enable **GPU fractioning** to set the GPU memory per device:
     * Select **GB** - Fraction of a GPU device’s memory
     * Set the memory **Request** - 20GB
   * Click **CREATE COMPUTE RESOURCE**
   * Select the newly created compute resource from the side pane
10. Open the **General** section and locate **Set annotation(s)**. Click **+ANNOTATION** and set the compute mode:
    * **Name** - `nvidia.com/container.<container-name>.gpu-compute.mode`
    * **Value** - `sm-sharing`
11. Click **CREATE WORKSPACE**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../../reference/cli/runai/):

```sh
runai project set "project-name"
runai workspace submit "workload-name-1" --image jupyter/scipy-notebook \
--gpu-memory-request 20G \
--annotation nvidia.com/container.workload-name-1.gpu-compute.mode=sm-sharing \
--external-url container=8888 --name-prefix jupyter \
--command -- start-notebook.sh \
--NotebookApp.token=
```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Workspaces](https://run-ai-docs.nvidia.com/api/2.26/workloads/workspaces) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/workspaces' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
    "name": "workload-name-1",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "command" : "start-notebook.sh",
        "args" : "--NotebookApp.token=''",
        "image": "jupyter/scipy-notebook",
        "annotations": [
            {
                "name": "nvidia.com/container.workload-name-1.gpu-compute.mode",
                "value": "sm-sharing"
            }
        ],
        "compute": {
            "gpuDevicesRequest": 1,
            "gpuRequestType": "memory",
            "gpuMemoryRequest": "20G"
        },
        "exposedUrls" : [
            {
                "container" : 8888,
                "toolType": "jupyter-notebook",
                "toolName": "Jupyter"
            }
        ]
    }
}
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](compute-modes-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.

</Tab>

</Tabs>

## Step 3: Submitting the Second Workspace

Repeat Step 2 with a different workload name, using the same compute resource and the same `sm-sharing` compute mode. Both workspaces request 20 GB of GPU memory, so both can be placed on the same GPU.

<Info>

**Note**

A workload requesting `time-slicing`, or a workload submitted without a compute mode annotation, is not placed on this GPU while the SM-sharing workspaces are running on it.

</Info>

## Step 4: Verifying the Compute Mode

<Tabs>

<Tab title="UI">

1. Go to the Workload manager → Workloads
2. Select one of the workspaces you created
3. Open the workload **details** and select the **YAML** view
4. Confirm that `nvidia.com/container.<container-name>.gpu-compute.mode` is present with the value `sm-sharing`

</Tab>

<Tab title="CLI v2">

Confirm the annotation is present on the pod:

```sh
kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.metadata.annotations}'
```

</Tab>

<Tab title="API">

Confirm the annotation is present on the pod:

```sh
kubectl get pod <pod-name> -n <namespace> -o jsonpath='{.metadata.annotations}'
```

</Tab>

</Tabs>

Both workspaces now run concurrently on separate streaming multiprocessors of the same GPU. Had the workloads used time-slicing, they would have taken sequential turns on the GPU compute engines instead.

## Next Steps

* Choose between the two modes for your own workloads using [GPU compute modes](../compute-modes.md).
* Manage and monitor your newly created workloads using the [Workloads](../../../../workloads-in-nvidia-run-ai/workloads.md) table.
