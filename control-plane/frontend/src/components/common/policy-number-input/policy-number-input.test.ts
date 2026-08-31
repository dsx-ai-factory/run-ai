/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import PolicyNumberInput from "@/components/common/policy-number-input/policy-number-input.vue";
import { describe, it, expect } from "vitest";
import { QInput, Quasar } from "quasar";
import { errorMessages } from "@/common/error-message.constant";
import RunaiTooltipWrapper from "../runai-tooltip-wrapper/runai-tooltip-wrapper.vue";

describe("policy-number-input.vue", () => {
  const createWrapper = (props = {}) => {
    return mount(PolicyNumberInput, {
      global: {
        plugins: [Quasar],
        components: {
          QInput,
          RunaiTooltipWrapper,
        },
      },
      props: {
        modelValue: 10,
        policyRules: {
          canEdit: true,
          min: 1,
          max: 100,
          step: 1,
          required: true,
        },
        label: "Test Label",
        tooltip: "Tooltip text",
        ...props,
      },
    });
  };

  describe("Html elements check", () => {
    it("renders q-input component", () => {
      const wrapper = createWrapper();
      expect(wrapper.findComponent(QInput).exists()).toBe(true);
    });
  });

  describe("Props", () => {
    describe("policyRules", () => {
      it("disables input if canEdit is false", async () => {
        const wrapper = createWrapper({ policyRules: { canEdit: false } });
        const input = wrapper.findComponent(QInput);
        expect(input.props().disable).toBe(true);
      });

      it("adds required rule if policyRules.required is true", async () => {
        const wrapper = createWrapper({ policyRules: { required: true }, rules: [] });
        const input = wrapper.findComponent(QInput);
        expect(input.props().rules.includes(wrapper.vm.requiredRule)).toBe(true);
      });

      it("adds max rule if this.max (either from policy or from prop) holds a value", async () => {
        let wrapper = createWrapper({ policyRules: {}, rules: [], maxValue: 10 });
        expect(wrapper.vm.rulesWithPolicy[0]).toEqual(wrapper.vm.maxRule);

        wrapper = createWrapper({ policyRules: { max: 10 }, rules: [] });
        expect(wrapper.vm.rulesWithPolicy[0]).toEqual(wrapper.vm.maxRule);
      });

      it("adds min rule if this.min (either from policy or from prop) holds a value", async () => {
        let wrapper = createWrapper({ policyRules: {}, rules: [], minValue: 10 });
        expect(wrapper.vm.rulesWithPolicy[0]).toEqual(wrapper.vm.minRule);

        wrapper = createWrapper({ policyRules: { min: 10 }, rules: [] });
        expect(wrapper.vm.rulesWithPolicy[0]).toEqual(wrapper.vm.minRule);
      });

      it("adds step rule if policy rule of step exists", async () => {
        const wrapper = createWrapper({ policyRules: { step: 10 }, rules: [] });
        expect(wrapper.vm.rulesWithPolicy[0]).toEqual(wrapper.vm.stepRule);
      });

      it("displays the correct tooltip text when disabled", () => {
        const wrapper = createWrapper({ policyRules: { canEdit: false } });
        expect(wrapper.vm.disabledTooltipText).toBe(errorMessages.CANT_BE_MODIFIED_POLICY);
      });
    });
    describe("rules", () => {
      it("displays custom rules in rulesWithPolicy computed property", () => {
        const noOddNumbers = (val: number) => val % 2 === 0 || "Cant be odd number";
        const wrapper = createWrapper({ rules: [noOddNumbers] });
        const input = wrapper.findComponent(QInput);
        expect(input.props().rules.includes(noOddNumbers)).toBe(true);
      });
    });
    describe("tooltip", () => {
      it("showExplanatoryTooltip is true when tooltip prop is provided", () => {
        const wrapper = createWrapper({ tooltip: "Tooltip text" });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(true);
      });
    });
  });

  describe("emits", () => {
    it("emits update:model-value event when value is updated", async () => {
      const wrapper = createWrapper();
      const input = wrapper.findComponent(QInput);
      await input.vm.$emit("update:modelValue", 50); // QInput (Quasar) declares emits in camelCase — must match to avoid Vue warn
      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([50]);
    });
  });

  describe("Computed", () => {
    describe("valueErrorMessage", () => {
      it("returns an empty string when no min, max, or step are defined", () => {
        const wrapper = createWrapper({ policyRules: null, minValue: undefined, maxValue: undefined, step: undefined });
        expect(wrapper.vm.valueErrorMessage).toBe("");
      });

      it("returns a combo message when min, max, and step are defined", () => {
        const wrapper = createWrapper({ minValue: 5, maxValue: 50, policyRules: { step: 5 } });
        const expectedMessage = errorMessages.ENTER_A_VALUE_BETWEEN_WITH_STEP.replace("${min}", "5")
          .replace("${max}", "50")
          .replace("${step}", "5");
        expect(wrapper.vm.valueErrorMessage).toBe(expectedMessage);
      });

      it("returns the min and max message when both are defined", () => {
        const wrapper = createWrapper({ policyRules: null, minValue: 10, maxValue: 40 });
        const expectedMessage = errorMessages.ENTER_A_VALUE_BETWEEN.replace("${min}", "10").replace("${max}", "40");
        expect(wrapper.vm.valueErrorMessage).toBe(expectedMessage);
      });

      it("returns the min and step message when both are defined", () => {
        const wrapper = createWrapper({ minValue: 2, policyRules: { step: 2 } });
        const expectedMessage = errorMessages.VALUE_MUST_FOLLOW_STEPS_STARTING_MIN.replace("${min}", "2").replace(
          "${step}",
          "2",
        );
        expect(wrapper.vm.valueErrorMessage).toBe(expectedMessage);
      });

      it("returns an empty string when only min is defined", () => {
        const wrapper = createWrapper({ policyRules: null, minValue: 5 });
        expect(wrapper.vm.valueErrorMessage).toBe("");
      });

      it("returns an empty string when only max is defined", () => {
        const wrapper = createWrapper({ policyRules: null, maxValue: 20 });
        expect(wrapper.vm.valueErrorMessage).toBe("");
      });

      it("returns an empty string when only step is defined", () => {
        const wrapper = createWrapper({ policyRules: { step: 5 } });
        expect(wrapper.vm.valueErrorMessage).toBe("");
      });
    });

    describe("max", () => {
      it("should return maxValue if policyRules is undefined", () => {
        const wrapper = createWrapper({ policyRules: undefined, maxValue: 10 });
        expect(wrapper.vm.max).toBe(10);
      });

      it("should return maxValue if policyRules is null", () => {
        const wrapper = createWrapper({ policyRules: null, maxValue: 10 });
        expect(wrapper.vm.max).toBe(10);
      });

      it("should return policyRules.max if it is a number", () => {
        const wrapper = createWrapper({ policyRules: { max: 15 }, maxValue: 10 });
        expect(wrapper.vm.max).toBe(15);
      });

      it("should return maxValue if policyRules.max is undefined", () => {
        const wrapper = createWrapper({ policyRules: { max: undefined }, maxValue: 10 });
        expect(wrapper.vm.max).toBe(10);
      });

      it("should return maxValue if policyRules.max is null", () => {
        const wrapper = createWrapper({ policyRules: { max: null }, maxValue: 10 });
        expect(wrapper.vm.max).toBe(10);
      });

      it("should return maxValue if policyRules.max is string", () => {
        const wrapper = createWrapper({ policyRules: { max: "str" }, maxValue: 10 });
        expect(wrapper.vm.max).toBe(10);
      });
    });

    describe("min", () => {
      it("should return minValue if policyRules is undefined", () => {
        const wrapper = createWrapper({ policyRules: undefined, minValue: 10 });
        expect(wrapper.vm.min).toBe(10);
      });

      it("should return minValue if policyRules is null", () => {
        const wrapper = createWrapper({ policyRules: null, minValue: 10 });
        expect(wrapper.vm.min).toBe(10);
      });

      it("should return policyRules.min if it is a number", () => {
        const wrapper = createWrapper({ policyRules: { min: 15 }, minValue: 10 });
        expect(wrapper.vm.min).toBe(15);
      });

      it("should return minValue if policyRules.min is undefined", () => {
        const wrapper = createWrapper({ policyRules: { min: undefined }, minValue: 10 });
        expect(wrapper.vm.min).toBe(10);
      });

      it("should return minValue if policyRules.min is null", () => {
        const wrapper = createWrapper({ policyRules: { min: null }, minValue: 10 });
        expect(wrapper.vm.min).toBe(10);
      });

      it("should return minValue if policyRules.min is string", () => {
        const wrapper = createWrapper({ policyRules: { min: "str" }, minValue: 10 });
        expect(wrapper.vm.min).toBe(10);
      });
    });

    describe("showExplanatoryTooltip", () => {
      it("should return false if policyRules.canEdit is false", () => {
        const wrapper = createWrapper({ policyRules: { canEdit: false } });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(false);
      });

      it("should return true if tooltip is defined and policyRules.canEdit true", () => {
        const wrapper = createWrapper({ policyRules: { canEdit: true }, tooltip: "Sample tooltip" });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(true);
      });

      it("should return false if tooltip is undefined and policyRules.canEdit true", () => {
        const wrapper = createWrapper({ policyRules: { canEdit: true }, tooltip: undefined });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(false);
      });

      it("should return true if tooltip is defined and policyRules is undefined", () => {
        const wrapper = createWrapper({ policyRules: undefined, tooltip: "Sample tooltip" });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(true);
      });

      it("should return false if both tooltip and policyRules are undefined", () => {
        const wrapper = createWrapper({ policyRules: undefined, tooltip: undefined });
        expect(wrapper.vm.showExplanatoryTooltip).toBe(false);
      });
    });
  });

  describe("general validations", () => {
    it("validates the required rule correctly", () => {
      const wrapper = createWrapper();

      expect(wrapper.vm.requiredRule(null as unknown as number)).toBe(
        "This field is required based on the policy your administrator set",
      );
      expect(wrapper.vm.requiredRule(10)).toBe(true);
    });

    it("validates the min rule correctly", () => {
      const wrapper = createWrapper({ policyRules: { min: 5 } });
      expect(wrapper.vm.minRule(3)).toBe(`${errorMessages.ENTER_A_VALUE_ABOVE_MIN}`.replace("${min}", "5"));
      expect(wrapper.vm.minRule(6)).toBe(true);
    });

    it("validates the max rule correctly", () => {
      const wrapper = createWrapper({ policyRules: { max: 50 } });
      expect(wrapper.vm.maxRule(51)).toBe(`${errorMessages.ENTER_A_VALUE_BELOW_MAX}`.replace("${max}", "50"));
      expect(wrapper.vm.maxRule(30)).toBe(true);
    });

    it("validates the step rule correctly", () => {
      const wrapper = createWrapper({ policyRules: { step: 5 } });
      expect(wrapper.vm.stepRule(7)).toBe("Enter a value that follows steps of 5");
      expect(wrapper.vm.stepRule(10)).toBe(true);
    });
  });
});
