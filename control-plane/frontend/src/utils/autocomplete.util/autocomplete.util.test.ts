/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { useClusterStore } from "@/stores/cluster.store";
import { autocompleteUtil } from "./autocomplete.util";
import { EFilterOperator } from "@/models/filter.model";
import { NodeFilterSortFields } from "@/swagger-models/cluster-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: vi.fn(),
}));

vi.mock("@/utils/filter.util/filter.util", () => ({
  filterUtil: {
    getEqualsFilterString: (field: string, value: string) => `${field}==${value}`,
    getContainsFilterString: (field: string, value: string) => `${field}${EFilterOperator.Contains}${value}`,
  },
}));

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getTenantNodes: vi.fn(),
  },
}));

const STORE_CLUSTER_ID = "store-cluster-abc";

describe("autocomplete.util", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useClusterStore).mockReturnValue({
      selectedClusterId: STORE_CLUSTER_ID,
    } as unknown as ReturnType<typeof useClusterStore>);
  });

  describe("createEntityAutocomplete", () => {
    const mockFetchFn = vi.fn();
    const extractName = (entity: { name: string }) => entity.name;

    beforeEach(() => {
      mockFetchFn.mockResolvedValue([{ name: "alpha" }, { name: "bravo" }, { name: "charlie" }]);
    });

    it("should add cluster filter from store when clusterScoped and clusterUuid is undefined", async () => {
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName);

      await fn("test");

      expect(mockFetchFn).toHaveBeenCalledWith(
        [`clusterId==${STORE_CLUSTER_ID}`, `name${EFilterOperator.Contains}test`],
        100,
      );
    });

    it("should add cluster filter with provided clusterUuid", async () => {
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName);

      await fn("test", "custom-cluster-123");

      expect(mockFetchFn).toHaveBeenCalledWith(
        ["clusterId==custom-cluster-123", `name${EFilterOperator.Contains}test`],
        100,
      );
    });

    it("should skip cluster filter when clusterUuid is empty string (All clusters)", async () => {
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName);

      await fn("test", "");

      expect(mockFetchFn).toHaveBeenCalledWith([`name${EFilterOperator.Contains}test`], 100);
    });

    it("should skip cluster filter when clusterScoped is false", async () => {
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName, false);

      await fn("test");

      expect(mockFetchFn).toHaveBeenCalledWith([`name${EFilterOperator.Contains}test`], 100);
    });

    it("should not add search filter when searchQuery is empty", async () => {
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName);

      await fn("");

      expect(mockFetchFn).toHaveBeenCalledWith([`clusterId==${STORE_CLUSTER_ID}`], 100);
    });

    it("should extract and deduplicate values", async () => {
      mockFetchFn.mockResolvedValue([{ name: "alpha" }, { name: "bravo" }, { name: "alpha" }]);
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", extractName);

      const result = await fn("");

      expect(result).toEqual(["alpha", "bravo"]);
    });

    it("should skip undefined values from extractFn", async () => {
      mockFetchFn.mockResolvedValue([{ name: "alpha" }, { name: undefined }, { name: "charlie" }]);
      const fn = autocompleteUtil.createEntityAutocomplete(
        mockFetchFn,
        "name",
        (e: { name: string | undefined }) => e.name,
      );

      const result = await fn("", "");

      expect(result).toEqual(["alpha", "charlie"]);
    });
  });

  describe("createNameAutocomplete", () => {
    const mockFetchFn = vi.fn();

    beforeEach(() => {
      mockFetchFn.mockResolvedValue([{ name: "project-a" }, { name: "project-b" }]);
    });

    it("should use default name extractor", async () => {
      const fn = autocompleteUtil.createNameAutocomplete(mockFetchFn, "name");

      const result = await fn("proj");

      expect(result).toEqual(["project-a", "project-b"]);
    });

    it("should respect clusterUuid context", async () => {
      const fn = autocompleteUtil.createNameAutocomplete(mockFetchFn, "name");

      await fn("proj", "specific-cluster");

      expect(mockFetchFn).toHaveBeenCalledWith(
        ["clusterId==specific-cluster", `name${EFilterOperator.Contains}proj`],
        100,
      );
    });

    it("should skip cluster filter for All clusters", async () => {
      const fn = autocompleteUtil.createNameAutocomplete(mockFetchFn, "name");

      await fn("proj", "");

      expect(mockFetchFn).toHaveBeenCalledWith([`name${EFilterOperator.Contains}proj`], 100);
    });
  });

  describe("createClientSideAutocomplete", () => {
    const mockFetchFn = vi.fn();
    const extractName = (entity: { label: string }) => entity.label;

    beforeEach(() => {
      mockFetchFn.mockResolvedValue([
        { label: "TensorBoard" },
        { label: "Jupyter Lab" },
        { label: "PyTorch" },
        { label: "tensorflow" },
      ]);
    });

    it("should filter results client-side (case-insensitive)", async () => {
      const fn = autocompleteUtil.createClientSideAutocomplete(mockFetchFn, extractName);

      const result = await fn("tensor");

      expect(result).toEqual(["TensorBoard", "tensorflow"]);
    });

    it("should return all values when searchQuery is empty", async () => {
      const fn = autocompleteUtil.createClientSideAutocomplete(mockFetchFn, extractName);

      const result = await fn("");

      expect(result).toEqual(["TensorBoard", "Jupyter Lab", "PyTorch", "tensorflow"]);
    });

    it("should pass clusterUuid to fetchFn", async () => {
      const fn = autocompleteUtil.createClientSideAutocomplete(mockFetchFn, extractName);

      await fn("tensor", "my-cluster");

      expect(mockFetchFn).toHaveBeenCalledWith("my-cluster");
    });

    it("should deduplicate results", async () => {
      mockFetchFn.mockResolvedValue([{ label: "Jupyter" }, { label: "Jupyter" }, { label: "VSCode" }]);
      const fn = autocompleteUtil.createClientSideAutocomplete(mockFetchFn, extractName);

      const result = await fn("");

      expect(result).toEqual(["Jupyter", "VSCode"]);
    });

    it("should skip undefined values", async () => {
      mockFetchFn.mockResolvedValue([{ label: "A" }, { label: undefined }, { label: "B" }]);
      const fn = autocompleteUtil.createClientSideAutocomplete(
        mockFetchFn,
        (e: { label: string | undefined }) => e.label,
      );

      const result = await fn("");

      expect(result).toEqual(["A", "B"]);
    });

    describe("entityFilterFn", () => {
      it("should exclude entities that fail the filter", async () => {
        mockFetchFn.mockResolvedValue([
          { label: "cluster-A-item", cluster: "cluster-A" },
          { label: "cluster-B-item", cluster: "cluster-B" },
          { label: "global-item", cluster: undefined },
        ]);
        const fn = autocompleteUtil.createClientSideAutocomplete(
          mockFetchFn,
          (e: { label: string }) => e.label,
          (e: { label: string; cluster?: string }, clusterUuid?: string) => {
            if (!clusterUuid) return true;
            return !e.cluster || e.cluster === clusterUuid;
          },
        );

        const result = await fn("", "cluster-A");

        expect(result).toEqual(["cluster-A-item", "global-item"]);
      });

      it("should include all entities when resolved clusterUuid is undefined (All clusters)", async () => {
        mockFetchFn.mockResolvedValue([
          { label: "item-A", cluster: "cluster-A" },
          { label: "item-B", cluster: "cluster-B" },
        ]);
        const fn = autocompleteUtil.createClientSideAutocomplete(
          mockFetchFn,
          (e: { label: string }) => e.label,
          (e: { label: string; cluster?: string }, clusterUuid?: string) => {
            if (!clusterUuid) return true;
            return e.cluster === clusterUuid;
          },
        );

        const result = await fn("", "");

        expect(result).toEqual(["item-A", "item-B"]);
      });

      it("should apply both entityFilterFn and text search", async () => {
        mockFetchFn.mockResolvedValue([
          { label: "nfs-share", cluster: "cluster-A" },
          { label: "nfs-backup", cluster: "cluster-B" },
          { label: "git-repo", cluster: "cluster-A" },
        ]);
        const fn = autocompleteUtil.createClientSideAutocomplete(
          mockFetchFn,
          (e: { label: string }) => e.label,
          (e: { label: string; cluster: string }, clusterUuid?: string) => {
            if (!clusterUuid) return true;
            return e.cluster === clusterUuid;
          },
        );

        const result = await fn("nfs", "cluster-A");

        expect(result).toEqual(["nfs-share"]);
      });
    });
  });

  describe("cluster resolution", () => {
    it("should fall back to store when clusterUuid is undefined", async () => {
      const mockFetchFn = vi.fn().mockResolvedValue([{ name: "item" }]);
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", (e: { name: string }) => e.name);

      await fn("q");

      const filterBy = mockFetchFn.mock.calls[0][0];
      expect(filterBy[0]).toBe(`clusterId==${STORE_CLUSTER_ID}`);
    });

    it("should use provided clusterUuid over store", async () => {
      const mockFetchFn = vi.fn().mockResolvedValue([{ name: "item" }]);
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", (e: { name: string }) => e.name);

      await fn("q", "explicit-cluster");

      const filterBy = mockFetchFn.mock.calls[0][0];
      expect(filterBy[0]).toBe("clusterId==explicit-cluster");
    });

    it("should omit cluster filter when clusterUuid is empty (All clusters)", async () => {
      const mockFetchFn = vi.fn().mockResolvedValue([{ name: "item" }]);
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", (e: { name: string }) => e.name);

      await fn("q", "");

      const filterBy = mockFetchFn.mock.calls[0][0];
      expect(filterBy).not.toContainEqual(expect.stringContaining("clusterId"));
    });
  });

  describe("nodes", () => {
    const mockGetTenantNodes = vi.mocked(clusterService.getTenantNodes);

    beforeEach(() => {
      mockGetTenantNodes.mockResolvedValue([
        { name: "node-1", nodePool: "pool-a", gpuInfo: { gpuType: "A100" } },
        { name: "node-2", nodePool: "pool-b", gpuInfo: { gpuType: "V100" } },
      ] as ReturnType<typeof clusterService.getTenantNodes> extends Promise<infer T> ? T : never);
    });

    it("should use clusterUuid (not clusterId) for cluster filter", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name);

      await fn("node");

      const filterBy = mockGetTenantNodes.mock.calls[0][0];
      expect(filterBy).toContainEqual(`clusterUuid==${STORE_CLUSTER_ID}`);
      expect(filterBy).not.toContainEqual(expect.stringContaining("clusterId=="));
    });

    it("should use provided clusterUuid over store", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name);

      await fn("node", "explicit-cluster-id");

      const filterBy = mockGetTenantNodes.mock.calls[0][0];
      expect(filterBy).toContainEqual("clusterUuid==explicit-cluster-id");
    });

    it("should skip cluster filter when clusterUuid is empty (All clusters)", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name);

      await fn("node", "");

      const filterBy = mockGetTenantNodes.mock.calls[0][0];
      expect(filterBy).not.toContainEqual(expect.stringContaining("clusterUuid=="));
    });

    it("should add contains filter for search query", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name);

      await fn("node", "");

      const filterBy = mockGetTenantNodes.mock.calls[0][0];
      expect(filterBy).toContainEqual(`name${EFilterOperator.Contains}node`);
    });

    it("should not add search filter when query is empty", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name);

      await fn("");

      const filterBy = mockGetTenantNodes.mock.calls[0][0];
      expect(filterBy).not.toContainEqual(expect.stringContaining(EFilterOperator.Contains));
    });

    it("should extract and deduplicate values", async () => {
      mockGetTenantNodes.mockResolvedValue([
        { name: "node-1", nodePool: "pool-a" },
        { name: "node-2", nodePool: "pool-a" },
        { name: "node-3", nodePool: "pool-b" },
      ] as ReturnType<typeof clusterService.getTenantNodes> extends Promise<infer T> ? T : never);

      const fn = autocompleteUtil.nodes(NodeFilterSortFields.NodePool, (n) => n.nodePool);

      const result = await fn("", "");

      expect(result).toEqual(["pool-a", "pool-b"]);
    });

    it("should pass fieldName as sortBy to API", async () => {
      const fn = autocompleteUtil.nodes(NodeFilterSortFields.GpuType, (n) => n.gpuInfo?.gpuType);

      await fn("A1");

      expect(mockGetTenantNodes).toHaveBeenCalledWith(
        expect.anything(),
        NodeFilterSortFields.GpuType,
        undefined,
        0,
        100,
      );
    });
  });

  describe("AUTOCOMPLETE_MAX_RESULTS", () => {
    it("should be 100", () => {
      expect(autocompleteUtil.AUTOCOMPLETE_MAX_RESULTS).toBe(100);
    });

    it("should cap results at max", async () => {
      const items = Array.from({ length: 150 }, (_, i) => ({ name: `item-${i}` }));
      const mockFetchFn = vi.fn().mockResolvedValue(items);
      const fn = autocompleteUtil.createEntityAutocomplete(mockFetchFn, "name", (e: { name: string }) => e.name);

      const result = await fn("", "");

      expect(result.length).toBe(100);
    });
  });
});
