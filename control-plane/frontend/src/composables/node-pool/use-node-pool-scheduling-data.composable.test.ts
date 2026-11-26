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
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", async (original) => {
  const actual = await original<{ orgUnitService: typeof orgUnitService }>();
  Object.assign(actual.orgUnitService, {
    getDepartment: vi.fn(),
    getDepartmentNodePoolResources: vi.fn(),
    countDepartmentNodePoolResources: vi.fn(),
    countProjectsNodePoolResources: vi.fn(),
    sumDepartmentNodePoolResources: vi.fn(),
    sumProjectNodePoolResources: vi.fn(),
  });
  return actual;
});
vi.mock("@/stores/permissions.store/index", () => ({
  usePermissionStore: () => ({ hasPermission: permissionMocks.hasPermission }),
}));
vi.mock("@/stores/setting.store/index", () => ({
  useSettingStore: () => ({ isLimitQuotaOverSubscriptionEnabled: false }),
}));
vi.mock("@/services/control-plane/org-unit-settings.service/org-unit-settings.service", () => ({
  orgUnitSettingsService: { getLimitQuotaOverSubscription: vi.fn().mockResolvedValue({ value: false }) },
}));
vi.mock("@/services/authorization/permission.service/permission.service", () => ({
  permissionService: {
    getPermittedScopesForAction: vi.fn().mockResolvedValue({ read: { system: true }, update: { system: true } }),
  },
}));
vi.mock("@/stores/auth.store/index", () => ({ useAuthStore: () => ({ currentUser: { tenantId: "tenant" } }) }));
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import {
  loadChildQuotaSum,
  loadNodePoolGpuLimits,
  loadSchedulingTotals,
  loadSchedulingDepartmentPage,
  loadSchedulingResources,
} from "@/composables/node-pool/node-pool-scheduling-data.util";
import { DEPARTMENT_PAGE_SIZE } from "@/models/node-pool-scheduling.model";
import { useNodePoolSchedulingData } from "@/composables/node-pool/use-node-pool-scheduling-data.composable";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client/index";
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
  it("loads one page initially, keeps full rank counts and shares subsequent scroll and complete reads", async () => {
    const rows = Array.from({ length: 158 }, (_, i) => ({
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
        next: (offset ?? 0) + DEPARTMENT_PAGE_SIZE < rows.length ? (offset ?? 0) + DEPARTMENT_PAGE_SIZE : undefined,
      }),
    );
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockImplementation(async (filters) =>
      filters?.includes("rank==High") ? rows.length : 0,
    );
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 316, deservedSum: 0 });
    const data = useNodePoolSchedulingData(ref("p"));
    await flushPromises();
    expect(data.baselineRows.value).toHaveLength(DEPARTMENT_PAGE_SIZE);
    expect(data.rankCounts.value.High).toBe(158);
    expect(data.aggregates.value.High.rawWeightsSum).toBe(316);
    expect(orgUnitService.getDepartmentNodePoolResources).toHaveBeenCalledTimes(1);
    await Promise.all([data.departmentPages.value.loadNext(), data.departmentPages.value.loadNext()]);
    expect(data.baselineRows.value).toHaveLength(2 * DEPARTMENT_PAGE_SIZE);
    await Promise.all([data.departmentPages.value.loadAll(), data.departmentPages.value.loadNext()]);
    expect(data.baselineRows.value).toEqual(rows);
    expect(data.departmentPages.value.nextOffset).toBeUndefined();
    expect(orgUnitService.getDepartmentNodePoolResources).toHaveBeenCalledTimes(
      Math.ceil(rows.length / DEPARTMENT_PAGE_SIZE),
    );
    expect(data.aggregates.value.High.rawWeightsSum).toBe(316);
    const visibleRows = data.baselineRows.value;
    const refresh = data.refresh();
    expect(data.baselineRows.value).toBe(visibleRows);
    await refresh;
    expect(data.baselineRows.value).toEqual(rows);
    expect(data.departmentPages.value.nextOffset).toBeUndefined();
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
  it("uses zero capacity for an explicitly empty node list without fetching metrics", async () => {
    vi.mocked(clusterService.getNodepoolMetrics).mockClear();
    expect(
      await loadSchedulingResources({ clusterId: "c", name: "no-nodes", phase: "Ready", status: { nodes: [] } }),
    ).toEqual({ gpus: 0, nodes: 0 });
    expect(clusterService.getNodepoolMetrics).not.toHaveBeenCalled();
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
  it("preserves measured zero and missing telemetry", async () => {
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    expect((await loadSchedulingResources({ clusterId: "c", name: "pool" })).gpus).toBeUndefined();
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({
      measurements: [{ type: "TOTAL_GPU", values: [{ value: "0", timestamp: "2026-11-08T00:00:00Z" }] }],
    });
    expect((await loadSchedulingResources({ clusterId: "c", name: "pool" })).gpus).toBe(0);
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
  it("rejects a page that repeats its cursor", async () => {
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValueOnce({ nodePoolResources: [], next: 50 });
    await expect(loadSchedulingDepartmentPage({ poolId: "p", clusterId: "c" }, 50)).rejects.toThrow(
      "pagination did not advance",
    );
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
    id.value = "b";
    await flushPromises();
    expect(data.snapshot.value).toBeUndefined();
    expect(data.loading.value).toBe(true);
    expect(data.baselineRows.value).toEqual([]);
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
    expect(data.baselineRows.value).toEqual([]);
    expect(data.loading.value).toBe(false);
    expect(vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls).toHaveLength(reads);
    expect(vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mock.calls).toHaveLength(sums);
    expect(vi.mocked(orgUnitService.countDepartmentNodePoolResources).mock.calls).toHaveLength(counts);
  });
  it("preserves restricted read visibility and individual department edit scope", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue({ id: "p", clusterId: "c", name: "pool" } as never);
    vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 9, deservedSum: 27 });
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
    expect(data.baselineRows.value.map((value) => value.orgUnitId)).toEqual(["visible"]);
    expect(data.totalsComplete.value).toBe(false);
    expect(data.aggregates.value.all).toMatchObject({ weightsSum: 9, deservedSum: 27 });
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
    expect(data.baselineRows.value).toEqual([]);
    expect(data.resources.value?.gpus).toBeUndefined();
    expect(data.loadError.value).toBe("");
  });
});

describe("node pool GPU limit check", () => {
  it.each([
    { departments: 0, projects: 0, expected: false },
    { departments: 1, projects: 0, expected: true },
    { departments: 0, projects: 1, expected: true },
  ])("checks all departments and projects: $departments/$projects", async ({ departments, projects, expected }) => {
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(departments);
    vi.mocked(orgUnitService.countProjectsNodePoolResources).mockResolvedValue(projects);
    expect(await loadNodePoolGpuLimits({ poolId: "pool", clusterId: "cluster", rank: "High" })).toBe(expected);
    const filters = ["nodePoolId==pool", "clusterId==cluster", "gpuLimit!=-1"];
    expect(orgUnitService.countDepartmentNodePoolResources).toHaveBeenLastCalledWith(filters);
    expect(orgUnitService.countProjectsNodePoolResources).toHaveBeenLastCalledWith(filters);
  });
  it("propagates a failed project check instead of treating it as no limits", async () => {
    vi.mocked(orgUnitService.countDepartmentNodePoolResources).mockResolvedValue(0);
    vi.mocked(orgUnitService.countProjectsNodePoolResources).mockRejectedValueOnce(new Error("Unavailable"));
    await expect(loadNodePoolGpuLimits({ poolId: "pool", clusterId: "cluster" })).rejects.toThrow("Unavailable");
  });
});
