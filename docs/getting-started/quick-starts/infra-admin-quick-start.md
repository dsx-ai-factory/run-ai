# Quick Start for Infrastructure Administrators

This guide is for infrastructure administrators responsible for installing, configuring, and operating NVIDIA Run:ai.

The quick start walks through the initial infrastructure setup lifecycle, including platform installation and the essential post-installation configuration required to prepare the cluster for onboarding and workload execution. It focuses on infrastructure-level concerns such as cluster readiness, control plane behavior, security boundaries, and operational stability.

<figure><img src="../../assets/infra-admin.png" alt=""/><figcaption></figcaption></figure>

## Prerequisites

Before you begin, ensure that:

* A Kubernetes cluster is up and running.
* [Helm](https://helm.sh/) 3.14 or later is installed.
* You have `kubectl` access to the cluster with admin-level permissions.

## Installation

The platform supports deployment using two primary methods, depending on your environment:

* [Install using Helm](../../installation-and-configuration/installation/install-using-helm/cp-system-requirements.md) - The standard installation method using Helm charts. Provides full control and flexibility over configuration and deployment.
* [Install using Base Command Manager (BCM)](../../installation-and-configuration/installation/bcm-install.md) - A guided installation method available through NVIDIA Base Command Manager intended to simplify deployment, employing defaults meant to enable most NVIDIA Run:ai capabilities on NVIDIA DGX SuperPOD systems.

## Getting Started: The Onboarding Wizard

After installation, sign in to the NVIDIA Run:ai UI. The onboarding wizard launches automatically and guides you through the required steps.

The wizard includes both infrastructure-level and organizational steps. As an infrastructure administrator, you are responsible for completing the infrastructure-related steps and then handing off the remaining organizational setup to a [platform administrator](platform-admin-quick-start.md#complete-the-onboarding-wizard).

<Info>

**Note**

Do not close the wizard before all steps are complete. The onboarding wizard cannot be reopened once dismissed.

</Info>

### Connect Your Cluster

<Info>

**Note**

If the NVIDIA Run:ai cluster is already deployed and connected to the control plane (e.g., via BCM installation or a pre-run Helm installation), the wizard will automatically detect the connection and skip this part.

</Info>

A cluster is your organization’s compute infrastructure, where AI workloads are executed. The wizard first directs you to review [system](../../installation-and-configuration/installation/install-using-helm/system-requirements.md) and [network](../../installation-and-configuration/installation/install-using-helm/network-requirements.md) requirements. It then generates a Helm command that you run on your Kubernetes cluster to install the required components and prepare the cluster for workload scheduling.

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../infrastructure-setup/procedures/clusters.md) table.

### Configure Platform Authentication

This step integrates NVIDIA Run:ai with your organization’s identity and access management system. Configure [Single Sign-On (SSO)](../../infrastructure-setup/authentication/overview.md#single-sign-on-sso) using SAML 2.0 or OpenID Connect (OIDC) to connect NVIDIA Run:ai to your corporate Identity Provider (IdP).

### Configure Email Server

Configure the [email server](../../settings/general-settings/notifications.md#email-notifications) used by NVIDIA Run:ai to send system notifications and user invitations. Email configuration ensures that users receive onboarding emails, password resets, and other platform notifications. This step prepares the platform for user onboarding and ongoing communication.

## Post Installation Infrastructure Setup

After installing NVIDIA Run:ai, complete the following foundational infrastructure configuration steps to ensure the platform is production-ready and can safely support organizational onboarding and workloads. These steps focus on cluster readiness, control plane behavior, and operational guardrails, rather than day-to-day platform usage:

* Validate node readiness and assign node roles as required
* Configure advanced control plane and cluster settings based on your environment requirements
* Enable required integrations and networking components
* Apply security and operational best practices
* Prepare the platform for scale, availability, and ongoing maintenance

The exact configuration required depends on your environment, scale, and operational model. Detailed procedures and advanced options are documented in the [Advanced setup](../../installation-and-configuration/post-installation-configuration/node-roles.md) and [Infrastructure procedures](../../infrastructure-setup/procedures/scaling.md) sections.
