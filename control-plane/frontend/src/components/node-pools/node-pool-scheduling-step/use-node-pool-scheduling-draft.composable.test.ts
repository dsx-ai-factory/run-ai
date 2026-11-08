/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { createSchedulingDraft, toSchedulingCapabilities } from "./node-pool-scheduling.model";
import { useNodePoolSchedulingDraft } from "./use-node-pool-scheduling-draft.composable";

const snapshot = () => ({ mode: "proportional" as const, automaticWeights: false, cpuQuota: false, departments: {} });
describe("scheduling draft", () => {
  it("maps each mode to capabilities", () => {
    const draft = createSchedulingDraft(snapshot());
    expect(toSchedulingCapabilities(draft)).toEqual({
      adjustableWeights: true,
      quota: false,
      ranks: false,
      cpuQuota: false,
    });
    draft.mode = "ranked";
    expect(toSchedulingCapabilities(draft).ranks).toBe(true);
    draft.mode = "guaranteed";
    draft.automaticWeights = true;
    expect(toSchedulingCapabilities(draft)).toEqual({
      adjustableWeights: false,
      quota: true,
      ranks: true,
      cpuQuota: false,
    });
  });
  it("cancels or discards field edits, with no per-mode cache", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.draft.value.departments.a = { id: "a", weight: 5 };
    state.requestMode("proportional");
    expect(state.pendingMode.value).toBeUndefined();
    state.requestMode("ranked");
    expect(state.draft.value.mode).toBe("proportional");
    state.cancelModeChange();
    expect(state.draft.value.departments.a.weight).toBe(5);
    state.requestMode("ranked");
    state.confirmModeChange();
    expect(state.draft.value.departments).toEqual({});
    expect(state.draft.value.mode).toBe("ranked");
    expect(state.resetRevision.value).toBe(1);
    state.requestMode("proportional");
    expect(state.pendingMode.value).toBeUndefined();
    expect(state.draft.value.departments).toEqual({});
  });
  it("switches repeatedly without a dialog when only modes changed", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.requestMode("guaranteed");
    state.requestMode("ranked");
    expect(state.pendingMode.value).toBeUndefined();
    expect(state.draft.value.mode).toBe("ranked");
  });
  it("discards to the latest saved checkpoint", () => {
    const state = useNodePoolSchedulingDraft(snapshot());
    state.acceptSnapshot({ ...snapshot(), departments: { a: { id: "a", weight: 3 } } });
    state.draft.value.departments.a.weight = 8;
    state.requestMode("ranked");
    state.confirmModeChange();
    expect(state.draft.value.departments.a.weight).toBe(3);
  });
});
