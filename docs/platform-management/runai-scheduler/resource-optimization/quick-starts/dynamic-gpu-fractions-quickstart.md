# Launching Workloads with Dynamic GPU Fractions

This quick start provides a step-by-step walkthrough for running a Jupyter Notebook with [dynamic GPU fractions](../dynamic-fractions.md).

NVIDIA Run:ai’s dynamic GPU fractions optimizes GPU utilization by enabling workloads to dynamically adjust their resource usage. It allows users to specify a guaranteed fraction of GPU memory and compute resources with a higher limit that can be dynamically utilized when additional resources are requested.

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 0.5 GPU.
* [Dynamic GPU fractions](../dynamic-fractions.md) is enabled.

<Info>

**Note**

* Flexible workload submission is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload submission.
* Dynamic GPU fractions is disabled by default in the NVIDIA Run:ai UI. To use dynamic GPU fractions, it must be enabled by your administrator, under **General Settings** → Resources → GPU resource optimization.

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

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Workspace**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workspace will run
5. Select **Start from scratch** to launch a new workspace quickly
6. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. To add a new environment:
   * Click the **+** icon to create a new environment
   * Enter quick-start as the **name** for the environment. The name must be unique.
   * Enter the **Image URL** - `gcr.io/run-ai-lab/pytorch-example-jupyter`
   * Tools - Set the connection for your tool:
     * Click **+TOOL**
     * Select **Jupyter** tool from the list
   *   Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:

       * Enter the command - `start-notebook.sh`
       * Enter the arguments - `--NotebookApp.token=''`

       **Note:** If [path-based routing](../../../../installation-and-configuration/configuration/networking/external-access-to-containers.md#access-to-the-running-workloads-container) is enabled on the cluster, enter `--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''`.
   * Click **CREATE ENVIRONMENT**
   * Select the newly created environment from the side pane
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. To add a new compute resource:
   * Click the **+** icon to create a new compute resource
   * Enter request-limit as the **name** for the compute resource. The name must be unique.
   * Set **GPU devices** per pod - 1
   * Enable **GPU fractioning** to set the GPU memory per device:
     * Select **GB** **-** Fraction of a GPU device’s memory
     * Set the memory **Request** - 4GB (the workload will allocate 4GB of the GPU memory)
     * Set the memory **Limit** - 12GB
   * Optional: set the **CPU compute per pod** - 0.1 cores (default)
   * Optional: set the **CPU memory per pod** - 100 MB (default)
   * Select **More settings** and toggle **Increase shared memory size**
   * Click **CREATE COMPUTE RESOURCE**
   * Select the newly created compute resource from the side pane
10. Click **CREATE WORKSPACE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Workspace**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workspace will run
5. Select **Start from scratch** to launch a new workspace quickly
6. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8.  Create an environment for your workspace

    * Click **+NEW ENVIRONMENT**
    * Enter quick-start as the **name** for the environment. The name must be unique.
    * Enter the **Image URL** - `gcr.io/run-ai-lab/pytorch-example-jupyter`
    * Tools - Set the connection for your tool
      * Click **+TOOL**
      * Select **Jupyter** tool from the list
    *   Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:

        * Enter the command - `start-notebook.sh`
        * Enter the arguments - `--NotebookApp.token=''`

        **Note:** If [path-based routing](../../../../installation-and-configuration/configuration/networking/external-access-to-containers.md#access-to-the-running-workloads-container) is enabled on the cluster, enter `--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''`.
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
9.  Create a new “**request-limit**” compute resource for your workspace

    * Click **+NEW COMPUTE RESOURCE**
    * Enter request-limit as the **name** for the compute resource. The name must be unique.
    * Set **GPU devices** per pod - 1
    * Enable **GPU fractioning** to set the GPU memory per device:
      * Select **GB** **-** Fraction of a GPU device’s memory
      * Set the memory **Request** - 4GB (the workload will allocate 4GB of the GPU memory)
      * Set the memory **Limit** - 12GB
    * Optional: set the **CPU compute per pod** - 0.1 cores (default)
    * Optional: set the **CPU memory per pod** - 100 MB (default)
    * Select **More settings** and toggle **Increase shared memory size**
    * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
10. Click **CREATE WORKSPACE**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../../reference/cli/runai/):

```sh
runai project set "project-name"
runai workspace submit "workload-name" \
--image gcr.io/run-ai-lab/pytorch-example-jupyter \
--gpu-memory-request 4G --gpu-memory-limit 12G --large-shm \
--external-url container=8888 --name-prefix jupyter  \
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
    "name": "workload-name", 
    "projectId": "<PROJECT-ID>", 
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "command" : "start-notebook.sh",
        "args" : "--NotebookApp.token=''",
        "image": "gcr.io/run-ai-lab/pytorch-example-jupyter",
        "compute": {
            "gpuDevicesRequest": 1,
            "gpuMemoryRequest": "4G",
            "gpuMemoryLimit": "12G",
            "largeShmRequest": true

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
* `<TOKEN>` - The API access token obtained in [Step 1](dynamic-gpu-fractions-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.
* `toolType` will show the Jupyter icon when connecting to the Jupyter tool via the user interface.
* `toolName` will show when connecting to the Jupyter tool via the user interface.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 3: Submitting the Second Workspace

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Workspace**
3. Select the **cluster** where the previous workspace was created
4. Select the **project** where the previous workspace was created
5. Select **Start from scratch** to launch a new workspace quickly
6. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the environment created in [Step 2](dynamic-gpu-fractions-quickstart.md#step-2-submitting-the-first-workspace).
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the compute resources created in [Step 2](dynamic-gpu-fractions-quickstart.md#step-2-submitting-the-first-workspace).
10. Click **CREATE WORKSPACE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Workspace**
3. Select the **cluster** where the previous workspace was created
4. Select the **project** where the previous workspace was created
5. Select **Start from scratch** to launch a new workspace quickly
6. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Select the environment created in [Step 2](dynamic-gpu-fractions-quickstart.md#step-2-submitting-the-first-workspace)
9. Select the compute resource created in [Step 2](dynamic-gpu-fractions-quickstart.md#step-2-submitting-the-first-workspace)
10. Click **CREATE WORKSPACE**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../../reference/cli/runai/):

```sh
runai project set "project-name"
runai workspace submit "workload-name" \
--image gcr.io/run-ai-lab/pytorch-example-jupyter --gpu-memory-request 4G \
--gpu-memory-limit 12G --large-shm --external-url container=8888 \
--name-prefix jupyter --command -- start-notebook.sh \
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
    "name": "workload-name", 
    "projectId": "<PROJECT-ID>", 
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "command" : "start-notebook.sh",
        "args" : "--NotebookApp.token=''",
        "image": "gcr.io/run-ai-lab/pytorch-example-jupyter",
        "compute": {
            "gpuDevicesRequest": 1,
            "gpuMemoryRequest": "4G",
            "gpuMemoryLimit": "12G",
            "largeShmRequest": true

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
* `<TOKEN>` - The API access token obtained in [Step 1](dynamic-gpu-fractions-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.
* `toolType` will show the Jupyter icon when connecting to the Jupyter tool via the user interface.
* `toolName` will show when connecting to the Jupyter tool via the user interface.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 4: Connecting to the Jupyter Notebook

<Tabs>

<Tab title="UI">

1. Select the newly created workspace with the Jupyter application that you want to connect to
2. Click **CONNECT**
3. Select the Jupyter tool. The selected tool is opened in a new tab on your browser.
4. Open a terminal and use the `watch nvidia-smi` to get a constant reading of the memory consumed by the pod. Note that the number shown in the memory box is the Limit and not the Request or Guarantee.
5. Open the file `Untitled.ipynb` and move the frame so you can see both tabs
6. Execute both cells in `Untitled.ipynb`. This will consume about **3 GB of GPU memory** and be well below the 4GB of the **GPU Memory Request** value.
7. In the second cell, edit the value after `--image-size` from 100 to 200 and run the cell. This will increase the GPU memory utilization to about 11.5 GB which is above the Request value.

</Tab>

<Tab title="CLI v2">

1. To connect to the Jupyter Notebook, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`
2. Open a terminal and use the `watch nvidia-smi` to get a constant reading of the memory consumed by the pod. Note that the number shown in the memory box is the Limit and not the Request or Guarantee.
3. Open the file `Untitled.ipynb` and move the frame so you can see both tabs
4. Execute both cells in `Untitled.ipynb`. This will consume about **3 GB of GPU memory** and be well below the 4GB of the **GPU Memory Request** value.
5. In the second cell, edit the value after `--image-size` from 100 to 200 and run the cell. This will increase the GPU memory utilization to about 11.5 GB which is above the Request value.

</Tab>

<Tab title="API">

1. To connect to the Jupyter Notebook, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`
2. Open a terminal and use the `watch nvidia-smi` to get a constant reading of the memory consumed by the pod. Note that the number shown in the memory box is the Limit and not the Request or Guarantee.
3. Open the file `Untitled.ipynb` and move the frame so you can see both tabs
4. Execute both cells in `Untitled.ipynb`. This will consume about **3 GB of GPU memory** and be well below the 4GB of the **GPU Memory Request** value.
5. In the second cell, edit the value after `--image-size` from 100 to 200 and run the cell. This will increase the GPU memory utilization to about 11.5 GB which is above the Request value.

</Tab>

</Tabs>

## Next Steps

Manage and monitor your newly created workload using the [Workloads](../../../../workloads-in-nvidia-run-ai/workloads.md) table.
