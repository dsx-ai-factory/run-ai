# Support, Lifecycle, and Deprecation Policy

## Supported Product

This page defines the support tiers, version lifecycle, and deprecation policy for the Support-entitled NVIDIA Run:ai self-hosted offering.

Support periods and coverage are governed by the applicable NVIDIA licensing agreement. This page describes the standard policy.

NVIDIA Run:ai support periods are governed by the terms of your applicable deployment model and the [NVIDIA Enterprise Support Policy](https://www.nvidia.com/en-us/agreements/enterprise-services/nvidia-enterprise-support-policy/).

The terms applicable to the different deployment models are listed below:

- [NVIDIA Cloud Agreement](https://www.nvidia.com/en-us/agreements/cloud-services/nvidia-cloud-agreement/) and Service Specific Terms govern the SaaS Run:ai Control Plane Service.
- [NVIDIA Software License Agreement](https://www.nvidia.com/en-us/agreements/enterprise-software/nvidia-software-license-agreement/) and [Product Specific Terms for NVIDIA AI Products](https://www.nvidia.com/en-us/agreements/cloud-services/runai-control-plane/) govern self-hosted Run:ai Control Plane Services.

## Overview

NVIDIA Run:ai follows a structured release process so customers receive regular updates, security patches, and predictable deployment stability. Every release moves through three lifecycle phases:

- Full support - The version receives both critical and important bug fixes.
- Extended support - The version receives critical bug fixes only.
- End of support - No further updates are issued; customers should upgrade.

Versions follow semantic versioning (Ma.Mi.Pa). Ma.Mi denotes a major release carrying features and security updates; Pa denotes a patch release carrying fixes only.

## Support Levels

| Level | Critical bug fixes | Important bug fixes | Coverage window |
| --- | --- | --- | --- |
| Full support | Yes | Yes | Months 0–9 after release |
| Extended support | Yes | — | Months 10–12 after release |
| End of support | — | — | Month 12 onward |

A critical bug is a severity-1 ticket under the NVIDIA Run:ai support agreement. An important bug is a severity-1 or severity-2 ticket under the same agreement.

Each version carries a 12-month total lifecycle - 9 months of full support followed by 3 months of extended support.

This policy applies as of October 13th, 2026 and does not reduce a support period or deprecation commitment previously announced for an already released version or during an existing Subscription.

## Component Version Compatibility

NVIDIA Run:ai ships three components: control plane, cluster, and CLI, released together with matching version numbers. For a supported configuration:

- The control plane version must be greater than or equal to the cluster version, and
- The CLI version must match the cluster version.

## Supported Versions

| Version | Release date | End of full support | End of extended support |
| --- | --- | --- | --- |
| 3.0 (latest) | October 13, 2026 | July 13, 2027 | October 13, 2027 |
| 2.26 | July 12, 2026 | April 12, 2027 | July 12, 2027 |
| 2.25 | May 3, 2026 | February 3, 2027 | May 3, 2027 |
| 2.24 | January 18, 2026 | October 18, 2026 | January 18, 2027 |

If a deployment is approaching the end of support, plan an upgrade to stay current on security fixes and functionality.

## Deprecation Policy

Deprecation is the formal, advance signal that a feature, API, CLI command, configuration option, or capability will be removed. It gives customers time to migrate before removal.

### Deprecation Window

A deprecated capability remains available for at least two subsequent releases from the date of the deprecation notice. It is removed no earlier than the end of that window, and removal is always announced in the [Deprecation Notifications](whats-new/deprecation-notifications.md) of the version that removes it.

### What Is Announced, and Where

- Deprecations are always announced in [Deprecation Notifications](whats-new/deprecation-notifications.md) for the release that introduces the deprecation.
- Each deprecation states what is deprecated, the reason, the earliest release in which it may be removed, and the recommended migration path or replacement.
- Existing deployments continue to function throughout the deprecation window. Deprecation is a signal to migrate, not an immediate loss of function.

### Breaking Changes

A breaking change is a change to a public API endpoint or parameter that requires customer action to preserve existing behavior after upgrade - that is, the removal or incompatible change of a public API endpoint or a request or response parameter. Breaking changes are introduced only at a major or minor (Ma.Mi) release, never in a patch (Pa) release, and are always called out in that release's [Deprecation Notifications](whats-new/deprecation-notifications.md).
