/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed } from "vue";
import {
  MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
  MIN_CLUSTER_VERSION_FOR_MNNVL_DETECTION,
  MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING,
} from "@/common/version.constant";
import type { ITopologySupport } from "@/models/node-pool-topology.model";
import { useClusterStore } from "@/stores/cluster.store";

export function useNodePoolTopologySupport(clusterId: () => string | undefined) {
  const clusters = useClusterStore();
  return computed<ITopologySupport>(() => {
    const id = clusterId() ?? null;
    return {
      mnnvl: clusters.isClusterVersionSufficient(id, MIN_CLUSTER_VERSION_FOR_MNNVL_DETECTION),
      numa: clusters.isClusterVersionSufficient(id, MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING),
      gpuDevice: clusters.isClusterVersionSufficient(id, MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY),
    };
  });
}
