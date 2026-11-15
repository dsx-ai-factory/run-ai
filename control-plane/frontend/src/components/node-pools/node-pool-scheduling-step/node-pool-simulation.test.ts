/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it, vi, afterEach } from "vitest";
import { effectScope, ref, nextTick } from "vue";
import { buildSimulationRequest, mapSimulationResponse, defaultSimulationDemand } from "./node-pool-simulation.util";
import { useNodePoolSimulation } from "./use-node-pool-simulation.composable";
import type { SimulationScenario } from "./node-pool-simulation.util";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
const draft = (): SchedulingDraft => ({
  mode: "proportional",
  automaticWeights: false,
  cpuQuota: false,
  departments: { hidden: { id: "hidden", weight: 9, rank: "High", limit: 12 } },
});
const scenario = () => ({ poolId: "17", capacity: 144, draft: draft(), ids: ["visible"], demand: { dragged: 23 } });
const response = {
  orgUnits: [
    {
      id: "visible",
      fairShare: { gpu: 12 },
      guaranteed: { gpu: 4 },
      overQuota: { gpu: 8 },
      unmet: { gpu: 3 },
      demand: { gpu: 15 },
      weightShareInRank: 0.25,
    },
  ],
};
afterEach(() => vi.useRealTimers());
describe("simulation contract", () => {
  it("includes offscreen edits and demands, omits inactive fields, and defaults to half capacity", () => {
    const input = scenario();
    const before = JSON.stringify(input);
    const result = buildSimulationRequest(input);
    expect(result).not.toHaveProperty("assumptions");
    expect(result).not.toHaveProperty("parent");
    expect(result.orgUnits).toEqual([
      { id: "visible" },
      { id: "hidden", gpu: { overQuotaWeight: 9 } },
      { id: "dragged", demand: { gpu: 23 } },
    ]);
    expect(result.nodePool?.schedulingCapabilities).toEqual({
      quota: false,
      ranks: false,
      adjustableWeights: true,
      limits: false,
    });
    expect(JSON.stringify(input)).toBe(before);
  });
  it("maps active edits in Guaranteed quota and keeps unsupported limits preview-only", () => {
    const input = scenario();
    input.draft.mode = "guaranteed";
    input.draft.maxAllocation = true;
    input.draft.departments.hidden.quota = 4;
    expect(buildSimulationRequest(input).orgUnits[1]).toEqual({
      id: "hidden",
      rank: "High",
      gpu: { deserved: 4, limit: 12, overQuotaWeight: 9 },
    });
    input.draft.mode = "ranked";
    expect(() => buildSimulationRequest(input)).toThrow(/preview-only/);
    expect(input.draft.departments.hidden.limit).toBe(12);
  });
  it("debounces rapid edits, aborts outstanding work, and reports ordinary errors", async () => {
    vi.useFakeTimers();
    const source = ref(scenario());
    const transport = vi.fn().mockRejectedValue({ response: { status: 500 } });
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    source.value.capacity = 100;
    await vi.advanceTimersByTimeAsync(100);
    source.value.capacity = 120;
    await vi.advanceTimersByTimeAsync(100);
    expect(transport).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(200);
    expect(transport).toHaveBeenCalledTimes(1);
    expect(state.status.value).toBe("error");
    expect(transport.mock.calls[0][1].nodePool.capacity.gpu).toBe(120);
    scope.stop();
    expect(transport.mock.calls[0][2].aborted).toBe(true);
  });
  it("keeps fractional half-capacity slider defaults without sending assumptions", () => {
    expect(defaultSimulationDemand(145)).toBe(72.5);
    expect(buildSimulationRequest({ ...scenario(), capacity: 145 })).not.toHaveProperty("assumptions");
  });
  it("refuses more than 500 scenario rows without dropping edits", () => {
    expect(() =>
      buildSimulationRequest({ ...scenario(), ids: Array.from({ length: 501 }, (_, i) => String(i)) }),
    ).toThrow(/500/);
  });
  it("maps effective demand and server percentages without calculating a schedule", () => {
    expect(mapSimulationResponse(response).visible).toEqual({
      allocated: 12,
      guaranteed: 4,
      overQuota: 8,
      unmet: 3,
      demand: 15,
      weightPercentage: 25,
    });
  });
  it.each([404, 501])("reports an unavailable backend for HTTP %s without changing the draft", async (status) => {
    vi.useFakeTimers();
    const source = ref(scenario());
    const before = JSON.stringify(source.value);
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, vi.fn().mockRejectedValue({ response: { status } })));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    expect(state.status.value).toBe("unavailable");
    expect(JSON.stringify(source.value)).toBe(before);
    scope.stop();
  });
  it("debounces, ignores obsolete responses, handles 501 and retries without changing drafts", async () => {
    vi.useFakeTimers();
    const source = ref(scenario());
    const before = JSON.stringify(source.value);
    let resolveFirst!: (value: typeof response) => void;
    const transport = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveFirst = resolve;
          }),
      )
      .mockResolvedValueOnce(response)
      .mockRejectedValueOnce({ response: { status: 501 } })
      .mockResolvedValue(response);
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    expect(transport).toHaveBeenCalledTimes(1);
    source.value = { ...source.value, capacity: 100 };
    await nextTick();
    await vi.advanceTimersByTimeAsync(300);
    expect(state.results.value?.visible.allocated).toBe(12);
    resolveFirst({ ...response, orgUnits: [] });
    await nextTick();
    expect(state.results.value?.visible.allocated).toBe(12);
    state.retry();
    await vi.advanceTimersByTimeAsync(300);
    expect(state.status.value).toBe("unavailable");
    expect(state.results.value).toBeUndefined();
    state.retry();
    await vi.advanceTimersByTimeAsync(300);
    expect(state.status.value).toBe("ready");
    expect(JSON.stringify({ ...source.value, capacity: 144 })).toBe(before);
    scope.stop();
  });

  it.each([
    { capacity: 0, demand: 0 },
    { capacity: 1, demand: 0.5 },
    { capacity: 145, demand: 72.5 },
    { capacity: 1000000, demand: 500000 },
  ])("uses slider placeholder $demand at capacity $capacity", ({ capacity, demand }) => {
    const request = buildSimulationRequest({ ...scenario(), capacity });
    expect(request.nodePool?.capacity?.gpu).toBe(capacity);
    expect(defaultSimulationDemand(capacity)).toBe(demand);
    expect(request).not.toHaveProperty("assumptions");
    expect(request.orgUnits[0]).not.toHaveProperty("demand");
  });
  it.each([-1, NaN, Infinity])("blocks unavailable capacity %s", (capacity) => {
    expect(() => buildSimulationRequest({ ...scenario(), capacity })).toThrow(/capacity/);
  });
  it.each([499, 500])("keeps all %s unique displayed and hidden department IDs", (size) => {
    const input = scenario();
    input.ids = Array.from({ length: size - 2 }, (_, i) => `visible-${i}`);
    input.ids.push("visible-0", "hidden", "dragged");
    const request = buildSimulationRequest(input);
    expect(request.orgUnits).toHaveLength(size);
    expect(request.orgUnits.at(-2)).toEqual({ id: "hidden", gpu: { overQuotaWeight: 9 } });
    expect(request.orgUnits.at(-1)).toEqual({ id: "dragged", demand: { gpu: 23 } });
  });
  it("blocks 501 combined IDs even with only 499 visible rows and preserves the hidden edits", () => {
    const input = scenario();
    input.ids = Array.from({ length: 499 }, (_, i) => `visible-${i}`);
    expect(() => buildSimulationRequest(input)).toThrow(/500/);
    expect(input.draft.departments.hidden.weight).toBe(9);
    expect(input.demand.dragged).toBe(23);
  });
  it("keeps zero demand, zero active edits, and explicit unlimited allocation", () => {
    const input = scenario();
    input.draft.mode = "guaranteed";
    input.draft.maxAllocation = true;
    input.draft.departments.hidden = { id: "hidden", weight: 0, quota: 0, limit: null, rank: "High" };
    input.demand.dragged = 0;
    expect(buildSimulationRequest(input).orgUnits).toEqual([
      { id: "visible" },
      { id: "hidden", rank: "High", gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 } },
      { id: "dragged", demand: { gpu: 0 } },
    ]);
    input.draft.automaticWeights = true;
    expect(buildSimulationRequest(input).orgUnits[1]).toEqual({
      id: "hidden",
      rank: "High",
      gpu: { deserved: 0, limit: -1 },
    });
  });
  it("preserves fractional response fields and percentage with zero allocations", () => {
    expect(
      mapSimulationResponse({
        orgUnits: [
          {
            id: "fractional",
            fairShare: { gpu: 0.375 },
            guaranteed: { gpu: 0 },
            overQuota: { gpu: 0.375 },
            unmet: { gpu: 0.125 },
            demand: { gpu: 0.5 },
            weightShareInRank: 0.125,
          },
        ],
      }),
    ).toEqual({
      fractional: {
        allocated: 0.375,
        guaranteed: 0,
        overQuota: 0.375,
        unmet: 0.125,
        demand: 0.5,
        weightPercentage: 12.5,
      },
    });
  });
  it.each([400, 401, 403, 404, 500, 501, undefined])(
    "clears stale results on failure %s and retries the same draft",
    async (code) => {
      vi.useFakeTimers();
      const source = ref(scenario());
      const before = JSON.stringify(source.value);
      const transport = vi
        .fn()
        .mockResolvedValueOnce(response)
        .mockRejectedValueOnce(code === undefined ? new Error("network down") : { response: { status: code } })
        .mockResolvedValueOnce(response);
      const scope = effectScope();
      const state = scope.run(() => useNodePoolSimulation(source, transport));
      if (!state) throw new Error("Simulation test scope is inactive.");
      await vi.advanceTimersByTimeAsync(300);
      expect(state.results.value?.visible.allocated).toBe(12);
      state.retry();
      expect(state.results.value).toBeUndefined();
      await vi.advanceTimersByTimeAsync(300);
      expect(state.status.value).toBe((code === 404 || code === 501) ? "unavailable" : "error");
      expect(state.message.value).toMatch(/preserved/);
      state.retry();
      await vi.advanceTimersByTimeAsync(300);
      expect(state.status.value).toBe("ready");
      expect(state.message.value).toBe("");
      expect(state.results.value?.visible.allocated).toBe(12);
      expect(JSON.stringify(source.value)).toBe(before);
      scope.stop();
    },
  );
  it("ignores an obsolete rejection after the replacement request succeeds", async () => {
    vi.useFakeTimers();
    let rejectOld!: (reason: unknown) => void;
    const transport = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((_resolve, reject) => {
            rejectOld = reject;
          }),
      )
      .mockResolvedValueOnce(response);
    const source = ref(scenario());
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    source.value.poolId = "other-pool";
    expect(transport.mock.calls[0][2].aborted).toBe(true);
    await vi.advanceTimersByTimeAsync(300);
    expect(transport.mock.calls[1][0]).toBe("other-pool");
    rejectOld({ response: { status: 501 } });
    await nextTick();
    expect(state.status.value).toBe("ready");
    expect(state.results.value?.visible.allocated).toBe(12);
    expect(state.message.value).toBe("");
    scope.stop();
  });
  it("ignores an obsolete success after the replacement request fails", async () => {
    vi.useFakeTimers();
    let resolveOld!: (value: typeof response) => void;
    const transport = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOld = resolve;
          }),
      )
      .mockRejectedValueOnce({ response: { status: 403 } });
    const source = ref(scenario());
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    source.value.capacity = 0;
    await vi.advanceTimersByTimeAsync(300);
    resolveOld(response);
    await nextTick();
    expect(state.status.value).toBe("error");
    expect(state.results.value).toBeUndefined();
    scope.stop();
  });
  it("cancels a pending debounce when disabled and ignores a late response after disposal", async () => {
    vi.useFakeTimers();
    let resolvePending!: (value: typeof response) => void;
    const transport = vi.fn().mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvePending = resolve;
        }),
    );
    const source = ref<SimulationScenario | undefined>(scenario());
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    source.value = undefined;
    await vi.advanceTimersByTimeAsync(300);
    expect(transport).not.toHaveBeenCalled();
    expect(state.status.value).toBe("idle");
    source.value = scenario();
    await vi.advanceTimersByTimeAsync(300);
    scope.stop();
    expect(transport.mock.calls[0][2].aborted).toBe(true);
    resolvePending(response);
    await nextTick();
    expect(state.results.value).toBeUndefined();
  });
});
