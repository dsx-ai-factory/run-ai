/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import { useClusterStore } from "@/stores/cluster.store";

import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { type Nodepool, type NodepoolMinimal, NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";

export const useNodePoolStore = defineStore("NodePool", {
  state: () => ({
    nodePools: [] as Array<Nodepool>,
    nodePoolsCount: 0 as number,
    clusterStore: useClusterStore(),
    nodePoolsByClusterId: new Map<string, NodepoolMinimal[]>(),
  }),
  getters: {
    isOnlyDefaultNodePool(): boolean {
      return this.nodePoolsCount === 1;
    },
  },
  actions: {
    async isOnlyDefaultNodePoolByClusterId(clusterId: string): Promise<boolean> {
      const existingNodePools = this.nodePoolsByClusterId.get(clusterId);

      if (existingNodePools) {
        return existingNodePools.length <= 1;
      }

      const fetchedNodePools = await clusterService.getNodepoolsMinimal([
        filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterId),
      ]);

      this.nodePoolsByClusterId.set(clusterId, fetchedNodePools);

      return fetchedNodePools.length <= 1;
    },
    removeNodePoolsByClusterId(clusterId: string): void {
      this.nodePoolsByClusterId.delete(clusterId);
    },
  },
});
