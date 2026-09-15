# What's New in Version 3.0

The NVIDIA Run:ai v3.0 what's new provides a detailed summary of the latest features, enhancements, and updates introduced in this version. It serves as a guide to help you understand the new capabilities and how to leverage them for improved workload management, resource optimization, and more.

For a full list of changes, see the GitHub changelog (TBD).

<Info>

**Note**

For a complete list of deprecated features, APIs, and capabilities, see [Deprecation Notifications](deprecation-notifications.md).

</Info>

## Feature Life Cycle

NVIDIA Run:ai uses life cycle labels to indicate the maturity and stability of features across releases:

* <Badge intent="success">Experimental</Badge> - This feature is in early development. It may not be stable and could be removed or changed significantly in future versions. Use with caution.
* <Badge intent="warning">Beta</Badge> - This feature is still being developed for official release in a future version and may have some limitations. Use with caution.
* <Badge intent="note">Legacy</Badge> - This feature is scheduled to be removed in future versions. We recommend using alternatives if available. Use only if necessary.

## Upgrade to 3.0

Upgrading to v3.0 requires administrators to perform additional steps beyond the base upgrade procedure. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for the specific changes to plan for, and [Upgrade NVIDIA Run:ai](../../installation-and-configuration/upgrade/upgrade.md) for the base upgrade procedure.

### Backup Before Upgrading

NVIDIA Run:ai v3.0 changes how third-party service images are sourced: instead of NVIDIA Run:ai bundling these images, you now pull them yourself from the upstream registry or your own mirror. Back up your cluster configurations, control plane configurations, and third-party services (PostgreSQL, Thanos) before upgrading. See [Back Up Before Upgrading](../../installation-and-configuration/upgrade/backup-before-upgrading.md) for the full backup procedure.

### GPU Fractions Upgrade

NVIDIA Run:ai v3.0 replaces the proprietary engine behind GPU fractions with GPU sharing built directly into NVIDIA CUDA. This makes the capability part of the standard NVIDIA stack, with broader compatibility and long-term maintenance by NVIDIA. Native workload types require no changes; supported workload types require changes only if their manifest sets a fraction limit. See [Migrate GPU Fractions](../../installation-and-configuration/upgrade/migrate-gpu-fractions.md) for prerequisites and the full procedure.

### Workload Model Unification

<Warning>

**TODO:** not sure what the name of this heading should be. Ofer, please advise.

</Warning>

The controllers behind native workload types (workspace, training, distributed training, standard inference, distributed inference) are removed in v3.0 and replaced by a single, unified controller.

Workloads submitted before upgrading through a legacy workload type - workspace, training, distributed training, standard inference, or distributed inference - continue running after the upgrade, but lose some capabilities until resubmitted. See [Migrate Legacy Workloads](../../installation-and-configuration/upgrade/migrate-legacy-workloads.md) for the affected capabilities and the resubmission procedure.

<Info>

**Note**

Resubmitting an affected workload through the NVIDIA Run:ai UI, API, or CLI binds it to the unified controller.

</Info>

### Dynamic Resource Allocation (DRA) Driver

A fresh GPU Operator installation installs the NVIDIA DRA driver automatically, but upgrading an existing GPU Operator deployment does not. If your cluster requires the DRA driver (for example, Multi-Node NVLink deployments such as GB200), install it manually after upgrading the GPU Operator. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for details.

### GPU Memory SWAP Disabled

Upgrading to v3.0 forces GPU Memory SWAP and the node-level scheduler off, even if they were previously enabled, and neither can be re-enabled afterward. The node-level scheduler was deprecated in v2.25 and GPU Memory SWAP in v2.26. Snapshot, the planned replacement capability, is targeted for v3.1 and is not available in v3.0. See [Plan the Upgrade](../../installation-and-configuration/upgrade/plan-the-upgrade.md) for what to check before upgrading.

### Air-Gapped Upgrades

Air-gapped upgrades now use the `runai-installer` binary's `airgap pack` and `airgap load` commands to rebuild your private registry mirror and generate fresh Helm values for the target version, replacing the previous NGC/JFrog `.tgz` download and hand-edited Helm command. See [Upgrade NVIDIA Run:ai](../../installation-and-configuration/upgrade/upgrade.md) for the full procedure.
