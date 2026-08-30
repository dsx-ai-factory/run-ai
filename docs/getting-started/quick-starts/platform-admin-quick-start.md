# Quick Start for Platform Administrators

This guide is for platform administrators responsible for configuring, governing, and operating NVIDIA Run:ai across the organization.

The quick start outlines the critical, high-level setup phases you must complete immediately after NVIDIA Run:ai is installed. It focuses on establishing authentication, organizational structure, resource governance, and operational visibility required to enable teams to run workloads.

## Prerequisites

* Access to the NVIDIA Run:ai tenant - You have the NVIDIA Run:ai tenant URL for your organization (for example, `https://<your-domain>`).
* Administrator credentials - You have credentials with system administrator permissions to sign in to the NVIDIA Run:ai UI.

## Complete the Onboarding Wizard

After the infrastructure administrator completes the infrastructure-related steps of the onboarding wizard (cluster connection, authentication, and email configuration), responsibility is handed off to the platform administrator to complete the organizational setup.

### Onboard Your First Research Team

This final step establishes the organizational structure. Provide a team name and add the email addresses of the first team members. If SSO is configured, you can create the team using an identity provider group name instead of adding individual email addresses. A project is then created for the team, initial quotas and permissions are applied, and invitations are sent.

## Ongoing Platform Management

After completing the onboarding wizard, continue managing the platform through the following core activities.

### Configure Platform Behavior and Admin Settings

NVIDIA Run:ai provides global configuration options to control system-wide features and functionality.

Use the [General settings](../../settings/general-settings/) in the Admin panel to tailor how the platform operates, including feature enablement and analytics behavior. These settings apply across all users and workloads and can be adjusted to align with organizational policies and operational requirements.

### Define Authorization and Access Control

Authorization ensures users can access only the features and resources required for their role. Define how users and teams access platform capabilities using [roles](../../infrastructure-setup/authentication/roles.md) and [access rules](../../infrastructure-setup/authentication/accessrules.md) to:

* Grant users the appropriate level of access
* Control who can submit workloads or manage resources
* Extend access by creating additional [custom roles](../../infrastructure-setup/authentication/roles.md#custom-role-management-api-only) as platform usage evolves

### Define Organizational Structure and Quota

Create additional [departments](../../platform-management/aiinitiatives/organization/departments.md) and [projects](../../platform-management/aiinitiatives/organization/projects.md) to logically partition resources. For each project, define guaranteed and over-quota resource allocations to control how resources are shared across the organization. At a high level, these constructs often align with an organization’s internal structure, such as business units, teams, or similar groupings. You can model them to reflect how your organization plans and manages AI initiatives. See [Adapting AI initiatives to your organization](../../platform-management/aiinitiatives/adapting-ai-initiatives.md) for more details.

### Configure Node Pools

Node pools allow you to translate organizational and business priorities into infrastructure-level scheduling decisions. By grouping worker nodes based on hardware type, capabilities, or location (for example, H100 vs. A100 GPUs), [node pools](../../platform-management/aiinitiatives/resources/node-pools.md) enable you to:

* Reserve high-end or scarce GPUs for business-critical or production workloads
* Provide predictable performance and scheduling behavior guarantees for prioritized projects
* Prevent lower-priority or experimental workloads from impacting production or revenue-generating use cases

Node pools form a foundational layer for enforcing resource access, workload placement, and scheduling policies across the organization.

### Define Policies (Governance)

Policies define how workloads are governed and scheduled across the platform. By combining workload policies with scheduling rules, you can standardize workload behavior and control how resources are used and shared.

Use policies to:

* Standardize workload behavior with [workload policies](../../platform-management/policies/native-workload-policies.md), enforcing best practices and organizational limits
* Define scheduling behavior with [scheduling rules](../../platform-management/policies/scheduling-rules.md) that determine how workloads are placed or how long they can run

Scheduling rules are applied at the project or department level and affect all matching workloads for that scope.

### Monitor and Optimize the Platform

Monitor platform usage and health to ensure efficient and reliable operation.

* Monitor usage - Use analytics dashboards to track GPU utilization, identify idle resources, and review consumption by department and project.
* Review logs and system health - Monitor control plane and cluster components to proactively troubleshoot issues and manage maintenance activities.
