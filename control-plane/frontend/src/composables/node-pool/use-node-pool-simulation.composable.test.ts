/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { BadRequestError, ForbiddenError } from "@/models/http-response.model";
import { describe, expect, it, vi, afterEach } from "vitest";
import { effectScope, ref, nextTick } from "vue";
import {
  buildSimulationRequest,
  mapSimulationResponse,
  defaultSimulationDemand,
} from "@/utils/node-pool-scheduling.util/node-pool-simulation.util";
import { useNodePoolSimulation as useSimulation } from "@/composables/node-pool/use-node-pool-simulation.composable";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: { simulateNodePoolFairShare: vi.fn() },
}));
function useNodePoolSimulation(
  source: Parameters<typeof useSimulation>[0],
  transport: typeof orgUnitService.simulateNodePoolFairShare,
) {
  vi.mocked(orgUnitService.simulateNodePoolFairShare).mockImplementation(transport);
  return useSimulation(source);
}
import type { ISimulationScenario } from "@/utils/node-pool-scheduling.util/node-pool-simulation.util";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
const draft = (): SchedulingDraft => ({
  mode: ESchedulingMode.Proportional,
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
  it.each([undefined, null])("handles a service rejection without an error payload: %s", async (error) => {
    vi.useFakeTimers();
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(ref(scenario()), vi.fn().mockRejectedValue(error)));
    if (!state) throw new Error("Missing simulation scope");
    await vi.advanceTimersByTimeAsync(300);
    expect(state.status.value).toBe("error");
    expect(state.message.value).toMatch(/Simulation failed/);
    scope.stop();
  });
  it.each([ESchedulingMode.Unknown, ESchedulingMode.Custom])("does not simulate unsupported mode %s", (mode) => {
    const input = scenario();
    input.draft.mode = mode;
    expect(() => buildSimulationRequest(input)).toThrow(/supported scheduling mode/);
  });
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
  it("maps Guaranteed limits with quota and rank", () => {
    const input = scenario();
    input.draft.mode = ESchedulingMode.Guaranteed;
    input.draft.maxAllocation = true;
    input.draft.departments.hidden.quota = 4;
    expect(buildSimulationRequest(input).orgUnits[1]).toEqual({
      id: "hidden",
      rank: "High",
      gpu: { deserved: 4, limit: 12, overQuotaWeight: 9 },
    });
  });
  it.each([ESchedulingMode.Proportional, ESchedulingMode.Ranked])(
    "sends only active limits and omits quota in %s",
    (mode) => {
      const input = scenario();
      input.draft.mode = mode;
      input.draft.maxAllocation = true;
      input.draft.departments.hidden.quota = 15;
      const active = buildSimulationRequest(input);
      expect(active.nodePool?.schedulingCapabilities).toMatchObject({ quota: false, limits: true });
      expect(active.orgUnits[1].gpu).toEqual({ overQuotaWeight: 9, limit: 12 });
      input.draft.maxAllocation = false;
      expect(buildSimulationRequest(input).nodePool?.schedulingCapabilities?.limits).toBe(false);
      expect(buildSimulationRequest(input).orgUnits[1].gpu).not.toHaveProperty("limit");
      expect(input.draft.departments.hidden.limit).toBe(12);
    },
  );
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
  it("does not request or interrupt simulation when unrelated settings change", async () => {
    vi.useFakeTimers();
    let resolve!: (value: typeof response) => void;
    const transport = vi.fn().mockImplementation(
      () =>
        new Promise((done) => {
          resolve = done;
        }),
    );
    const source = ref(scenario());
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    source.value.draft.cpuQuota = true;
    source.value.draft.timeFairshare = true;
    source.value.draft.minimumRuntime = { Seconds: 30 };
    await vi.advanceTimersByTimeAsync(500);
    expect(transport).toHaveBeenCalledTimes(1);
    expect(transport.mock.calls[0][2].aborted).toBe(false);
    resolve(response);
    await nextTick();
    expect(state.results.value?.visible.allocated).toBe(12);
    source.value.draft.historicalWeight = 5;
    source.value = { ...source.value, draft: { ...source.value.draft } };
    expect(state.status.value).toBe("ready");
    await vi.advanceTimersByTimeAsync(500);
    expect(transport).toHaveBeenCalledTimes(1);
    state.retry();
    await vi.advanceTimersByTimeAsync(300);
    expect(transport).toHaveBeenCalledTimes(2);
    scope.stop();
  });
  it("reschedules a previous payload after an intervening edit or disable", async () => {
    vi.useFakeTimers();
    const transport = vi.fn().mockResolvedValue(response);
    const source = ref<ISimulationScenario | undefined>(scenario());
    const scope = effectScope();
    scope.run(() => useNodePoolSimulation(source, transport));
    await vi.advanceTimersByTimeAsync(300);
    if (!source.value) throw new Error("Simulation scenario is missing.");
    source.value.demand.dragged = 40;
    await vi.advanceTimersByTimeAsync(300);
    source.value.demand.dragged = 23;
    await vi.advanceTimersByTimeAsync(300);
    expect(transport).toHaveBeenCalledTimes(3);
    source.value = undefined;
    source.value = scenario();
    await vi.advanceTimersByTimeAsync(300);
    expect(transport).toHaveBeenCalledTimes(4);
    scope.stop();
  });
  it.each([
    {
      code: 400,
      server: "Maximum allocation must be at least guaranteed GPUs.",
      text: /Maximum allocation must be at least guaranteed GPUs/,
    },
    { code: 403, server: "internal authorization detail", text: /permission/i },
  ])("shows actionable HTTP $code feedback without suggesting an identical retry", async ({ code, server, text }) => {
    vi.useFakeTimers();
    const source = ref(scenario());
    const before = JSON.stringify(source.value);
    const scope = effectScope();
    const state = scope.run(() =>
      useNodePoolSimulation(
        source,
        vi.fn().mockRejectedValue({ response: { status: code, data: { message: server } } }),
      ),
    );
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    expect(state.message.value).toMatch(text);
    expect(state.canRetry.value).toBe(false);
    expect(state.message.value).not.toContain("Try again");
    expect(JSON.stringify(source.value)).toBe(before);
    scope.stop();
  });
  it.each([new BadRequestError("Limit below quota"), new ForbiddenError()])(
    "retains normalized service error feedback for $statusCode",
    async (error) => {
      vi.useFakeTimers();
      const scope = effectScope();
      const state = scope.run(() => useNodePoolSimulation(ref(scenario()), vi.fn().mockRejectedValue(error)));
      if (!state) throw new Error("Missing scope");
      await vi.advanceTimersByTimeAsync(300);
      expect(state.canRetry.value).toBe(false);
      expect(state.message.value).toMatch(error.statusCode === 400 ? /Limit below quota/ : /permission/i);
      scope.stop();
    },
  );
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
  it.each([404])("reports an unavailable backend for HTTP %s without changing the draft", async (status) => {
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
  it("debounces, ignores obsolete responses and retries server failures without changing drafts", async () => {
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
    expect(state.status.value).toBe("error");
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
    input.draft.mode = ESchedulingMode.Guaranteed;
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
  it("starts mode switches immediately and restores only matching results while refreshing", async () => {
    vi.useFakeTimers();
    const source = ref(scenario());
    const transport = vi.fn().mockResolvedValue(response);
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(0);
    expect(state.results.value?.visible.allocated).toBe(12);

    transport.mockResolvedValue({ orgUnits: [{ ...response.orgUnits[0], fairShare: { gpu: 20 } }] });
    source.value.draft.mode = ESchedulingMode.Ranked;
    expect(state.results.value).toBeUndefined();
    await vi.advanceTimersByTimeAsync(0);
    expect(state.results.value?.visible.allocated).toBe(20);

    source.value.draft.mode = ESchedulingMode.Proportional;
    expect(state.status.value).toBe("loading");
    expect(state.results.value?.visible.allocated).toBe(12);
    source.value.demand.dragged++;
    source.value.draft.mode = ESchedulingMode.Ranked;
    expect(state.results.value).toBeUndefined();
    await vi.advanceTimersByTimeAsync(0);
    expect(transport).toHaveBeenCalledTimes(3);
    scope.stop();
  });

  it("keeps the previous result throughout debounce and request, then replaces it", async () => {
    vi.useFakeTimers();
    let resolveNext!: (value: typeof response) => void;
    const transport = vi
      .fn()
      .mockResolvedValueOnce(response)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveNext = resolve;
          }),
      );
    const source = ref(scenario());
    const scope = effectScope();
    const state = scope.run(() => useNodePoolSimulation(source, transport));
    if (!state) throw new Error("Simulation test scope is inactive.");
    await vi.advanceTimersByTimeAsync(300);
    const previous = state.results.value;
    source.value.demand.dragged = 40;
    expect(state.status.value).toBe("loading");
    expect(state.results.value).toBe(previous);
    await vi.advanceTimersByTimeAsync(300);
    expect(state.results.value).toBe(previous);
    resolveNext({ orgUnits: [{ ...response.orgUnits[0], fairShare: { gpu: 20 } }] });
    await nextTick();
    expect(state.status.value).toBe("ready");
    expect(state.results.value?.visible.allocated).toBe(20);
    scope.stop();
  });
  it.each(["pool", "capacity", "mode", "disabled"])(
    "clears previous results when simulation context changes: %s",
    async (change) => {
      vi.useFakeTimers();
      const source = ref<ISimulationScenario | undefined>(scenario());
      const scope = effectScope();
      const state = scope.run(() => useNodePoolSimulation(source, vi.fn().mockResolvedValue(response)));
      if (!state || !source.value) throw new Error("Simulation test scope is inactive.");
      await vi.advanceTimersByTimeAsync(300);
      expect(state.results.value?.visible.allocated).toBe(12);
      if (change === "pool") source.value.poolId = "other";
      if (change === "capacity") source.value.capacity = 64;
      if (change === "mode") source.value.draft.mode = ESchedulingMode.Ranked;
      if (change === "disabled") source.value = undefined;
      expect(state.results.value).toBeUndefined();
      scope.stop();
    },
  );
  it.each([
    [400, "Simulation inputs are invalid."],
    [401, "Simulation failed."],
    [403, "You’re missing permissions to simulate GPU allocation."],
    [404, "Simulation is unavailable."],
    [500, "Simulation failed."],
    [501, "Simulation failed."],
    [undefined, "Simulation failed."],
  ] as const)("clears stale results on failure %s and retries the same draft", async (code, message) => {
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
    expect(state.results.value?.visible.allocated).toBe(12);
    await vi.advanceTimersByTimeAsync(300);
    expect(state.results.value).toBeUndefined();
    expect(state.status.value).toBe(code === 404 ? "unavailable" : "error");
    expect(state.message.value).toBe(message);
    state.retry();
    await vi.advanceTimersByTimeAsync(300);
    expect(state.status.value).toBe("ready");
    expect(state.message.value).toBe("");
    expect(state.results.value?.visible.allocated).toBe(12);
    expect(JSON.stringify(source.value)).toBe(before);
    scope.stop();
  });
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
    const source = ref<ISimulationScenario | undefined>(scenario());
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
