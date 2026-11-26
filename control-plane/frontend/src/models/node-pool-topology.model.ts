/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import {
  NodepoolGPUNetworkAccelerationDetection,
  type GpuPodsPlacementStrategy,
  type NodepoolUpdateFields,
  type PlacementStrategy,
  type SchedulingConfiguration,
} from "@/swagger-models/cluster-service-client";
import {
  type INodePoolSummaryConfiguration,
  DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
  NODE_POOL_CONFIGURATION_DEFAULTS as defaults,
} from "@/models/node-pool.model";

export interface ITopologyDraft {
  networkTopologyId: string | null;
  gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection;
  gpuNetworkAccelerationLabelKey: string;
  numaAwareScheduling: boolean;
  gpuNodePlacement: PlacementStrategy;
  gpuDevicePlacement: PlacementStrategy;
  cpuNodePlacement: PlacementStrategy;
}

export interface ITopologySupport {
  mnnvl: boolean;
  numa: boolean;
  gpuDevice: boolean;
}

export function hasMnnvlChanges(initial: ITopologyDraft, draft: ITopologyDraft): boolean {
  return (
    draft.gpuNetworkAccelerationDetection !== initial.gpuNetworkAccelerationDetection ||
    draft.gpuNetworkAccelerationLabelKey !== initial.gpuNetworkAccelerationLabelKey
  );
}

export function topologyDraftFromPool(pool: INodePoolSummaryConfiguration): ITopologyDraft {
  const configuration = pool.schedulingConfiguration;
  return {
    networkTopologyId: pool.networkTopologyId ?? null,
    gpuNetworkAccelerationDetection: pool.gpuNetworkAccelerationDetection ?? defaults.gpuNetworkAccelerationDetection,
    gpuNetworkAccelerationLabelKey: pool.gpuNetworkAccelerationLabelKey ?? "",
    numaAwareScheduling: configuration?.numaAwareScheduling ?? defaults.numaAwareScheduling,
    gpuNodePlacement:
      configuration?.gpuPodsPlacementStrategy?.node ??
      configuration?.placementStrategy?.gpu ??
      defaults.gpuNodePlacement,
    gpuDevicePlacement: configuration?.gpuPodsPlacementStrategy?.device ?? defaults.gpuDevicePlacement,
    cpuNodePlacement:
      configuration?.cpuPodsPlacementStrategy?.node ??
      configuration?.placementStrategy?.cpu ??
      defaults.cpuNodePlacement,
  };
}

export function buildTopologyPatch(
  initial: ITopologyDraft,
  draft: ITopologyDraft,
  supports: ITopologySupport,
  canReadTopologies: boolean,
): NodepoolUpdateFields {
  const patch: NodepoolUpdateFields = {};
  if (canReadTopologies && draft.networkTopologyId !== initial.networkTopologyId) {
    patch.networkTopologyId = draft.networkTopologyId;
  }
  if (supports.mnnvl && hasMnnvlChanges(initial, draft)) {
    if (draft.gpuNetworkAccelerationDetection !== initial.gpuNetworkAccelerationDetection) {
      patch.gpuNetworkAccelerationDetection = draft.gpuNetworkAccelerationDetection;
    }
    if (draft.gpuNetworkAccelerationDetection !== NodepoolGPUNetworkAccelerationDetection.DontUse) {
      const initialLabel = initial.gpuNetworkAccelerationLabelKey || DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY;
      const label = draft.gpuNetworkAccelerationLabelKey.trim() || DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY;
      if (label !== initialLabel) patch.gpuNetworkAccelerationLabelKey = label;
    }
  }

  const scheduling: SchedulingConfiguration = {};
  const gpuPlacement: GpuPodsPlacementStrategy = {};
  if (draft.gpuNodePlacement !== initial.gpuNodePlacement) gpuPlacement.node = draft.gpuNodePlacement;
  if (supports.gpuDevice && draft.gpuDevicePlacement !== initial.gpuDevicePlacement) {
    gpuPlacement.device = draft.gpuDevicePlacement;
  }
  if (Object.keys(gpuPlacement).length) scheduling.gpuPodsPlacementStrategy = gpuPlacement;
  if (draft.cpuNodePlacement !== initial.cpuNodePlacement) {
    scheduling.cpuPodsPlacementStrategy = { node: draft.cpuNodePlacement };
  }
  if (supports.numa && draft.numaAwareScheduling !== initial.numaAwareScheduling) {
    scheduling.numaAwareScheduling = draft.numaAwareScheduling;
  }
  if (Object.keys(scheduling).length) patch.schedulingConfiguration = scheduling;
  return patch;
}
