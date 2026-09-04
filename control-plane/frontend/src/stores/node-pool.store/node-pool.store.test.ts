/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useNodePoolStore } from "./node-pool.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepoolsMinimal: vi.fn(),
  },
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({}),
}));

const mockNodePools: NodepoolMinimal[] = [
  { id: "1", name: "default", clusterId: "cluster-1" },
  { id: "2", name: "gpu-pool", clusterId: "cluster-1" },
];

describe("NodePoolStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  describe("isOnlyDefaultNodePoolByClusterId", () => {
    it("should call getNodepoolsMinimal with correct filter", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledWith(["clusterId==cluster-1"]);
    });

    it("should return true when only one node pool exists", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([mockNodePools[0]]);
      const store = useNodePoolStore();

      const result = await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(result).toBe(true);
    });

    it("should return true when no node pools exist", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue([]);
      const store = useNodePoolStore();

      const result = await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(result).toBe(true);
    });

    it("should return false when multiple node pools exist", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      const result = await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(result).toBe(false);
    });

    it("should cache node pools and not call API on second request", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(1);
    });

    it("should fetch separately for different cluster IDs", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-2");

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(2);
      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledWith(["clusterId==cluster-1"]);
      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledWith(["clusterId==cluster-2"]);
    });
  });

  describe("removeNodePoolsByClusterId", () => {
    it("should clear cache and cause API to be called again", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");
      store.removeNodePoolsByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");

      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(2);
    });

    it("should only clear cache for specified cluster ID", async () => {
      vi.mocked(clusterService.getNodepoolsMinimal).mockResolvedValue(mockNodePools);
      const store = useNodePoolStore();

      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-2");
      store.removeNodePoolsByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-1");
      await store.isOnlyDefaultNodePoolByClusterId("cluster-2");

      // cluster-1: initial + after clear = 2, cluster-2: initial only = 1
      expect(clusterService.getNodepoolsMinimal).toHaveBeenCalledTimes(3);
    });
  });
});
