/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { RunaiToggle } from "@/components/common/runai-toggle";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import Settings from "./node-pool-scheduling-settings.vue";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
const draft: SchedulingDraft = {
  mode: "guaranteed",
  automaticWeights: true,
  cpuQuota: false,
  maxAllocation: true,
  departments: { a: { id: "a", limit: 8, weight: 2, quota: 3, rank: "High" } },
};
describe("scheduling settings draft interactions", () => {
  it("keeps max edits on cancel and only discards limits on explicit Disable", async () => {
    const wrapper = mount(Settings, { props: { draft, quotaPolicy: true }, global: { plugins: [Quasar] } });
    const toggle = wrapper
      .findAllComponents(RunaiToggle)
      .find((toggle) => toggle.props("aid") === "max-allocation-toggle");
    if (!toggle) throw new Error("Maximum allocation toggle is missing");
    toggle.vm.$emit("update:model-value", false);
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted("patch")).toBeUndefined();
    wrapper.findComponent(RunaiBaseDialog).vm.$emit("close");
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted("patch")).toBeUndefined();
    toggle.vm.$emit("update:model-value", false);
    await wrapper.vm.$nextTick();
    (document.querySelector('[aid="confirm-disable-setting"]') as HTMLButtonElement).click();
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted("patch")?.[0]).toEqual([
      { maxAllocation: false, departments: { a: { id: "a", weight: 2, quota: 3, rank: "High" } } },
    ]);
    expect(draft.departments.a.limit).toBe(8);
    wrapper.unmount();
  });
  it("never offers an organization policy mutation", () => {
    const wrapper = mount(Settings, { props: { draft, quotaPolicy: true }, global: { plugins: [Quasar] } });
    const policy = wrapper
      .findAllComponents(RunaiToggle)
      .find((toggle) => toggle.props("aid") === "project-quota-policy");
    if (!policy) throw new Error("Project quota policy toggle is missing");
    expect(policy.props("disable")).toBe(true);
    expect(policy.props("modelValue")).toBe(true);
    expect(wrapper.emitted("patch")).toBeUndefined();
    wrapper.unmount();
  });
});
