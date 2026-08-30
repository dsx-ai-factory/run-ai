# Run Your First Custom Inference Workload

This quick start provides a step-by-step walkthrough for running and querying a custom inference workload.

An inference workload provides the setup and configuration needed to deploy your trained model for real-time or batch predictions. It includes specifications for the container image, data sets, network settings, and resource requests required to serve your models.

<Info>

**Note**

Before running this example, verify that the CUDA version supported by [Triton](https://docs.nvidia.com/deeplearning/triton-inference-server/user-guide/docs/introduction/compatibility.html) is compatible with your target [GPU hardware](https://developer.nvidia.com/cuda/gpus).

</Info>

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../platform-management/aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 1 GPU.
* [Knative](../../../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference) is properly installed by your administrator.

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

## Step 2: Submitting an Inference Workload

<Info>

**Note**

The container images used in this example are publicly available and do not require authentication to pull.

</Info>

<Tabs>

<Tab title="UI - Flexible">

1. Go to Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Under **Workload architecture**, select **Standard**
6. Select **Custom** from Inference server
7. Enter a unique **name** for the workload. If the name already exists in the project, you will be requested to submit a different name.
8.  Click **CONTINUE**

    In the next step:
9. Under **Environment**, enter the **Image URL** - `nvcr.io/nvidia/tritonserver:26.02-py3`
10. Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:
    * Enter the command - `/bin/bash -c`
    * Enter the arguments (include the surrounding quotation marks) - `"mkdir -p /models/mnist/1 && wget -q -O /models/mnist/1/model.onnx https://github.com/onnx/models/raw/main/validated/vision/classification/mnist/model/mnist-12.onnx && tritonserver --model-repository=/models --strict-model-config=false --allow-metrics=true --allow-gpu-metrics=true"`
11. Set the inference **serving endpoint** to **HTTP** and the container port to `8000`
12. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the **’half-gpu’** compute resource for your workload.
    * If ‘half-gpu’ is not displayed, follow the below steps to create a one-time compute resource configuration:
      * Set **GPU devices** per pod - 1
      * Enable **GPU fractioning** to set the GPU memory per device
        * Select **% (of device) -** Fraction of a GPU device’s memory
        * Set the memory **Request** - 50 (the workload will allocate 50% of the GPU memory)
      * Optional: set the **CPU compute** per pod - 0.1 cores (default)
      * Optional: set the **CPU memory** per pod - 100 MB (default)
13. Under **Replica autoscaling**:
    * Set a **minimum** of 0 replicas and **maximum** of 2 replicas
    * Set the conditions for creating a new replica to **Concurrency (Requests)** and the **value** to 3
    * Set when the replicas should be automatically scaled down to zero to **After 5 minutes of inactivity**
14. Click **CREATE INFERENCE**

This would start a triton inference server with a maximum of 2 instances, each instance consumes half a GPU.

</Tab>

<Tab title="UI - Original">

1. Go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Select **custom** inference from **Inference type** (if applicable)
6. Enter a unique **name** for the workload. If the name already exists in the project, you will be requested to submit a different name.
7.  Click **CONTINUE**

    In the next step:
8.  Create an environment for your workload

    * Click **+NEW ENVIRONMENT**
    * Enter a **name** for the environment. The name must be unique.
    * Enter the **Image URL** - `nvcr.io/nvidia/tritonserver:26.02-py3`
    * Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:
      * Enter the command: `/bin/bash -c`
      * Enter the arguments (include the surrounding quotation marks): `"mkdir -p /models/mnist/1 && wget -q -O /models/mnist/1/model.onnx https://github.com/onnx/models/raw/main/validated/vision/classification/mnist/model/mnist-12.onnx && tritonserver --model-repository=/models --strict-model-config=false --allow-metrics=true --allow-gpu-metrics=true"`
    * Set the inference **serving endpoint** to **HTTP** and the container port to `8000`
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
9.  Select the **‘half-gpu’** compute resource for your workload

    * If ‘half-gpu’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW COMPUTE RESOURCE**
      * Enter a **name** for the compute resource. The name must be unique.
      * Set **GPU devices** per pod - 1
      * Enable **GPU fractioning** to set the GPU memory per device
        * Select **% (of device) -** Fraction of a GPU device’s memory
        * Set the memory **Request** - 50 (the workload will allocate 50% of the GPU memory)
      * Optional: set the **CPU compute** per pod - 0.1 cores (default)
      * Optional: set the **CPU memory** per pod - 100 MB (default)
      * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
10. Under **Replica autoscaling**:
    * Set a **minimum** of 0 replicas and **maximum** of 2 replicas
    * Set the conditions for creating a new replica to **Concurrency (Requests)** and set the **value** to 3
    * Set when the replicas should be automatically scaled down to zero to **After 5 minutes of inactivity**
11. Click **CREATE INFERENCE**

This would start a triton inference server with a maximum of 2 instances, each instance consumes half a GPU.

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../../reference/cli/runai/):

```sh
runai project set <project_name>
runai inference submit <WORKLOAD-NAME> \
--image nvcr.io/nvidia/tritonserver:26.02-py3 \
--command \
-- /bin/bash -c "mkdir -p /models/mnist/1 && wget -q -O /models/mnist/1/model.onnx https://github.com/onnx/models/raw/main/validated/vision/classification/mnist/model/mnist-12.onnx && tritonserver --model-repository=/models --strict-model-config=false --allow-metrics=true --allow-gpu-metrics=true" \
-gpu-devices-request 1 --gpu-memory-request 50M \
--serving-port=8000 --min-replicas=0 --max-replicas=2 \
--metric=concurrency --metric-threshold 3 \
--scale-to-zero-retention-seconds 300
```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Inferences](https://run-ai-docs.nvidia.com/api/2.26/workloads/inferences) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/inferences' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
    "name": "<WORKLOAD-NAME>",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "image": "nvcr.io/nvidia/tritonserver:26.02-py3",
        "command": "/bin/bash -c '\''mkdir -p /models/mnist/1 && wget -q -O /models/mnist/1/model.onnx https://github.com/onnx/models/raw/main/validated/vision/classification/mnist/model/mnist-12.onnx && tritonserver --model-repository=/models --strict-model-config=false --allow-metrics=true --allow-gpu-metrics=true'\''",
        "servingPort": {
            "protocol": "http",
            "container": 8000
        },
        "autoscaling": {
            "minReplicas": 0,
            "maxReplicas": 2,
            "metric": "concurrency",
            "metricThreshold": 3,
            "scaleToZeroRetentionSeconds": 300
        },
        },
        "compute": {
            "cpuCoreRequest": 0.1,
            "gpuRequestType": "portion",
            "cpuMemoryRequest": "100M",
            "gpuDevicesRequest": 1,
            "gpuPortionRequest": 0.5
        }
    }
}'      
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](inference-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 3: Querying the Inference Server

<Info>

**Note**

The container images used in this example are publicly available and do not require authentication to pull.

</Info>

In this step, you'll test the deployed model by sending a request to the inference server. To do this, you'll launch a general-purpose workload, typically a **Training** or **Workspace** workload, to run the Triton demo client. You'll first retrieve the workload address, which serves as the model’s inference serving endpoint. Then, use the client to send a sample request and verify that the model is responding correctly.

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads.
2. Click COLUMNS and select **Connections.**
3. Select the link under the Connections column for the inference workload created in [Step 2](inference-quickstart.md#step-2-submitting-an-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Training**
6. Select the **cluster** and **project** where the inference workload was created
7. Under **Workload architecture**, select **Standard**
8. Select **Start from scratch** to launch a new workload quickly
9. Enter a unique **name** for the workload. If the name already exists in the project, you will be requested to submit a different name.
10. Click **CONTINUE**

    In the next step:
11. Under **Environment**, enter the **Image URL** - `nvcr.io/nvidia/tritonserver:26.02-py3-sdk`
12. Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:
    * Enter the command - `perf_analyzer`
    * Enter the arguments - `-m mnist -p 3600000 -u <INFERENCE-ENDPOINT> -i http`. Make sure to replace the inference endpoint with the **Address** you retrieved above.
13. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select **‘cpu-only’** from the list.
    * If ‘cpu-only’ is not displayed, follow the below steps to create a one-time compute resource configuration:
      * Set **GPU devices** per pod - 0
      * Set **CPU compute** per pod - 0.1 cores
      * Set the **CPU memory** per pod - 100 MB (default)
14. Click **CREATE TRAINING**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads.
2. Click COLUMNS and select **Connections.**
3. Select the link under the Connections column for the inference workload created in [Step 2](inference-quickstart.md#step-2-submitting-an-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Training**
6. Select the **cluster** and **project** where the inference workload was created
7. Under **Workload architecture**, select **Standard**
8. Select **Start from scratch** to launch a new workload quickly
9. Enter a unique **name** for the workload. If the name already exists in the project, you will be requested to submit a different name.
10. Click **CONTINUE**

    In the next step:
11. Create an environment for your workload

    * Click **+NEW ENVIRONMENT**
    * Enter quick-start as the **name** for the environment. The name must be unique.
    * Enter the **Image URL** - `nvcr.io/nvidia/tritonserver:26.02-py3-sdk`
    * Set the runtime settings for the environment. Click **+COMMAND & ARGUMENTS** and add the following:
      * Enter the command: `perf_analyzer`
      * Enter the arguments: `-m mnist -p 3600000 -u <INFERENCE-ENDPOINT> -i http`. Make sure to replace the inference endpoint with the **Address** you retrieved above.
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
12. Select the **‘cpu-only’** compute resource for your workspace

    * If ‘cpu-only’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW COMPUTE RESOURCE**
      * Enter cpu-only as the **name** for the compute resource. The name must be unique.
      * Set **GPU devices** per pod - 0
      * Set **CPU compute** per pod - 0.1 cores
      * Set the **CPU memory** per pod - 100 MB (default)
      * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
13. Click **CREATE TRAINING**

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. To retrieve the inference endpoint, use the [runai inference describe](../../../reference/cli/runai/runai_inference_describe.md) command. For more details, see [CLI reference](../../../reference/cli/runai/):

```sh
runai project set <project_name>
runai training submit <WORKLOAD-NAME> \
--image nvcr.io/nvidia/tritonserver:26.02-py3-sdk \
--command -- perf_analyzer -m mnist -p 3600000 -u <INFERENCE-ENDPOINT> -i http
```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters according. For more details, see [Trainings](https://run-ai-docs.nvidia.com/api/2.26/workloads/trainings) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/trainings' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
    "name": "<WORKLOAD-NAME>",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "image": "nvcr.io/nvidia/tritonserver:26.02-py3-sdk",
        "command": "perf_analyzer",
        "args": "-m mnist -p 3600000 -u <INFERENCE-ENDPOINT> -i http",
        "compute": {
            "cpuCoreRequest": 0.1,
            "cpuMemoryRequest": "100M"
        }
    }
}'
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](inference-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.
* `<INFERENCE-ENDPOINT>` - You can get the inference endpoint from the urls parameter via the [Get Workloads](https://run-ai-docs.nvidia.com/api/2.26/workloads/workloads#get-api-v1-workloads) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Next Steps

* Select the inference workload you created in [Step 2](inference-quickstart.md#step-2-submitting-an-inference-workload) and go to the [Metrics](../../workloads.md#show-hide-details) tab to see various GPU and inference metrics graphs rise.
* Manage and monitor your newly created workload using the [Workloads](../../workloads.md) table.
