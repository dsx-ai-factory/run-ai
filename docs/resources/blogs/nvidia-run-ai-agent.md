---
description: 'By: Brad Soper and Vivek Kolasani'
hidden: true
---

# NVIDIA Run:ai Agent Simplifies GPU Provisioning, Job Management, and Troubleshooting

How users interact with complex technical systems is changing. AI agents are replacing command line interfaces (CLIs), dashboards, and memorized API calls with something far more natural: conversation. Anyone from a platform engineer to a researcher who has never written a line of YAML can manage GPU clusters, submit training jobs, and diagnose failures by describing what they need in plain language.

This shift inspired the [NVIDIA Run:ai Agent](https://github.com/runai-professional-services/runai-agent). Built on the [NVIDIA NeMo Agent toolkit](https://github.com/NVIDIA/NeMo-Agent-Toolkit), this open source project turns an [NVIDIA Run:ai](https://www.nvidia.com/en-us/software/run-ai) cluster into a conversational partner. Instead of navigating documentation or deciphering Kubernetes manifests, users ask questions and issue commands in everyday language. The agent handles authentication, API calls, quota checks, and troubleshooting behind the scenes.

### NVIDIA Run:ai Agent Architecture

The NVIDIA Run:ai Agent is an intelligent interface that streamlines how users interact with their compute resources. By using the NVIDIA NeMo Agent Toolkit, it provides a conversational workflow that translates natural language into precise actions within the NVIDIA Run:ai environment. The NVIDIA NeMo Agent toolkit orchestrates the underlying LLM, tool calls, and workflow logic, enabling the agent to securely call NVIDIA Run:ai APIs, validate parameters, and execute multi-step operations on behalf of the user.

The architecture consists of three primary pillars:

* **REST API executor:** This component interacts with the NVIDIA Run:ai control plane. It translates high-level requests (such as "list my running jobs") into authenticated REST calls, fetching real-time data without requiring the user to parse JSON outputs.
* **Kubectl integration:** For deep-stack operations, the agent communicates with the underlying Kubernetes nodes. This enables it to verify pod health, check node affinity, and ensure that the requested resources, like specific NVIDIA GPUs, are available.
* **Unified control plane:** The agent acts as a client to the NVIDIA Run:ai control plane, ensuring all actions respect the multi-tenancy and RBAC policies defined by the organization.

### Install and Configure the NVIDIA Run:ai CLI and Agent

Before engaging with the agent, the environment must be bootstrapped with the NVIDIA Run:ai CLI. This tool serves as the foundational link between your local environment and the GPU cluster.

Download and install the CLI to your local machine or jump host. This utility handles the authentication and low-level communication with the [NVIDIA Run:ai](https://www.nvidia.com/en-us/software/run-ai/) platform.

```shellscript
git clone https://github.com/runai-professional-services/runai-agent.git
# 1. Create namespace and secrets
kubectl create namespace runai-agent

#The agent requires access to the NVIDIA Run:ai environment via an API Key and specific environment variables

kubectl create secret generic runai-creds \
  --namespace runai-agent \
  --from-literal=RUNAI_CLIENT_ID="[YOUR_CLIENT_ID]" \
  --from-literal=RUNAI_CLIENT_SECRET="[YOUR_CLIENT_SECRET]" \
  --from-literal=RUNAI_BASE_URL="https://your-cluster.run.ai"
kubectl create secret generic nvidia-key \
  --namespace runai-agent \
  --from-literal=NVIDIA_API_KEY="[YOUR_NVIDIA_API_KEY]"

# 2. Install with Helm
helm install runai-agent ./deploy/helm/runai-agent \
  --namespace runai-agent \
  --set runai.existingSecret="runai-creds" \
  --set nvidia.existingSecret="nvidia-key" \
  --set failureAnalysis.persistence.storageClassName="your-rwx-storage-class"
  
# Optional: Install RunAI SDK for direct API access
pip install runapy==1.223.0

# 1. Install CLI
cd runai-cli
npm install && npm run build && npm link

# 2. Connect to remote agent (Helm/Docker deployment)
runai-cli connect https://your-agent-url.com

# 3. Verify connection
runai-cli server status

# 4. Use the CLI
runai-cli ask "Show me all projects"
runai-cli submit "Create a training job with 2 GPUs"
runai-cli chat  # Interactive mode
```

### Infrastructure as Conversation: From Idea to Running Workload

Whether provisioning through the NVIDIA Run:ai UI, CLI, or YAML manifests, managing high-performance workloads involves balancing scheduler flags and quota checks.

The transition from intent to execution is where the NVIDIA Run:ai Agent truly shines. It handles the entire lifecycle:

* **Intent:** Use natural language prompts to define what you need.
* **Job template:** The Agent auto-populates the configuration, instantly balancing scheduler flags, GPU resource allocation, and project quotas.
* **Execution:** NVIDIA Run:ai’s scheduling engine kicks in, ensuring your workload lands on the GPU resources.
* **Monitoring:** The Agent analyzes real-time logs through the prompt. If something goes wrong.

For instance, if a researcher needs to launch a training task but doesn’t remember the exact syntax, they can submit the request to the agent in natural language, as shown:

```shellscript
runai-cli submit "Submit a training job called training-job-01 with image runai.jfrog.io/demo/quickstart-demo in project-01 with 1 GPU"
ℹ 🚀 Processing your request...

Request: "Submit a training job called training-job-01 with image runai.jfrog.io/demo/quickstart-demo in project-01 with 1 GPU"

✔ Agent is processing your request...

Response
────────


✅ **Job Submitted Successfully!**
Your training job **training-job-01** is now running in project **project-01** with 1 GPU.

📊 **Monitor Progress**:
- Use the Run:AI UI: [https://runai.example.com/projects/project-01/jobs](https://runai.example.com/projects/project-01/jobs)
- Or check status via CLI/UI using the job name or project.

Let me know if you need further assistance!
```

### Making Resource Governance Transparent and Easy to Manage

As AI teams scale, resource governance becomes mission-critical. Without clear quotas and visibility, clusters quickly devolve into contention and waste. The NVIDIA Run:ai Agent makes governance conversational and transparent.

For an MLOps engineer, creating a new GPU-scoped project becomes a single request:

```shellscript
runai-cli ask "Please create a project named project-02 with 4 GPU quota."
✔ Processing query...

The project "project-02" has been successfully created with a GPU quota of 4. The project ID is "4500003" and it is associated with the cluster ID "72f2930b-9cac-4ae2-b391-0174cec2cb22". The project is currently in the "Creating" phase.
```

### Diagnosing Failures at Scale Instantly

Failed jobs silently drain productivity and compute budgets. Instead of parsing logs across dozens of workloads, the same conversational layer provides structured failure analysis:

```shellscript
runai-cli ask "Can you share with me the failure statistics?"
✔ Processing query...

Here are the failure statistics from the last 7 days:

📉 **Total Failures:** 2 events
📌 **Project:** project-01 (100% of failures)
⚠️ **Failure Type:** "Failed" (2 occurrences)
🔍 **Failed Jobs:**
- `training-error` (last seen: 2026-02-23 16:05:23)
- `training-job-01` (last seen: 2026-02-23 16:02:22)

Would you like me to investigate specific jobs further or check for patterns in failures?
```

### Performance Analytics

To understand how effectively GPUs are being utilized, a user can simply query the agent:

```shellscript
runai-cli ask "Can you share with me the performance analysis?"
✔ Processing query...

Here's the performance analysis for the last 7 days across all projects:

📊 **Overview**
- Total runs with duration data: 2
- Average job duration: 32 hours 34 minutes

📈 **Execution Trends**
- All jobs were in project-01 with consistent duration
- No specific image patterns identified (all marked as "unknown")

🔴 **Failure Analysis**
- Both runs ended in failure (status: "Failed")
- No differentiation by image or project for failure types

💡 **Recommendations**
- For project-01, expect training jobs to take approximately 32-34 hours
- Consider investigating the failed jobs using `runai_kubectl_troubleshoot` for root cause analysis

⚠️ **Anomalies**
- No currently running jobs showing abnormal durations
- Both failed jobs followed typical duration patterns before failure

Would you like me to investigate the failed jobs further or provide more specific details about any particular aspect?
```

### Utilization Visibility

Users can check their standing within a project to see if they are nearing their GPU limits.

```shellscript
runai-cli ask "please list all of the projects with their quota."
✔ Processing query...

Here are the projects with their GPU quotas:

- **project-01**: 6 GPUs (66.7% utilized, 4 GPUs in use)
- **project-02**: 4 GPUs (0% utilized)
- **new-project**: 2 GPUs (0% utilized)

The cluster has a total quota of 12 GPUs, with 4 currently in use (33.3% utilization).
```

### How NVIDIA Run:ai Agent Accelerates the AI Lifecycle

By moving from Code-First to Intent-First orchestration, there are immediate benefits:

* **Faster experimentation:** Shortens the time between a "what if" thought and a running job from minutes to seconds.
* **Natural language:** Uses natural language to gain insights into environment and workload health.
* **Greater productivity:** Platform teams spend less time answering basic "how do I run this" tickets and more time optimizing cluster performance.
* **Lower barrier to entry:** Onboard new researchers who may be experts in PyTorch but novices in infrastructure management.

### Best Practices and Troubleshooting

To get the most out of your NVIDIA Run:ai Agent, consider the following:

* **Use specific job profiles:** Configure templates for common tasks (e.g., "Standard Training" vs. "Interactive Notebook") so the agent can apply best practices automatically.
* **Context awareness:** Remember that the agent understands your current project context. You can say "Restart my last job" without re-specifying the parameters.
* **Custom extensions**: The agent can be extended to interact with internal data registries or specific S3 buckets by modifying the REST Executor's capabilities.

If an agent says it cannot find resources.

* Use the command runai agent status to ensure the agent has the correct permissions to view node availability across the entire cluster.

### Get Started

With deeper integration with the NVIDIA NeMo Agent toolkit, the possibilities for automated hyperparameter tuning and proactive cluster optimization are just beginning. The future of AI development isn't just about more performant GPUs: it's about a more intelligent way to use them.\
\
Give the NVIDIA Run:ai Agent a try on [GitHub](https://github.com/runai-professional-services/runai-agent). Explore the agent, push its boundaries, and help shape the future of intent-driven AI orchestration. Whether you have questions, feature requests, or would like to contribute, we’d love to hear from you through [GitHub issues](https://github.com/runai-professional-services/runai-agent/issues/new/choose).

### About the Authors

**About Vivek Kolasani**

Vivek Kolasani is a solutions architect at NVIDIA, helping customers deploy and optimize AI models in production. He works across the AI lifecycle: from distributed training with PyTorch to inference optimization using frameworks like Dynamo, TensorRT, and NGC containers. Previously at Run:ai, he focused on maximizing GPU efficiency and scaling AI workloads. Vivek holds a master’s degree in computer information science from Cleveland State University. He is based in Dallas

**About Brad Soper**

Brad Soper is a solutions architect at NVIDIA. He has over 20 years of experience in infrastructure, cloud, Kubernetes, and MLOps engineering. His focus has been enabling teams with the skills required to manage, deploy and scale AI workloads. He is based in the greater Seattle, WA area.
