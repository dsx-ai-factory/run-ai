# Compliance

This guide details the data privacy and compliance considerations for deploying NVIDIA Run:ai. It is intended to help administrators and compliance teams understand the data management practices involved with NVIDIA Run:ai. This ensures the permissions align with organizational policies and regulatory requirements before installation and during integration and onboarding of the various teams.

## Data Privacy

When using the NVIDIA Run:ai SaaS cluster, the Control plane operates through the NVIDIA Run:ai cloud, requiring the transmission of certain data for control and analytics. Below is a detailed breakdown of the specific data sent to the NVIDIA Run:ai cloud in the SaaS offering.

<Callout intent="info">
For organizations where data privacy policies do not align with this data transmission, NVIDIA Run:ai offers a self-hosted version. This version includes the control plane on premise and does not communicate with the cloud.
</Callout>

## Data Sent to the NVIDIA Run:ai Cloud

| Asset                  | Details                                                                                                                  |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Workload Metrics       | Includes workload names, CPU, GPU, and memory metrics, as well as parameters provided during the `runai submit` command. |
| Workload Assets        | Covers environments, compute resources, and data resources associated with workloads.                                    |
| Resource Credentials   | Credentials for cluster resources, encrypted with a SHA-512 algorithm specific to each tenant.                           |
| Node Metrics           | Node-specific data including names, IPs, and performance metrics (CPU, GPU, memory).                                     |
| Cluster Metrics        | Cluster-wide metrics such as names, CPU, GPU, and memory usage.                                                          |
| Projects & Departments | Includes names and quota information for projects and departments.                                                       |
| Users                  | User roles within NVIDIA Run:ai, email addresses, and passwords.                                                         |

## Key Consideration

NVIDIA Run:ai ensures that no deep-learning artifacts, such as code, images, container logs, training data, models, or checkpoints, are transmitted to the cloud. These assets remain securely within your organization's firewalls, safeguarding sensitive intellectual property and data.
