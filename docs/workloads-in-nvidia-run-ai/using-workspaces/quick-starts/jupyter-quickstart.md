# Running Jupyter Notebooks Using Workspaces

This quick start provides a step-by-step walkthrough for running a Jupyter Notebook using workspaces.

A workspace contains the setup and configuration needed for building your model, including the container, images, data sets, and resource requests, as well as the required tools for the research, all in one place. See [Running workspaces](../running-workspace.md) for more information.

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../platform-management/aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 1 GPU.

<Info>

**Note**

Flexible workload submission is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload submission.

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

## Step 2: Submitting a Workspace

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
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the **‘jupyter-lab’** environment for your workspace (Image URL: `jupyter/scipy-notebook)`
   * If ‘jupyter-lab’ is not displayed in the gallery, follow the below steps to create a one-time environment configuration::
     * Enter the jupyter-lab **Image URL** - `jupyter/scipy-notebook`
     * Tools - Set the connection for your tool
       * Click **+TOOL**
       * Select **Jupyter** tool from the list
     *   Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:

         * Enter the command - `start-notebook.sh`
         * Enter the arguments - `--NotebookApp.token=''`

         **Note:** If [path-based routing](../../../installation-and-configuration/configuration/networking/external-access-to-containers.md#access-to-the-running-workloads-container) is enabled on the cluster, enter `--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''`.
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the **‘one-gpu’** compute resource for your workload.
   * If ‘one-gpu’ is not displayed, follow the below steps to create a one-time compute resource configuration:
     * Set **GPU devices** per pod - 1
     * Optional: set the **CPU compute** per pod - 0.1 cores (default)
     * Optional: set the **CPU memory** per pod - 100 MB (default)
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
8.  Select the **‘jupyter-lab’** environment for your workspace (Image URL: `jupyter/scipy-notebook`)

    * If ‘jupyter-lab’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW ENVIRONMENT**
      * Enter jupyter-lab as the **name** for the environment. The name must be unique.
      * Enter the jupyter-lab **Image URL** - `jupyter/scipy-notebook`
      * Tools - Set the connection for your tool
        * Click **+TOOL**
        * Select **Jupyter** tool from the list
      *   Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:

          * Enter the command - `start-notebook.sh`
          * Enter the arguments - `--NotebookApp.token=''`

          **Note:** If [path-based routing](../../../installation-and-configuration/configuration/networking/external-access-to-containers.md#access-to-the-running-workloads-container) is enabled on the cluster, enter `--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''`.
      * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
9.  Select the **‘one-gpu’** compute resource for your workspace

    * If ‘one-gpu’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW COMPUTE RESOURCE**
      * Enter one-gpu as the **name** for the compute resource. The name must be unique
      * Set **GPU devices** per pod - 1
      * Optional: set the **CPU compute per pod** - 0.1 cores (default)
      * Optional: set the **CPU memory per pod** - 100 MB (default)
      * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
10. Click **CREATE WORKSPACE**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../reference/cli/runai/):

```shell
runai project set "project-name"
runai workspace submit "workload-name" \
--image jupyter/scipy-notebook --gpu-devices-request 0 \
--command --external-url container=8888 -- start-notebook.sh \
--NotebookApp.token=
```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Workspaces](https://run-ai-docs.nvidia.com/api/2.26/workloads/workspaces) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/workspaces' \ 
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \ 
    "name": "workload-name", 
    "projectId": "<PROJECT-ID>",  
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "command" : "start-notebook.sh",
        "args" : "--NotebookApp.token=''",
        "image": "jupyter/scipy-notebook",
        "compute": {
            "gpuDevicesRequest": 1
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
* `<TOKEN>` - The API access token obtained in [Step 1](jupyter-quickstart.md#step-1-logging-in)
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

## Step 3: Connecting to the Jupyter Notebook

<Tabs>

<Tab title="UI">

1. Select the newly created workspace with the Jupyter application that you want to connect to
2. Click **CONNECT**
3. Select the Jupyter tool. The selected tool is opened in a new tab on your browser.

</Tab>

<Tab title="CLI v2">

To connect to the Jupyter Notebook, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`

</Tab>

<Tab title="CLI v1 (Deprecated)">

To connect to the Jupyter Notebook, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`

</Tab>

<Tab title="API">

To connect to the Jupyter Notebook, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`

</Tab>

</Tabs>

## Next Steps

Manage and monitor your newly created workload using the [Workloads](../../workloads.md) table.
