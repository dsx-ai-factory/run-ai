/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { VueWrapper } from "@vue/test-utils";
import TerminationGracePeriodSection from "./termination-grace-period-section.vue";
import { RunaiDurationInput } from "@/components/common/runai-duration-input";
import type { IntegerRules } from "@/swagger-models/assets-service-client";

// Mock policy rules for testing
const mockPolicyRules = {
  editable: {
    canEdit: true,
    min: 30,
    max: 300,
    step: 10,
  } as IntegerRules,
  nonEditable: {
    canEdit: false,
    min: 30,
    max: 300,
    step: 10,
  } as IntegerRules,
  withConstraints: {
    canEdit: true,
    min: 60,
    max: 180,
    step: 15,
  } as IntegerRules,
};

describe("termination-grace-period-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminationGracePeriodSection>>;

  const createWrapper = (props = {}) => {
    return mount(TerminationGracePeriodSection, {
      global: { plugins: [Quasar] },
      props: { modelValue: null, policyRules: null, readOnly: false, ...props },
    });
  };

  beforeEach(() => {
    wrapper = createWrapper();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("Component Structure and Props", () => {
    it("Should handle model value and emit events correctly", async () => {
      await wrapper.setProps({ modelValue: 120 });
      expect(wrapper.findComponent(RunaiDurationInput).props("modelValue")).toBe(120);

      const input = wrapper.findComponent(RunaiDurationInput);
      await input.vm.$emit("update:model-value", 90);
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([90]);
    });
  });

  describe("Policy Rules and Restrictions", () => {
    it("Should handle canEdit policy restrictions correctly", async () => {
      await wrapper.setProps({ policyRules: mockPolicyRules.editable });
      expect(wrapper.vm.isDisabled).toBe(false);
      expect(wrapper.vm.isDisabledByPolicy).toBe(false);

      await wrapper.setProps({ policyRules: mockPolicyRules.nonEditable });
      expect(wrapper.vm.isDisabled).toBe(true);
      expect(wrapper.vm.isDisabledByPolicy).toBe(true);
      expect(wrapper.findComponent({ name: "q-tooltip" }).exists()).toBe(true);
    });

    it("Should combine readOnly and policy restrictions properly", async () => {
      await wrapper.setProps({ readOnly: true, policyRules: mockPolicyRules.editable });
      expect(wrapper.vm.isDisabled).toBe(true);
      expect(wrapper.vm.isDisabledByPolicy).toBe(false);

      await wrapper.setProps({ readOnly: false, policyRules: mockPolicyRules.nonEditable });
      expect(wrapper.vm.isDisabled).toBe(true);
      expect(wrapper.vm.isDisabledByPolicy).toBe(true);
    });
  });

  describe("Validation Rules and Behavior", () => {
    it("Should not apply any validation when no policy rules exist", async () => {
      await wrapper.setProps({ modelValue: null });
      expect(wrapper.vm.validGracePeriod()).toBe(true);

      await wrapper.setProps({ modelValue: 250 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);

      await wrapper.setProps({ modelValue: 350 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);

      await wrapper.setProps({ modelValue: 10000 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);
    });

    it("Should apply policy constraints correctly", async () => {
      await wrapper.setProps({ policyRules: mockPolicyRules.withConstraints });

      await wrapper.setProps({ modelValue: 30 });
      expect(wrapper.vm.validGracePeriod()).not.toBe(true);

      await wrapper.setProps({ modelValue: 90 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);

      expect(wrapper.vm.notSmallerThanRule(30)).not.toBe(true);
      expect(wrapper.vm.notSmallerThanRule(90)).toBe(true);

      expect(wrapper.vm.validStepRule(70)).not.toBe(true);
      expect(wrapper.vm.validStepRule(75)).toBe(true);
    });

    it("Should respect policy max value higher than default 300 seconds", async () => {
      const highMaxPolicy: IntegerRules = {
        canEdit: true,
        min: 0,
        max: 259200, // 4 days in seconds
      };

      await wrapper.setProps({ policyRules: highMaxPolicy });

      await wrapper.setProps({ modelValue: 7 * 60 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);

      await wrapper.setProps({ modelValue: 259200 });
      expect(wrapper.vm.validGracePeriod()).toBe(true);
      expect(wrapper.vm.notLargerThanRule(259200)).toBe(true);

      await wrapper.setProps({ modelValue: 259201 });
      expect(wrapper.vm.validGracePeriod()).not.toBe(true);
      expect(wrapper.vm.notLargerThanRule(259201)).not.toBe(true);
    });
  });

  describe("Edge Cases and Boundary Values", () => {
    it("Should handle various policy rule scenarios", async () => {
      await wrapper.setProps({ policyRules: undefined });
      expect(wrapper.vm.isDisabled).toBe(false);

      await wrapper.setProps({ policyRules: null });
      expect(wrapper.vm.isDisabled).toBe(false);

      await wrapper.setProps({ policyRules: { min: 30, max: 300 } });
      expect(wrapper.vm.isDisabledByPolicy).toBe(false);
    });

    it("Should handle boundary values correctly", async () => {
      await wrapper.setProps({
        policyRules: mockPolicyRules.withConstraints,
        modelValue: 0,
      });
      expect(wrapper.findComponent(RunaiDurationInput).props("modelValue")).toBe(0);

      expect(wrapper.vm.notSmallerThanRule(60)).toBe(true);
      expect(wrapper.vm.notLargerThanRule(180)).toBe(true);
    });
  });
});
