/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { useWizardLayout } from "./use-wizard-layout.composable";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

describe("useWizardLayout", () => {
  let mockSteps: IRunaiWizardLayoutStep[];

  beforeEach(() => {
    mockSteps = [
      { name: "step-1", title: "Step 1", done: false },
      { name: "step-2", title: "Step 2", done: false },
      { name: "step-3", title: "Step 3", done: false },
    ];
  });

  describe("Initialization", () => {
    it("initializes with correct default state", () => {
      const { steps, activeStep } = useWizardLayout(mockSteps);

      expect(steps.value).toHaveLength(3);
      expect(activeStep.value).toBe("step-1");
      expect(steps.value[0]).toEqual(mockSteps[0]);
    });

    it("initializes with custom active step", () => {
      const { activeStep } = useWizardLayout(mockSteps, "step-2");

      expect(activeStep.value).toBe("step-2");
    });

    it("creates a new array reference (not mutating the original array reference)", () => {
      const { steps } = useWizardLayout(mockSteps);

      steps.value[0].done = true;

      // Note: The composable does a shallow copy, so the step objects are still references
      // This is intentional to allow the parent to observe changes
      // What's protected is the array reference itself
      expect(steps.value).not.toBe(mockSteps);
    });
  });

  describe("Navigation Methods", () => {
    describe("goToStep", () => {
      it("navigates to a valid step", () => {
        const { activeStep, goToStep } = useWizardLayout(mockSteps);

        const result = goToStep("step-3");

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-3");
      });

      it("returns false and warns for invalid step name", () => {
        const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
        const { activeStep, goToStep } = useWizardLayout(mockSteps);

        const result = goToStep("invalid-step");

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-1"); // Falls back to first step
        expect(consoleWarnSpy).toHaveBeenCalledWith(
          '[useWizardLayout] Step "invalid-step" does not exist. Falling back to first step.',
        );

        consoleWarnSpy.mockRestore();
      });

      it("returns false for disabled step", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false },
        ];
        const { activeStep, goToStep } = useWizardLayout(stepsWithDisabled);

        const result = goToStep("step-2");

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-1");
      });

      it("navigates to an enabled step successfully", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false },
        ];
        const { activeStep, goToStep } = useWizardLayout(stepsWithDisabled);

        const result = goToStep("step-3");

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-3");
      });
    });

    describe("nextStep", () => {
      it("moves to the next step successfully", () => {
        const { activeStep, nextStep } = useWizardLayout(mockSteps);

        const result = nextStep();

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-2");
      });

      it("returns false when already on last step", () => {
        const { activeStep, nextStep } = useWizardLayout(mockSteps);
        activeStep.value = "step-3";

        const result = nextStep();

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-3");
      });

      it("skips disabled steps when moving forward", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false },
        ];
        const { activeStep, nextStep } = useWizardLayout(stepsWithDisabled);

        const result = nextStep();

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-3"); // Skipped step-2
      });

      it("returns false when all remaining steps are disabled", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false, disabled: true },
        ];
        const { activeStep, nextStep } = useWizardLayout(stepsWithDisabled);

        const result = nextStep();

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-1");
      });

      it("skips multiple consecutive disabled steps", () => {
        const stepsWithMultipleDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false, disabled: true },
          { name: "step-4", title: "Step 4", done: false },
        ];
        const { activeStep, nextStep } = useWizardLayout(stepsWithMultipleDisabled);

        const result = nextStep();

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-4"); // Skipped step-2 and step-3
      });
    });

    describe("previousStep", () => {
      it("moves to the previous step successfully", () => {
        const { activeStep, previousStep } = useWizardLayout(mockSteps, "step-3");

        const result = previousStep();

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-2");
      });

      it("returns false when already on first step", () => {
        const { activeStep, previousStep } = useWizardLayout(mockSteps);

        const result = previousStep();

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-1");
      });

      it("skips disabled steps when moving backward", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false },
        ];
        const { activeStep, previousStep } = useWizardLayout(stepsWithDisabled, "step-3");

        const result = previousStep();

        expect(result).toBe(true);
        expect(activeStep.value).toBe("step-1"); // Skipped step-2
      });

      it("returns false when all previous steps are disabled", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false, disabled: true },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
          { name: "step-3", title: "Step 3", done: false },
        ];
        const { activeStep, previousStep } = useWizardLayout(stepsWithDisabled, "step-3");

        const result = previousStep();

        expect(result).toBe(false);
        expect(activeStep.value).toBe("step-3");
      });
    });

    describe("canNavigateToStep", () => {
      it("returns true for valid enabled step", () => {
        const { canNavigateToStep } = useWizardLayout(mockSteps);

        const result = canNavigateToStep("step-2");

        expect(result).toBe(true);
      });

      it("returns false for invalid step name", () => {
        const { canNavigateToStep } = useWizardLayout(mockSteps);

        expect(canNavigateToStep("invalid-step")).toBe(false);
      });

      it("returns false for disabled step", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        const { canNavigateToStep } = useWizardLayout(stepsWithDisabled);

        const result = canNavigateToStep("step-2");

        expect(result).toBe(false);
      });
    });
  });

  describe("Step Mutation Methods", () => {
    describe("markStepComplete", () => {
      it("marks a step as complete", () => {
        const { steps, markStepComplete } = useWizardLayout(mockSteps);

        markStepComplete("step-1");

        expect(steps.value[0].done).toBe(true);
      });

      it("handles invalid step name gracefully", () => {
        const { steps, markStepComplete } = useWizardLayout(mockSteps);

        markStepComplete("invalid-step");

        expect(steps.value[0].done).toBe(false);
      });

      it("can mark multiple steps as complete", () => {
        const { steps, markStepComplete } = useWizardLayout(mockSteps);

        markStepComplete("step-1");
        markStepComplete("step-2");

        expect(steps.value[0].done).toBe(true);
        expect(steps.value[1].done).toBe(true);
        expect(steps.value[2].done).toBe(false);
      });
    });

    describe("markStepIncomplete", () => {
      it("marks a step as incomplete", () => {
        const { steps, markStepComplete, markStepIncomplete } = useWizardLayout(mockSteps);
        markStepComplete("step-1");

        markStepIncomplete("step-1");

        expect(steps.value[0].done).toBe(false);
      });

      it("handles invalid step name gracefully", () => {
        const { steps, markStepIncomplete } = useWizardLayout(mockSteps);

        markStepIncomplete("invalid-step");

        // Should not throw or cause issues
        expect(steps.value).toHaveLength(3);
      });
    });

    describe("setStepError", () => {
      it("sets error state on a step", () => {
        const { steps, setStepError } = useWizardLayout(mockSteps);

        setStepError("step-1", true);

        expect(steps.value[0].hasError).toBe(true);
      });

      it("clears error state on a step", () => {
        const { steps, setStepError } = useWizardLayout(mockSteps);
        setStepError("step-1", true);

        setStepError("step-1", false);

        expect(steps.value[0].hasError).toBe(false);
      });

      it("handles invalid step name gracefully", () => {
        const { steps, setStepError } = useWizardLayout(mockSteps);

        setStepError("invalid-step", true);

        expect(steps.value[0].hasError).toBeUndefined();
      });
    });

    describe("setStepDisabled", () => {
      it("disables a step", () => {
        const { steps, setStepDisabled } = useWizardLayout(mockSteps);

        setStepDisabled("step-2", true);

        expect(steps.value[1].disabled).toBe(true);
      });

      it("enables a step", () => {
        const { steps, setStepDisabled } = useWizardLayout(mockSteps);
        setStepDisabled("step-2", true);

        setStepDisabled("step-2", false);

        expect(steps.value[1].disabled).toBe(false);
      });

      it("handles invalid step name gracefully", () => {
        const { steps, setStepDisabled } = useWizardLayout(mockSteps);

        setStepDisabled("invalid-step", true);

        expect(steps.value[0].disabled).toBeUndefined();
      });
    });

    describe("updateStep", () => {
      it("updates multiple properties at once", () => {
        const { steps, updateStep } = useWizardLayout(mockSteps);

        updateStep("step-1", { done: true, hasError: false, title: "Updated Title" });

        expect(steps.value[0].done).toBe(true);
        expect(steps.value[0].hasError).toBe(false);
        expect(steps.value[0].title).toBe("Updated Title");
      });

      it("preserves other properties when updating", () => {
        const { steps, updateStep } = useWizardLayout(mockSteps);

        updateStep("step-1", { done: true });

        expect(steps.value[0].name).toBe("step-1");
        expect(steps.value[0].title).toBe("Step 1");
        expect(steps.value[0].done).toBe(true);
      });

      it("handles invalid step name gracefully", () => {
        const { steps, updateStep } = useWizardLayout(mockSteps);

        updateStep("invalid-step", { done: true });

        expect(steps.value[0].done).toBe(false);
      });
    });
  });

  describe("Computed Properties", () => {
    describe("canGoNext", () => {
      it("returns true when not on last step", () => {
        const { canGoNext } = useWizardLayout(mockSteps);

        expect(canGoNext.value).toBe(true);
      });

      it("returns false when on last step", () => {
        const { canGoNext } = useWizardLayout(mockSteps, "step-3");

        expect(canGoNext.value).toBe(false);
      });

      it("returns false when next steps are disabled", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        const { canGoNext } = useWizardLayout(stepsWithDisabled);

        expect(canGoNext.value).toBe(false);
      });
    });

    describe("canGoPrevious", () => {
      it("returns false when on first step", () => {
        const { canGoPrevious } = useWizardLayout(mockSteps);

        expect(canGoPrevious.value).toBe(false);
      });

      it("returns true when not on first step", () => {
        const { canGoPrevious } = useWizardLayout(mockSteps, "step-2");

        expect(canGoPrevious.value).toBe(true);
      });

      it("returns false when previous steps are disabled", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false, disabled: true },
          { name: "step-2", title: "Step 2", done: false },
        ];
        const { canGoPrevious } = useWizardLayout(stepsWithDisabled, "step-2");

        expect(canGoPrevious.value).toBe(false);
      });
    });

    describe("isFirstStep", () => {
      it("returns true when on first step", () => {
        const { isFirstStep } = useWizardLayout(mockSteps);

        expect(isFirstStep.value).toBe(true);
      });

      it("returns false when not on first step", () => {
        const { isFirstStep } = useWizardLayout(mockSteps, "step-2");

        expect(isFirstStep.value).toBe(false);
      });
    });

    describe("isLastStep", () => {
      it("returns false when not on last step", () => {
        const { isLastStep } = useWizardLayout(mockSteps);

        expect(isLastStep.value).toBe(false);
      });

      it("returns true when on last step", () => {
        const { isLastStep } = useWizardLayout(mockSteps, "step-3");

        expect(isLastStep.value).toBe(true);
      });
    });

    describe("currentStep", () => {
      it("returns the current step object", () => {
        const { currentStep } = useWizardLayout(mockSteps);

        expect(currentStep.value).toEqual(mockSteps[0]);
      });

      it("updates when activeStep changes", () => {
        const { activeStep, currentStep } = useWizardLayout(mockSteps);

        activeStep.value = "step-2";

        expect(currentStep.value).toEqual(mockSteps[1]);
      });

      it("returns undefined for invalid activeStep", () => {
        const { activeStep, currentStep } = useWizardLayout(mockSteps);

        activeStep.value = "invalid-step";

        expect(currentStep.value).toBeUndefined();
      });
    });
  });

  describe("Event Handler", () => {
    describe("handleStepClick", () => {
      it("navigates to clicked step", () => {
        const { activeStep, handleStepClick } = useWizardLayout(mockSteps);

        handleStepClick({ step: mockSteps[2], name: "step-3" });

        expect(activeStep.value).toBe("step-3");
      });

      it("does not navigate to disabled step", () => {
        const stepsWithDisabled = [
          { name: "step-1", title: "Step 1", done: false },
          { name: "step-2", title: "Step 2", done: false, disabled: true },
        ];
        const { activeStep, handleStepClick } = useWizardLayout(stepsWithDisabled);

        handleStepClick({ step: stepsWithDisabled[1], name: "step-2" });

        expect(activeStep.value).toBe("step-1");
      });

      it("navigates to enabled step successfully", () => {
        const { activeStep, handleStepClick } = useWizardLayout(mockSteps);

        handleStepClick({ step: mockSteps[1], name: "step-2" });

        expect(activeStep.value).toBe("step-2");
      });
    });
  });

  describe("Edge Cases", () => {
    it("handles empty steps array gracefully", () => {
      const { steps, activeStep } = useWizardLayout([]);

      expect(steps.value).toHaveLength(0);
      expect(activeStep.value).toBe("");
    });

    it("handles single step array", () => {
      const singleStep = [{ name: "only-step", title: "Only Step", done: false }];
      const { steps, canGoNext, canGoPrevious, isFirstStep, isLastStep } = useWizardLayout(singleStep);

      expect(steps.value).toHaveLength(1);
      expect(canGoNext.value).toBe(false);
      expect(canGoPrevious.value).toBe(false);
      expect(isFirstStep.value).toBe(true);
      expect(isLastStep.value).toBe(true);
    });

    it("handles all steps disabled", () => {
      const allDisabled = [
        { name: "step-1", title: "Step 1", done: false, disabled: true },
        { name: "step-2", title: "Step 2", done: false, disabled: true },
      ];
      const { canGoNext, canGoPrevious, nextStep, previousStep } = useWizardLayout(allDisabled);

      expect(canGoNext.value).toBe(false);
      expect(canGoPrevious.value).toBe(false);
      expect(nextStep()).toBe(false);
      expect(previousStep()).toBe(false);
    });
  });

  describe("Dynamic Step Management", () => {
    it("allows adding steps dynamically", () => {
      const { steps } = useWizardLayout(mockSteps);

      steps.value.push({ name: "step-4", title: "Step 4", done: false });

      expect(steps.value).toHaveLength(4);
      expect(steps.value[3].name).toBe("step-4");
    });

    it("allows removing steps dynamically", () => {
      const { steps } = useWizardLayout(mockSteps);

      steps.value.pop();

      expect(steps.value).toHaveLength(2);
    });

    it("allows inserting steps in the middle", () => {
      const { steps } = useWizardLayout(mockSteps);

      steps.value.splice(1, 0, { name: "step-1.5", title: "Step 1.5", done: false });

      expect(steps.value).toHaveLength(4);
      expect(steps.value[1].name).toBe("step-1.5");
    });

    it("navigation works correctly after dynamic step addition", () => {
      const { steps, nextStep, activeStep } = useWizardLayout(mockSteps);

      steps.value.push({ name: "step-4", title: "Step 4", done: false });
      activeStep.value = "step-3";
      const result = nextStep();

      expect(result).toBe(true);
      expect(activeStep.value).toBe("step-4");
    });
  });

  describe("Complex Scenarios", () => {
    it("handles navigation with mixed disabled steps", () => {
      const complexSteps = [
        { name: "step-1", title: "Step 1", done: false },
        { name: "step-2", title: "Step 2", done: false, disabled: true },
        { name: "step-3", title: "Step 3", done: false },
        { name: "step-4", title: "Step 4", done: false, disabled: true },
        { name: "step-5", title: "Step 5", done: false },
      ];
      const { activeStep, nextStep } = useWizardLayout(complexSteps);

      // From step-1 to step-3 (skip step-2)
      let result = nextStep();
      expect(result).toBe(true);
      expect(activeStep.value).toBe("step-3");

      // From step-3 to step-5 (skip step-4)
      result = nextStep();
      expect(result).toBe(true);
      expect(activeStep.value).toBe("step-5");

      // No more steps
      result = nextStep();
      expect(result).toBe(false);
      expect(activeStep.value).toBe("step-5");
    });

    it("maintains step state through navigation", () => {
      const { activeStep, markStepComplete, setStepError, nextStep, previousStep, steps } = useWizardLayout(mockSteps);

      // Mark first step complete and move forward
      markStepComplete("step-1");
      nextStep();

      // Set error on second step
      setStepError("step-2", true);

      // Go back and check first step is still complete
      previousStep();

      expect(steps.value[0].done).toBe(true);
      expect(steps.value[1].hasError).toBe(true);
      expect(activeStep.value).toBe("step-1");
    });
  });
});
