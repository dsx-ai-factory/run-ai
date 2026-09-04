# Model-Free NIM Service Deployment

This tutorial demonstrates how to deploy the Llama 3.2 1B Instruct model as a NIM Service workload on the NVIDIA Run:ai platform. NIM Service workloads use the [NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/) under the hood. You can use this workflow as a reference and adapt it for other models, model sources, container images, and hardware configurations.

<Info>

**Note**

For the NIM example, verify that the selected NIM is supported on your target hardware according to the [NIM support matrix](https://docs.nvidia.com/nim/large-language-models/latest/supported-models.html).

</Info>

In this tutorial, you will learn how to:

* Set up environment prerequisites for NIM Services
* Create a user access key for API integrations with NVIDIA Run:ai
* Create credentials to authenticate with NGC and pull the NIM image
* Deploy a NIM Service inference workload using the NVIDIA Run:ai UI, API, or CLI
* Access the inference endpoint to send requests

## Prerequisites

Before you start, make sure the following requirements are met:

* Your administrator has:
  * Installed the [NVIDIA NIM Operator](https://docs.nvidia.com/nim-operator/latest/) on the cluster.
  * Created a [project](../../platform-management/aiinitiatives/organization/projects.md) for you.
* You have:
  * An NGC account with an active NGC API key. To obtain a key, go to [NGC](https://catalog.ngc.nvidia.com/) → Setup → API Keys, then generate or copy an existing key.
  * A Hugging Face account with an active access token and an accepted license for [Llama 3.2 1B Instruct](https://huggingface.co/meta-llama/Llama-3.2-1B-Instruct). The token is required only because this tutorial's example model is gated; public models and other model sources do not need one.

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [Creating a user access key](nim-service.md#step-2-creating-a-user-access-key).

</Tab>

<Tab title="CLI v2">

Run the below --help command to obtain the login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

</Tabs>

## Step 2: Creating a User Access Key

<Info>

**Note**

This step is only required if you intend to follow the API steps in this tutorial.

</Info>

Access keys are used for API integrations with NVIDIA Run:ai. An access key contains a client ID and a client secret. With the client credentials, you can obtain a token and use it within subsequent API calls.

In the NVIDIA Run:ai user interface:

1. Click the user avatar at the top right corner, then select **Settings**
2. Click **+ACCESS KEY**
3. Enter the access key's **name** and click **CREATE**
4. Copy the **Client ID** and **Client secret** and store securely
5. Click **DONE**

To request an API access token, use the client credentials to get a token to access NVIDIA Run:ai using the [Tokens](https://run-ai-docs.nvidia.com/api/2.26/authentication-and-authorization/tokens) API. For example:

```bash
curl -X POST \ 
  # Replace <COMPANY_URL> below with:
  # For self-hosted use the NVIDIA Run:ai user interface URL.
  'https://<COMPANY_URL>/api/v1/token' \ 
  --header 'Accept: */*' \ 
  --header 'Content-Type: application/json' \ 
  --data-raw '{ 
  "grantType":"client_credentials", 
  "clientId":"<CLIENT ID>", 
  "clientSecret" : "<CLIENT SECRET>" 
}'
```

## Step 3: Creating the NGC API Key Credential

This [credential](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md) stores your NGC API key for authenticating with the NGC catalog.

<Info>

**Note**

[User credentials](../../settings/user-settings/user-credentials.md) cannot be used with NIM Service workloads. If you do not have permission to create credentials, contact your administrator.

</Info>

1. Go to Workload manager → Credentials.
2. Click **+NEW CREDENTIAL** and select **NGC API key**.
3. Select a [**scope**](../../workloads-in-nvidia-run-ai/workload-assets/overview.md#asset-scope) covering the project you will submit the workload to.
4. Enter a **name** for the credential. The name must be unique.
5. Enter your **NGC API key**.
6. Click **CREATE CREDENTIAL**.

Once created, note the **Kubernetes name** from the [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#credentials-table) table - you will need it when submitting the workload in Step 5.

## Step 4: Creating the Docker Registry Credential

This [credential](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md) authenticates image pulls from `nvcr.io` so the workload can fetch the NIM container image.

<Info>

**Note**

[User credentials](../../settings/user-settings/user-credentials.md) cannot be used with NIM Service workloads. If you do not have permission to create credentials, contact your administrator.

</Info>

1. Go to Workload manager → Credentials.
2. Click **+NEW CREDENTIAL** and select **Docker registry**.
3. Select the same [**scope**](../../workloads-in-nvidia-run-ai/workload-assets/overview.md#asset-scope) used for the NGC API key credential in Step 3.
4. Enter a **name** for the credential. The name must be unique.
5. Select **New secret** and enter:
   * **Username**: `$oauthtoken`
   * **Password**: your **NGC API key** (the same key used in Step 3)
   * **Docker registry URL**: `nvcr.io`
6. Click **CREATE CREDENTIAL**.

Once created, note the **Kubernetes name** from the [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#credentials-table) table - you will need it when submitting the workload in Step 5.

## Step 5: Creating the Workload

This tutorial uses a [model-free NIM](https://docs.nvidia.com/nim/large-language-models/latest/about-nim-llm/overview.html#model-free-and-model-specific-containers) container, which serves a model you configure at deploy time via the `NIM_MODEL_NAME` environment variable rather than a model bundled into the image. In this tutorial, that model is Llama 3.2 1B Instruct. NIM Services can also use model-specific NIMs from NGC. The workload provisions a PVC inline at submission time and mounts it at `/opt/nim/.cache` to cache downloaded model weights and engine artifacts so subsequent restarts skip the download.

<Info>

**Note**

The first time the workload is submitted, the PVC is provisioned and the model is downloaded, so startup takes longer. Subsequent restarts reuse the cached weights and start much faster.

</Info>

### How the Configuration Works

<Info>

**Note**

`NIM_MODEL_NAME` accepts models from several sources, set using the source prefix. For example:

* Hugging Face: `hf://openai-community/gpt2`
* NGC: `ngc://nim/meta/llama-3.3-70b-instruct:hf`
* Amazon S3: `s3://my-bucket/my-model`
* Google Cloud Storage: `gs://my-bucket/my-model`
* Local directory or PVC: `/mnt/models/my-llama`

A Hugging Face token (`HF_TOKEN`) is required only for gated Hugging Face models. Public Hugging Face models and the other sources can be pulled without one.

</Info>

* **Container image** - The model-free NIM container image (`llm-nim`), which can serve any supported model. The specific model is selected at runtime via the `NIM_MODEL_NAME` environment variable.
* **Image pull secret** - References the Docker registry Kubernetes secret created in [Step 4](nim-service.md#step-4-creating-the-docker-registry-credential), used to pull the NIM image from `nvcr.io`.
* **NGC authentication secret** - References the NGC API key Kubernetes secret created in [Step 3](nim-service.md#step-3-creating-the-ngc-api-key-credential), used to authenticate with the NGC catalog for model access.
* `NIM_MODEL_NAME` - Selects the model and its source. A model-free NIM can load models from NGC (`ngc://`), Hugging Face (`hf://`), Amazon S3 (`s3://`), Google Cloud Storage (`gs://`), or a local directory / PVC path. In this tutorial, the value is `hf://meta-llama/Llama-3.2-1B-Instruct`.
* `NIM_SERVER_PORT` / `NIM_HTTP_API_PORT` - Set the port on which the NIM server listens for inference requests.
* `HF_TOKEN` - Required only for gated Hugging Face models like Llama 3.2, combined with license acceptance on the model page on Hugging Face. Public Hugging Face models and models loaded from other sources do not require it.
* **Model store PVC** - Provisioned inline at workload submission time and mounted at `/opt/nim/.cache`. Caches the downloaded model weights so subsequent restarts skip the download.
* **Serving port** - The port exposed for serving inference requests. Port 8000 is the default for NIM's OpenAI-compatible API server.
* **External access** - When enabled, the inference endpoint is accessible from outside the cluster. Without it, the endpoint is only reachable inside the cluster.
* **GPU allocation** - The number of GPUs requested for the workload. Llama 3.2 1B fits on a single GPU.
* **Replicas** - The number of NIM Service replicas to run. Set to `1` for this tutorial.

### Submitting the Workload

<Tabs>

<Tab title="UI">

1. To create a workload, go to Workload manager → Workloads.
2. Click **+ NEW WORKLOAD** and select **Via YAML** from the dropdown.
3. In the YAML submission form, select the **cluster** where the workload will run.
4. Upload or paste the YAML manifest below. To upload a file, click **UPLOAD YAML FILE** and choose your YAML. To paste the YAML, insert it directly into the editor. Before submitting, update the following fields in the manifest:
   * `<image-pull-secret-name>` - The Kubernetes name of the Docker registry credential created in [Step 4](nim-service.md#step-4-creating-the-docker-registry-credential).
   * `<ngc-auth-secret-name>` - The Kubernetes name of the NGC API key credential created in [Step 3](nim-service.md#step-3-creating-the-ngc-api-key-credential).
   * `<HF_TOKEN>` - Your Hugging Face access token.

```yaml
apiVersion: apps.nvidia.com/v1alpha1
kind: NIMService
metadata:
  name: nim-llama-svc
spec:
  image:
    repository: nvcr.io/nim/nvidia/llm-nim
    tag: "1.12"
    pullPolicy: IfNotPresent
    pullSecrets:
      - <image-pull-secret-name>
  authSecret: <ngc-auth-secret-name>
  replicas: 1
  storage:
    pvc:
      create: true
      size: 20Gi
      volumeAccessMode: ReadWriteOnce
  resources:
    limits:
      nvidia.com/gpu: 1
  expose:
    service:
      type: ClusterIP
      port: 8000
  env:
    - name: NIM_MODEL_NAME
      value: hf://meta-llama/Llama-3.2-1B-Instruct
    - name: NIM_SERVER_PORT
      value: "8000"
    - name: NIM_HTTP_API_PORT
      value: "8000"
    - name: HF_TOKEN
      value: <HF_TOKEN>
```

5. Select a **project** from the submission form.
6. Click **CREATE WORKLOAD**.

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters. For more details, see the [NVIDIA NIM](https://run-ai-docs.nvidia.com/api/2.26/workloads/nvidia-nim) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](nim-service.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](https://run-ai-docs.nvidia.com/api/2.26/organizations/projects#get-api-v1-org-unit-projects) API.
* `<image-pull-secret-name>` - The Kubernetes name of the Docker registry credential created in [Step 4](nim-service.md#step-4-creating-the-docker-registry-credential).
* `<ngc-auth-secret-name>` - The Kubernetes name of the NGC API key credential created in [Step 3](nim-service.md#step-3-creating-the-ngc-api-key-credential).
* `<HF_TOKEN>` - Your Hugging Face access token.

```bash
curl -L 'https://<COMPANY-URL>/api/v2/workloads/nim-services' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "metadata": {
      "name": "<workload-name>",
      "projectId": "<PROJECT-ID>"
    },
    "spec": {
      "image": "nvcr.io/nim/nvidia/llm-nim:1.12",
      "imagePullSecrets": [
        {"name": "<image-pull-secret-name>"}
      ],
      "ngcAuthSecret": "<ngc-auth-secret-name>",
      "environmentVariables": [
        {"name": "NIM_MODEL_NAME", "value": "hf://meta-llama/Llama-3.2-1B-Instruct"},
        {"name": "NIM_SERVER_PORT", "value": "8000"},
        {"name": "NIM_HTTP_API_PORT", "value": "8000"},
        {"name": "HF_TOKEN", "value": "<HF_TOKEN>"}
      ],
      "modelStore": {
        "pvc": {
          "existingPvc": false,
          "claimInfo": {
            "size": "20Gi",
            "accessModes": {
              "readWriteOnce": true
            }
          }
        }
      },
      "servingPort": {
        "port": 8000,
        "exposeExternally": true
      },
      "compute": {
        "gpuDevicesRequest": 1
      },
      "replicas": 1
    }
  }'
```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the parameters below with the name of your project, workload, credentials (from Steps 3 and 4), and Hugging Face token. For more details, see the [CLI reference](../../reference/cli/runai/runai_inference_nim_submit.md):

```sh
runai inference nim submit <workload-name> \
  -p <project-name> \
  -i nvcr.io/nim/nvidia/llm-nim:1.12 \
  --image-pull-secret <image-pull-secret-name> \
  --ngc-auth-secret <ngc-auth-secret-name> \
  --model-new-pvc "size=20Gi,accessmode-rwo" \
  --serving-port 8000 \
  -g 1 \
  -e NIM_MODEL_NAME=hf://meta-llama/Llama-3.2-1B-Instruct \
  -e NIM_SERVER_PORT=8000 \
  -e NIM_HTTP_API_PORT=8000 \
  -e HF_TOKEN=<HF_TOKEN>
```

</Tab>

</Tabs>

## Step 6: Verifying the Workload Status

After submitting the workload, wait for it to reach the **Running** status in the [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) table. A workload becomes ready to accept inference requests only after all its pods have fully initialized, including model loading.

The first startup takes longer because the model weights are downloaded to the PVC. During this time, the workload may remain in **Initializing** even though the pod is already running. Subsequent restarts reuse the cached weights and start much faster.

To monitor progress:

* Select the workload and click the **SHOW DETAILS** button at the upper-right side of the action bar. The details pane appears, presenting the **Logs** tab to track model-download and model-loading progress.
* The workload transitions to **Running** only when the pod finishes loading the model and all readiness checks pass.

Once the workload reaches **Running** and shows an available **Connection**, you can proceed to access the inference endpoint.

## Step 7: Accessing the Inference Workload

You can programmatically consume an inference workload via API by making direct calls to the serving endpoint, typically from other workloads or external integrations. Once an inference workload is deployed, the serving endpoint URL appears in the **Connections** column of the [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) table. To retrieve the service endpoint programmatically, use the [Get Workloads](https://run-ai-docs.nvidia.com/api/2.26/workloads/workloads#get-api-v1-workloads) API. The endpoint URL will be available in the response body under `endpoints`.

<Info>

**Note**

For clusters below version 2.25, the endpoint URL will be available in the response body under `urls`.

</Info>

```bash
#replace <serving-endpoint-url> and <model-name> (e.g. "meta-llama/Llama-3.2-1B-Instruct")
curl <serving-endpoint-url>/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<model-name>",
    "messages": [{
      "role": "user",
      "content": "What is Kubernetes?"
    }]
  }'
```
