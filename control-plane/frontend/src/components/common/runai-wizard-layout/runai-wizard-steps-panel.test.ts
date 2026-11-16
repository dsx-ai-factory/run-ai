/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import { Quasar, QItem, QItemSection, QItemLabel, QIcon, QTooltip } from "quasar";
import RunaiWizardStepsPanel from "./runai-wizard-steps-panel.vue";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

describe("RunaiWizardStepsPanel", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiWizardStepsPanel>>;
  let mockSteps: IRunaiWizardLayoutStep[];

  const createWrapper = (props?: Partial<{ steps: IRunaiWizardLayoutStep[]; activeStep: string }>) => {
    return mount(RunaiWizardStepsPanel, {
      props: {
        steps: props?.steps || mockSteps,
        activeStep: props?.activeStep ?? "step-1",
      },
      global: {
        plugins: [Quasar],
        components: { QItem, QItemSection, QItemLabel, QIcon, QTooltip },
      },
    });
  };

  beforeEach(() => {
    mockSteps = [
      { name: "step-1", title: "Step 1", done: false },
      { name: "step-2", title: "Step 2", done: false },
      { name: "step-3", title: "Step 3", done: false },
    ];
  });

  describe("Orientation", () => {
    it("defaults to vertical orientation (no horizontal class)", () => {
      wrapper = createWrapper();

      const panel = wrapper.find(".wizard-steps-panel");
      expect(panel.classes()).not.toContain("wizard-steps-panel--horizontal");
    });

    it("applies horizontal class when orientation is horizontal", () => {
      wrapper = mount(RunaiWizardStepsPanel, {
        props: { steps: mockSteps, activeStep: "step-1", orientation: "horizontal" },
        global: {
          plugins: [Quasar],
          components: { QItem, QItemSection, QItemLabel, QIcon, QTooltip },
        },
      });

      const panel = wrapper.find(".wizard-steps-panel");
      expect(panel.classes()).toContain("wizard-steps-panel--horizontal");
    });

    it("still renders all steps and dividers in horizontal orientation", () => {
      wrapper = mount(RunaiWizardStepsPanel, {
        props: { steps: mockSteps, activeStep: "step-1", orientation: "horizontal" },
        global: {
          plugins: [Quasar],
          components: { QItem, QItemSection, QItemLabel, QIcon, QTooltip },
        },
      });

      expect(wrapper.findAll(".wizard-step")).toHaveLength(3);
      expect(wrapper.findAll(".wizard-step-divider")).toHaveLength(2);
    });
  });

  describe("Component Rendering", () => {
    it("renders the component correctly", () => {
      wrapper = createWrapper();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".wizard-steps-panel").exists()).toBe(true);
    });

    it("renders all steps", () => {
      wrapper = createWrapper();

      const stepItems = wrapper.findAll(".wizard-step");
      expect(stepItems).toHaveLength(3);
    });

    it("renders step titles correctly", () => {
      wrapper = createWrapper();

      const titles = wrapper.findAll(".wizard-step-title");
      expect(titles[0].text()).toBe("Step 1");
      expect(titles[1].text()).toBe("Step 2");
      expect(titles[2].text()).toBe("Step 3");
    });

    it("renders step indicators", () => {
      wrapper = createWrapper();

      const indicators = wrapper.findAll(".wizard-step-indicator");
      expect(indicators).toHaveLength(3);
    });

    it("renders step numbers for default state", () => {
      wrapper = createWrapper({ activeStep: "step-2" });

      const stepNumbers = wrapper.findAll(".wizard-step-number");
      expect(stepNumbers.length).toBeGreaterThan(0);
      expect(stepNumbers[0].text()).toBe("1");
    });

    it("renders dividers between steps", () => {
      wrapper = createWrapper();

      const dividers = wrapper.findAll(".wizard-step-divider");
      expect(dividers).toHaveLength(2); // n-1 dividers for n steps
    });

    it("does not render divider after last step", () => {
      wrapper = createWrapper();

      const lastStep = wrapper.findAll(".wizard-step").at(-1);
      expect(lastStep?.find(".wizard-step-divider").exists()).toBe(false);
    });
  });

  describe("Step States", () => {
    describe("Active/Selected State", () => {
      it("applies selected class to active step", () => {
        wrapper = createWrapper({ activeStep: "step-2" });

        const steps = wrapper.findAll(".wizard-step");
        expect(steps[1].classes()).toContain("wizard-step--selected");
      });

      it("shows pen icon for active step", () => {
        wrapper = createWrapper({ activeStep: "step-1" });

        const icons = wrapper.findAllComponents(QIcon);
        const penIcon = icons.find((icon) => icon.props("name") === "ph-fill ph-pencil-simple");
        expect(penIcon).toBeTruthy();
      });

      it("applies primary background to active step indicator", () => {
        wrapper = createWrapper({ activeStep: "step-2" });

        const indicators = wrapper.findAll(".wizard-step-indicator");
        expect(indicators[1].classes()).toContain("bg-primary");
      });
    });

    describe("Completed State", () => {
      it("applies completed class to done steps", () => {
        const stepsWithCompleted = [
          { name: "step-1", title: "Step 1", done: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithCompleted, activeStep: "step-2" });

        const steps = wrapper.findAll(".wizard-step");
        expect(steps[0].classes()).toContain("wizard-step--completed");
      });

      it("shows check icon for completed steps", () => {
        const stepsWithCompleted = [
          { name: "step-1", title: "Step 1", done: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithCompleted, activeStep: "step-2" });

        const icons = wrapper.findAllComponents(QIcon);
        const checkIcon = icons.find((icon) => icon.props("name") === "ph-bold ph-check");
        expect(checkIcon).toBeTruthy();
      });

      it("applies primary background to completed step indicator", () => {
        const stepsWithCompleted = [
          { name: "step-1", title: "Step 1", done: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithCompleted, activeStep: "step-2" });

        const indicators = wrapper.findAll(".wizard-step-indicator");
        expect(indicators[0].classes()).toContain("bg-primary");
      });
    });

    describe("Error State", () => {
      it("applies error class to steps with errors", () => {
        const stepsWithError = [
          { name: "step-1", title: "Step 1", done: false, hasError: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithError });

        const steps = wrapper.findAll(".wizard-step");
        expect(steps[0].classes()).toContain("wizard-step--error");
      });

      it("shows exclamation icon for error steps", () => {
        const stepsWithError = [
          { name: "step-1", title: "Step 1", done: false, hasError: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithError, activeStep: "step-2" });

        const icons = wrapper.findAllComponents(QIcon);
        const errorIcon = icons.find((icon) => icon.props("name") === "ph-bold ph-warning-circle");
        expect(errorIcon).toBeTruthy();
      });

      it("applies negative background to error step indicator", () => {
        const stepsWithError = [
          { name: "step-1", title: "Step 1", done: false, hasError: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithError, activeStep: "step-2" });

        const indicators = wrapper.findAll(".wizard-step-indicator");
        expect(indicators[0].classes()).toContain("bg-negative");
      });

      it("applies negative color to error step title", () => {
        const stepsWithError = [
          { name: "step-1", title: "Step 1", done: false, hasError: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        wrapper = createWrapper({ steps: stepsWithError, activeStep: "step-2" });

        const titles = wrapper.findAll(".wizard-step-title");
        expect(titles[0].classes()).toContain("text-negative");
      });
    });

    describe("Disabled State", () => {
      it("applies disabled class to disabled steps", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        wrapper = createWrapper({ steps: stepsWithDisabled });

        const steps = wrapper.findAll(".wizard-step");
        expect(steps[1].classes()).toContain("wizard-step--disabled");
      });

      it("disables clickability for disabled steps", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        wrapper = createWrapper({ steps: stepsWithDisabled });

        const qItems = wrapper.findAllComponents(QItem);
        expect(qItems[1].props("disable")).toBe(true);
      });

      it("applies grey background to disabled step indicator", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        wrapper = createWrapper({ steps: stepsWithDisabled });

        const indicators = wrapper.findAll(".wizard-step-indicator");
        expect(indicators[1].classes()).toContain("bg-grey-5");
      });

      it("applies grey color to disabled step title", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        wrapper = createWrapper({ steps: stepsWithDisabled });

        const titles = wrapper.findAll(".wizard-step-title");
        expect(titles[1].classes()).toContain("text-grey-5");
      });
    });

    describe("Default State", () => {
      it("applies default class to unselected, uncompleted steps", () => {
        wrapper = createWrapper({ activeStep: "step-1" });

        const steps = wrapper.findAll(".wizard-step");
        expect(steps[1].classes()).toContain("wizard-step--default");
      });

      it("shows step number for default steps", () => {
        wrapper = createWrapper({ activeStep: "step-1" });

        const stepNumbers = wrapper.findAll(".wizard-step-number");
        expect(stepNumbers.length).toBeGreaterThan(0);
      });

      it("applies grey background to default step indicator", () => {
        wrapper = createWrapper({ activeStep: "step-1" });

        const indicators = wrapper.findAll(".wizard-step-indicator");
        expect(indicators[1].classes()).toContain("bg-grey-6");
      });

      it("applies grey color to default step title", () => {
        wrapper = createWrapper({ activeStep: "step-1" });

        const titles = wrapper.findAll(".wizard-step-title");
        expect(titles[1].classes()).toContain("text-grey-6");
      });
    });
  });

  describe("Step Icons", () => {
    it("determines icon visibility correctly", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
        { name: "step-3", title: "Step 3", done: false, hasError: true },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-2" });

      // Step 0 (done): should show icon
      expect(wrapper.vm.shouldShowIcon(stepsWithMixedStates[0])).toBe(true);

      // Step 1 (active): should show icon
      expect(wrapper.vm.shouldShowIcon(stepsWithMixedStates[1])).toBe(true);

      // Step 2 (error): should show icon
      expect(wrapper.vm.shouldShowIcon(stepsWithMixedStates[2])).toBe(true);
    });

    it("returns correct icon name based on state", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
        { name: "step-3", title: "Step 3", done: false, hasError: true },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-2" });

      expect(wrapper.vm.getStepIcon(stepsWithMixedStates[0])).toBe("ph-bold ph-check");
      expect(wrapper.vm.getStepIcon(stepsWithMixedStates[1])).toBe("ph-fill ph-pencil-simple");
      expect(wrapper.vm.getStepIcon(stepsWithMixedStates[2])).toBe("ph-bold ph-warning-circle");
    });

    it("returns reviewed icon size based on state", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-2" });

      expect(wrapper.vm.getStepIconSize(stepsWithMixedStates[0])).toBe("16px");
      expect(wrapper.vm.getStepIconSize(stepsWithMixedStates[1])).toBe("16px");
    });

    it("active step uses reviewed sidebar icon size", () => {
      wrapper = createWrapper({ activeStep: "step-1" });

      expect(wrapper.vm.getStepIconSize(mockSteps[0])).toBe("16px");
    });

    it("completed and error steps use reviewed icon sizes", () => {
      const stepsWithStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false, hasError: true },
      ];
      wrapper = createWrapper({ steps: stepsWithStates, activeStep: "step-3" });

      expect(wrapper.vm.getStepIconSize(stepsWithStates[0])).toBe("16px");
      expect(wrapper.vm.getStepIconSize(stepsWithStates[1])).toBe("20px");
    });
  });

  describe("User Interactions", () => {
    it("emits step-click event when enabled step is clicked", async () => {
      wrapper = createWrapper();

      const stepItems = wrapper.findAllComponents(QItem);
      await stepItems[1].trigger("click");

      expect(wrapper.emitted("step-click")).toBeTruthy();
      expect(wrapper.emitted("step-click")?.[0]).toEqual([{ step: mockSteps[1], name: "step-2" }]);
    });

    it("does not emit step-click for disabled steps", async () => {
      const stepsWithDisabled = [
        { name: "step-1", title: "Step 1", done: false },
        { name: "step-2", title: "Step 2", done: false, disabled: true },
      ];
      wrapper = createWrapper({ steps: stepsWithDisabled });

      const stepItems = wrapper.findAllComponents(QItem);
      await stepItems[1].trigger("click");

      expect(wrapper.emitted("step-click")).toBeFalsy();
    });

    it("emits correct payload on step click", async () => {
      wrapper = createWrapper();

      const stepItems = wrapper.findAllComponents(QItem);
      await stepItems[2].trigger("click");

      const emitted = wrapper.emitted("step-click");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([{ step: mockSteps[2], name: "step-3" }]);
    });

    it("allows clicking on completed steps", async () => {
      const stepsWithCompleted = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithCompleted, activeStep: "step-2" });

      const stepItems = wrapper.findAllComponents(QItem);
      await stepItems[0].trigger("click");

      expect(wrapper.emitted("step-click")).toBeTruthy();
    });

    it("allows clicking on error steps", async () => {
      const stepsWithError = [
        { name: "step-1", title: "Step 1", done: false, hasError: true },
        { name: "step-2", title: "Step 2", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithError, activeStep: "step-2" });

      const stepItems = wrapper.findAllComponents(QItem);
      await stepItems[0].trigger("click");

      expect(wrapper.emitted("step-click")).toBeTruthy();
    });
  });

  describe("Props", () => {
    it("requires steps prop", () => {
      expect(RunaiWizardStepsPanel.props.steps.required).toBe(true);
    });

    it("requires activeStep prop", () => {
      expect(RunaiWizardStepsPanel.props.activeStep.required).toBe(true);
    });

    it("activeStep is string type", () => {
      expect(RunaiWizardStepsPanel.props.activeStep.type).toBe(String);
    });

    it("accepts custom steps array", () => {
      const customSteps = [
        { name: "custom-a", title: "Custom A", done: false },
        { name: "custom-b", title: "Custom B", done: true },
      ];
      wrapper = createWrapper({ steps: customSteps });

      expect(wrapper.props("steps")).toEqual(customSteps);
    });

    it("updates when activeStep prop changes", async () => {
      wrapper = createWrapper({ activeStep: "step-1" });

      await wrapper.setProps({ activeStep: "step-3" });

      const steps = wrapper.findAll(".wizard-step");
      expect(steps[2].classes()).toContain("wizard-step--selected");
    });
  });

  describe("Component Methods", () => {
    it("handleStepClick calls emit for enabled steps", () => {
      wrapper = createWrapper();

      wrapper.vm.handleStepClick(mockSteps[1]);

      expect(wrapper.emitted("step-click")).toBeTruthy();
    });

    it("handleStepClick does not emit for disabled steps", () => {
      const stepsWithDisabled = [
        { name: "step-1", title: "Step 1", done: false },
        { name: "step-2", title: "Step 2", done: false, disabled: true },
      ];
      wrapper = createWrapper({ steps: stepsWithDisabled });

      wrapper.vm.handleStepClick(stepsWithDisabled[1]);

      expect(wrapper.emitted("step-click")).toBeFalsy();
    });

    it("getStepClasses returns correct classes for each state", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
        { name: "step-3", title: "Step 3", done: false, disabled: true },
        { name: "step-4", title: "Step 4", done: false, hasError: true },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-2" });

      const classes0 = wrapper.vm.getStepClasses(stepsWithMixedStates[0]);
      expect(classes0["wizard-step--completed"]).toBe(true);

      const classes1 = wrapper.vm.getStepClasses(stepsWithMixedStates[1]);
      expect(classes1["wizard-step--selected"]).toBe(true);

      const classes2 = wrapper.vm.getStepClasses(stepsWithMixedStates[2]);
      expect(classes2["wizard-step--disabled"]).toBe(true);

      const classes3 = wrapper.vm.getStepClasses(stepsWithMixedStates[3]);
      expect(classes3["wizard-step--error"]).toBe(true);
    });

    it("getIndicatorClasses returns correct indicator classes", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: true },
        { name: "step-2", title: "Step 2", done: false },
        { name: "step-3", title: "Step 3", done: false, hasError: true },
        { name: "step-4", title: "Step 4", done: false, disabled: true },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-2" });

      expect(wrapper.vm.getIndicatorClasses(stepsWithMixedStates[0])).toBe("bg-primary");
      expect(wrapper.vm.getIndicatorClasses(stepsWithMixedStates[1])).toBe("bg-primary");
      expect(wrapper.vm.getIndicatorClasses(stepsWithMixedStates[2])).toBe("bg-negative");
      expect(wrapper.vm.getIndicatorClasses(stepsWithMixedStates[3])).toBe("bg-grey-5");
    });

    it("getTitleClasses returns correct title classes", () => {
      const stepsWithMixedStates = [
        { name: "step-1", title: "Step 1", done: false },
        { name: "step-2", title: "Step 2", done: false, hasError: true },
        { name: "step-3", title: "Step 3", done: false, disabled: true },
      ];
      wrapper = createWrapper({ steps: stepsWithMixedStates, activeStep: "step-1" });

      expect(wrapper.vm.getTitleClasses(stepsWithMixedStates[0])).toBe("text-grey-9");
      expect(wrapper.vm.getTitleClasses(stepsWithMixedStates[1])).toBe("text-negative");
      expect(wrapper.vm.getTitleClasses(stepsWithMixedStates[2])).toBe("text-grey-5");
    });
  });

  describe("Edge Cases", () => {
    it("handles empty steps array", () => {
      wrapper = createWrapper({ steps: [] });

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.findAll(".wizard-step")).toHaveLength(0);
    });

    it("handles single step", () => {
      const singleStep = [{ name: "only-step", title: "Only Step", done: false }];
      wrapper = createWrapper({ steps: singleStep, activeStep: "only-step" });

      expect(wrapper.findAll(".wizard-step")).toHaveLength(1);
      expect(wrapper.findAll(".wizard-step-divider")).toHaveLength(0);
    });

    it("handles many steps", () => {
      const manySteps = Array.from({ length: 10 }, (_, i) => ({
        name: `step-${i + 1}`,
        title: `Step ${i + 1}`,
        done: false,
      }));
      wrapper = createWrapper({ steps: manySteps, activeStep: "step-1" });

      expect(wrapper.findAll(".wizard-step")).toHaveLength(10);
      expect(wrapper.findAll(".wizard-step-divider")).toHaveLength(9);
    });

    it("handles long step titles", () => {
      const stepsWithLongTitles = [
        { name: "long-step", title: "This is a very long step title that might wrap", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithLongTitles, activeStep: "long-step" });

      const title = wrapper.find(".wizard-step-title");
      expect(title.text()).toBe("This is a very long step title that might wrap");
    });

    it("handles activeStep that does not exist", () => {
      wrapper = createWrapper({ activeStep: "non-existent-step" });

      expect(wrapper.exists()).toBe(true);
      // No step should have selected class
      const selectedSteps = wrapper.findAll(".wizard-step--selected");
      expect(selectedSteps).toHaveLength(0);
    });
  });

  describe("Accessibility", () => {
    it("renders with proper structure", () => {
      wrapper = createWrapper();

      const stepsPanel = wrapper.find(".wizard-steps-panel");
      expect(stepsPanel.exists()).toBe(true);
    });

    it("uses Quasar accessible components", () => {
      wrapper = createWrapper();

      expect(wrapper.findAllComponents(QItem).length).toBeGreaterThan(0);
    });

    it("sets clickable prop on QItem for interactivity", () => {
      wrapper = createWrapper();

      const qItems = wrapper.findAllComponents(QItem);
      qItems.forEach((item, index) => {
        if (!mockSteps[index].disabled) {
          expect(item.props("clickable")).toBe(true);
        }
      });
    });
  });

  describe("Visual States Priority", () => {
    it("active state takes precedence over done state", () => {
      const steps = [{ name: "step-1", title: "Step 1", done: true }];
      wrapper = createWrapper({ steps, activeStep: "step-1" });

      const stepElement = wrapper.find(".wizard-step");
      expect(stepElement.classes()).toContain("wizard-step--selected");
      expect(stepElement.classes()).not.toContain("wizard-step--completed");
    });

    it("error state is independent from other states", () => {
      const steps = [{ name: "step-1", title: "Step 1", done: true, hasError: true }];
      wrapper = createWrapper({ steps, activeStep: "step-2" });

      const stepElement = wrapper.find(".wizard-step");
      expect(stepElement.classes()).toContain("wizard-step--error");
      expect(stepElement.classes()).toContain("wizard-step--completed");
    });
  });

  describe("Tooltips", () => {
    it("renders tooltip when step has tooltip property", () => {
      const stepsWithTooltip = [
        { name: "step-1", title: "Step 1", done: false, tooltip: "This is step 1" },
        { name: "step-2", title: "Step 2", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithTooltip });

      const tooltips = wrapper.findAllComponents(QTooltip);
      expect(tooltips).toHaveLength(1);
    });

    it("does not render tooltip when step.tooltip is undefined", () => {
      const stepsWithoutTooltip = [
        { name: "step-1", title: "Step 1", done: false },
        { name: "step-2", title: "Step 2", done: false },
      ];
      wrapper = createWrapper({ steps: stepsWithoutTooltip });

      const tooltips = wrapper.findAllComponents(QTooltip);
      expect(tooltips).toHaveLength(0);
    });

    it("renders tooltip component when step has tooltip property", () => {
      const stepsWithTooltip = [{ name: "step-1", title: "Step 1", done: false, tooltip: "Test tooltip" }];
      wrapper = createWrapper({ steps: stepsWithTooltip });

      const tooltip = wrapper.findComponent(QTooltip);
      expect(tooltip.exists()).toBe(true);
    });

    it("renders correct number of tooltips based on tooltip property", () => {
      const stepsWithTooltips = [
        { name: "step-1", title: "Step 1", done: false, tooltip: "Tooltip 1" },
        { name: "step-2", title: "Step 2", done: false, tooltip: "Tooltip 2" },
        { name: "step-3", title: "Step 3", done: false }, // No tooltip
      ];
      wrapper = createWrapper({ steps: stepsWithTooltips });

      const tooltips = wrapper.findAllComponents(QTooltip);
      expect(tooltips).toHaveLength(2); // Only 2 tooltips for 3 steps
    });

    it("applies correct anchor and self props to tooltip", () => {
      const stepsWithTooltip = [{ name: "step-1", title: "Step 1", done: false, tooltip: "Test" }];
      wrapper = createWrapper({ steps: stepsWithTooltip });

      const tooltip = wrapper.findComponent(QTooltip);
      expect(tooltip.props("anchor")).toBe("center right");
      expect(tooltip.props("self")).toBe("top middle");
    });

    it("renders tooltips for all step states (active, completed, error, disabled)", () => {
      const steps = [
        { name: "step-1", title: "Active Step", done: false, tooltip: "Active tooltip" },
        { name: "step-2", title: "Completed Step", done: true, tooltip: "Completed tooltip" },
        { name: "step-3", title: "Error Step", done: false, hasError: true, tooltip: "Error tooltip" },
        { name: "step-4", title: "Disabled Step", done: false, disabled: true, tooltip: "Disabled tooltip" },
      ];
      wrapper = createWrapper({ steps, activeStep: "step-1" });

      const tooltips = wrapper.findAllComponents(QTooltip);
      expect(tooltips).toHaveLength(4);
    });

    it("handles empty tooltip string", () => {
      const stepsWithEmptyTooltip = [{ name: "step-1", title: "Step 1", done: false, tooltip: "" }];
      wrapper = createWrapper({ steps: stepsWithEmptyTooltip });

      // Empty string is still truthy for v-if, so tooltip should not render
      // But we should check if the component handles it gracefully
      const tooltips = wrapper.findAllComponents(QTooltip);
      // Empty string is falsy in v-if context when checking just the string
      // But technically "" is still a defined value, so it may render
      // Let's check what actually happens
      expect(tooltips.length).toBe(0);
    });
  });
});
