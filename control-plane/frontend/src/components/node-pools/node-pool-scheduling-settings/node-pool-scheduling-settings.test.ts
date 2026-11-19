/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { QInput, QOptionGroup, Quasar } from "quasar";
import { RunaiToggle } from "@/components/common/runai-toggle/index";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/index";
import Settings from "@/components/node-pools/node-pool-scheduling-settings/node-pool-scheduling-settings.vue";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
const draft: SchedulingDraft = {
  mode: ESchedulingMode.Guaranteed,
  automaticWeights: true,
  cpuQuota: false,
  maxAllocation: true,
  departments: { a: { id: "a", limit: 8, weight: 2, quota: 3, rank: "High" } },
};
describe("scheduling settings draft interactions", () => {
  it("changes custom weights through the local radio group and prevents readonly patches", async () => {
    const wrapper = mount(Settings, {
      props: { draft },
      global: { plugins: [Quasar] },
    });
    const options = wrapper.findComponent(QOptionGroup);
    expect(options.attributes("aria-label")).toBe("Weight within rank");
    expect(options.props("modelValue")).toBe("automatic");
    options.vm.$emit("update:model-value", "custom");
    expect(wrapper.emitted("patch")).toEqual([[{ automaticWeights: false }]]);
    await wrapper.setProps({ readonly: true });
    expect(options.props("disable")).toBe(true);
    options.vm.$emit("update:model-value", "automatic");
    expect(wrapper.emitted("patch")).toHaveLength(1);
    wrapper.unmount();
  });
  it.each([ESchedulingMode.Proportional, ESchedulingMode.Ranked] as const)(
    "hides quota-only advanced settings in %s even with stale enabled values",
    (mode) => {
      const wrapper = mount(Settings, {
        props: { draft: { ...draft, mode, cpuQuota: true } },
        global: { plugins: [Quasar] },
      });
      expect(wrapper.find('[aid="max-allocation-toggle"]').exists()).toBe(false);
      expect(wrapper.find('[aid="cpu-quota-toggle"]').exists()).toBe(false);
      expect(wrapper.text()).not.toContain("Max GPU allocation");
      expect(wrapper.text()).not.toContain("Preview only");
      expect(wrapper.find('[aid="time-fairshare-toggle"]').exists()).toBe(true);
      expect(wrapper.find(".runtime-section").exists()).toBe(true);
      wrapper.unmount();
    },
  );
  it("reveals historical usage settings when fairshare is enabled and validates negative values", async () => {
    const wrapper = mount(Settings, {
      props: { draft, savedMaxAllocation: true, savedCpuQuota: false, canResetResources: true },
      global: { plugins: [Quasar] },
    });
    expect(wrapper.find(".historical-fields").exists()).toBe(false);
    await wrapper.setProps({
      draft: {
        ...draft,
        timeFairshare: true,
        historicalWeight: -1,
        historicalWindow: { Days: 7, Hours: 0, Minutes: -1 },
      },
    });
    const weight = wrapper.findAllComponents(QInput).find((input) => input.props("label") === "Historical usage weight");
    expect(weight?.props("error")).toBe(true);
    expect(weight?.props("errorMessage")).toBe("Enter a value equal to or above 0.0");
    const minutes = wrapper.findAllComponents(QInput).find((input) => input.props("label") === "Minutes");
    expect(minutes?.props("error")).toBe(true);
    expect(minutes?.props("errorMessage")).toBe("Enter a timeframe of at least 0 minutes");
    await wrapper.setProps({
      draft: { ...draft, timeFairshare: true, historicalWeight: 0, historicalWindow: { Days: 7, Hours: 0, Minutes: 0 } },
    });
    expect(wrapper.find(".historical-fields .q-field--error").exists()).toBe(false);
    wrapper.unmount();
  });
  it("does not convert incomplete numeric input to an empty or default value", async () => {
    const wrapper = mount(Settings, {
      props: { draft: { ...draft, timeFairshare: true, historicalWeight: 1 } },
      global: { plugins: [Quasar] },
    });
    const input = wrapper.find(".historical-weight-input input");
    Object.defineProperty(input.element, "validity", { configurable: true, value: { badInput: true } });
    await input.setValue("");
    expect(wrapper.emitted("patch")?.at(-1)).toEqual([{ historicalWeight: NaN }]);
    Object.defineProperty(input.element, "validity", { configurable: true, value: { badInput: false } });
    await input.setValue("1.5");
    expect(wrapper.emitted("patch")?.at(-1)).toEqual([{ historicalWeight: 1.5 }]);
    wrapper.unmount();
  });
  it("keeps unsaved limits when disabling the capability; backend resets stored limits", async () => {
    const wrapper = mount(Settings, {
      props: { draft, savedMaxAllocation: true, savedCpuQuota: false, canResetResources: true },
      global: { plugins: [Quasar] },
    });
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
    expect(wrapper.emitted("patch")?.[0]).toEqual([{ maxAllocation: false }]);
    expect(draft.departments.a.limit).toBe(8);
    wrapper.unmount();
  });
  it("disables reset-sensitive controls without cluster-wide update scope", () => {
    const wrapper = mount(Settings, {
      props: { draft, savedMaxAllocation: true, savedCpuQuota: false, canResetResources: false },
      global: { plugins: [Quasar] },
    });
    const toggle = wrapper
      .findAllComponents(RunaiToggle)
      .find((toggle) => toggle.props("aid") === "max-allocation-toggle");
    if (!toggle) throw new Error("Missing toggle");
    expect(toggle.props("disable")).toBe(true);
    toggle.vm.$emit("update:model-value", false);
    expect(wrapper.emitted("patch")).toBeUndefined();
    wrapper.unmount();
  });
  it("allows enabling and reverting unsaved capabilities without reset scope", async () => {
    const wrapper = mount(Settings, {
      props: {
        draft: { ...draft, maxAllocation: false },
        canResetResources: false,
        savedMaxAllocation: false,
        savedCpuQuota: false,
      },
      global: { plugins: [Quasar] },
    });
    for (const aid of ["max-allocation-toggle", "cpu-quota-toggle"]) {
      const toggle = wrapper.findAllComponents(RunaiToggle).find((toggle) => toggle.props("aid") === aid);
      if (!toggle) throw new Error("Missing toggle");
      expect(toggle.props("disable")).toBe(false);
      toggle.vm.$emit("update:model-value", true);
    }
    expect(wrapper.emitted("patch")).toHaveLength(2);
    await wrapper.setProps({ draft: { ...draft, cpuQuota: true } });
    const max = wrapper.findAllComponents(RunaiToggle).find((toggle) => toggle.props("aid") === "max-allocation-toggle");
    if (!max) throw new Error("Missing max toggle");
    expect(max.props("disable")).toBe(false);
    max.vm.$emit("update:model-value", false);
    expect(wrapper.emitted("patch")?.at(-1)).toEqual([{ maxAllocation: false }]);
    wrapper.unmount();
  });
  it("never offers an organization policy mutation", () => {
    const wrapper = mount(Settings, {
      props: { draft, savedMaxAllocation: true, savedCpuQuota: false, canResetResources: true },
      global: { plugins: [Quasar] },
    });
    const policy = wrapper
      .findAllComponents(RunaiToggle)
      .find((toggle) => toggle.props("aid") === "project-quota-policy");
    expect(policy).toBeUndefined();
    expect(wrapper.emitted("patch")).toBeUndefined();
    wrapper.unmount();
  });
});
