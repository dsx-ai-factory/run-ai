# Launching Workloads with GPU Memory Swap

This quick start provides a step-by-step walkthrough for running multiple LLMs (inference workload) on a single GPU using [GPU memory swap](../memory-swap.md).

GPU memory swap expands the GPU physical memory to the CPU memory, allowing NVIDIA Run:ai to place and run more workloads on the same GPU physical hardware. This provides a smooth workload context switching between GPU memory and CPU memory, eliminating the need to kill workloads when the memory requirement is larger than what the GPU physical memory can provide.

## Prerequisites

Before you start, make sure:

* You have created a [project](../../../aiinitiatives/organization/projects.md) or have one created for you.
* The project has an assigned quota of at least 1 GPU.
* [Dynamic GPU fractions](../dynamic-fractions.md) is enabled.
* GPU memory swap is enabled on at least one free node as detailed [here](../memory-swap.md#enabling-and-configuring-gpu-memory-swap).
* [Host-based routing](../../../../installation-and-configuration/installation/install-using-helm/system-requirements.md#host-based-routing) is configured.

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

<Tab title="API">

To use the API, you will need to obtain a token as shown in [API authentication](https://run-ai-docs.nvidia.com/api/2.26/getting-started/how-to-authenticate-to-the-api).

</Tab>

</Tabs>

## Step 2: Submitting the First Inference Workload

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Select **custom** inference from **Inference type** (if applicable)
6. Enter a **name** for the workload (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. To add a new environment:
   * Click the **+** icon to create a new environment
   * Enter quick-start as the **name** for the environment. The name must be unique.
   * Enter the NVIDIA Run:ai vLLM **Image URL** - `runai.jfrog.io/core-llm/runai-vllm:v0.6.4-0.10.0`
   * Set the inference **serving endpoint** to **HTTP** and the container port to `8000`
   * Set the runtime settings for the environment. Click **+ENVIRONMENT VARIABLE** and add the following:
     * **Name:** RUNAI\_MODEL **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct` (you can choose any vLLM supporting model from Hugging Face)
     * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `Llama-3.2-1B-Instruct`
     * **Name:** HF\_TOKEN **Source:** Custom **Value:** \<Your Hugging Face token> (only needed for gated models)
     * **Name:** VLLM_RPC_TIMEOUT **Source:** Custom **Value:** 60000
   * Click **CREATE ENVIRONMENT**
   * Select the newly created environment from the side pane
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. To add a new compute resource:
   * Click the **+** icon to create a new compute resource
   * Enter request-limit as the **name** for the compute resource. The name must be unique.
   * Set **GPU devices** per pod - 1
   * Enable **GPU fractioning** to set the GPU memory per device:
     * Select **% (of device)** - Fraction of a GPU device’s memory
     * Set the memory **Request** - 50 (the workload will allocate 50% of the GPU memory)
     * Set the memory **Limit** - 100%
   * Optional: set the **CPU compute** per pod - 0.1 cores (default)
   * Optional: set the **CPU memory** per pod - 100 MB (default)
   * Select **More settings** and toggle **Increase shared memory size**
   * Click **CREATE COMPUTE RESOURCE**
   * Select the newly created compute resource from the side pane
10. Click **CREATE INFERENCE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select under which **cluster** to create the workload
4. Select the **project** in which your workload will run
5. Select **custom** inference from **Inference type** (if applicable)
6. Enter a **name** for the workload (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8.  Create an environment for your workload

    * Click **+NEW ENVIRONMENT**
    * Enter quick-start as the **name** for the environment. The name must be unique.
    * Enter the NVIDIA Run:ai vLLM **Image URL** - `runai.jfrog.io/core-llm/runai-vllm:v0.6.4-0.10.0`
    * Set the runtime settings for the environment. Click **+ENVIRONMENT VARIABLE** and add the following:
      * **Name:** RUNAI\_MODEL **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct` (you can choose any vLLM supporting model from Hugging Face)
      * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `Llama-3.2-1B-Instruct`
      * **Name:** HF\_TOKEN **Source:** Custom **Value:** \<Your Hugging Face token> (only needed for gated models)
      * **Name:** VLLM_RPC_TIMEOUT **Source:** Custom **Value:** 60000
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
9.  Create a new “**request-limit**” compute resource

    * Click **+NEW COMPUTE RESOURCE**
    * Enter request-limit as the **name** for the compute resource. The name must be unique.
    * Set **GPU devices** per pod - 1
    * Enable **GPU fractioning** to set the GPU memory per device:
      * Select **% (of device)** - Fraction of a GPU device’s memory
      * Set the memory **Request** - 50 (the workload will allocate 50% of the GPU memory)
      * Set the memory **Limit** - 100%
    * Optional: set the **CPU compute per pod** - 0.1 cores (default)
    * Optional: set the **CPU memory per pod** - 100 MB (default)
    * Select **More settings** and toggle **Increase shared memory size**
    * Click **CREATE COMPUTE RESOURCE**

    The newly created compute resource will be selected automatically
10. Click **CREATE INFERENCE**

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Inferences](https://run-ai-docs.nvidia.com/api/2.26/workloads/inferences) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/inferences' \ 
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{ 
    "name": "workload-name", 
    "useGivenNameAsPrefix": true,
    "projectId": "<PROJECT-ID>", 
    "clusterId": "<CLUSTER-UUID>", 
    "spec": {
        "image": "runai.jfrog.io/core-llm/runai-vllm:v0.6.4-0.10.0",
        "imagePullPolicy":"IfNotPresent",
        "environmentVariables": [
          {
            "name": "RUNAI_MODEL",
            "value": "meta-lama/Llama-3.2-1B-Instruct"
          },
          {
            "name": "VLLM_RPC_TIMEOUT",
            "value": "60000"
          },
          {
            "name": "HF_TOKEN",
            "value":"<INSERT HUGGINGFACE TOKEN>"
          }
        ],
        "compute": {
            "gpuDevicesRequest": 1,
            "gpuRequestType": "portion",
            "gpuPortionRequest": 0.1,
            "gpuPortionLimit": 1,
            "cpuCoreRequest":0.2,
            "cpuMemoryRequest": "200M",
            "largeShmRequest": false

        },
        "servingPort": {
            "container": 8000,
            "protocol": "http",
            "authorizationType": "public"
        }
    }
}       
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](gpu-memory-swap-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 3: Submitting the Second Inference Workload

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select the **cluster** where the previous inference workload was created
4. Select the **project** where the previous inference workload was created
5. Select **custom** inference from **Inference type** (if applicable)
6. Enter a **name** for the workload (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the environment created in [Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload).
9. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the compute resources created in [Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload).
10. Click **CREATE INFERENCE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click **+NEW WORKLOAD** and select **Inference**
3. Select the **cluster** where the previous inference workload was created
4. Select the **project** where the previous inference workload was created
5. Select **custom** inference from **Inference type** (if applicable)
6. Enter a **name** for the workload (if the name already exists in the project, you will be requested to submit a different name)
7.  Click **CONTINUE**

    In the next step:
8. Select the environment created in [Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload)
9. Select the compute resource created in [Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload)
10. Click **CREATE INFERENCE**

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Inferences](https://run-ai-docs.nvidia.com/api/2.26/workloads/inferences) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/inferences' \ 
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \ 
-d '{ 
    "name": "workload-name", 
    "useGivenNameAsPrefix": true,
    "projectId": "<PROJECT-ID>",  
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
        "image": "runai.jfrog.io/core-llm/runai-vllm:v0.6.4-0.10.0",
        "imagePullPolicy":"IfNotPresent",
        "environmentVariables": [
          {
            "name": "RUNAI_MODEL",
            "value": "meta-lama/Llama-3.2-1B-Instruct"
          },
          {
            "name": "VLLM_RPC_TIMEOUT",
            "value": "60000"
          },
          {
            "name": "HF_TOKEN",
            "value":"<INSERT HUGGINGFACE TOKEN>"
          }
        ],
        "compute": {
            "gpuDevicesRequest": 1,
            "gpuRequestType": "portion",
            "gpuPortionRequest": 0.1,
            "gpuPortionLimit": 1,
            "cpuCoreRequest":0.2,
            "cpuMemoryRequest": "200M",
            "largeShmRequest": false

        },
        "servingPort": {
            "container": 8000,
            "protocol": "http",
            "authorizationType": "public"
        }
    }
}       
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](gpu-memory-swap-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 4: Submitting the First Workspace

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click COLUMNS and select **Connections**
3. Select the link under the Connections column for the first inference workload created in[ Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Workspace**
6. Select the **cluster** where the previous inference workloads were created
7. Select the **project** where the previous inference workloads were created
8. Select **Start from scratch** to launch a new workspace quickly
9. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
10. Click **CONTINUE**

    In the next step:
11. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the **‘chatbot-ui’** environment for your workspace (Image URL: `runai.jfrog.io/core-llm/llm-app)`
    * Set the runtime settings for the environment with the following **environment variables**:
      * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
      * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the address link from Step 4
      * Delete the **PATH\_PREFIX** environment variable if you are using host-based routing.
    * If ‘chatbot-ui’ is not displayed in the gallery, follow the below steps:
      * Click the **+** icon to create a new environment
      * Enter chatbot-ui as the **name** for the environment. The name must be unique.
      * Enter the chatbot-ui **Image URL** - `runai.jfrog.io/core-llm/llm-app`
      * Tools - Set the connection for your tool
        * Click **+TOOL**
        * Select **Chatbot UI** tool from the list
      * Set the runtime settings for the environment. Click **+ENVIRONMENT VARIABLE** and add the following:
        * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
        * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the **Address** link
        * **Name:** RUNAI_MODEL_TOKEN\_LIMIT **Source:** Custom **Value:** 8192
        * **Name:** RUNAI_MODEL_MAX\_LENGTH **Source:** Custom **Value:** 16384
      * Click **CREATE ENVIRONMENT**
      * Select the newly created environment from the side pane
12. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select **‘cpu-only’** from the list.
    * If ‘cpu-only’ is not displayed, follow the below steps:
      * Click the **+** icon to create a new compute resource
      * Enter cpu-only as the **name** for the compute resource. The name must be unique.
      * Set **GPU devices** per pod - 0
      * Set **CPU compute** per pod - 0.1 cores
      * Set the **CPU memory** per pod - 100 MB (default)
      * Click **CREATE COMPUTE RESOURCE**
      * Select the newly created compute resource from the side pane
13. Click **CREATE WORKSPACE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click COLUMNS and select **Connections**
3. Select the link under the Connections column for the first inference workload created in[ Step 2](gpu-memory-swap-quickstart.md#step-2-submitting-the-first-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Workspace**
6. Select the **cluster** where the previous inference workloads were created
7. Select the **project** where the previous inference workloads were created
8. Select **Start from scratch** to launch a new workspace quickly
9. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
10. Click **CONTINUE**

    In the next step:
11. Select the **‘chatbot-ui’** environment for your workspace (Image URL: `runai.jfrog.io/core-llm/llm-app`)

    * Set the runtime settings for the environment with the following **environment variables**:
      * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
      * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the address link from Step 4
      * Delete the **PATH\_PREFIX** environment variable if you are using host-based routing.
    * If ‘chatbot-ui’ is not displayed in the gallery, follow the below steps:
      * Click **+NEW ENVIRONMENT**
      * Enter chatbot-ui as the **name** for the environment. The name must be unique.
      * Enter the chatbot-ui **Image URL** - `runai.jfrog.io/core-llm/llm-app`
      * Tools - Set the connection for your tool
        * Click **+TOOL**
        * Select **Chatbot UI** tool from the list
      * Set the runtime settings for the environment. Click **+ENVIRONMENT VARIABLE** and add the following:
        * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
        * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the **Address** link
        * **Name:** RUNAI_MODEL_TOKEN\_LIMIT **Source:** Custom **Value:** 8192
        * **Name:** RUNAI_MODEL_MAX\_LENGTH **Source:** Custom **Value:** 16384
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
13. Click **CREATE WORKSPACE**

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
        "image": "runai.jfrog.io/core-llm/llm-app",
        "environmentVariables": [
          {
            "name": "RUNAI_MODEL_NAME",
            "value": "meta-llama/Llama-3.2-1B-Instruct"
          },
          {
            "name": "RUNAI_MODEL_BASE_URL",
            "value": "<URL>" 
          }
        ],
        "compute": {
            "cpuCoreRequest":0.1,
            "cpuMemoryRequest": "100M",
        }
    }
}
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](gpu-memory-swap-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.
* `<URL>` - The URL for connecting an external service related to the workload. You can get the URL via the [List Workloads](https://run-ai-docs.nvidia.com/api/2.26/workloads/workloads#get-api-v1-workloads) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 5: Submitting the Second Workspace

<Tabs>

<Tab title="UI - Flexible">

1. Go to the Workload manager → Workloads
2. Click COLUMNS and select **Connections**
3. Select the link under the Connections column for the second inference workload created in [Step 3](gpu-memory-swap-quickstart.md#step-3-submitting-the-second-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Workspace**
6. Select the **cluster** where the previous inference workloads were created
7. Select the **project** where the previous inference workloads were created
8. Select **Start from scratch** to launch a new workspace quickly
9. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
10. Click **CONTINUE**

    In the next step:
11. Under **Environment**, click the **load** icon. A side pane appears, displaying a list of available environments. Select the environment created in [Step 4](gpu-memory-swap-quickstart.md#step-4-submitting-the-first-workspace).
    * Set the runtime settings for the environment with the following **environment variables**:
      * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
      * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the **Address** link
      * Delete the **PATH\_PREFIX** environment variable if you are using host-based routing.
12. Under **Compute resources**, click the **load** icon. A side pane appears, displaying a list of available compute resources. Select the compute resources created in [Step 4](gpu-memory-swap-quickstart.md#step-4-submitting-the-first-workspace).
13. Click **CREATE WORKSPACE**

</Tab>

<Tab title="UI - Original">

1. Go to the Workload manager → Workloads
2. Click COLUMNS and select **Connections**
3. Select the link under the Connections column for the second inference workload created in [Step 3](gpu-memory-swap-quickstart.md#step-3-submitting-the-second-inference-workload)
4. In the **Connections Associated with Workload form,** copy the URL under the **Address** column
5. Click **+NEW WORKLOAD** and select **Workspace**
6. Select the **cluster** where the previous inference workloads were created
7. Select the **project** where the previous inference workloads were created
8. Select **Start from scratch** to launch a new workspace quickly
9. Enter a **name** for the workspace (if the name already exists in the project, you will be requested to submit a different name)
10. Click **CONTINUE**

    In the next step:
11. Select the environment created in [Step 4](gpu-memory-swap-quickstart.md#step-4-submitting-the-first-workspace)
    * Set the runtime settings for the environment with the following **environment variables**:
      * **Name:** RUNAI_MODEL_NAME **Source:** Custom **Value:** `meta-llama/Llama-3.2-1B-Instruct`
      * **Name:** RUNAI_MODEL_BASE\_URL **Source:** Custom **Value:** Add the **Address** link
      * Delete the **PATH\_PREFIX** environment variable if you are using host-based routing.
12. Select the compute resource created in [Step 4](gpu-memory-swap-quickstart.md#step-4-submitting-the-first-workspace)
13. Click **CREATE WORKSPACE**

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the below parameters. For more details, see [Workspaces](https://run-ai-docs.nvidia.com/api/2.26/workloads/workspaces) API:

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/workspaces' \ 
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \ 
-d '{ 
    "name": "workload-name", 
    "projectId": "<PROJECT-ID>", '\ 
    "clusterId": "<CLUSTER-UUID>", \ 
    "spec": {  
        "image": "runai.jfrog.io/core-llm/llm-app",
        "environmentVariables": [
          {
            "name": "RUNAI_MODEL_NAME",
            "value": "meta-llama/Llama-3.2-1B-Instruct"
          },
          {
            "name": "RUNAI_MODEL_BASE_URL",
            "value": "<URL>" 
          }
        ],
        "compute": {
            "cpuCoreRequest":0.1,
            "cpuMemoryRequest": "100M",
        }
    }
}
```

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - The API access token obtained in [Step 1](gpu-memory-swap-quickstart.md#step-1-logging-in)
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](https://run-ai-docs.nvidia.com/api/2.26/organizations/clusters#get-api-v1-clusters) API.
* `<URL>` - The URL for connecting an external service related to the workload. You can get the URL via the [List Workloads](https://run-ai-docs.nvidia.com/api/2.26/workloads/workloads#get-api-v1-workloads) API.

<Info>

**Note**

The above API snippet runs with NVIDIA Run:ai clusters of 2.18 and above only.

</Info>

</Tab>

</Tabs>

## Step 6: Connecting to Chatbot-UI

<Tabs>

<Tab title="UI">

1. Select the newly created workspace that you want to connect to
2. Click **CONNECT**
3. Select the ChatbotUI tool. The selected tool is opened in a new tab on your browser.
4. Query both workspaces simultaneously and see them both responding. The one on CPU RAM at the time will take longer as it switches back to the GPU and vice versa.

</Tab>

<Tab title="API">

1. To connect to the ChatbotUI tool, browse directly to `https://<COMPANY-URL>/<PROJECT-NAME>/<WORKLOAD-NAME>`
2. Query both workspaces simultaneously and see them both responding. The one on CPU RAM at the time will take longer as it switches back to the GPU and vice versa.

</Tab>

</Tabs>

## Next Steps

Manage and monitor your newly created workloads using the [Workloads](../../../../workloads-in-nvidia-run-ai/workloads.md) table.
