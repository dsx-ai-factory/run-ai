/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import {
  NodepoolGPUNetworkAccelerationDetection,
  type Nodepool,
  type NodepoolUpdateFields,
  type PlacementStrategy,
} from "@/swagger-models/cluster-service-client";
import {
  DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
  NODE_POOL_CONFIGURATION_DEFAULTS as defaults,
} from "@/models/node-pool.model";

export interface TopologyDraft {
  networkTopologyId: string | null;
  gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection;
  gpuNetworkAccelerationLabelKey: string;
  numaAwareScheduling: boolean;
  gpuNodePlacement: PlacementStrategy;
  gpuDevicePlacement: PlacementStrategy;
  cpuNodePlacement: PlacementStrategy;
}

export interface TopologySupport {
  mnnvl: boolean;
  numa: boolean;
  gpuDevice: boolean;
  networkTopology: boolean;
}

export function topologyDraftFromPool(pool: Nodepool): TopologyDraft {
  const configuration = pool.schedulingConfiguration;
  return {
    networkTopologyId: pool.networkTopologyId ?? null,
    gpuNetworkAccelerationDetection: pool.gpuNetworkAccelerationDetection ?? defaults.gpuNetworkAccelerationDetection,
    gpuNetworkAccelerationLabelKey: pool.gpuNetworkAccelerationLabelKey ?? "",
    numaAwareScheduling: configuration?.numaAwareScheduling ?? defaults.numaAwareScheduling,
    gpuNodePlacement: configuration?.gpuPodsPlacementStrategy?.node ?? defaults.gpuNodePlacement,
    gpuDevicePlacement: configuration?.gpuPodsPlacementStrategy?.device ?? defaults.gpuDevicePlacement,
    cpuNodePlacement: configuration?.cpuPodsPlacementStrategy?.node ?? defaults.cpuNodePlacement,
  };
}

export function validateTopologyLabelKey(value: string): true | string {
  if (!value.trim()) return true;
  const parts = value.split("/");
  const name = parts.at(-1) ?? "";
  const validName =
    name.length <= 63 && /^[a-zA-Z0-9]/.test(name) && /[a-zA-Z0-9]$/.test(name) && !/[^a-zA-Z0-9_.-]/.test(name);
  const prefix = parts.length === 2 ? parts[0] : undefined;
  const validPrefix =
    prefix === undefined ||
    (prefix.length <= 253 &&
      prefix.split(".").every((part) => /^[a-z0-9]/.test(part) && /[a-z0-9]$/.test(part) && !/[^a-z0-9-]/.test(part)));
  return parts.length <= 2 && validName && validPrefix
    ? true
    : "Enter a valid Kubernetes label key: a name up to 63 characters with an optional lowercase DNS prefix.";
}

export function buildTopologyPatch(draft: TopologyDraft, supports: TopologySupport): NodepoolUpdateFields {
  const usesLabel =
    supports.mnnvl && draft.gpuNetworkAccelerationDetection !== NodepoolGPUNetworkAccelerationDetection.DontUse;
  if (usesLabel) {
    const validation = validateTopologyLabelKey(draft.gpuNetworkAccelerationLabelKey);
    if (validation !== true) throw new Error(validation);
  }
  return {
    // The backend clears the association when this field is absent, even if its control is version-gated.
    networkTopologyId: draft.networkTopologyId,
    ...(supports.mnnvl ? { gpuNetworkAccelerationDetection: draft.gpuNetworkAccelerationDetection } : {}),
    ...(usesLabel
      ? {
          gpuNetworkAccelerationLabelKey:
            draft.gpuNetworkAccelerationLabelKey.trim() || DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
        }
      : {}),
    schedulingConfiguration: {
      gpuPodsPlacementStrategy: {
        node: draft.gpuNodePlacement,
        ...(supports.gpuDevice ? { device: draft.gpuDevicePlacement } : {}),
      },
      cpuPodsPlacementStrategy: { node: draft.cpuNodePlacement },
      ...(supports.numa ? { numaAwareScheduling: draft.numaAwareScheduling } : {}),
    },
  };
}
