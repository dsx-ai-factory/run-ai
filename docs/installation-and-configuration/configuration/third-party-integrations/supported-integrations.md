---
---

# Supported Integrations

NVIDIA Run:ai includes built-in support for the following Kubernetes components and external tools. You may still need cluster-side installation (for example, install an operator and its CRDs) before you can use the integration.

## Frameworks

<table><thead><tr><th>Framework</th><th>Category</th><th>Supported Version</th><th>Additional Information</th></tr></thead><tbody><tr><td>Dynamo operator</td><td>Distributed inference</td><td>1.5.x</td><td><p>Dynamo operator is a Kubernetes operator that simplifies the deployment, configuration, and lifecycle management of DynamoGraphs.</p><p>NVIDIA Run:ai provides out of the box support for submitting Dynamo workloads <a href="../../../workloads-in-nvidia-run-ai/submit-via-yaml.md">via YAML</a>. See <a href="https://docs.nvidia.com/dynamo/latest/">Dynamo operator</a> documentation for more details.</p></td></tr><tr><td>NIM operator</td><td>Model Serving</td><td>3.0.x</td><td><p>The NVIDIA NIM Operator enables Kubernetes cluster administrators to operate the software components and services necessary to deploy NVIDIA NIMs and NVIDIA NeMo microservices in Kubernetes.</p><p>NVIDIA Run:ai provides out of the box support for submitting NIM operator workloads <a href="../../../workloads-in-nvidia-run-ai/submit-via-yaml.md">via YAML</a>. See <a href="https://docs.nvidia.com/nim-operator/latest/">NIM operator</a> documentation for more details.</p></td></tr><tr><td>LeaderWorkerSet (LWS)</td><td>Distributed inference</td><td>0.6.0 or higher</td><td>NVIDIA Run:ai provides out of the box support for submitting <a href="https://github.com/kubernetes-sigs/lws">LWS</a> workloads <a href="../../../workloads-in-nvidia-run-ai/submit-via-yaml.md">via YAML</a> and NVIDIA Run:ai native <a href="../../../workloads-in-nvidia-run-ai/using-inference/distributed-inference.md">distributed inference</a> workloads using LWS.</td></tr><tr><td>Kubeflow MPI</td><td>Distributed training</td><td>MPI Operator v0.6.0 or higher</td><td>NVIDIA Run:ai provides out of the box support for submitting MPI workloads via API, CLI or UI. See <a href="../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md">Distributed training</a> for more details.</td></tr><tr><td>PyTorch</td><td>Distributed training</td><td>Kubeflow Training Operator v1.9.2</td><td>NVIDIA Run:ai provides out of the box support for submitting PyTorch workloads via API, CLI or UI. See <a href="../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md">Distributed training</a> for more details.</td></tr><tr><td>TensorFlow</td><td>Distributed training</td><td>Kubeflow Training Operator v1.9.2</td><td>NVIDIA Run:ai provides out of the box support for submitting TensorFlow workloads via API, CLI or UI. See <a href="../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md">Distributed training</a> for more details.</td></tr><tr><td>XGBoost</td><td>Distributed training</td><td>Kubeflow Training Operator v1.9.2</td><td>NVIDIA Run:ai provides out of the box support for submitting XGBoost via API, CLI or UI. See <a href="../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md">Distributed training</a> for more details.</td></tr><tr><td>JAX</td><td>Distributed training</td><td>Kubeflow Training Operator v1.9.2</td><td>NVIDIA Run:ai provides out of the box support for submitting JAX workloads via API, CLI or UI. See <a href="../../../workloads-in-nvidia-run-ai/using-training/distributed-training-models.md">Distributed training</a> for more details.</td></tr><tr><td>Triton</td><td>Orchestration</td><td>Any version</td><td>Usage via docker base image</td></tr></tbody></table>

## Development Tools

| Tool | Category | Additional Information |
| --- | --- | --- |
| Jupyter Notebook | Development | NVIDIA Run:ai provides integrated support with Jupyter Notebooks. See [Jupyter Notebook quick start](../../../workloads-in-nvidia-run-ai/using-workspaces/quick-starts/jupyter-quickstart.md) example. |
| PyCharm | Development | Containers created by NVIDIA Run:ai can be accessed via PyCharm. |
| VScode | Development | Containers created by NVIDIA Run:ai can be accessed via Visual Studio Code. You can automatically launch Visual Studio code web from the NVIDIA Run:ai console. |

## Storage and Registries

| Tool | Category | Additional Information |
| --- | --- | --- |
| Docker Registry | Repositories | NVIDIA Run:ai allows using a docker registry as a [Credential](../../../workloads-in-nvidia-run-ai/workload-assets/credentials.md) asset |
| GitHub | Storage | NVIDIA Run:ai communicates with GitHub by defining it as a [data source](../../../workloads-in-nvidia-run-ai/workload-assets/datasources.md) asset |
| S3 | Storage | NVIDIA Run:ai communicates with S3 by defining a [data source](../../../workloads-in-nvidia-run-ai/workload-assets/datasources.md) asset |

## Experiment Tracking and Monitoring

| Tool | Category | Additional Information |
| --- | --- | --- |
| TensorBoard | Experiment tracking | NVIDIA Run:ai comes with a preset TensorBoard [Environment](../../../workloads-in-nvidia-run-ai/workload-assets/environments.md) asset |

## Infrastructure and Cost Optimization

| Tool | Category | Additional Information |
| --- | --- | --- |
| Karpenter | Cost Optimization | NVIDIA Run:ai provides out of the box support for Karpenter to save cloud costs. Integration notes with Karpenter can be found [here](karpenter.md). |
