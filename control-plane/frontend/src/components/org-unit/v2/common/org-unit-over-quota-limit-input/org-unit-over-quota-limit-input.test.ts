/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import OrgUnitOverQuotaLimitInput from "./org-unit-over-quota-limit-input.vue";
import { EQuotaEntity } from "@/models/resource.model";
import { EQuotaInputState } from "@/models/org-unit.model";

// Mock the QuotaInput component to capture props passed to it
vi.mock("@/components/org-unit/v2/common/quota-input", () => ({
  QuotaInput: {
    name: "QuotaInput",
    template: `
      <div class="mock-quota-input" :data-state="state" :data-message="message">
        <input 
          type="number" 
          :value="modelValue" 
          @input="$emit('update:model-value', $event.target.value === '' ? null : Number($event.target.value))"
        />
        <span class="state">{{ state }}</span>
        <span class="message">{{ message }}</span>
      </div>
    `,
    props: ["modelValue", "state", "message", "autofocus", "label", "rules", "placeholder"],
    emits: ["update:model-value"],
  },
}));

describe("OrgUnitOverQuotaLimitInput", () => {
  let wrapper: VueWrapper<InstanceType<typeof OrgUnitOverQuotaLimitInput>>;

  const mountComponent = (props: { modelValue: number | null; deserved: number; entityType?: EQuotaEntity }) => {
    wrapper = mount(OrgUnitOverQuotaLimitInput, {
      global: {
        plugins: [Quasar],
      },
      props: {
        entityType: EQuotaEntity.project,
        ...props,
      },
    });
    return wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders the component correctly", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
      });
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.findComponent({ name: "QuotaInput" }).exists()).toBe(true);
    });

    it("passes modelValue to QuotaInput", () => {
      const wrapper = mountComponent({
        modelValue: 15,
        deserved: 10,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("modelValue")).toBe(15);
    });

    it("displays correct label for project entity", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
        entityType: EQuotaEntity.project,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("label")).toBe("Max GPU allocation");
    });

    it("displays correct label for department entity", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
        entityType: EQuotaEntity.department,
      });
      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      expect(quotaInput.props("label")).toBe("Max GPU allocation");
    });
  });

  describe("Validation - Valid States", () => {
    it("shows default state when value is above deserved", () => {
      const wrapper = mountComponent({
        modelValue: 20,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
      expect(wrapper.vm.quotaMessage).toBe("");
    });

    it("shows error state when value equals deserved (must be strictly greater)", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.quotaMessage).toContain("must be greater than the assigned quota");
    });

    it("shows default state when value is null (unlimited)", () => {
      const wrapper = mountComponent({
        modelValue: null,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
      expect(wrapper.vm.quotaMessage).toBe("");
    });

    it("shows error state when value is 0 and deserved is 0 (must be strictly greater)", () => {
      const wrapper = mountComponent({
        modelValue: 0,
        deserved: 0,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.quotaMessage).toContain("must be greater than the assigned quota");
    });
  });

  describe("Validation - Error States", () => {
    it("shows error state when value is negative", () => {
      const wrapper = mountComponent({
        modelValue: -5,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.isNegative).toBe(true);
    });

    it("shows error message when value is negative (and below deserved takes precedence)", () => {
      // Note: When value is negative AND below deserved, the "must be greater" message is shown
      // because isBelowDeserved is checked first in quotaMessage
      const wrapper = mountComponent({
        modelValue: -1,
        deserved: 10,
      });
      // Since -1 <= 10, isBelowDeserved is true, so it shows the below-quota message
      expect(wrapper.vm.quotaMessage).toContain("must be greater than the assigned quota");
    });

    it("shows negative error message when value is negative and deserved is 0", () => {
      // When deserved is 0, a negative value triggers isBelowDeserved (since -1 <= 0)
      const wrapper = mountComponent({
        modelValue: -1,
        deserved: 0,
      });
      // Since -1 <= 0, isBelowDeserved is still true
      expect(wrapper.vm.quotaMessage).toContain("must be greater than the assigned quota");
    });

    it("shows 'must be 0 or greater' message when negative but above deserved", () => {
      // Edge case: if deserved is also negative, a less negative value
      // would make isBelowDeserved false but isNegative true
      const wrapper = mountComponent({
        modelValue: -5,
        deserved: -10, // -5 > -10, so isBelowDeserved is false
      });
      // isNegative is true (since -5 < 0), isBelowDeserved is false
      expect(wrapper.vm.isBelowDeserved).toBe(false);
      expect(wrapper.vm.isNegative).toBe(true);
      expect(wrapper.vm.quotaMessage).toContain("must be 0 or greater");
    });

    it("shows error state when value is below deserved", () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });

    it("shows below-deserved error message with correct threshold", () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 15,
      });
      expect(wrapper.vm.quotaMessage).toContain("must be greater than the assigned quota");
      expect(wrapper.vm.quotaMessage).toContain("15 GPUs");
    });
  });

  describe("Validation Rules", () => {
    it("returns true when state is not error", () => {
      const wrapper = mountComponent({
        modelValue: 20,
        deserved: 10,
      });
      const rules = wrapper.vm.rules;
      expect(rules).toHaveLength(1);
      expect(rules[0]()).toBe(true);
    });

    it("returns empty string (error) when state is error", () => {
      const wrapper = mountComponent({
        modelValue: -5,
        deserved: 10,
      });
      const rules = wrapper.vm.rules;
      expect(rules).toHaveLength(1);
      expect(rules[0]()).toBe("");
    });

    it("returns empty string (error) when below deserved", () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });
      const rules = wrapper.vm.rules;
      expect(rules[0]()).toBe("");
    });
  });

  describe("Value Emission", () => {
    it("emits update:model-value when value changes", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
      });

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      await quotaInput.vm.$emit("update:model-value", 20);

      expect(wrapper.emitted("update:model-value")).toBeTruthy();
      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([20]);
    });

    it("emits null when cleared (unlimited)", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
      });

      const quotaInput = wrapper.findComponent({ name: "QuotaInput" });
      await quotaInput.vm.$emit("update:model-value", null);

      expect(wrapper.emitted("update:model-value")?.[0]).toEqual([null]);
    });
  });

  describe("Computed Properties", () => {
    it("isBelowDeserved is false when value is null", () => {
      const wrapper = mountComponent({
        modelValue: null,
        deserved: 100,
      });
      expect(wrapper.vm.isBelowDeserved).toBe(false);
    });

    it("isBelowDeserved is true when value is less than deserved", () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });

    it("isBelowDeserved is true when value equals deserved (must be strictly greater)", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 10,
      });
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });

    it("isBelowDeserved is false when value is greater than deserved", () => {
      const wrapper = mountComponent({
        modelValue: 20,
        deserved: 10,
      });
      expect(wrapper.vm.isBelowDeserved).toBe(false);
    });

    it("isNegative includes below deserved check", () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });
      // Based on current implementation, isNegative includes isBelowDeserved
      expect(wrapper.vm.isNegative).toBe(true);
    });

    it("isNegative is true for actual negative values", () => {
      const wrapper = mountComponent({
        modelValue: -1,
        deserved: 0,
      });
      expect(wrapper.vm.isNegative).toBe(true);
    });
  });

  describe("Reactivity", () => {
    it("updates validation state when modelValue changes", async () => {
      const wrapper = mountComponent({
        modelValue: 20,
        deserved: 10,
      });

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);

      await wrapper.setProps({ modelValue: 5 });
      await flushPromises();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });

    it("updates validation state when deserved changes", async () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 5,
      });

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);

      await wrapper.setProps({ deserved: 20 });
      await flushPromises();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });

    it("clears error when value becomes valid", async () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);

      await wrapper.setProps({ modelValue: 15 });
      await flushPromises();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });

    it("clears error when value becomes null (unlimited)", async () => {
      const wrapper = mountComponent({
        modelValue: 5,
        deserved: 10,
      });

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);

      await wrapper.setProps({ modelValue: null });
      await flushPromises();

      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });
  });

  describe("Edge Cases", () => {
    it("handles zero deserved correctly (value must be > 0)", () => {
      const wrapper = mountComponent({
        modelValue: 1,
        deserved: 0,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });

    it("handles large values correctly", () => {
      const wrapper = mountComponent({
        modelValue: 999999,
        deserved: 1000,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });

    it("handles decimal values correctly", () => {
      const wrapper = mountComponent({
        modelValue: 10.5,
        deserved: 10,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Default);
    });

    it("handles decimal deserved values correctly", () => {
      const wrapper = mountComponent({
        modelValue: 10,
        deserved: 10.5,
      });
      expect(wrapper.vm.quotaState).toBe(EQuotaInputState.Error);
      expect(wrapper.vm.isBelowDeserved).toBe(true);
    });
  });
});
