import { describe, it, expect } from "vitest";
import { ref, nextTick } from "vue";
import { useOrgUnitNodePools } from "./use-org-unit-node-pools.composable";
import type { Resources } from "@/swagger-models/org-unit-service-client";
import { DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";

describe("useOrgUnitNodePools", () => {
  const createMockResources = (): Resources[] => [
    { nodePool: { id: "pool-1", name: "Pool 1" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
    { nodePool: { id: "pool-2", name: "Pool 2" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
    { nodePool: { id: "pool-3", name: "Pool 3" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
  ];

  describe("allNodePools", () => {
    it("extracts node pools from resources", () => {
      const resources = createMockResources();
      const { allNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
      });

      expect(allNodePools.value).toEqual([
        { id: "pool-1", name: "Pool 1" },
        { id: "pool-2", name: "Pool 2" },
        { id: "pool-3", name: "Pool 3" },
      ]);
    });

    it("filters out pools without names", () => {
      const resources: Resources[] = [
        { nodePool: { id: "pool-1", name: "Pool 1" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
        { nodePool: { id: "pool-2", name: "" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
        { nodePool: { id: "pool-3" }, gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 } },
      ];

      const { allNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
      });

      expect(allNodePools.value).toEqual([{ id: "pool-1", name: "Pool 1" }]);
    });

    it("handles empty resources", () => {
      const { allNodePools } = useOrgUnitNodePools({
        resources: () => [],
        defaultNodePools: () => [],
      });

      expect(allNodePools.value).toEqual([]);
    });
  });

  describe("selectedNodePools", () => {
    it("maps defaultNodePools to NodePoolOption format", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1", "Pool 3"],
      });

      expect(selectedNodePools.value).toEqual([
        { id: "pool-1", name: "Pool 1" },
        { id: "pool-3", name: "Pool 3" },
      ]);
    });

    it("filters out pools not found in resources", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1", "NonExistent"],
      });

      expect(selectedNodePools.value).toEqual([{ id: "pool-1", name: "Pool 1" }]);
    });

    it("preserves order from defaultNodePools", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 3", "Pool 1", "Pool 2"],
      });

      expect(selectedNodePools.value).toEqual([
        { id: "pool-3", name: "Pool 3" },
        { id: "pool-1", name: "Pool 1" },
        { id: "pool-2", name: "Pool 2" },
      ]);
    });
  });

  describe("fallback logic", () => {
    it("uses defaultNodePools when available", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
        parentDefaultNodePools: () => ["Pool 2"],
      });

      expect(selectedNodePools.value).toEqual([{ id: "pool-1", name: "Pool 1" }]);
    });

    it("falls back to parentDefaultNodePools when defaultNodePools is empty", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => [],
        parentDefaultNodePools: () => ["Pool 2"],
      });

      expect(selectedNodePools.value).toEqual([{ id: "pool-2", name: "Pool 2" }]);
    });

    it("falls back to parentDefaultNodePools when defaultNodePools is undefined", () => {
      const resources = createMockResources();
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => undefined,
        parentDefaultNodePools: () => ["Pool 2"],
      });

      expect(selectedNodePools.value).toEqual([{ id: "pool-2", name: "Pool 2" }]);
    });

    it("falls back to DEFAULT_NODE_POOL_NAME when both are empty", () => {
      const resources: Resources[] = [
        {
          nodePool: { id: "default-id", name: DEFAULT_NODE_POOL_NAME },
          gpu: { deserved: 0, limit: -1, overQuotaWeight: 2 },
        },
      ];
      const { selectedNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => undefined,
        parentDefaultNodePools: () => undefined,
      });

      expect(selectedNodePools.value).toEqual([{ id: "default-id", name: "default" }]);
    });
  });

  describe("selectedNodePoolId and selection", () => {
    it("auto-selects first pool on initialization", () => {
      const resources = createMockResources();
      const { selectedNodePoolId } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 2", "Pool 1"],
      });

      expect(selectedNodePoolId.value).toBe("pool-2");
    });

    it("provides selectedNodePool object", () => {
      const resources = createMockResources();
      const { selectedNodePool } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1", "Pool 2"],
      });

      expect(selectedNodePool.value).toEqual({ id: "pool-1", name: "Pool 1" });
    });

    it("provides selectedNodePoolName", () => {
      const resources = createMockResources();
      const { selectedNodePoolName } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
      });

      expect(selectedNodePoolName.value).toBe("Pool 1");
    });

    it("allows manual selection change", async () => {
      const resources = createMockResources();
      const { selectedNodePoolId, selectedNodePoolName } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1", "Pool 2"],
      });

      selectedNodePoolId.value = "pool-2";
      await nextTick();

      expect(selectedNodePoolName.value).toBe("Pool 2");
    });
  });

  describe("editable mode", () => {
    it("allows updating node pools via updateNodePools", async () => {
      const resources = createMockResources();
      const { localDefaultNodePools, updateNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
        editable: true,
      });

      expect(localDefaultNodePools.value).toEqual(["Pool 1"]);

      updateNodePools(["Pool 2", "Pool 3"]);
      await nextTick();

      expect(localDefaultNodePools.value).toEqual(["Pool 2", "Pool 3"]);
    });

    it("tracks dirty state when pools are modified", async () => {
      const resources = createMockResources();
      const { isNodePoolListDirty, updateNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
        editable: true,
      });

      expect(isNodePoolListDirty.value).toBe(false);

      updateNodePools(["Pool 1", "Pool 2"]);
      await nextTick();

      expect(isNodePoolListDirty.value).toBe(true);
    });

    it("tracks dirty state when order changes", async () => {
      const resources = createMockResources();
      const { isNodePoolListDirty, updateNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1", "Pool 2"],
        editable: true,
      });

      expect(isNodePoolListDirty.value).toBe(false);

      updateNodePools(["Pool 2", "Pool 1"]);
      await nextTick();

      expect(isNodePoolListDirty.value).toBe(true);
    });

    it("resets to baseline via resetNodePoolsToBaseline", async () => {
      const resources = createMockResources();
      const { localDefaultNodePools, isNodePoolListDirty, updateNodePools, resetNodePoolsToBaseline } =
        useOrgUnitNodePools({
          resources: () => resources,
          defaultNodePools: () => ["Pool 1"],
          editable: true,
        });

      updateNodePools(["Pool 2", "Pool 3"]);
      await nextTick();
      expect(isNodePoolListDirty.value).toBe(true);

      resetNodePoolsToBaseline();
      await nextTick();

      expect(localDefaultNodePools.value).toEqual(["Pool 1"]);
      expect(isNodePoolListDirty.value).toBe(false);
    });

    it("selectedNodePools reflects local changes in editable mode", async () => {
      const resources = createMockResources();
      const { selectedNodePools, updateNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
        editable: true,
      });

      expect(selectedNodePools.value).toEqual([{ id: "pool-1", name: "Pool 1" }]);

      updateNodePools(["Pool 3", "Pool 2"]);
      await nextTick();

      expect(selectedNodePools.value).toEqual([
        { id: "pool-3", name: "Pool 3" },
        { id: "pool-2", name: "Pool 2" },
      ]);
    });
  });

  describe("read-only mode", () => {
    it("isNodePoolListDirty is always false", async () => {
      const resources = createMockResources();
      const { isNodePoolListDirty, localDefaultNodePools } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => ["Pool 1"],
        editable: false,
      });

      expect(isNodePoolListDirty.value).toBe(false);

      // Even if we directly modify local (not recommended), dirty should be false
      localDefaultNodePools.value = ["Pool 2"];
      await nextTick();

      expect(isNodePoolListDirty.value).toBe(false);
    });
  });

  describe("selection auto-correction", () => {
    it("selects first pool when current selection is removed", async () => {
      const defaultPoolsRef = ref<string[]>(["Pool 1", "Pool 2"]);
      const resources = createMockResources();

      const { selectedNodePoolId } = useOrgUnitNodePools({
        resources: () => resources,
        defaultNodePools: () => defaultPoolsRef.value,
        editable: false,
      });

      // Initial selection
      expect(selectedNodePoolId.value).toBe("pool-1");

      // Switch to Pool 2
      selectedNodePoolId.value = "pool-2";
      await nextTick();
      expect(selectedNodePoolId.value).toBe("pool-2");

      // Remove Pool 2 from defaults - should auto-select Pool 1
      defaultPoolsRef.value = ["Pool 1"];
      await nextTick();

      expect(selectedNodePoolId.value).toBe("pool-1");
    });
  });
});
