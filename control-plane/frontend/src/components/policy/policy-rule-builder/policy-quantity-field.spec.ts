/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, it, expect, vi } from "vitest";
import PolicyQuantityField from "./policy-quantity-field.vue";
import { K8S_QUANTITY_UNIT_OPTIONS } from "./policy-rule-builder.constant";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiSelect } from "@/components/common/runai-select";
import type { TEnforcementValueRule } from "@/utils/gvk-policy.util";

function mountField(modelValue: string, rules: TEnforcementValueRule[] = []) {
  return shallowMount(PolicyQuantityField, {
    props: { modelValue, rules, unitOptions: K8S_QUANTITY_UNIT_OPTIONS, aid: "q", label: "Min" },
    global: { plugins: [Quasar] },
  });
}

function lastEmit(wrapper: ReturnType<typeof mountField>): unknown {
  const events = wrapper.emitted("update:model-value");
  return events?.[events.length - 1]?.[0];
}

describe("policy-quantity-field.vue", () => {
  it("splits a stored quantity into its number and unit (edit case)", () => {
    const wrapper = mountField("1Gi");
    expect(wrapper.findComponent(PolicyNumberInput).props("modelValue")).toBe(1);
    expect(wrapper.findComponent(RunaiSelect).props("modelValue")).toBe("Gi");
  });

  it("shows the None unit (sentinel) for a unitless quantity", () => {
    const wrapper = mountField("2");
    expect(wrapper.findComponent(PolicyNumberInput).props("modelValue")).toBe(2);
    expect(wrapper.findComponent(RunaiSelect).props("modelValue")).toBe("none");
  });

  it("emits the combined string when the number changes", () => {
    const wrapper = mountField("1Gi");
    wrapper.findComponent(PolicyNumberInput).vm.$emit("update:model-value", 2);
    expect(lastEmit(wrapper)).toBe("2Gi");
  });

  it("keeps the number when only the unit changes", () => {
    const wrapper = mountField("1Gi");
    wrapper.findComponent(RunaiSelect).vm.$emit("update:model-value", "Mi");
    expect(lastEmit(wrapper)).toBe("1Mi");
  });

  it("drops the unit suffix when the unit is set to None (back to unitless)", () => {
    const wrapper = mountField("1Gi");
    wrapper.findComponent(RunaiSelect).vm.$emit("update:model-value", "none");
    expect(lastEmit(wrapper)).toBe("1");
  });

  it("emits a bare number when the unit is empty", () => {
    const wrapper = mountField("");
    wrapper.findComponent(PolicyNumberInput).vm.$emit("update:model-value", 2);
    expect(lastEmit(wrapper)).toBe("2");
  });

  it("emits an empty string when the number is cleared", () => {
    const wrapper = mountField("5Gi");
    wrapper.findComponent(PolicyNumberInput).vm.$emit("update:model-value", null);
    expect(lastEmit(wrapper)).toBe("");
  });

  it("runs the caller's rules against the recombined quantity string", () => {
    const rule = vi.fn().mockReturnValue("bad");
    const wrapper = mountField("1Gi", [rule]);
    const wrapped = wrapper.findComponent(PolicyNumberInput).props("rules") as Array<() => unknown>;
    expect(wrapped[0]()).toBe("bad");
    expect(rule).toHaveBeenCalledWith("1Gi");
  });

  it("re-syncs the inputs when the value is replaced externally", async () => {
    const wrapper = mountField("5Gi");
    await wrapper.setProps({ modelValue: "" });
    expect(wrapper.findComponent(PolicyNumberInput).props("modelValue")).toBe(null);
    expect(wrapper.findComponent(RunaiSelect).props("modelValue")).toBe("none");
  });
});
