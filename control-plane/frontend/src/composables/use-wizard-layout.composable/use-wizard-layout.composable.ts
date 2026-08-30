import { ref, computed } from "vue";
import type { Ref, ComputedRef } from "vue";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";

export interface IUseWizardLayoutReturn {
  // State (mutable for dynamic step management)
  steps: Ref<IRunaiWizardLayoutStep[]>;
  activeStep: Ref<string>; // Changed from number to string

  // Navigation methods
  goToStep: (stepName: string) => boolean;
  nextStep: () => boolean;
  previousStep: () => boolean;
  canNavigateToStep: (stepName: string) => boolean;

  // Step mutation methods
  markStepComplete: (stepName: string) => void;
  markStepIncomplete: (stepName: string) => void;
  setStepError: (stepName: string, hasError: boolean) => void;
  setStepDisabled: (stepName: string, disabled: boolean) => void;
  updateStep: (stepName: string, updates: Partial<IRunaiWizardLayoutStep>) => void;
  setSteps: (newSteps: IRunaiWizardLayoutStep[]) => void;

  // Computed properties
  canGoNext: ComputedRef<boolean>;
  canGoPrevious: ComputedRef<boolean>;
  isFirstStep: ComputedRef<boolean>;
  isLastStep: ComputedRef<boolean>;
  currentStep: ComputedRef<IRunaiWizardLayoutStep | undefined>;

  // Event handler
  handleStepClick: (payload: { step: IRunaiWizardLayoutStep; name: string }) => void;
}

/**
 * Composable for managing wizard layout state and navigation
 *
 * @param initialSteps - Array of initial wizard steps
 * @param initialActiveStepName - Optional name of the initial active step (defaults to first step)
 * @returns Object containing wizard state, navigation methods, and computed properties
 *
 * @example
 * ```ts
 * const {
 *   steps,
 *   activeStep,
 *   nextStep,
 *   previousStep,
 *   markStepComplete,
 *   handleStepClick
 * } = useWizardLayout([
 *   { name: 'step-1', title: 'Step 1', done: false },
 *   { name: 'step-2', title: 'Step 2', done: false },
 * ], 'step-1');
 * ```
 */
export function useWizardLayout(
  initialSteps: IRunaiWizardLayoutStep[],
  initialActiveStepName?: string,
): IUseWizardLayoutReturn {
  // State
  const steps = ref<IRunaiWizardLayoutStep[]>([...initialSteps]);
  const activeStep = ref<string>(initialActiveStepName || (steps.value[0]?.name ?? ""));

  // Helper: Find step by name
  const findStepByName = (name: string): IRunaiWizardLayoutStep | undefined => {
    return steps.value.find((step) => step.name === name);
  };

  // Helper: Find step index by name
  const findStepIndexByName = (name: string): number => {
    return steps.value.findIndex((step) => step.name === name);
  };

  // Helper: Validate step name exists
  const isValidStepName = (name: string): boolean => {
    return findStepByName(name) !== undefined;
  };

  // Helper: Find next non-disabled step name
  const findNextEnabledStepName = (fromStepName: string): string | null => {
    const currentIndex = findStepIndexByName(fromStepName);
    if (currentIndex === -1) return null;

    for (let i = currentIndex + 1; i < steps.value.length; i++) {
      if (!steps.value[i].disabled) {
        return steps.value[i].name;
      }
    }
    return null;
  };

  // Helper: Find previous non-disabled step name
  const findPreviousEnabledStepName = (fromStepName: string): string | null => {
    const currentIndex = findStepIndexByName(fromStepName);
    if (currentIndex === -1) return null;

    for (let i = currentIndex - 1; i >= 0; i--) {
      if (!steps.value[i].disabled) {
        return steps.value[i].name;
      }
    }
    return null;
  };

  // Navigation Methods

  /**
   * Check if navigation to a specific step is allowed
   * @param stepName - Target step name
   * @returns true if navigation is allowed
   */
  const canNavigateToStep = (stepName: string): boolean => {
    const step = findStepByName(stepName);
    if (!step) return false;
    if (step.disabled) return false;
    return true;
  };

  /**
   * Navigate to a specific step
   * @param stepName - Target step name
   * @returns true if navigation was successful
   */
  const goToStep = (stepName: string): boolean => {
    if (!isValidStepName(stepName)) {
      console.warn(`[useWizardLayout] Step "${stepName}" does not exist. Falling back to first step.`);
      if (steps.value.length > 0) {
        activeStep.value = steps.value[0].name;
      }
      return false;
    }

    if (!canNavigateToStep(stepName)) return false;
    activeStep.value = stepName;
    return true;
  };

  /**
   * Navigate to the next step (skips disabled steps)
   * @returns true if navigation was successful
   */
  const nextStep = (): boolean => {
    const nextStepName = findNextEnabledStepName(activeStep.value);
    if (nextStepName !== null) {
      activeStep.value = nextStepName;
      return true;
    }
    return false;
  };

  /**
   * Navigate to the previous step (skips disabled steps)
   * @returns true if navigation was successful
   */
  const previousStep = (): boolean => {
    const prevStepName = findPreviousEnabledStepName(activeStep.value);
    if (prevStepName !== null) {
      activeStep.value = prevStepName;
      return true;
    }
    return false;
  };

  // Step Mutation Methods

  /**
   * Mark a step as complete
   * @param stepName - Step name to mark complete
   */
  const markStepComplete = (stepName: string): void => {
    const step = findStepByName(stepName);
    if (step) {
      step.done = true;
    }
  };

  /**
   * Mark a step as incomplete
   * @param stepName - Step name to mark incomplete
   */
  const markStepIncomplete = (stepName: string): void => {
    const step = findStepByName(stepName);
    if (step) {
      step.done = false;
    }
  };

  /**
   * Set or unset error state on a step
   * @param stepName - Step name
   * @param hasError - Whether step has error
   */
  const setStepError = (stepName: string, hasError: boolean): void => {
    const step = findStepByName(stepName);
    if (step) {
      step.hasError = hasError;
    }
  };

  /**
   * Enable or disable a step
   * @param stepName - Step name
   * @param disabled - Whether step should be disabled
   */
  const setStepDisabled = (stepName: string, disabled: boolean): void => {
    const step = findStepByName(stepName);
    if (step) {
      step.disabled = disabled;
    }
  };

  /**
   * Update step properties
   * @param stepName - Step name
   * @param updates - Partial step object with properties to update
   */
  const updateStep = (stepName: string, updates: Partial<IRunaiWizardLayoutStep>): void => {
    const index = findStepIndexByName(stepName);
    if (index !== -1) {
      steps.value[index] = { ...steps.value[index], ...updates };
    }
  };

  /**
   * Replace all steps with new steps array
   * Preserves the active step if it exists in the new steps, otherwise resets to first step
   * @param newSteps - New array of steps
   */
  const setSteps = (newSteps: IRunaiWizardLayoutStep[]): void => {
    const currentActiveStep = activeStep.value;
    steps.value = [...newSteps];

    // If current active step doesn't exist in new steps, reset to first step
    const activeStepExists = newSteps.some((step) => step.name === currentActiveStep);
    if (!activeStepExists && newSteps.length > 0) {
      activeStep.value = newSteps[0].name;
    }
  };

  // Computed Properties

  /**
   * Whether navigation to next step is possible
   */
  const canGoNext = computed<boolean>(() => {
    return findNextEnabledStepName(activeStep.value) !== null;
  });

  /**
   * Whether navigation to previous step is possible
   */
  const canGoPrevious = computed<boolean>(() => {
    return findPreviousEnabledStepName(activeStep.value) !== null;
  });

  /**
   * Whether currently on first step
   */
  const isFirstStep = computed<boolean>(() => {
    return steps.value.length > 0 && activeStep.value === steps.value[0].name;
  });

  /**
   * Whether currently on last step
   */
  const isLastStep = computed<boolean>(() => {
    return steps.value.length > 0 && activeStep.value === steps.value[steps.value.length - 1].name;
  });

  /**
   * Current active step object
   */
  const currentStep = computed<IRunaiWizardLayoutStep | undefined>(() => {
    return findStepByName(activeStep.value);
  });

  // Event Handler

  /**
   * Handler for step-click event from wizard layout component
   * @param payload - Event payload with step and name
   */
  const handleStepClick = (payload: { step: IRunaiWizardLayoutStep; name: string }): void => {
    goToStep(payload.name);
  };

  return {
    // State (mutable for flexibility)
    steps,
    activeStep,

    // Navigation methods
    goToStep,
    nextStep,
    previousStep,
    canNavigateToStep,

    // Step mutation methods
    markStepComplete,
    markStepIncomplete,
    setStepError,
    setStepDisabled,
    updateStep,
    setSteps,

    // Computed properties
    canGoNext,
    canGoPrevious,
    isFirstStep,
    isLastStep,
    currentStep,

    // Event handler
    handleStepClick,
  };
}
