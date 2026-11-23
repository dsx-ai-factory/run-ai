/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { describe, it, expect, vi } from "vitest";
import { buildSchedulingSave } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-save.util";
import type { Nodepool } from "@/swagger-models/cluster-service-client/index";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
import { useNodePoolSchedulingSave } from "@/composables/node-pool/use-node-pool-scheduling-save.composable";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
function createSchedulingSaver(transport: {
  patchPool: typeof clusterService.patchNodepool;
  patchDepartments: typeof orgUnitService.batchPatchDepartmentNodePoolResources;
}) {
  vi.spyOn(clusterService, "patchNodepool").mockImplementation(transport.patchPool);
  vi.spyOn(orgUnitService, "batchPatchDepartmentNodePoolResources").mockImplementation(transport.patchDepartments);
  return useNodePoolSchedulingSave();
}
const pool = {
  id: "17",
  schedulingCapabilities: { quota: false, ranks: false, adjustableWeights: true, cpuQuota: false, limits: false },
  schedulingConfiguration: { gpuPodsPlacementStrategy: { node: "spread" } },
} as Nodepool;
const draft = (): SchedulingDraft => ({
  mode: ESchedulingMode.Proportional,
  automaticWeights: false,
  cpuQuota: false,
  departments: { a: { id: "a", weight: 3, quota: 8, limit: 9, rank: "High" } },
});
const success = (ids: string[]) => ({
  id: "batch",
  succeeded: ids.map((orgUnitId) => ({ orgUnitId, nodePoolId: "17" })),
  totalRequested: ids.length,
  totalSuccessful: ids.length,
  totalFailed: 0,
});
describe("scheduling persistence", () => {
  it.each(["saved-topology", null, undefined])(
    "omits topology from a scheduling-only PATCH even when the form loaded %s",
    (networkTopologyId) => {
      expect(buildSchedulingSave({ ...pool, networkTopologyId }, draft()).pool).not.toHaveProperty("networkTopologyId");
    },
  );
  it.each([ESchedulingMode.Unknown, ESchedulingMode.Custom])("does not save unsupported mode %s", (mode) => {
    expect(() => buildSchedulingSave(pool, { ...draft(), mode })).toThrow(/supported scheduling mode/);
  });
  it("maps only active resource fields and preserves unrelated pool settings", () => {
    const result = buildSchedulingSave(pool, draft());
    expect(result.departments).toEqual([
      { orgUnitId: "a", nodePoolId: "17", resources: { gpu: { overQuotaWeight: 3 } } },
    ]);
    expect(result.pool.schedulingConfiguration).not.toHaveProperty("gpuPodsPlacementStrategy");
    expect(result.pool).not.toHaveProperty("labelKey");
  });
  it("omits inactive limits and rejects downgrade or invalid durations before any writes", () => {
    const inactive = buildSchedulingSave(pool, { ...draft(), maxAllocation: true });
    expect(inactive.pool.schedulingCapabilities?.limits).toBe(false);
    expect(inactive.departments[0].resources?.gpu).not.toHaveProperty("limit");
    expect(() =>
      buildSchedulingSave(
        {
          ...pool,
          schedulingCapabilities: { adjustableWeights: true, ranks: false, quota: true, cpuQuota: false, limits: false },
        },
        draft(),
      ),
    ).toThrow(/downgrade/);
    expect(() => buildSchedulingSave(pool, { ...draft(), minimumRuntime: { Days: -1 } })).toThrow(/duration/);
  });
  it("omits stale fairshare fields when disabled and rejects negative fractional limits", () => {
    const saved = {
      ...pool,
      schedulingConfiguration: { timeBasedFairShare: { enabled: true, historicalUsageWeight: 2, windowDuration: "7d" } },
    };
    expect(
      buildSchedulingSave(saved, { ...draft(), timeFairshare: false, historicalWeight: 2 }).pool.schedulingConfiguration
        ?.timeBasedFairShare,
    ).toEqual({ enabled: false });
    expect(() =>
      buildSchedulingSave(pool, {
        ...draft(),
        mode: ESchedulingMode.Guaranteed,
        maxAllocation: true,
        departments: { a: { id: "a", limit: -0.5 } },
      }),
    ).toThrow(/Maximum allocation/);
  });
  it("saves pool first and retries only unsuccessful department patches, preserving draft", async () => {
    const update = vi.fn().mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    const batch = vi
      .fn()
      .mockResolvedValueOnce({
        ...success(["a"]),
        totalRequested: 2,
        totalFailed: 1,
        failed: [{ payload: { orgUnitId: "b", nodePoolId: "17" }, message: "quota conflict" }],
      })
      .mockResolvedValueOnce(success(["b"]));
    const saver = createSchedulingSaver({ patchPool: update, patchDepartments: batch });
    const value = draft();
    value.departments.b = { id: "b", weight: 4 };
    const before = JSON.stringify(value);
    await expect(saver.save(pool, value)).rejects.toThrow(/quota conflict/);
    await saver.save(pool, value);
    expect(update).toHaveBeenCalledTimes(1);
    expect(batch.mock.calls[1][0].patchRequests.map((row: { orgUnitId: string }) => row.orgUnitId)).toEqual(["b"]);
    expect(update.mock.invocationCallOrder[0]).toBeLessThan(batch.mock.invocationCallOrder[0]);
    expect(JSON.stringify(value)).toBe(before);
  });
  it("chunks department patches at 100 and does not write them when pool update fails", async () => {
    const value = draft();
    value.departments = Object.fromEntries(
      Array.from({ length: 201 }, (_, i) => [String(i), { id: String(i), weight: 2 }]),
    );
    const update = vi
      .fn()
      .mockRejectedValueOnce(new Error("permission denied"))
      .mockImplementation(async (_id, payload) => ({ ...pool, ...payload }));
    const batch = vi
      .fn()
      .mockImplementation(async (request) =>
        success(request.patchRequests.map((r: { orgUnitId: string }) => r.orgUnitId)),
      );
    const saver = createSchedulingSaver({ patchPool: update, patchDepartments: batch });
    await expect(saver.save(pool, value)).rejects.toThrow(/permission denied/);
    expect(batch).not.toHaveBeenCalled();
    await saver.save(pool, value);
    expect(batch.mock.calls.map((c) => c[0].patchRequests.length)).toEqual([100, 100, 1]);
  });
  it("rejects a server that ignores capabilities and suppresses duplicate submissions", async () => {
    let finish!: (value: Nodepool) => void;
    const update = vi.fn(() => new Promise<Nodepool>((resolve) => (finish = resolve)));
    const batch = vi.fn();
    const saver = createSchedulingSaver({ patchPool: update, patchDepartments: batch });
    const value = { ...draft(), mode: ESchedulingMode.Ranked as const };
    const pending = saver.save(pool, value);
    await expect(saver.save(pool, value)).rejects.toThrow(/already/);
    finish(pool);
    await expect(pending).rejects.toThrow(/capabilities/);
    expect(batch).not.toHaveBeenCalled();
  });

  it.each([
    [{ Days: 7, Hours: 0, Minutes: 0 }, "7d"],
    [{ Days: 0, Hours: 2, Minutes: 30 }, "2h30m"],
    [{ Days: 0, Hours: 0, Minutes: 0 }, "0"],
  ])("saves historical windows using the API-supported minute/hour/day units: %j", (historicalWindow, expected) => {
    const actual = buildSchedulingSave(pool, {
      ...draft(),
      timeFairshare: true,
      historicalWeight: 1,
      historicalWindow,
    }).pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration;
    expect(actual).toBe(expected);
  });
  it("does not silently discard nonzero seconds in a historical window", () => {
    expect(() =>
      buildSchedulingSave(pool, {
        ...draft(),
        timeFairshare: true,
        historicalWindow: { Minutes: 1, Seconds: 30 },
      }),
    ).toThrow(/historical usage window/i);
  });
  it("persists a fractional runtime change from saved whole seconds", () => {
    const saved = { ...pool, schedulingConfiguration: { minGuaranteedRuntime: "5s" } };
    const actual = buildSchedulingSave(saved, { ...draft(), minimumRuntime: { Seconds: 0.5 } });
    expect(actual.pool.schedulingConfiguration?.minGuaranteedRuntime).toBe("0.5s");
  });
  it.each([
    ["0.5h", { Minutes: 30 }],
    ["2w", { Days: 14 }],
    ["500ms", { Seconds: 0.5 }],
  ])("keeps a semantically unchanged saved runtime %s untouched", (minGuaranteedRuntime, minimumRuntime) => {
    const saved = { ...pool, schedulingConfiguration: { minGuaranteedRuntime } };
    expect(buildSchedulingSave(saved, { ...draft(), minimumRuntime }).pool.schedulingConfiguration).not.toHaveProperty(
      "minGuaranteedRuntime",
    );
  });
  it("persists changing a saved runtime to zero", () => {
    const saved = { ...pool, schedulingConfiguration: { minGuaranteedRuntime: "1h" } };
    expect(
      buildSchedulingSave(saved, { ...draft(), minimumRuntime: { Hours: 0 } }).pool.schedulingConfiguration,
    ).toHaveProperty("minGuaranteedRuntime", "0");
  });

  it("preserves an equivalent saved historical window without rewriting its units", () => {
    const saved = {
      ...pool,
      schedulingConfiguration: { timeBasedFairShare: { enabled: true, windowDuration: "0.5h" } },
    };
    const actual = buildSchedulingSave(saved, {
      ...draft(),
      timeFairshare: true,
      historicalWindow: { Minutes: 30 },
    });
    expect(actual.pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration).toBe("0.5h");
  });
  it("retains fractional minutes in the historical window instead of truncating them", () => {
    const actual = buildSchedulingSave(pool, {
      ...draft(),
      timeFairshare: true,
      historicalWindow: { Minutes: 0.5 },
    });
    expect(actual.pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration).toBe("0.5m");
  });
});
