/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { INodeGpu } from "@/models/node-gpu.model";

import { type NodeStatusConditionDetails, NodeStatusEnum } from "@/swagger-models/cluster-service-client";
import type { IStatusColOptions } from "@/models/table.model";

export enum EMetricsType {
  Default = "Resource utilization",
  Advanced = "GPU profiling",
  Inference = "Inference",
  NvidiaNim = "NVIDIA NIM",
  Vllm = "vLLM",
}

export const NrtHealthyConditionType = "MissingNrtHealthyPrerequisite";
export const GpuFractioningConditionType = "MissingGpuFractioningPrerequisite";
export const GpuFractioningNotManagedReason = "NotManaged";
export const GpuFractioningNotManagedPrerequisite = "The node is not covered by the GPU fractioning operator.";
export const NodeStatusMap: Record<NodeStatusEnum, IStatusColOptions> = {
  [NodeStatusEnum.Ready]: {
    status: NodeStatusEnum.Ready,
    color: "success",
    displayAnimation: false,
  },
  [NodeStatusEnum.NotReady]: {
    status: "Not Ready",
    color: "negative",
    displayAnimation: false,
  },
  [NodeStatusEnum.Unknown]: {
    status: "Unknown",
    displayAnimation: false,
  },
  [NodeStatusEnum.MissingPrerequisites]: {
    status: "Missing Prerequisites",
    color: "warning",
    displayAnimation: false,
  },
};
export const NodeConditionMap: Record<string, NodeStatusConditionDetails> = {
  ["DiskPressure"]: {
    type: "DiskPressure",
    reason: "Disk pressure",
    message: "Disk pressure: Disk capacity is low.",
  },
  ["MemoryPressure"]: {
    type: "MemoryPressure",
    reason: "Memory pressure",
    message: "Memory pressure: Node memory is low.",
  },
  ["PidPressure"]: {
    type: "PidPressure",
    reason: "PID pressure",
    message: "PID pressure: Too many processes on the node.",
  },
  ["NetworkUnavailable"]: {
    type: "NetworkUnavailable",
    reason: "Network unavailable",
    message: "Network unavailable: Network is not configured correctly for the node.",
  },
  ["SchedulingDisabled"]: {
    type: "SchedulingDisabled",
    reason: "Scheduling disabled",
    message:
      "Scheduling disabled: Node might be cordoned and marked as unavailable to the scheduler. Uncordon it to make it available.",
  },
  ["UndrainedMigratedNode"]: {
    type: "UndrainedMigratedNode",
    reason: "Undrained migrated node",
    message: "Undrained migrated node: For the node to be ready again, evict all pods.",
  },
  ["MissingNvidiaContainerToolkit"]: {
    type: "MissingNvidiaContainerToolkit",
    reason: "Missing NVIDIA Container Toolkit",
    message: "Missing NVIDIA container toolkit: For the node to be ready again, install NVIDIA container toolkit.",
  },
  ["MissingNvidiaDcgmExporter"]: {
    type: "MissingNvidiaDcgmExporter",
    reason: "Missing nvidia DCGM Exporter",
    message: "Missing NVIDIA DCGM: Some metrics are missing. For the full list, install the NVIDIA DCGM.",
  },
  ["NodeNotReady"]: {
    type: "NodeNotReady",
    reason: "Node not ready",
    message: "Node not ready.",
  },
  [GpuFractioningConditionType]: {
    type: GpuFractioningConditionType,
    reason: "Missing GPU fractioning prerequisites",
    message: "The GPU fractioning operator did not report a reason.",
  },
};
export interface INode {
  id: string;
  node: string;
  nodeStatus?: string;
  utilization: string;
  gpuType?: string;
  totalGpus?: string;
  allocatedGpus: string;
  usedGpus: string;
  totalCpus: string;
  allocatedCpus: string;
  usedCpus: string;
  totalGpuMemory: string;
  usedGpuMemory: string;
  totalCpuMemory: string;
  allocatedMemory: string;
  usedCpuMemory: string;
  swapCpuMemory: string;
  freeGpus?: string | null;
  nodePool: string;
  gpuDevices?: Array<INodeGpu>;
  conditions?: Array<NodeStatusConditionDetails>;
  clusterId: string;
}

export interface INodeGpusModalOptions {
  header: string;
  gpuDevices: INodeGpu[];
}

export interface INodeModalOptions {
  clusterId: string;
  nodePoolName: string;
}
