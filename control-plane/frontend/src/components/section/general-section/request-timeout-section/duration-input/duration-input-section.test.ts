/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import DurationInputSection, { EEmitDurationInput } from "./duration-input-section.vue";
import { errorMessages, errorMessagesUtils } from "@/common/error-message.constant";
import type { IntegerRules } from "@/swagger-models/policy-service-client/api";

describe("duration-input-section.vue", () => {
  const defaultProps = {
    modelValue: 120,
    label: "Test Label",
    tooltipText: "Test Tooltip",
    min: 5, // 5 seconds
    max: 600, // 10 minutes
  };

  const mountCmp = (props = {}, policyRules: IntegerRules | null = null) =>
    mount(DurationInputSection, {
      global: { plugins: [Quasar] },
      props: {
        ...defaultProps,
        ...props,
        policyRules,
      },
    });

  it("renders label and tooltip", () => {
    const wrapper = mountCmp();
    expect(wrapper.text()).toContain(defaultProps.label);
    expect(wrapper.findComponent({ name: "runai-tooltip" }).exists()).toBe(true);
  });

  it("passes modelValue and disables correctly", () => {
    const wrapper = mountCmp({ disable: true });
    const input = wrapper.findComponent({ name: "runai-duration-input" });
    expect(input.props("modelValue")).toBe(defaultProps.modelValue);
    expect(input.props("disable")).toBe(true);
  });

  it("emits update:model-value when updateDurationTime is called", () => {
    const wrapper = mountCmp();
    wrapper.vm.updateDurationTime(300);
    expect(wrapper.emitted()[EEmitDurationInput.UpdateModelValue]).toBeTruthy();
    expect(wrapper.emitted()[EEmitDurationInput.UpdateModelValue][0]).toEqual([300]);
  });

  it("shows q-tooltip if policyRules.canEdit is false", async () => {
    const wrapper = mountCmp({}, { canEdit: false });
    const tooltips = wrapper.findAllComponents({ name: "q-tooltip" });
    expect(tooltips.length).toBe(2);
    expect(tooltips[1].vm.$slots.default()[0].children).toContain(errorMessages.CANT_BE_MODIFIED_POLICY);
  });

  it("computed isDisable returns true if disable or policyRules.canEdit is false", () => {
    const wrapper1 = mountCmp({ disable: true });
    expect(wrapper1.vm.isDisable).toBe(true);
    const wrapper2 = mountCmp({}, { canEdit: false });
    expect(wrapper2.vm.isDisable).toBe(true);
    const wrapper3 = mountCmp({}, { canEdit: true });
    expect(wrapper3.vm.isDisable).toBe(false);
  });

  it("rangeRule returns true if modelValue is null", () => {
    const wrapper = mountCmp({ modelValue: null });
    expect(wrapper.vm.rangeRule()).toBe(true);
  });

  it("rangeRule returns true if modelValue is in range", () => {
    const wrapper = mountCmp({ modelValue: defaultProps.max - 1 });
    expect(wrapper.vm.rangeRule()).toBe(true);
  });

  it("rangeRule returns error if modelValue is out of range", () => {
    // less then min
    const wrapper2 = mountCmp({ modelValue: defaultProps.min - 1 });
    expect(wrapper2.vm.rangeRule()).toBe(wrapper2.vm.timeframeErrorMessage);
    // more then max
    const wrapper3 = mountCmp({ modelValue: defaultProps.max + 1 });
    expect(wrapper3.vm.rangeRule()).toBe(wrapper2.vm.timeframeErrorMessage);
  });

  it("validStepRule returns error if value does not match step", () => {
    const policyRules: IntegerRules = { min: 60, max: 600, step: 30 };
    const wrapper = mountCmp({ modelValue: 120 }, policyRules);
    expect(wrapper.vm.validStepRule(65)).toContain("steps");
    expect(wrapper.vm.validStepRule(90)).toBe(true);
  });

  it("minSecondsAllowedValue and maxSecondsAllowedValue respect policyRules", () => {
    const policyRules: IntegerRules = { min: 100, max: 200 };
    const wrapper = mountCmp({}, policyRules);
    expect(wrapper.vm.minSecondsAllowedValue).toBe(100);
    expect(wrapper.vm.maxSecondsAllowedValue).toBe(200);
  });

  it("minSecondsAllowedValue and maxSecondsAllowedValue falls back to defaultProps when policyRules is undefined", () => {
    const wrapper = mountCmp({});
    expect(wrapper.vm.minSecondsAllowedValue).toBe(defaultProps.min);
    expect(wrapper.vm.maxSecondsAllowedValue).toBe(defaultProps.max);
  });

  it("rules computed includes validStepRule if step is set in policyRules", () => {
    const policyRules: IntegerRules = { min: 60, max: 600, step: 30 };
    const wrapper = mountCmp({}, policyRules);
    expect(wrapper.vm.rules.length).toBe(2);
  });

  it("rules computed only includes rangeRule if no step in policyRules", () => {
    const policyRules: IntegerRules = { min: 60, max: 600 };
    const wrapper = mountCmp({}, policyRules);
    expect(wrapper.vm.rules.length).toBe(1);
  });

  it("timeframeErrorMessage uses errorMessagesUtils", () => {
    const maxInSeconds = 600;
    const wrapper = mountCmp({ min: 60, max: maxInSeconds });
    const maxInMinutes = Math.floor(maxInSeconds / 60);
    const expected = errorMessagesUtils.TIMEFRAME_BETWEEN_MIN_SECONDS_AND_MAX_MINUTES(60, maxInMinutes);
    expect(wrapper.vm.timeframeErrorMessage).toBe(expected);
  });
});
