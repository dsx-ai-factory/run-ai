/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingModeId as ESchedulingMode } from "@/models/node-pool.model";
import { describe, expect, it } from "vitest";
import { createSchedulingDraft, toSchedulingCapabilities } from "./node-pool-scheduling.model";
import { useNodePoolSchedulingDraft } from "./use-node-pool-scheduling-draft.composable";

const snapshot = () => ({
  mode: ESchedulingMode.Proportional as const,
  automaticWeights: false,
  cpuQuota: false,
  departments: {},
});
describe("scheduling draft", () => {
  it("maps each mode to capabilities", () => {
    const draft = createSchedulingDraft(snapshot());
    expect(toSchedulingCapabilities(draft)).toEqual({
      adjustableWeights: true,
      quota: false,
      ranks: false,
      cpuQuota: false,
      limits: false,
    });
    draft.mode = ESchedulingMode.Ranked;
    expect(toSchedulingCapabilities(draft).ranks).toBe(true);
    draft.mode = ESchedulingMode.Guaranteed;
    draft.automaticWeights = true;
    expect(toSchedulingCapabilities(draft)).toEqual({
      adjustableWeights: false,
      quota: true,
      ranks: true,
      cpuQuota: false,
      limits: false,
    });
  });
  it("preserves non-limit entered values when switching modes", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.draft.value.departments.a = { id: "a", weight: 5, quota: 8, limit: 10, rank: "High" };
    state.requestMode(ESchedulingMode.Ranked);
    state.requestMode(ESchedulingMode.Guaranteed);
    state.requestMode(ESchedulingMode.Proportional);
    expect(state.pendingMode.value).toBeUndefined();
    expect(state.draft.value.departments.a).toEqual({ id: "a", weight: 5, quota: 8, rank: "High" });
  });
  it.each([ESchedulingMode.Proportional, ESchedulingMode.Ranked] as const)(
    "clears maximum allocation and every draft limit on leaving Guaranteed for %s",
    (mode) => {
      const state = useNodePoolSchedulingDraft({
        ...snapshot(),
        mode: ESchedulingMode.Guaranteed,
        maxAllocation: true,
        departments: { a: { id: "a", limit: 10, weight: 5, quota: 8, rank: "High" }, b: { id: "b", limit: -1 } },
      });
      state.requestMode(mode);
      expect(state.draft.value.maxAllocation).toBe(false);
      expect(state.draft.value.departments.a).toEqual({ id: "a", weight: 5, quota: 8, rank: "High" });
      expect(state.draft.value.departments.b).toEqual({ id: "b" });
      state.requestMode(ESchedulingMode.Guaranteed);
      expect(state.draft.value.maxAllocation).toBe(false);
      expect(state.draft.value.departments.a).not.toHaveProperty("limit");
    },
  );
  it("switches repeatedly without a dialog when only modes changed", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.requestMode(ESchedulingMode.Guaranteed);
    state.requestMode(ESchedulingMode.Ranked);
    expect(state.pendingMode.value).toBeUndefined();
    expect(state.draft.value.mode).toBe(ESchedulingMode.Ranked);
  });
  it("preserves edits after accepting a saved checkpoint", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.acceptSnapshot({ ...snapshot(), departments: { a: { id: "a", weight: 3 } } });
    state.draft.value.departments.a.weight = 8;
    state.requestMode(ESchedulingMode.Ranked);
    state.confirmModeChange();
    expect(state.draft.value.departments.a.weight).toBe(8);
  });
});
