/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IStatusColOptions } from "@/models/table.model";
import { NodepoolGPUNetworkAccelerationDetection, NodepoolPhase } from "@/swagger-models/cluster-service-client";
import type { ISelectOption } from "@/models/global.model";
import { docsUtil } from "@/utils/docs.util/docs.util";
import DOCS from "@/constants/docs";

export enum PlacementStrategy {
  Binpack = "binpack",
  Spread = "spread",
}

export const binPackLabel = "Bin-pack";
export const spreadLabel = "Spread";

export const GPU_NODE_PLACEMENT_TOOLTIP = "Determines how workloads are distributed across nodes in the node pool.";
export const GPU_DEVICE_PLACEMENT_TOOLTIP =
  "Determines how workloads are distributed across GPU devices within each node.";
export const CPU_NODE_PLACEMENT_TOOLTIP = "Determines how CPU workloads are distributed across nodes in the node pool.";

export const PlacementOptions = {
  gpu: [
    {
      label: binPackLabel,
      value: PlacementStrategy.Binpack,
      toolTip:
        "Place as many workloads as possible in each GPU and node to use fewer resources and maximize GPU and node vacancy.",
    },
    {
      label: spreadLabel,
      value: PlacementStrategy.Spread,
      toolTip:
        "Spread workloads across as many GPUs and nodes as possible to minimize the load and maximize the available resources per workload.",
    },
  ],
  gpuDevice: [
    {
      label: binPackLabel,
      value: PlacementStrategy.Binpack,
      toolTip: "Place as many workloads as possible on each GPU device to minimize the number of GPU devices used.",
    },
    {
      label: spreadLabel,
      value: PlacementStrategy.Spread,
      toolTip: "Spread workloads across GPU devices within each node to maximize the available resources per workload.",
    },
  ],
  cpu: [
    {
      label: binPackLabel,
      value: PlacementStrategy.Binpack,
      toolTip:
        "Place as many workloads as possible in each CPU and node to use fewer resources and maximize CPU and node vacancy.",
    },
    {
      label: spreadLabel,
      value: PlacementStrategy.Spread,
      toolTip:
        "Spread workloads across as many CPUs and nodes as possible to minimize the load and maximize the available resources per workload.",
    },
  ],
};

export const NodePoolStatusMessageMap = new Map([
  [
    "Error Deleting NodePool",
    {
      newMessage: DOCS.SUPPORT.URL
        ? `This node pool can't be deleted. If this issue persists, contact ${docsUtil.getSupportAnchor()}`
        : "This node pool can't be deleted.",
    },
  ],
  [
    "Scheduler is not ready",
    {
      newMessage: DOCS.SUPPORT.URL
        ? `The Run:ai scheduler is not ready. If the issue persists contact ${docsUtil.getSupportAnchor()}`
        : "The Run:ai scheduler is not ready.",
    },
  ],
  [
    "Nodes unavailable",
    {
      newMessage: DOCS.SUPPORT.URL
        ? `Nodes are unavailable, if the issue persists contact ${docsUtil.getSupportAnchor()}`
        : "Nodes are unavailable.",
    },
  ],
]);

export const NodePoolStatusMessagePrefix = [
  {
    originalMessage: "Scheduler is not ready, reason:",
    newMessage: "The Run:ai scheduler isn't ready because:\n",
  },
  {
    originalMessage: "Unschedulable Nodes:",
    newMessage: "The following nodes within the node pool aren't ready:\n",
  },
  {
    originalMessage: "Nodes Being Drained:",
    newMessage:
      "Seems there are workloads on the nodes under this node pool which are not associated with this node pool. Please delete all these running/failed workloads.\nThe nodes are:",
  },
  {
    originalMessage:
      "The following node(s) within the node pool aren't ready because they are assigned to a different node pool but haven't been drained yet:",
    newMessage: "The following nodes are assigned to this node pool and will be added once\n they finish draining:\n",
  },
  {
    originalMessage: "The following node(s) are assigned to this node pool and will be added once they finish draining:",
    newMessage:
      "The following nodes within the node pool aren't ready because they\n are assigned to a different node pool but haven't been drained yet:\n",
  },
];

export const NodePoolStatusMessageAllowedPrefixesToKeep: string[] = ["The following node(s)"];

export const NRT_NODEPOOL_PREREQ_MESSAGE_PREFIX =
  "The following prerequisites are not met on one or more nodes in the node pool:";

export interface INodePoolsNameAndId {
  name: string;
  id: string;
}

export const NodePoolPhaseMap: Record<NodepoolPhase, IStatusColOptions> = {
  [NodepoolPhase.Creating]: {
    status: "Creating...",
    displayAnimation: true,
  },
  [NodepoolPhase.Updating]: {
    status: "Updating...",
    displayAnimation: true,
  },
  [NodepoolPhase.Deleting]: {
    status: "Deleting...",
    displayAnimation: true,
  },
  [NodepoolPhase.Ready]: {
    status: "Ready",
    color: "success",
    displayAnimation: false,
  },
  [NodepoolPhase.Empty]: {
    status: "Empty",
    displayAnimation: false,
  },
  [NodepoolPhase.Deleted]: {
    status: "Deleted",
    displayAnimation: false,
  },
  [NodepoolPhase.Unschedulable]: {
    status: "Unschedulable",
    color: "negative",
    displayAnimation: false,
  },
  [NodepoolPhase.MissingPrerequisites]: {
    status: "Missing Prerequisites",
    color: "warning",
    displayAnimation: false,
  },
};

//Network acceleration - MNNVL Multi Node NVLink
export const DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY = "nvidia.com/gpu.clique";

export const NODE_POOL_NETWORK_DISCOVERY_OPTIONS: ISelectOption[] = [
  {
    label: "Auto-detect",
    value: NodepoolGPUNetworkAccelerationDetection.Auto,
    toolTip:
      "Automatically detect whether the node pool contains any MNNVL nodes. MNNVL nodes that share the same ID are part of the same NVL rack.",
  },
  {
    label: "MNNVL is present",
    value: NodepoolGPUNetworkAccelerationDetection.Use,
  },
  {
    label: "MNNVL is not present",
    value: NodepoolGPUNetworkAccelerationDetection.DontUse,
  },
];

export const EMPTY_PRIORITY_LABEL = "empty";

import type { NodepoolCreateFields, SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client";
// TODO: include limits: false after #4369 merges, including isProportional.
export const PROPORTIONAL_CAPABILITIES = Object.freeze({
  adjustableWeights: true,
  quota: false,
  ranks: false,
  cpuQuota: false,
} satisfies SchedulingCapabilitiesNullable);

// Mirror defaults applied to omitted fields by cluster-service's nodepoolSetDefaults.
// Keep optional version-gated fields out of the create payload until explicitly selected.
export const NODE_POOL_CONFIGURATION_DEFAULTS = {
  gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto,
  numaAwareScheduling: false,
  gpuNodePlacement: PlacementStrategy.Binpack,
  gpuDevicePlacement: PlacementStrategy.Binpack,
  cpuNodePlacement: PlacementStrategy.Binpack,
} as const;

export function isProportional(capabilities?: SchedulingCapabilitiesNullable): boolean {
  return (
    capabilities?.adjustableWeights === true &&
    capabilities.quota === false &&
    capabilities.ranks === false &&
    capabilities.cpuQuota === false
  );
}
export interface INodePoolSummaryConfiguration extends NodepoolCreateFields {
  clusterName?: string;
  networkTopologyName?: string | null;
}

// Supplied by inventory/telemetry once available, never inferred from cluster-wide capacity.
export interface INodePoolSummaryResources {
  nodes?: number;
  gpus?: number;
  gpuMemoryBytes?: number;
  gpuTypes?: string[];
}

export const NETWORK_ACCELERATION_LABELS: Record<NodepoolGPUNetworkAccelerationDetection, string> = {
  [NodepoolGPUNetworkAccelerationDetection.Auto]: "Auto detect",
  [NodepoolGPUNetworkAccelerationDetection.Use]: "Is present",
  [NodepoolGPUNetworkAccelerationDetection.DontUse]: "Is not present",
};

export enum ENodePoolSchedulingMode {
  Unknown = "Unknown",
  Guaranteed = "Guaranteed quota",
  Ranked = "Ranked",
  Proportional = "Proportional",
  Custom = "Custom",
}

export function getSchedulingMode(capabilities?: SchedulingCapabilitiesNullable): ENodePoolSchedulingMode {
  if (!capabilities) return ENodePoolSchedulingMode.Unknown;
  if (capabilities.quota) return ENodePoolSchedulingMode.Guaranteed;
  if (capabilities.quota === false && capabilities.ranks) return ENodePoolSchedulingMode.Ranked;
  if (isProportional(capabilities)) return ENodePoolSchedulingMode.Proportional;
  return ENodePoolSchedulingMode.Custom;
}

export const SCHEDULING_MODE_DESCRIPTIONS: Partial<Record<ENodePoolSchedulingMode, string>> = {
  Proportional: "GPU resources are distributed proportionally according to each department's and project's weight.",
};

export const NODE_POOL_LABEL_REGEX = {
  labelKey: "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$",
  labelValue: "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$",
};
