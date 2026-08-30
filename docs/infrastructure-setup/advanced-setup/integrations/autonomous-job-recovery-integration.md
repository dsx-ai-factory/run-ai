# Autonomous Job Recovery Integration

NVIDIA Run:ai integrates with Autonomous Job Recovery (AJR) to enhance fault detection and autonomous recovery for large-scale distributed training workloads. Large training runs spanning thousands of GPUs and running for days or weeks are highly susceptible to hardware failures, network issues, and software bugs. Without automated recovery, these failures require manual intervention, resulting in significant downtime and wasted resources.

With this integration, NVIDIA Run:ai reports distributed training workloads to AJR, which analyzes workload and cluster data and returns recommendations - such as restarting a workload on different nodes to avoid faulty hardware.

<Info>

**Note**

This integration is available on [NVIDIA Mission Control (NMC)](https://docs.nvidia.com/mission-control/index.html) deployments only. AJR must be deployed on the cluster before enabling this integration.

</Info>

## Prerequisites

* NMC 2.5 or later, with AJR deployed on the cluster. See [Installation Guide for Autonomous Job Recovery](https://docs.nvidia.com/mission-control/docs/nmc-software-installation-guide/2.2.0/are-install/are-install.html) for more details.
* NVIDIA Run:ai cluster v2.25 or later.

## How it Works

When the integration is enabled, the following sequence occurs:

* **Workload labeling** - NVIDIA Run:ai automatically adds labels to native distributed training workloads. These labels allow AJR to identify and track the workloads it should monitor.
* **Continuous analysis** - AJR continuously monitors cluster, node, and job health data to detect faults and anomalies.
* **Recommendation** - When a fault is detected, AJR sends a recommendation to NVIDIA Run:ai to restart the workload. The recommendation may also include excluding specific nodes from future scheduling.
* **Action** - Depending on the configured integration mode, NVIDIA Run:ai either logs the recommendation for visibility or acts on it automatically by restarting the workload.
* **Node affinity updates** - When AJR recommends excluding specific nodes, NVIDIA Run:ai adds a node affinity rule to the workload before restarting it, preventing it from being scheduled on the identified faulty nodes. Users can also set their own node affinity preferences; NVIDIA Run:ai manages these alongside the AJR-derived affinity without overriding user configuration. Updated node affinity is reflected in both the Workload API and the NVIDIA Run:ai UI.

## Configuring the Integration

The integration is configured per cluster via the `RunaiConfig` custom resource under `spec.advice-restart-controller`. For instructions on editing the `RunaiConfig`, see [Advanced cluster configurations](../../../installation-and-configuration/post-installation-configuration/cluster-config.md).

<Info>

**Note**

When installed via BCM, the integration is enabled in `action` mode by default.

</Info>

| Field | Description |
| --- | --- |
| `enabled` | Enables the AJR integration. When set to `true`, Run:ai deploys the services required for the integration. |
| `mode` | Sets the integration mode. Accepted values: `log` or `action`. See [Integration Modes](autonomous-job-recovery-integration.md#integration-modes) for details. |
| `restartLimit` | The maximum number of times a workload can be restarted based on AJR recommendations. Once reached, recommendations are still logged but no restart action is taken. |
| `maxRetriesPerAdvice` | The maximum number of times NVIDIA Run:ai will try to perform an AJR advice (for example, to restart the workload). Once reached, NVIDIA Run:ai will no longer try for this specific advice |

## Integration Modes

AJR recommendations and restart actions are surfaced as events in the workload Event history tab, for example: a recommendation received from AJR or a workload restart triggered. See [Workloads](../../../workloads-in-nvidia-run-ai/workloads.md#event-history) for more details.

### Log Mode

In log mode, AJR recommendations are received and recorded in the workload event history for visibility and troubleshooting, but no action is taken. Use this mode to evaluate AJR recommendations before enabling automated recovery.

### Action Mode

In action mode, NVIDIA Run:ai automatically acts on AJR recommendations by restarting the affected workload. Restarts are triggered for every AJR recommendation, up to the restart limit of the workload.

When a restart is triggered, NVIDIA Run:ai fully deletes and recreates the workload's job and pods while preserving the workload identity, restart count, and existing resources (such as persistent storage). If AJR recommends restarting on different nodes, NVIDIA Run:ai updates the node affinity accordingly before restarting, merging it with any existing node affinity on the workload

Workloads in **Suspended**, **Suspending**, or **Completed** states are not restarted. AJR recommendations for these workloads are still logged.

## Disabling the Integration per Workload

By default, all workloads follow the cluster-level integration configuration set by the administrator. To disable the AJR integration for a specific workload, add the following annotation when submitting it:

```bash
run.ai/disable-ajr-integration=true
```

When this annotation is set to `true`, NVIDIA Run:ai will not operate the AJR integration for that workload. If the annotation is set to `false` or is not present, the workload follows the cluster-level configuration.
