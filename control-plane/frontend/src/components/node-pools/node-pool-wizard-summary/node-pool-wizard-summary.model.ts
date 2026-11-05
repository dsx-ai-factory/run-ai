/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { isProportional } from "@/components/node-pools/node-pool-create-flow/node-pool-create-flow.model";
import type {
  Nodepool,
  NodepoolCreateFields,
  SchedulingCapabilitiesNullable,
} from "@/swagger-models/cluster-service-client";
import { NodepoolGPUNetworkAccelerationDetection } from "@/swagger-models/cluster-service-client";

export type NodePoolSummaryConfiguration = NodepoolCreateFields &
  Partial<Pick<Nodepool, "clusterName" | "networkTopologyName">>;

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

export type NodePoolSchedulingMode = "Unknown" | "Guaranteed quota" | "Ranked" | "Proportional" | "Custom";

export function getSchedulingMode(capabilities?: SchedulingCapabilitiesNullable): NodePoolSchedulingMode {
  if (!capabilities) return "Unknown";
  if (capabilities.quota) return "Guaranteed quota";
  if (capabilities.quota === false && capabilities.ranks) return "Ranked";
  if (isProportional(capabilities)) return "Proportional";
  return "Custom";
}

export const SCHEDULING_MODE_DESCRIPTIONS: Partial<Record<NodePoolSchedulingMode, string>> = {
  Proportional: "GPU resources are distributed proportionally according to each department's and project's weight.",
};
