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
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { loadDepartmentPage, loadSchedulingTotals, telemetryResources } from "./node-pool-scheduling.service";
import { useNodePoolSchedulingData } from "./use-node-pool-scheduling-data.composable";
describe("scoped scheduling reads", () => {
  it("scopes pages and unfiltered aggregates", async () => {
    vi.mocked(orgUnitService.getDepartmentNodePoolResources).mockResolvedValue({ nodePoolResources: [] });
    vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mockResolvedValue({ weightsSum: 10, deservedSum: 3 });
    await loadDepartmentPage({ poolId: "p", clusterId: "c", search: "abc", rank: "High", offset: 50, limit: 50 });
    const args = vi.mocked(orgUnitService.getDepartmentNodePoolResources).mock.calls.at(-1)!;
    expect(args[0]).toEqual(expect.arrayContaining(["nodePoolId==p", "clusterId==c", "rank==High"]));
    expect(args.slice(3)).toEqual([50, 50, "abc"]);
    expect(await loadSchedulingTotals({ poolId: "p", clusterId: "c" })).toEqual({ weightsSum: 10, deservedSum: 3 });
    expect(vi.mocked(orgUnitService.sumDepartmentNodePoolResources).mock.calls.at(-1)).toEqual([
      "p",
      ["nodePoolId==p", "clusterId==c"],
    ]);
  });
  it("preserves measured zero and missing telemetry", () => {
    expect(telemetryResources({ measurements: [] }).gpus).toBeUndefined();
    expect(telemetryResources({ measurements: [{ type: "TOTAL_GPU", values: [{ value: "0" }] }] }).gpus).toBe(0);
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
});
