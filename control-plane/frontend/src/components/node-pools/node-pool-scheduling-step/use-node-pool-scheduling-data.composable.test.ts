/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, ref } from "vue";
import { flushPromises } from "@vue/test-utils";
const permissionMocks = vi.hoisted(() => ({ hasPermission: vi.fn(() => true) }));
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepool: vi.fn(), getNodepoolMetrics: vi.fn() },
}));
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getDepartment: vi.fn(),
    getDepartmentNodePoolResources: vi.fn(),
    countDepartmentNodePoolResources: vi.fn(),
    sumDepartmentNodePoolResources: vi.fn(),
    sumProjectNodePoolResources: vi.fn(),
  },
}));
vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({ hasPermission: permissionMocks.hasPermission }),
}));
vi.mock("@/stores/setting.store", () => ({ useSettingStore: () => ({ isLimitQuotaOverSubscriptionEnabled: false }) }));
vi.mock("@/services/control-plane/org-unit-settings.service/org-unit-settings.service", () => ({
  orgUnitSettingsService: { getLimitQuotaOverSubscription: vi.fn().mockResolvedValue({ value: false }) },
}));
vi.mock("@/services/authorization/permission.service/permission.service", () => ({
  permissionService: {
    getPermittedScopesForAction: vi.fn().mockResolvedValue({ read: { system: true }, update: { system: true } }),
  },
}));
vi.mock("@/stores/auth.store", () => ({ useAuthStore: () => ({ currentUser: { tenantId: "tenant" } }) }));
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import {
  loadChildQuotaSum,
  loadDepartmentPage,
  loadSchedulingTotals,
  loadSchedulingWeightTotals,
  telemetryResources,
  loadSchedulingResources,
} from "./node-pool-scheduling.service";
import { useNodePoolSchedulingData } from "./use-node-pool-scheduling-data.composable";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
beforeEach(() => {
  permissionMocks.hasPermission.mockReturnValue(true);
  vi.mocked(permissionService.getPermittedScopesForAction).mockResolvedValue({
    read: { system: true },
    update: { system: true },
    create: {},
    delete: {},
    sync: {},
  });
});
describe("scoped scheduling reads", () => {
  it("reuses the complete baseline for initial rows and rank data without duplicate requests", async () => {
    const rows = Array.from({ length: 58 }, (_, i) => ({
      id: String(i),
      orgUnitId: String(i),
      orgUnitName: `Dept ${i}`,
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank: "High",
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    }));
    vi.mocked(clusterService.getNodepool).mockResolvedValue({
      id: "p",
      clusterId: "c",
      name: "pool",
      phase: "Empty",
      schedulingCapabilities: { quota: false, ranks: false, adjustableWeights: true },
    } as never);
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation(
      async (_filters, _sort, _order, offset, limit) => ({
        nodePoolResources: rows.slice(offset ?? 0, (offset ?? 0) + (limit ?? 50)),
      }),
    );
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(58);
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 116, deservedSum: 0 });
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.baselineRows.value).toHaveLength(58);
    expect(data.pages.value).toHaveLength(50);
    expect(data.rankCounts.value.High).toBe(58);
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources)).toHaveBeenCalledTimes(1);
    expect(vi.mocked(orgUnitService.sumDepartmentNodePoolResources)).toHaveBeenCalledTimes(1);
    expect(vi.mocked(orgUnitService.countDepartmentNodePoolResources)).toHaveBeenCalledTimes(1);
    const first = data.pages.value.slice();
    await data.loadMore();
    expect(data.pages.value).toHaveLength(58);
    expect(data.pages.value.slice(0, 50)).toEqual(first);
    expect(data.hasMore.value).toBe(false);
  });
  it("treats an explicit empty department child list as zero project quota", async () => {
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue({ clusterId: "c", children: [] } as never);
    expect(await loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "empty")).toBe(0);
  });
  it.each([undefined, null, [{ id: "child", name: "Child" }]])(
    "keeps missing aggregates unavailable when the child list is not explicitly empty (%s)",
    async (children) => {
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue({ clusterId: "c", children } as never);
      await expect(loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "unknown")).rejects.toThrow(
        "Project quota totals are unavailable.",
      );
    },
  );
  it("checks the department cluster before accepting an empty child list", async () => {
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue({ clusterId: "other", children: [] } as never);
    await expect(loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "other")).rejects.toThrow(
      "Department does not belong to this cluster.",
    );
  });
  it("reads the selected pool's finite project quota and keeps unlimited aggregates unavailable", async () => {
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue({
      clusterId: "c",
      children: [{ id: "child", name: "Child" }],
      projectsAggregatedResources: [
        { nodePool: { id: "other" }, gpu: { deserved: 100 } },
        { nodePool: { id: "p" }, gpu: { deserved: 7 } },
      ],
    } as never);
    expect(await loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "known")).toBe(7);
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue({
      clusterId: "c",
      children: [{ id: "child", name: "Child" }],
      projectsAggregatedResources: [{ nodePool: { id: "p" }, gpu: { deserved: -1 } }],
    } as never);
    await expect(loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "unlimited")).rejects.toThrow(
      "Project quota totals are unavailable.",
    );
  });
  it("keeps returned finite or invalid aggregates authoritative even with an empty child list", async () => {
    vi.mocked(orgUnitService.getDepartment).mockResolvedValue({
      clusterId: "c",
      children: [],
      projectsAggregatedResources: [{ nodePool: { id: "p" }, gpu: { deserved: 5 } }],
    } as never);
    expect(await loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "known")).toBe(5);
    for (const deserved of [-1, NaN, Infinity]) {
      vi.mocked(orgUnitService.getDepartment).mockResolvedValue({
        clusterId: "c",
        children: [],
        projectsAggregatedResources: [{ nodePool: { id: "p" }, gpu: { deserved } }],
      } as never);
      await expect(loadChildQuotaSum({ poolId: "p", clusterId: "c" }, "invalid")).rejects.toThrow(
        "Project quota totals are unavailable.",
      );
    }
  });
  it("scopes pages and unfiltered aggregates", async () => {
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 10, deservedSum: 3 });
    await loadDepartmentPage({ poolId: "p", clusterId: "c", search: "abc", rank: "High", offset: 50, limit: 50 });
    const args = vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.at(-1) ?? [];
    expect(args[0]).toEqual(expect.arrayContaining(["nodePoolId==p", "clusterId==c", "rank==High"]));
    expect(args.slice(3)).toEqual([50, 50, "abc"]);
    expect(await loadSchedulingTotals({ poolId: "p", clusterId: "c" })).toEqual({ weightsSum: 10, deservedSum: 3 });
    expect(vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mock.calls.at(-1)).toEqual(["p", ["clusterId==c"]]);
  });
  it("uses zero capacity for confirmed Empty pools without metrics", async () => {
    vi.mocked(clusterService.getNodepoolMetrics).mockClear();
    expect(await loadSchedulingResources({ clusterId: "c", name: "empty", phase: "Empty" })).toEqual({
      gpus: 0,
      nodes: 0,
    });
    expect(clusterService.getNodepoolMetrics).not.toHaveBeenCalled();
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    expect((await loadSchedulingResources({ clusterId: "c", name: "ready", phase: "Ready" })).gpus).toBeUndefined();
  });
  it("uses the saved pool inventory before telemetry is available", async () => {
    vi.mocked(clusterService.getNodepoolMetrics).mockClear();
    expect(
      await loadSchedulingResources({
        clusterId: "c",
        name: "ready",
        phase: "Ready",
        status: { nodes: ["a", "b"] },
        gpuTypes: [{ type: "P100", totalDevices: 32, gpuMemoryBytes: 100, maxDevicesPerNode: 16 }],
      }),
    ).toEqual({ gpus: 32, nodes: 2, gpuTypes: ["P100"], gpuMemoryBytes: 3200 });
    expect(clusterService.getNodepoolMetrics).not.toHaveBeenCalled();
  });
  it("retries capacity without replacing the scheduling snapshot and stops after disposal", async () => {
    vi.useFakeTimers();
    const scope = effectScope();
    try {
      vi.mocked(clusterService.getNodepool).mockResolvedValue({
        id: "p",
        clusterId: "c",
        name: "pool",
        phase: "Creating",
      } as never);
      vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
      vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [] });
      vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
      vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 0, deservedSum: 0 });
      const data = scope.run(() => useNodePoolSchedulingData(ref("p")));
      if (!data) throw new Error("Scheduling test scope is inactive.");
      await flushPromises();
      const snapshot = data.snapshot.value;
      vi.mocked(clusterService.getNodepool).mockResolvedValue({
        id: "p",
        clusterId: "c",
        name: "pool",
        phase: "Ready",
        gpuTypes: [{ type: "P100", totalDevices: 64, gpuMemoryBytes: 100 }],
        status: { nodes: ["a", "b", "c", "d"] },
      } as never);
      await vi.advanceTimersByTimeAsync(5000);
      await flushPromises();
      expect(data.resources.value?.gpus).toBe(64);
      expect(data.snapshot.value).toBe(snapshot);
      scope.stop();
      const calls = vi.mocked(clusterService.getNodepool).mock.calls.length;
      await vi.advanceTimersByTimeAsync(10000);
      expect(clusterService.getNodepool).toHaveBeenCalledTimes(calls);
    } finally {
      scope.stop();
      vi.useRealTimers();
    }
  });
  it("preserves measured zero and missing telemetry", () => {
    expect(telemetryResources({ measurements: [] }).gpus).toBeUndefined();
    expect(
      telemetryResources({
        measurements: [{ type: "TOTAL_GPU", values: [{ value: "0", timestamp: "2026-11-08T00:00:00Z" }] }],
      }).gpus,
    ).toBe(0);
  });
  it("keeps totals when search changes during initial aggregate loading", async () => {
    let release!: (value: { weightsSum: number; deservedSum: number }) => void;
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [] });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockImplementation(
      () =>
        new Promise((resolve) => {
          release = resolve;
        }),
    );
    // All rank reads share a deferred promise so the full load remains pending.
    const waiting = new Promise<{ weightsSum: number; deservedSum: number }>((resolve) => {
      release = resolve;
    });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockReturnValue(waiting);
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    data.search.value = "changed";
    await flushPromises();
    release({ weightsSum: 10, deservedSum: 3 });
    await flushPromises();
    expect(data.aggregates.value.all?.weightsSum).toBe(10);
  });
  it("rejects stale pool results and surfaces errors", async () => {
    let release!: (value: never) => void;
    vi.mocked(clusterService.getNodepool).mockImplementation((id) =>
      id === "a"
        ? new Promise((resolve) => {
            release = resolve;
          })
        : Promise.reject(new Error("Denied")),
    );
    const id = ref("a");
    const data = useNodePoolSchedulingData(id);
    id.value = "b";
    await flushPromises();
    release({ id: "a", clusterId: "c" } as never);
    await flushPromises();
    expect(data.snapshot.value).toBeUndefined();
    expect(data.loadError.value).toContain("Denied");
  });
  it("loads complete hidden-page raw and automatic denominators with pool scope", async () => {
    const row = {
      id: "r",
      orgUnitId: "a",
      orgUnitName: "Alpha",
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank: "Low",
      resources: { gpu: { overQuotaWeight: 2, deserved: 5 } },
    };
    vi.mocked(orgUnitService.getDepartmentNodePoolResources)
      .mockResolvedValueOnce({ nodePoolResources: [row], next: 1 })
      .mockResolvedValueOnce({
        nodePoolResources: [
          { ...row, id: "s", orgUnitId: "b", resources: { gpu: { overQuotaWeight: 8, deserved: 3 } } },
        ],
      });
    expect((await loadSchedulingWeightTotals({ poolId: "p", clusterId: "c" })).all).toEqual({
      rawWeightsSum: 10,
      automaticWeightsSum: 8,
    });
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.at(-1)?.[3]).toBe(1);
  });
  it("keeps filtered pagination independent of unrelated rank reads", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({
      id: "p",
      clusterId: "c",
      name: "pool",
      schedulingCapabilities: { adjustableWeights: true },
    } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 120, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockImplementation(async (filters) =>
      filters?.includes("rank==High") ? 60 : filters?.includes("rank==MediumLow") ? 10 : 70,
    );
    const make = (id: number, rank: string) => ({
      id: String(id),
      orgUnitId: String(id),
      orgUnitName: `match ${id}`,
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank,
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    });
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation(
      async (filters, _sort, _order, offset, limit) => {
        if (limit === 100) return { nodePoolResources: [] };
        if (filters?.includes("rank==MediumLow"))
          return { nodePoolResources: Array.from({ length: 10 }, (_, i) => make(100 + i, "MediumLow")) };
        return {
          nodePoolResources: Array.from({ length: offset ? 10 : 50 }, (_, i) => make((offset ?? 0) + i, "High")),
        };
      },
    );
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    data.search.value = "match";
    data.rank.value = "High";
    await flushPromises();
    await data.loadRank("MediumLow");
    expect(data.pages.value.every((row) => row.rank === "High")).toBe(true);
    expect(data.hasMore.value).toBe(true);
    await data.loadMore();
    expect(data.pages.value).toHaveLength(60);
    expect(data.hasMore.value).toBe(false);
  });
  it("ignores delayed rank rows after query replacement and permits fresh rank reads", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 0, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
    let resolveOld!: (value: { nodePoolResources: never[] }) => void;
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation(
      async (filters, _sort, _order, _offset, limit, search) => {
        if (limit === 50 && filters?.includes("rank==MediumLow") && !search)
          return new Promise((resolve) => {
            resolveOld = resolve;
          });
        return { nodePoolResources: [] };
      },
    );
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    const pending = data.loadRank("MediumLow");
    await flushPromises();
    data.search.value = "new query";
    await flushPromises();
    const before = vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.length;
    await data.loadRank("MediumLow");
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.length).toBe(before + 1);
    resolveOld({ nodePoolResources: [{ orgUnitId: "stale", rank: "MediumLow" } as never] });
    await pending;
    expect(data.pages.value).toEqual([]);
  });
  it("waits for a replacement first page before appending an expanded selected rank", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 300, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockImplementation(async (filters) =>
      filters?.includes("rank==High") ? 100 : 150,
    );
    const make = (i: number, rank: string) => ({
      id: `${rank}${i}`,
      orgUnitId: `${rank}${i}`,
      orgUnitName: `${rank} ${i}`,
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank,
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    });
    let resolveFirst!: (value: { nodePoolResources: ReturnType<typeof make>[] }) => void;
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation(
      async (filters, _sort, _order, offset, limit) => {
        if (limit === 100) return { nodePoolResources: [] };
        if (filters?.includes("rank==High") && offset === 0)
          return new Promise((resolve) => {
            resolveFirst = resolve;
          });
        const rank = filters?.includes("rank==High") ? "High" : "Low";
        return { nodePoolResources: Array.from({ length: 50 }, (_, i) => make((offset ?? 0) + i, rank)) };
      },
    );
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.pages.value).toHaveLength(50);
    data.rank.value = "High";
    await flushPromises();
    const calls = vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.length;
    await data.loadRank("High");
    await data.loadMore();
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.length).toBe(calls);
    expect(data.hasMore.value).toBe(false);
    resolveFirst({ nodePoolResources: Array.from({ length: 50 }, (_, i) => make(i, "High")) });
    await flushPromises();
    expect(data.pages.value.map((row) => row.orgUnitId)).toEqual(Array.from({ length: 50 }, (_, i) => `High${i}`));
    expect(data.hasMore.value).toBe(true);
    await data.loadRank("High");
    expect(data.pages.value.map((row) => row.orgUnitId)).toEqual(Array.from({ length: 100 }, (_, i) => `High${i}`));
    expect(data.hasMore.value).toBe(false);
  });

  it("clears all prior-pool department state while a replacement pool is pending", async () => {
    const oldRow = {
      id: "resource-a",
      orgUnitId: "department-a",
      orgUnitName: "Alpha",
      clusterId: "c",
      nodePoolId: "a",
      nodePoolName: "Pool A",
      rank: "High",
      resources: { gpu: { overQuotaWeight: 2, deserved: 3 } },
    };
    vi.mocked(clusterService.getNodepool).mockImplementation((id) =>
      id === "a" ? Promise.resolve({ id: "a", clusterId: "c", name: "Pool A" } as never) : new Promise(() => {}),
    );
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 2, deservedSum: 3 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(1);
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [oldRow] });
    const id = ref("a");
    const data = useNodePoolSchedulingData(id);
    await flushPromises();
    expect(data.baselineRows.value.map((row) => row.orgUnitId)).toEqual(["department-a"]);
    expect(data.count.value).toBe(1);
    id.value = "b";
    await flushPromises();
    expect(data.snapshot.value).toBeUndefined();
    expect(data.loading.value).toBe(true);
    expect(data.pages.value).toEqual([]);
    expect(data.baselineRows.value).toEqual([]);
    expect(data.count.value).toBe(0);
  });

  it("loads pool metadata and capacity without requesting forbidden department data", async () => {
    permissionMocks.hasPermission.mockReturnValue(false);
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({
      measurements: [{ type: "TOTAL_GPU", values: [{ value: "8", timestamp: "2026-11-14T00:00:00Z" }] }],
    });
    const reads = vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.length;
    const sums = vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mock.calls.length;
    const counts = vi.mocked(orgUnitService.countDepartmentNodePoolResources).mock.calls.length;
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.snapshot.value?.id).toBe("p");
    expect(data.resources.value?.gpus).toBe(8);
    expect(data.permissions.value).toEqual({ view: false, edit: false });
    expect(data.pages.value).toEqual([]);
    expect(data.baselineRows.value).toEqual([]);
    expect(data.loading.value).toBe(false);
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls).toHaveLength(reads);
    expect(vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mock.calls).toHaveLength(sums);
    expect(vi.mocked(orgUnitService.countDepartmentNodePoolResources).mock.calls).toHaveLength(counts);
  });
  it("preserves restricted read visibility and individual department edit scope", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 2, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(1);
    const row = {
      id: "resource",
      orgUnitId: "visible",
      orgUnitName: "Visible",
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank: "High",
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    };
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [row] });
    vi.mocked(permissionService.getPermittedScopesForAction).mockImplementation(async (resource) => ({
      read: resource === ResourceType.Department ? { departments: ["visible"] } : {},
      update: resource === ResourceType.Department ? { departments: ["visible"] } : {},
      create: {},
      delete: {},
      sync: {},
    }));
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockClear();
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.pages.value.map((value) => value.orgUnitId)).toEqual(["visible"]);
    expect(data.totalsComplete.value).toBe(false);
    expect(
      vi
        .mocked(orgUnitService.sumDepartmentNodePoolResources)
        .mock.calls.filter(([, filters]) => filters?.some((value) => value.startsWith("rank=="))),
    ).toHaveLength(9);
    expect(data.canEditDepartment("visible")).toBe(true);
    expect(data.canEditDepartment("hidden")).toBe(false);
    expect(data.canResetResources.value).toBe(false);
  });
  it.each([
    { department: { system: true }, project: { clusters: ["c"] }, canReset: true },
    { department: { tenants: ["tenant"] }, project: { system: true }, canReset: true },
    { department: { clusters: ["c"] }, project: { departments: ["visible"] }, canReset: false },
    { department: { clusters: ["other"] }, project: { system: true }, canReset: false },
  ])("requires both scopes to cover the pool cluster: $canReset", async ({ department, project, canReset }) => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 0, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [] });
    vi.mocked(permissionService.getPermittedScopesForAction).mockImplementation(async (resource) => ({
      [Action.Read]: { clusters: ["c"] },
      [Action.Update]: resource === ResourceType.Department ? department : project,
      create: {},
      delete: {},
      sync: {},
    }));
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.totalsComplete.value).toBe(true);
    expect(data.canResetResources.value).toBe(canReset);
  });
  it("retains the first page on append failure and retries the same offset without duplicate rows", async () => {
    const make = (id: string) => ({
      id,
      orgUnitId: id,
      orgUnitName: id,
      clusterId: "c",
      nodePoolId: "p",
      nodePoolName: "pool",
      rank: "High",
      resources: { gpu: { overQuotaWeight: 2, deserved: 0 } },
    });
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 6, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(4);
    let appendAttempts = 0;
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation(
      async (_filters, _sort, _order, offset, limit) => {
        if (limit === 100) return { nodePoolResources: [make("a"), make("b"), make("c")] };
        if (!offset) return { nodePoolResources: [make("a"), make("b")] };
        if (appendAttempts++ === 0) throw new Error("Temporary page failure");
        return { nodePoolResources: [make("b"), make("c")] };
      },
    );
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.hasMore.value).toBe(true);
    await data.loadMore();
    expect(data.pages.value.map((row) => row.orgUnitId)).toEqual(["a", "b"]);
    expect(data.loadError.value).toContain("Temporary page failure");
    expect(data.hasMore.value).toBe(true);
    expect(data.loading.value).toBe(false);
    await data.loadMore();
    expect(data.pages.value.map((row) => row.orgUnitId)).toEqual(["a", "b", "c"]);
    expect(data.loadError.value).toBe("");
    expect(data.hasMore.value).toBe(false);
    const offsets = vi
      .mocked(orgUnitService.getDepartmentNodePoolResources)
      .mock.calls.filter((call) => call[4] === 50)
      .slice(-3)
      .map((call) => call[3]);
    expect(offsets).toEqual([0, 2, 2]);
  });
  it("ignores old department and telemetry responses after a new pool has loaded", async () => {
    vi.mocked(clusterService.getNodepool).mockImplementation(async (id) => ({ id, clusterId: "c", name: id }) as never);
    let resolveOldRows!: (value: { nodePoolResources: never[] }) => void;
    let resolveOldMetrics!: (value: never) => void;
    vi.mocked(clusterService.getNodepoolMetrics).mockImplementation((_cluster, pool) =>
      pool === "a"
        ? new Promise((resolve) => {
            resolveOldMetrics = resolve;
          })
        : Promise.resolve({ measurements: [] }),
    );
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 0, deservedSum: 0 });
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockImplementation((filters) =>
      filters?.includes("nodePoolId==a")
        ? new Promise((resolve) => {
            resolveOldRows = resolve;
          })
        : Promise.resolve({ nodePoolResources: [] }),
    );
    const id = ref("a");
    const data = useNodePoolSchedulingData(id);
    await flushPromises();
    id.value = "b";
    await flushPromises();
    expect(data.snapshot.value?.id).toBe("b");
    expect(data.loading.value).toBe(false);
    resolveOldRows({
      nodePoolResources: [
        {
          id: "old-resource",
          orgUnitId: "obsolete",
          orgUnitName: "Obsolete",
          clusterId: "c",
          nodePoolId: "a",
          nodePoolName: "a",
          rank: "High",
          resources: { gpu: { overQuotaWeight: 9, deserved: 0 } },
        } as never,
      ],
    });
    resolveOldMetrics({
      measurements: [{ type: "TOTAL_GPU", values: [{ value: "999", timestamp: "2026-11-14T00:00:00Z" }] }],
    } as never);
    await flushPromises();
    expect(data.snapshot.value?.id).toBe("b");
    expect(data.pages.value).toEqual([]);
    expect(data.baselineRows.value).toEqual([]);
    expect(data.resources.value?.gpus).toBeUndefined();
    expect(data.loadError.value).toBe("");
  });
});
