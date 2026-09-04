/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import WorkerElasticitySection from "./worker-elasticity-section.vue";
import { Quasar } from "quasar";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { IntegerRules } from "@/swagger-models/assets-service-client";

const defaultPolicyRulesMin: IntegerRules = { min: 1, max: 10 };
const defaultPolicyRulesMax: IntegerRules = { min: 1, max: 10000 };

const createWrapper = (props = {}) => {
  return mount(WorkerElasticitySection, {
    global: { plugins: [Quasar] },
    props: {
      minWorkers: undefined,
      maxWorkers: undefined,
      numWorkers: 1,
      sectionDisabled: false,
      policyRulesMin: defaultPolicyRulesMin,
      policyRulesMax: defaultPolicyRulesMax,
      ...props,
    },
  });
};

describe("worker-elasticity-section.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkerElasticitySection>>;

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Html elements check", () => {
    it("renders worker elasticity section with title and inputs", () => {
      wrapper = createWrapper();
      expect(wrapper.html()).toContain("Worker elasticity");
      expect(wrapper.html()).toContain("Set the minimum and maximum workers for your workload");
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs).toHaveLength(2);
    });

    it("renders minimum and maximum labels", () => {
      wrapper = createWrapper();
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs[0].props("label")).toBe("Minimum");
      expect(inputs[1].props("label")).toBe("Maximum");
    });

    it("renders tooltip with correct text", () => {
      wrapper = createWrapper();
      const tooltip = wrapper.findComponent({ name: "runai-tooltip" });
      expect(tooltip.exists()).toBe(true);
      expect(tooltip.props("tooltipText")).toContain(
        "The workload can use more or fewer workers within this range based on available resources",
      );
    });

    it("renders folding sub-section", () => {
      wrapper = createWrapper();
      const foldingSection = wrapper.findComponent({ name: "runai-folding-sub-section" });
      expect(foldingSection.exists()).toBe(true);
      expect(foldingSection.props("label")).toBe("Worker elasticity");
    });
  });

  describe("Props", () => {
    describe("minWorkers", () => {
      it("handles undefined minWorkers", () => {
        wrapper = createWrapper({ minWorkers: undefined });
        expect(wrapper.vm.minWorkersValue).toBe(null);
      });

      it("handles null minWorkers", () => {
        wrapper = createWrapper({ minWorkers: null });
        expect(wrapper.vm.minWorkersValue).toBe(null);
      });

      it("handles numeric minWorkers", () => {
        wrapper = createWrapper({ minWorkers: 2 });
        expect(wrapper.vm.minWorkersValue).toBe(2);
      });
    });

    describe("maxWorkers", () => {
      it("handles undefined maxWorkers", () => {
        wrapper = createWrapper({ maxWorkers: undefined });
        expect(wrapper.vm.maxWorkersValue).toBe(null);
      });

      it("handles null maxWorkers", () => {
        wrapper = createWrapper({ maxWorkers: null });
        expect(wrapper.vm.maxWorkersValue).toBe(null);
      });

      it("handles numeric maxWorkers", () => {
        wrapper = createWrapper({ maxWorkers: 5 });
        expect(wrapper.vm.maxWorkersValue).toBe(5);
      });
    });

    describe("numWorkers", () => {
      it("defaults to 1 when undefined", () => {
        wrapper = createWrapper({ numWorkers: undefined });
        expect(wrapper.vm.numWorkersValue).toBe(1);
      });

      it("uses provided numWorkers value", () => {
        wrapper = createWrapper({ numWorkers: 3 });
        expect(wrapper.vm.numWorkersValue).toBe(3);
      });
    });

    describe("sectionDisabled", () => {
      it("disables inputs when sectionDisabled is true", () => {
        wrapper = createWrapper({ sectionDisabled: true });
        const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
        expect(inputs[0].props("disable")).toBe(true);
        expect(inputs[1].props("disable")).toBe(true);
      });

      it("enables inputs when sectionDisabled is false", () => {
        wrapper = createWrapper({ sectionDisabled: false });
        const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
        expect(inputs[0].props("disable")).toBe(false);
        expect(inputs[1].props("disable")).toBe(false);
      });
    });

    describe("policyRules", () => {
      it("passes policyRulesMin to minimum input", () => {
        const customRules = { min: 2, max: 20 };
        wrapper = createWrapper({ policyRulesMin: customRules });
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        expect(minInput.props("policyRules")).toEqual(customRules);
      });

      it("passes policyRulesMax to maximum input", () => {
        const customRules = { min: 1, max: 5000 };
        wrapper = createWrapper({ policyRulesMax: customRules });
        const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];
        expect(maxInput.props("policyRules")).toEqual(customRules);
      });
    });
  });

  describe("Computed properties", () => {
    describe("minValueForMaxWorkers", () => {
      it("returns 1 when minWorkers is null", () => {
        wrapper = createWrapper({ minWorkers: null });
        expect(wrapper.vm.minValueForMaxWorkers).toBe(1);
      });

      it("returns minWorkers value when set", () => {
        wrapper = createWrapper({ minWorkers: 3 });
        expect(wrapper.vm.minValueForMaxWorkers).toBe(3);
      });
    });

    describe("minWorkersRules", () => {
      it("returns validation rules array", () => {
        wrapper = createWrapper();
        const rules = wrapper.vm.minWorkersRules;
        expect(Array.isArray(rules)).toBe(true);
        expect(rules.length).toBeGreaterThanOrEqual(1);
        expect(typeof rules[0]).toBe("function");
      });

      it("validates null/undefined as valid", () => {
        wrapper = createWrapper();
        const rule = wrapper.vm.minWorkersRules[0];
        expect(rule(null)).toBe(true);
        expect(rule(undefined)).toBe(true);
      });

      it("validates value over 0", () => {
        wrapper = createWrapper({ numWorkers: 5 });
        const rule = wrapper.vm.minWorkersRules[0];
        expect(rule(0)).toContain("Enter a value over 0");
        expect(rule(-1)).toContain("Enter a value over 0");
        expect(rule(1)).toBe(true);
      });

      it("validates value up to numWorkers", () => {
        wrapper = createWrapper({ numWorkers: 5 });
        const rule = wrapper.vm.minWorkersRules[0];
        expect(rule(6)).toContain("up to the number of workers");
        expect(rule(5)).toBe(true);
        expect(rule(3)).toBe(true);
      });

      it("updates validation when numWorkers changes", async () => {
        wrapper = createWrapper({ numWorkers: 3 });
        let rule = wrapper.vm.minWorkersRules[0];
        expect(rule(4)).toContain("up to the number of workers");

        await wrapper.setProps({ numWorkers: 10 });
        rule = wrapper.vm.minWorkersRules[0];
        expect(rule(4)).toBe(true);
        expect(rule(11)).toContain("up to the number of workers");
      });
    });

    describe("maxWorkersRules", () => {
      it("returns validation rules array", () => {
        wrapper = createWrapper();
        const rules = wrapper.vm.maxWorkersRules;
        expect(Array.isArray(rules)).toBe(true);
        expect(rules.length).toBeGreaterThanOrEqual(1);
        expect(typeof rules[0]).toBe("function");
      });

      it("validates null/undefined as valid", () => {
        wrapper = createWrapper();
        const rule = wrapper.vm.maxWorkersRules[0];
        expect(rule(null)).toBe(true);
        expect(rule(undefined)).toBe(true);
      });

      it("validates value equal to or above numWorkers", () => {
        wrapper = createWrapper({ numWorkers: 5 });
        const rule = wrapper.vm.maxWorkersRules[0];
        expect(rule(4)).toContain("Enter a value equal to or above the number of workers");
        expect(rule(5)).toBe(true);
        expect(rule(6)).toBe(true);
      });

      it("validates value up to 10,000", () => {
        wrapper = createWrapper({ numWorkers: 5 });
        const rule = wrapper.vm.maxWorkersRules[0];
        expect(rule(10001)).toContain("Enter a value equal to or above the number of workers, up to 10,000");
        expect(rule(10000)).toBe(true);
        expect(rule(9999)).toBe(true);
      });

      it("updates validation when numWorkers changes", async () => {
        wrapper = createWrapper({ numWorkers: 3 });
        let rule = wrapper.vm.maxWorkersRules[0];
        expect(rule(2)).toContain("equal to or above the number of workers");

        await wrapper.setProps({ numWorkers: 1 });
        rule = wrapper.vm.maxWorkersRules[0];
        expect(rule(2)).toBe(true);
      });
    });
  });

  describe("Emits", () => {
    describe("update:min-workers", () => {
      it("emits update:min-workers when minimum input changes", async () => {
        wrapper = createWrapper();
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        await minInput.vm.$emit("update:model-value", "3");
        expect(wrapper.emitted("update:min-workers")).toBeTruthy();
        expect(wrapper.emitted("update:min-workers")?.[0]).toEqual([3]);
      });

      it("emits null when minimum input is cleared", async () => {
        wrapper = createWrapper({ minWorkers: 2 });
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        await minInput.vm.$emit("update:model-value", "");
        expect(wrapper.emitted("update:min-workers")).toBeTruthy();
        expect(wrapper.emitted("update:min-workers")?.[0]).toEqual([null]);
      });

      it("auto-adjusts max when min exceeds max", async () => {
        wrapper = createWrapper({ minWorkers: 2, maxWorkers: 5 });
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        await minInput.vm.$emit("update:model-value", "7");

        // Should emit both min and max updates
        expect(wrapper.emitted("update:min-workers")).toBeTruthy();
        expect(wrapper.emitted("update:max-workers")).toBeTruthy();
        expect(wrapper.emitted("update:min-workers")?.[0]).toEqual([7]);
        expect(wrapper.emitted("update:max-workers")?.[0]).toEqual([7]);
      });

      it("does not auto-adjust max when min is less than max", async () => {
        wrapper = createWrapper({ minWorkers: 2, maxWorkers: 5 });
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        await minInput.vm.$emit("update:model-value", "3");

        expect(wrapper.emitted("update:min-workers")).toBeTruthy();
        expect(wrapper.emitted("update:max-workers")).toBeFalsy();
      });

      it("auto-adjusts max when max is null", async () => {
        wrapper = createWrapper({ minWorkers: 2, maxWorkers: null });
        const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
        await minInput.vm.$emit("update:model-value", "5");

        expect(wrapper.emitted("update:max-workers")).toBeTruthy();
        expect(wrapper.emitted("update:max-workers")?.[0]).toEqual([5]);
      });
    });

    describe("update:max-workers", () => {
      it("emits update:max-workers when maximum input changes", async () => {
        wrapper = createWrapper();
        const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];
        await maxInput.vm.$emit("update:model-value", "10");
        expect(wrapper.emitted("update:max-workers")).toBeTruthy();
        expect(wrapper.emitted("update:max-workers")?.[0]).toEqual([10]);
      });

      it("emits null when maximum input is cleared", async () => {
        wrapper = createWrapper({ maxWorkers: 5 });
        const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];
        await maxInput.vm.$emit("update:model-value", "");
        expect(wrapper.emitted("update:max-workers")).toBeTruthy();
        expect(wrapper.emitted("update:max-workers")?.[0]).toEqual([null]);
      });
    });
  });

  describe("Input configuration", () => {
    it("sets isInteger prop to true for both inputs", () => {
      wrapper = createWrapper();
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs[0].props("isInteger")).toBe(true);
      expect(inputs[1].props("isInteger")).toBe(true);
    });

    it("sets reactive-rules prop to true for both inputs", () => {
      wrapper = createWrapper();
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs[0].props("reactiveRules")).toBe(true);
      expect(inputs[1].props("reactiveRules")).toBe(true);
    });

    it("sets stack-label prop to true for both inputs", () => {
      wrapper = createWrapper();
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs[0].props("stackLabel")).toBe(true);
      expect(inputs[1].props("stackLabel")).toBe(true);
    });

    it("sets correct min-value for minimum input", () => {
      wrapper = createWrapper();
      const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
      expect(minInput.props("minValue")).toBe(1);
    });

    it("sets correct min-value for maximum input based on minWorkers", async () => {
      wrapper = createWrapper({ minWorkers: 3 });
      const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];
      expect(maxInput.props("minValue")).toBe(3);

      await wrapper.setProps({ minWorkers: null });
      expect(maxInput.props("minValue")).toBe(1);
    });

    it("sets max-value to 10000 for maximum input", () => {
      wrapper = createWrapper();
      const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];
      expect(maxInput.props("maxValue")).toBe(10000);
    });

    it("sets correct aids for accessibility", () => {
      wrapper = createWrapper();
      const inputs = wrapper.findAllComponents({ name: "policy-number-input" });
      expect(inputs[0].props("aid")).toBe("min-workers-input");
      expect(inputs[1].props("aid")).toBe("max-workers-input");
    });
  });

  describe("Edge cases", () => {
    it("handles zero numWorkers gracefully", () => {
      wrapper = createWrapper({ numWorkers: 0 });
      expect(wrapper.vm.numWorkersValue).toBe(1); // Falls back to 1
    });

    it("handles very large numWorkers values", () => {
      wrapper = createWrapper({ numWorkers: 9999 });
      const rule = wrapper.vm.minWorkersRules[0];
      expect(rule(10000)).toContain("up to the number of workers");
    });

    it("handles min equals max scenario", async () => {
      wrapper = createWrapper({ minWorkers: 5, maxWorkers: 5 });
      const minInput = wrapper.findAllComponents({ name: "policy-number-input" })[0];
      const maxInput = wrapper.findAllComponents({ name: "policy-number-input" })[1];

      expect(minInput.props("modelValue")).toBe(5);
      expect(maxInput.props("modelValue")).toBe(5);
    });

    it("handles rapid updates to numWorkers", async () => {
      wrapper = createWrapper({ numWorkers: 1 });

      await wrapper.setProps({ numWorkers: 5 });
      await wrapper.setProps({ numWorkers: 10 });
      await wrapper.setProps({ numWorkers: 3 });

      const rule = wrapper.vm.minWorkersRules[0];
      expect(rule(4)).toContain("up to the number of workers");
    });
  });
});
