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
