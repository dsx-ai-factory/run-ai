# Run Your First Standard Training

This quick start provides a step-by-step walkthrough for running a standard training workload.

A training workload contains the setup and configuration needed for building your model, including the container, images, data sets, and resource requests, as well as the required tools for the research, all in a single place.

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../platform-management/aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 1 GPU.

<Note>

**Note**

Flexible workload submission is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload submission.

</Note>

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

To use the API, you will need to obtain a token as shown in [API authentication](../../../rest-api/how-to-authenticate-to-the-api.md).

</Tab>

</Tabs>

## Step 2: Submitting a Standard Training Workload

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Training**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Under **Workload architecture**, select **Standard**
6. Select **Start from scratch** to launch a new workload quickly
7. Enter a **name** for the standard training workload (if the name already exists in the project, you will be requested to submit a different name)
8.  Click **CONTINUE**

    In the next step:
9. Under **Environment**, enter the **Image URL** - `kubeflow/pytorch-dist-mnist:latest`
10. Under **Runtime settings**, click **+COMMAND & ARGUMENTS**, then set the **Arguments** to `--epochs 100`.
11. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the **‘one-gpu’** compute resource for your workload.
    * If ‘one-gpu’ is not displayed, follow the below steps to create a one-time compute resource configuration:
      * Set **GPU devices** per pod - 1
      * Optional: set the **CPU compute** per pod - 0.1 cores (default)
      * Optional: set the **CPU memory** per pod - 100 MB (default)
12. Click **CREATE TRAINING**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Training**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Under **Workload architecture**, select **Standard**
6. Select **Start from scratch** to launch a new workload quickly
7. Enter a **name** for the standard training workload (if the name already exists in the project, you will be requested to submit a different name)
8.  Click **CONTINUE**

    In the next step:
9.  Create an environment for your workload

    * Click **+NEW ENVIRONMENT**
    * Enter quick-start as the **name** for the environment. The name must be unique.
    * Enter `kubeflow/pytorch-dist-mnist:latest` as the **Image URL**
    * Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and enter the arguments - `--epochs 100`
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
10. Select the **‘one-gpu’** compute resource for your workload

    * If ‘one-gpu’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW COMPUTE RESOURCE**
      * Enter one-gpu as the **name** for the compute resource. The name must be unique.
      * Set **GPU devices** per pod - 1
      * Optional: set the **CPU compute per pod** - 0.1 cores (default)
      * Optional: set the **CPU memory per pod** - 100 MB (default)
      * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
11. Click **CREATE TRAINING**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../reference/cli/runai/runai_training_submit.md):

```sh
runai project set "project-name"
runai training submit "workload-name" -i kubeflow/pytorch-dist-mnist:latest -g 1 -- --epochs 100
```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters according. For more details, see [Trainings](api:main-api:POST/api/v1/workloads/trainings) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/training' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{  
    "name": "workload-name",  
    "projectId": "<PROJECT-ID>", 
    "clusterId": "<CLUSTER-UUID>", 
    "spec": {
        "image": "kubeflow/pytorch-dist-mnist:latest",
        "args": "--epochs 100",
        "compute": { 
            "gpuDevicesRequest": 1
        } 
    } 
 }
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](standard-training-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](api:main-api:GET/api/v1/clusters) API.

<Note>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Note>

</Tab>

</Tabs>

## Next Steps

* Manage and monitor your newly created workload using the [Workloads](../../workloads.md) table.
* After validating your training performance and results, deploy your model using [inference.](../../using-inference/nvidia-run-ai-inference-overview.md)
