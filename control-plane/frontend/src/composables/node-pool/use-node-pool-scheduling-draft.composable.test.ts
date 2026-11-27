/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { describe, expect, it } from "vitest";
import { createSchedulingDraft, toSchedulingCapabilities } from "@/models/node-pool-scheduling.model";
import { useNodePoolSchedulingDraft } from "@/composables/node-pool/use-node-pool-scheduling-draft.composable";

const snapshot = () => ({
  mode: ESchedulingMode.Proportional as const,
  automaticWeights: false,
  cpuQuota: false,
  departments: {},
});
describe("scheduling draft", () => {
  it.each([false, true])("round-trips a Guaranteed pool with ranks set to %s", (ranks) => {
    const capabilities = { quota: true, ranks, adjustableWeights: true, limits: false, cpuQuota: false };
    const draft = createSchedulingDraft({ ...snapshot(), mode: ESchedulingMode.Guaranteed, capabilities });
    expect(toSchedulingCapabilities(draft)).toEqual(capabilities);
  });
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
  it("preserves the maximum allocation setting and all department edits across mode changes", () => {
    const departments = { a: { id: "a", limit: 10, weight: 5, quota: 8, rank: "High" }, b: { id: "b", limit: -1 } };
    const state = useNodePoolSchedulingDraft({ ...snapshot(), maxAllocation: true, departments });
    for (const mode of [ESchedulingMode.Guaranteed, ESchedulingMode.Ranked, ESchedulingMode.Proportional]) {
      state.requestMode(mode);
      expect(state.draft.value).toMatchObject({ mode, maxAllocation: true, departments });
      expect(toSchedulingCapabilities(state.draft.value).limits).toBe(true);
    }
  });
  it("switches repeatedly without a dialog when only modes changed", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.requestMode(ESchedulingMode.Guaranteed);
    state.requestMode(ESchedulingMode.Ranked);
    expect(state.draft.value.mode).toBe(ESchedulingMode.Ranked);
  });
});
