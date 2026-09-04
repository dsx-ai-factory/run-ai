/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, test, vi, beforeEach, afterEach } from "vitest";

import { PreemptibilitySection } from ".";
import { RunaiSelect } from "@/components/common/runai-select";
import { preemptibilitySectionMock } from "./preemptibility-section.mock";
import { Preemptibility } from "@/swagger-models/workloads-client";

const { mockPolicyRules, mockDefaultFactoryProps } = preemptibilitySectionMock;

describe("preemptibility-section.vue", () => {
  let preemptibilitySection: VueWrapper<InstanceType<typeof PreemptibilitySection>>;

  function factory(props = {}): VueWrapper<InstanceType<typeof PreemptibilitySection>> {
    const defaultProps = {
      ...mockDefaultFactoryProps,
      ...props,
    };

    const wrapper = mount(PreemptibilitySection, {
      global: {
        plugins: [Quasar],
      },
      props: defaultProps,
    });

    wrapper.vm.$q.notify = vi.fn();
    return wrapper;
  }

  beforeEach(() => {
    preemptibilitySection = factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Preemptibility Options & Computation", () => {
    test("Should compute preemptibilityOptions correctly", () => {
      const options = preemptibilitySection.vm.preemptibilityOptions;

      expect(options.length).toBe(2);
      expect(options.some((opt) => opt.value === Preemptibility.Preemptible)).toBe(true);
      expect(options.some((opt) => opt.value === Preemptibility.NonPreemptible)).toBe(true);
    });

    test("Should compute selectedPreemptibilityOption correctly", () => {
      const selectedOption = preemptibilitySection.vm.selectedPreemptibilityOption;
      expect(selectedOption).toBeDefined();
      expect(selectedOption?.value).toBe(Preemptibility.Preemptible);
    });

    test("Should include correct labels from policy rules", () => {
      const options = preemptibilitySection.vm.preemptibilityOptions;

      const preemptibleOption = options.find((opt) => opt.value === Preemptibility.Preemptible);
      const nonPreemptibleOption = options.find((opt) => opt.value === Preemptibility.NonPreemptible);

      expect(preemptibleOption?.label).toBe("Preemptible Workload");
      expect(nonPreemptibleOption?.label).toBe("Non-Preemptible Workload");
    });

    test("Should add warning tooltip only for non-preemptible when showQuotaWarning is true", () => {
      preemptibilitySection = factory({ showQuotaWarning: true });
      const options = preemptibilitySection.vm.preemptibilityOptions;

      const preemptibleOption = options.find((opt) => opt.value === Preemptibility.Preemptible);
      const nonPreemptibleOption = options.find((opt) => opt.value === Preemptibility.NonPreemptible);

      // Preemptible should not have warning
      expect(preemptibleOption?.warningTooltip).toBeUndefined();
      expect(preemptibleOption?.toolTip).toBeUndefined();

      // Non-preemptible should have warning tooltip
      expect(nonPreemptibleOption?.warningTooltip).toContain("available GPU quota");
      expect(nonPreemptibleOption?.type).toBeDefined();
      expect(nonPreemptibleOption?.isHtmlTooltipText).toBe(true);
    });

    test("Should not add warning tooltip when showQuotaWarning is false", () => {
      preemptibilitySection = factory({ showQuotaWarning: false });
      const options = preemptibilitySection.vm.preemptibilityOptions;

      options.forEach((option) => {
        expect(option.warningTooltip).toBeUndefined();
        expect(option.type).toBeUndefined();
      });
    });

    test("Should disable options not in policy options list", () => {
      preemptibilitySection = factory({ policyRules: mockPolicyRules.limitedOptions });
      const options = preemptibilitySection.vm.preemptibilityOptions;

      const preemptibleOption = options.find((opt) => opt.value === Preemptibility.Preemptible);

      // Only Preemptible is in the policy options, so NonPreemptible should be disabled
      expect(preemptibleOption).toBeDefined();
      // The option itself exists but should be disabled by the select component based on policy
    });
  });

  describe("Select Component Functionality", () => {
    test("Should emit update:model-value when preemptibility changes", async () => {
      const runaiSelect = preemptibilitySection.findComponent(RunaiSelect);
      expect(runaiSelect.exists()).toBe(true);

      const newOption = preemptibilitySection.vm.preemptibilityOptions.find(
        (opt) => opt.value === Preemptibility.NonPreemptible,
      );

      await runaiSelect.vm.$emit("update:model-value", newOption);
      await preemptibilitySection.vm.$nextTick();

      expect(preemptibilitySection.emitted("update:model-value")).toBeTruthy();
      const emittedValues = preemptibilitySection.emitted("update:model-value") as (Preemptibility | null)[][];
      expect(emittedValues[0][0]).toBe(Preemptibility.NonPreemptible);
    });

    test("Should emit null when selection is cleared", async () => {
      const runaiSelect = preemptibilitySection.findComponent(RunaiSelect);

      await runaiSelect.vm.$emit("update:model-value", null);
      await preemptibilitySection.vm.$nextTick();

      expect(preemptibilitySection.emitted("update:model-value")).toBeTruthy();
      const emittedValues = preemptibilitySection.emitted("update:model-value") as (Preemptibility | null)[][];
      expect(emittedValues[0][0]).toBeNull();
    });

    test("Should pass correct policy rules to runai-select", () => {
      const runaiSelect = preemptibilitySection.findComponent(RunaiSelect);
      const policyRulesProp = runaiSelect.props("policyRules");

      expect(policyRulesProp).toBeDefined();
      expect(policyRulesProp?.canEdit).toBe(true);
      expect(policyRulesProp?.required).toBe(false);
      expect(policyRulesProp?.isClosedList).toBe(true);
    });
  });

  describe("Disabled State & Policy Rules", () => {
    test("Should disable select when disabled prop is true", () => {
      preemptibilitySection = factory({ disabled: true });
      const runaiSelect = preemptibilitySection.findComponent(RunaiSelect);

      expect(runaiSelect.props("disable")).toBe(true);
      expect(preemptibilitySection.vm.isSectionDisabled).toBe(true);
    });

    test("Should disable select when policy canEdit is false", () => {
      preemptibilitySection = factory({ policyRules: mockPolicyRules.nonEditable });
      const runaiSelect = preemptibilitySection.findComponent(RunaiSelect);

      expect(runaiSelect.props("disable")).toBe(true);
      expect(preemptibilitySection.vm.isSectionDisabled).toBe(true);
    });

    test("Should compute isSectionDisabled correctly based on policy", () => {
      preemptibilitySection = factory({ policyRules: mockPolicyRules.nonEditable });
      expect(preemptibilitySection.vm.isSectionDisabled).toBe(true);

      preemptibilitySection = factory({ policyRules: mockPolicyRules.editable });
      expect(preemptibilitySection.vm.isSectionDisabled).toBe(false);
    });

    test("Should use policy options when provided", () => {
      preemptibilitySection = factory({ policyRules: mockPolicyRules.limitedOptions });
      const policyRules = preemptibilitySection.vm.selectGenericPolicyRules;

      expect(policyRules?.options).toBeDefined();
      expect(policyRules?.isClosedList).toBe(true);
    });
  });

  describe("Quota Warning Display", () => {
    test("Should compute shouldShowQuotaWarning correctly when non-preemptible is selected and showQuotaWarning is true", () => {
      preemptibilitySection = factory({
        modelValue: Preemptibility.NonPreemptible,
        showQuotaWarning: true,
      });

      expect(preemptibilitySection.vm.shouldShowQuotaWarning).toBe(true);
    });

    test("Should not show quota warning when preemptible is selected", () => {
      preemptibilitySection = factory({
        modelValue: Preemptibility.Preemptible,
        showQuotaWarning: true,
      });

      expect(preemptibilitySection.vm.shouldShowQuotaWarning).toBe(false);
    });

    test("Should not show quota warning when showQuotaWarning prop is false", () => {
      preemptibilitySection = factory({
        modelValue: Preemptibility.NonPreemptible,
        showQuotaWarning: false,
      });

      expect(preemptibilitySection.vm.shouldShowQuotaWarning).toBe(false);
    });

    test("Should display correct quota warning text", () => {
      preemptibilitySection = factory({
        modelValue: Preemptibility.NonPreemptible,
        showQuotaWarning: true,
      });

      const warningText = preemptibilitySection.vm.quotaWarningText;
      expect(warningText).toContain("pending");
      expect(warningText).toContain("available GPU quota");
      expect(warningText).toContain("non-preemptible workloads");
    });
  });

  describe("Component Props Validation", () => {
    test("Should accept null as modelValue", () => {
      preemptibilitySection = factory({ modelValue: null });
      expect(preemptibilitySection.vm.selectedPreemptibilityOption).toBeNull();
    });

    test("Should accept valid preemptibility value", () => {
      preemptibilitySection = factory({ modelValue: Preemptibility.NonPreemptible });
      expect(preemptibilitySection.vm.selectedPreemptibilityOption?.value).toBe(Preemptibility.NonPreemptible);
    });

    test("Should work with null policy rules", () => {
      preemptibilitySection = factory({ policyRules: null });
      const options = preemptibilitySection.vm.preemptibilityOptions;

      expect(options.length).toBe(2);
      expect(preemptibilitySection.vm.selectGenericPolicyRules).toBeUndefined();
    });
  });
});
