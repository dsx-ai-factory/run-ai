/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it, vi } from "vitest";
import { ref } from "vue";
import { flushPromises } from "@vue/test-utils";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepool: vi.fn(), getNodepoolMetrics: vi.fn() },
}));
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getDepartmentNodePoolResources: vi.fn(),
    countDepartmentNodePoolResources: vi.fn(),
    sumDepartmentNodePoolResources: vi.fn(),
    sumProjectNodePoolResources: vi.fn(),
  },
}));
vi.mock("@/stores/permissions.store", () => ({ usePermissionStore: () => ({ hasPermission: () => true }) }));
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
  loadDepartmentPage,
  loadSchedulingTotals,
  loadSchedulingWeightTotals,
  telemetryResources,
} from "./node-pool-scheduling.service";
import { useNodePoolSchedulingData } from "./use-node-pool-scheduling-data.composable";
describe("scoped scheduling reads", () => {
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
});
