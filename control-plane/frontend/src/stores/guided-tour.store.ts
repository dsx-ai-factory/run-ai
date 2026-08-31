/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import type {
  TourConfig,
  TourState,
  TourStep,
  TourContext,
  PopoverProps,
  GuidedTour,
} from "@/components/common/guided-tour-popup/guided-tour.types";
import { guidedTourService } from "@/services/guided-tour.service/guided-tour.service";
import {
  getWorkloadsGuidedTourSteps,
  OVERVIEW_GUIDED_TOUR_STEPS,
  getProjectGuidedTourSteps,
  getDepartmentGuidedTourSteps,
  type ProjectTourOptions,
  type DepartmentTourOptions,
} from "@/services/guided-tour.service/guided-tours";
import {
  DEFAULT_CONFIG,
  OVERVIEW_TOUR_CONFIG,
  PROJECT_TOUR_CONFIG,
  DEPARTMENT_TOUR_CONFIG,
} from "@/services/guided-tour.service/guided-tour-configs";
import { usePermissionStore } from "./permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION } from "@/common/version.constant";
import { useClusterStore } from "./cluster.store";
import { useSettingStore } from "./setting.store";
import { SettingKeys } from "@/models/setting.model";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// Initial state
const INITIAL_STATE: TourState = {
  isInitialized: false,
  activeIndex: undefined,
  activeElement: null,
  activeStep: null,
  previousElement: null,
  previousStep: null,
  overlayElement: null,
  isTransitioning: false,
  animationDuration: 250,
  transitionCallback: null,
};

export const useGuidedTourStore = defineStore("GuidedTour", {
  state: () => ({
    // Configuration
    config: { ...DEFAULT_CONFIG } as TourConfig,

    // Tour state
    tourState: { ...INITIAL_STATE } as TourState,

    // UI state
    isActive: false,
    isOverviewTourActive: false,
    isProjectTourActive: false,
    showPopover: false,
    showCompletionModal: false,
    showExitConfirmModal: false,

    // Callback for project tour exit (to navigate to scheduling rules step)
    projectTourExitCallback: null as (() => void) | null,

    // Callback for project tour step navigation (to navigate wizard between steps)
    projectTourStepCallback: null as ((stepName: string) => void) | null,

    // Callback for project tour step preparation (to prepare DOM before showing step)
    projectTourPrepCallback: null as ((stepName: string) => void | Promise<void>) | null,

    // Callback for project tour creation (to trigger "Create Project & Continue")
    // Returns Promise<boolean> - true if project created successfully
    projectTourCreateCallback: null as (() => Promise<boolean>) | null,

    // Callback for project tour completion (to navigate wizard after tour ends)
    projectTourCompletionCallback: null as (() => void) | null,

    // Flag to track if project was created successfully (used by nextOn eventCondition)
    projectCreatedSuccessfully: false,

    // Department tour state (mirroring project tour)
    isDepartmentTourActive: false,

    // Callback for department tour exit (to navigate to scheduling rules step)
    departmentTourExitCallback: null as (() => void) | null,

    // Callback for department tour step navigation (to navigate wizard between steps)
    departmentTourStepCallback: null as ((stepName: string) => void) | null,

    // Callback for department tour step preparation (to prepare DOM before showing step)
    departmentTourPrepCallback: null as ((stepName: string) => void | Promise<void>) | null,

    // Callback for department tour creation (to trigger "Create Department & Continue")
    // Returns Promise<boolean> - true if department created successfully
    departmentTourCreateCallback: null as (() => Promise<boolean>) | null,

    // Callback for department tour completion (to navigate wizard after tour ends)
    departmentTourCompletionCallback: null as (() => void) | null,

    // Flag to track if department was created successfully (used by nextOn eventCondition)
    departmentCreatedSuccessfully: false,

    // Current step data
    currentStep: null as TourStep | null,
    currentElement: null as Element | null,
    popoverProps: {} as Partial<PopoverProps>,

    // Event listeners tracking
    eventListeners: {} as { [key: string]: (event: Event) => void },
    nextOnListeners: [] as { element: Element; event: string; handler: (event: Event) => void }[],

    // Guided tour conditions
    workloadTourAllowed: false as boolean,
    workloadTourTemplateName: "" as string,
  }),

  getters: {
    isInitialized(): boolean {
      return this.tourState.isInitialized;
    },
    activeIndex(): number | undefined {
      return this.tourState.activeIndex;
    },
    activeStep(): TourStep | null {
      return this.tourState.activeStep;
    },
    activeElement(): Element | null {
      return this.tourState.activeElement;
    },
    steps(): TourStep[] {
      return this.config.steps || [];
    },
    hasNext(): boolean {
      const currentIndex = this.tourState.activeIndex;
      return currentIndex !== undefined && !!this.steps[currentIndex + 1];
    },
    hasPrev(): boolean {
      const currentIndex = this.tourState.activeIndex;
      return currentIndex !== undefined && !!this.steps[currentIndex - 1];
    },
    isFirstStep(): boolean {
      return this.tourState.activeIndex === 0;
    },
    isLastStep(): boolean {
      return this.tourState.activeIndex !== undefined && this.tourState.activeIndex === this.steps.length - 1;
    },
    // Context for callbacks - computed for efficiency
    context(): TourContext {
      return {
        config: this.config,
        state: this.tourState,
      };
    },
  },
  actions: {
    // Configuration methods
    setConfig(config: TourConfig): void {
      this.config = { ...config };
    },
    setSteps(steps: TourStep[]): void {
      this.config = { ...this.config, steps };
    },
    // State management methods
    initialize(): void {
      if (this.tourState.isInitialized) return;

      this.tourState.isInitialized = true;
      this.isActive = true;
      document.body.classList.add("guided-tour-active");
      guidedTourService.lockScrolling(this.config, this.tourState);
      this.setupEventListeners();
    },
    reset(): void {
      this.clearEventListeners();
      this.clearNextOnListeners();

      // Stop dynamic element monitoring
      this.stopDynamicElementMonitoring();

      // Reset tour-specific flags
      this.isOverviewTourActive = false;
      this.isProjectTourActive = false;
      this.projectTourExitCallback = null;
      this.projectTourStepCallback = null;
      this.projectTourPrepCallback = null;
      this.projectTourCreateCallback = null;
      this.projectTourCompletionCallback = null;
      this.projectCreatedSuccessfully = false;

      // Reset department tour flags
      this.isDepartmentTourActive = false;
      this.departmentTourExitCallback = null;
      this.departmentTourStepCallback = null;
      this.departmentTourPrepCallback = null;
      this.departmentTourCreateCallback = null;
      this.departmentTourCompletionCallback = null;
      this.departmentCreatedSuccessfully = false;

      // Clean up current state
      guidedTourService.removeOverlay(this.tourState);

      // Clean up all tour-related event handlers
      guidedTourService.cleanupTourEventHandlers(this.tourState);

      // Reset DOM elements
      guidedTourService.cleanupActiveElements();
      document.body.classList.remove("guided-tour-active");
      document.getElementById("guided-tour-dummy-element")?.remove();

      // Unlock scrolling (this will also clean up any remaining handlers)
      guidedTourService.unlockScrolling(this.config, this.tourState);

      // Reset state
      this.tourState = { ...INITIAL_STATE };
      this.currentStep = null;
      this.currentElement = null;
      this.popoverProps = {};
      this.isActive = false;
      this.showPopover = false;
      this.showCompletionModal = false;
    },
    // Step navigation
    async highlightStep(stepIndex: number): Promise<void> {
      const step = this.steps[stepIndex];
      if (!step) {
        this.destroy();
        return;
      }

      // Handle beforeShowStep setup
      if (step.beforeShowStep) {
        try {
          await guidedTourService.handleBeforeShowStep(step, this.config, this.tourState);
        } catch (error) {
          console.error("beforeShowStep failed, cancelling tour:", error);
          // Cancel the tour if critical elements are not found
          this.destroy();
          return;
        }
      }

      // Get elements and update state
      const fromElement = this.tourState.activeElement;
      const fromStep = this.tourState.activeStep;

      let targetElement: Element;
      try {
        targetElement = guidedTourService.getTargetElement(step.element);
      } catch (error) {
        console.error("[Guided Tour Store] Critical: Cannot find target element for step:", error);
        console.log("[Guided Tour Store] Tour shutting down completely due to missing target element");
        this.destroy();
        return;
      }

      this.updateStepState(stepIndex, step, targetElement, fromElement, fromStep);

      // Handle deselection callback
      if (fromStep && fromElement) {
        const onDeselected = fromStep.onDeselected || this.config.onDeselected;
        if (onDeselected) {
          onDeselected(fromElement, fromStep, this.context);
        }
      }

      // Handle highlight started callback
      const onHighlightStarted = step.onHighlightStarted || this.config.onHighlightStarted;
      if (onHighlightStarted) {
        await onHighlightStarted(targetElement, step, this.context);
      }

      // Bring element into view
      await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
      if (targetElement) {
        guidedTourService.bringIntoView(targetElement, this.config);
      }

      // Handle animation transition
      if (this.config.animate && fromElement && targetElement && fromElement !== targetElement) {
        try {
          await this.animateTransition(fromElement, targetElement, step);
        } catch (error) {
          console.error("[Guided Tour Store] Animated transition failed:", error);
          console.log("[Guided Tour Store] Tour shutting down completely due to critical error");
          this.destroy();
          return;
        }
      } else {
        try {
          await this.completeTransition(targetElement, step);
        } catch (error) {
          console.error("[Guided Tour Store] Immediate transition failed:", error);
          console.log("[Guided Tour Store] Tour shutting down completely due to critical error");
          this.destroy();
          return;
        }
      }
    },
    updateStepState(
      stepIndex: number,
      step: TourStep,
      targetElement: Element,
      fromElement: Element | null,
      fromStep: TourStep | null,
    ): void {
      this.clearNextOnListeners();

      // Stop dynamic element monitoring from previous step
      this.stopDynamicElementMonitoring();

      // Clean up scroll allowance for previous element
      if (fromElement) {
        guidedTourService.removeElementScrollAllowance(fromElement, this.tourState);
      }

      this.tourState.previousElement = fromElement;
      this.tourState.previousStep = fromStep;
      this.tourState.activeIndex = stepIndex;
      this.tourState.activeStep = step;
      this.tourState.activeElement = targetElement;

      this.currentStep = step;
      this.currentElement = targetElement;
    },
    async animateTransition(fromElement: Element, toElement: Element, step: TourStep): Promise<void> {
      const duration = this.tourState.animationDuration;
      this.showPopover = false;

      return new Promise((resolve) => {
        const callback = async () => {
          await this.completeTransition(toElement, step);
          resolve();
        };

        guidedTourService.animateTransition(fromElement, toElement, duration, this.config, this.tourState, callback);
      });
    },
    async completeTransition(targetElement: Element, step: TourStep): Promise<void> {
      this.tourState.transitionCallback = null;

      // Don't show tour yet - wait for setupNextOn to complete first

      // Handle overlay visibility based on step.hideOverlay
      if (step.hideOverlay) {
        // Hide overlay, unlock scrolling, and enable full interaction for this step
        guidedTourService.removeOverlay(this.tourState);
        guidedTourService.unlockScrolling(this.config, this.tourState);
        document.body.classList.remove("guided-tour-active");
      } else {
        // Restore guided-tour-active class if it was previously removed
        document.body.classList.add("guided-tour-active");
        // Restore scroll locking if it was previously unlocked
        guidedTourService.lockScrolling(this.config, this.tourState);
        // Create or update overlay
        if (!this.tourState.overlayElement) {
          guidedTourService.createOverlay(targetElement, this.config, this.tourState, this.handleOverlayClick, step);
        } else {
          guidedTourService.updateOverlay(targetElement, this.config, this.tourState, step);
        }
      }

      // Setup popover
      if (step.popover) {
        this.setupPopover(targetElement, step);
      } else {
        this.showPopover = false;
      }

      // Apply element styling
      guidedTourService.applyElementStyling(targetElement, step, this.config);

      // Setup scroll allowance for the highlighted element
      guidedTourService.setupElementScrollAllowance(targetElement, this.config, this.tourState);

      // Handle onHighlighted callback
      const onHighlighted = step.onHighlighted || this.config.onHighlighted;
      if (onHighlighted) {
        onHighlighted(targetElement, step, this.context);
      }

      // Setup nextOn listeners and wait for completion
      try {
        const listeners = await guidedTourService.setupNextOn(
          targetElement,
          step,
          this.config,
          this.tourState,
          () => this.clearNextOnListeners(),
          () => this.hideTour(),
          () => this.moveNext(),
        );

        this.nextOnListeners.push(...listeners);

        // Initialize tracked elements set
        this.tourState.trackedElements = new Set();
        listeners.forEach(({ element }) => {
          this.tourState.trackedElements?.add(element);
        });

        // Start dynamic element monitoring if nextOn has selectors
        if (step.nextOn?.selector) {
          this.startDynamicElementMonitoring(step);
        }
      } catch (error) {
        console.error("[Guided Tour Store] setupNextOn failed:", error);

        // Show error to user and fail the tour
        this.showPopover = false;
        this.destroy();

        // You could also show a user-friendly error message here
        // For example: this.showErrorModal("Tour step failed to load properly");

        throw error; // Re-throw to prevent tour from continuing in broken state
      }

      // Now show the tour after everything is ready
      this.showTour();
    },
    setupPopover(targetElement: Element, step: TourStep): void {
      if (!step.popover) return;

      const currentIndex = this.tourState.activeIndex || 0;
      const hasNext = currentIndex < this.steps.length - 1;
      const hasPrev = currentIndex > 0;
      const showButtons = step.popover.showButtons || this.config.showButtons || [];

      let progressText = step.popover.progressText || this.config.progressText || "";
      progressText = progressText
        .replace("{{current}}", String(currentIndex + 1))
        .replace("{{total}}", String(this.steps.length));

      const targetSelector = guidedTourService.generateElementSelector(targetElement);
      const animationClass = this.config.animate ? "guided-tour-fade" : "";
      const popoverClass = this.config.popoverClass || "";
      const customClass = [animationClass, popoverClass].filter(Boolean).join(" ");

      this.popoverProps = {
        modelValue: true,
        anchor: guidedTourService.mapSideToAnchor(step.popover.side || "bottom", step.popover.align || "center"),
        self: guidedTourService.mapSideToSelf(step.popover.side || "bottom", step.popover.align || "center"),
        offset: step.popover.offset || this.config.popoverOffset || [10, 10],
        noParentEvent: true,
        target: targetSelector,
        step: step,
        title: step.popover.title,
        description: step.popover.description,
        showProgress: step.popover.showProgress ?? this.config.showProgress ?? false,
        progressText,
        showButtons,
        hasNext,
        hasPrev,
        nextBtnText: hasNext
          ? step.popover.nextBtnText || this.config.nextBtnText
          : step.popover.doneBtnText || this.config.doneBtnText,
        prevBtnText: step.popover.prevBtnText || this.config.prevBtnText,
        customClass: customClass,
      };

      this.showPopover = true;
    },
    setupNextOn(element: Element | null, step: TourStep): void {
      if (!step.nextOn) return;
    },
    // Navigation methods
    async drive(stepIndex = 0): Promise<void> {
      if (!this.steps || !this.steps[stepIndex]) {
        this.destroy();
        return;
      }

      if (!this.tourState.isInitialized) {
        this.initialize();
      }

      await this.highlightStep(stepIndex);
    },
    moveNext(): void {
      // Clear popup immediately for better UX
      this.showPopover = false;

      this.clearNextOnListeners();
      const currentIndex = this.tourState.activeIndex;
      if (currentIndex === undefined) return;

      const nextIndex = currentIndex + 1;
      if (this.steps[nextIndex]) {
        this.drive(nextIndex);
      } else {
        this.destroy();
      }
    },
    movePrevious(): void {
      // Clear popup immediately for better UX
      this.showPopover = false;

      this.clearNextOnListeners();
      const currentIndex = this.tourState.activeIndex;
      if (currentIndex === undefined) return;

      const prevIndex = currentIndex - 1;
      if (prevIndex >= 0) {
        this.drive(prevIndex);
      }
    },
    moveTo(index: number): void {
      // Clear popup immediately for better UX
      this.showPopover = false;

      this.clearNextOnListeners();
      if (this.steps[index]) {
        this.drive(index);
      }
    },
    // Event handlers
    onNext(): void {
      const element = this.tourState.activeElement;
      const step = this.tourState.activeStep;
      if (element && step) {
        this.handleNext(element, step);
      }
    },
    onPrevious(): void {
      const element = this.tourState.activeElement;
      const step = this.tourState.activeStep;
      if (element && step) {
        this.handlePrevious(element, step);
      }
    },
    onClose(): void {
      const element = this.tourState.activeElement;
      const step = this.tourState.activeStep;
      if (element && step) {
        this.handleClose(element, step);
      }
    },
    handleNext(element: Element | null, step: TourStep): void {
      const onNextClick = step.popover?.onNextClick || this.config.onNextClick;
      if (onNextClick) {
        onNextClick(element, step, this.context);
      } else {
        this.moveNext();
      }
    },
    handlePrevious(element: Element | null, step: TourStep): void {
      const onPrevClick = step.popover?.onPrevClick || this.config.onPrevClick;
      if (onPrevClick) {
        onPrevClick(element, step, this.context);
      } else {
        this.movePrevious();
      }
    },
    handleClose(element: Element | null, step: TourStep): void {
      const onCloseClick = step.popover?.onCloseClick || this.config.onCloseClick;
      if (onCloseClick) {
        onCloseClick(element, step, this.context);
      } else {
        this.showExitConfirmation();
      }
      // // cleanup active elements
      // guidedTourService.cleanupActiveElements();
    },
    handleOverlayClick(): void {
      const behavior = this.config.overlayClickBehavior || "ignore";

      if (behavior === "close" && this.config.allowClose) {
        this.destroy();
      } else if (behavior === "nextStep") {
        this.moveNext();
      }
    },
    // Event listeners management
    setupEventListeners(): void {
      if (this.config.allowKeyboardControl) {
        this.eventListeners.keydown = (event: Event) => this.handleKeydown(event);
        document.addEventListener("keydown", this.eventListeners.keydown);
      }
    },
    clearEventListeners(): void {
      if (this.config.allowKeyboardControl && this.eventListeners.keydown) {
        document.removeEventListener("keydown", this.eventListeners.keydown);
      }
      this.eventListeners = {};
    },
    clearNextOnListeners(): void {
      this.nextOnListeners.forEach(({ element, event, handler }) => {
        element.removeEventListener(event, handler);
      });
      this.nextOnListeners = [];
    },
    startDynamicElementMonitoring(step: TourStep): void {
      // Stop any existing monitoring
      this.stopDynamicElementMonitoring();

      // Start new monitoring
      this.tourState.elementMonitor = guidedTourService.startElementMonitoring(
        step,
        this.config,
        this.tourState,
        (newElements, removedElements) => this.handleElementChange(newElements, removedElements, step),
      );
    },
    stopDynamicElementMonitoring(): void {
      if (this.tourState.elementMonitor) {
        guidedTourService.stopElementMonitoring(this.tourState.elementMonitor);
        this.tourState.elementMonitor = null;
      }
    },
    handleElementChange(newElements: Element[], removedElements: Element[], step: TourStep): void {
      // Remove styling and listeners from removed elements
      if (removedElements.length > 0) {
        guidedTourService.removeStylingFromElements(removedElements);

        // Remove listeners from removed elements
        this.nextOnListeners = this.nextOnListeners.filter(({ element, event, handler }) => {
          const shouldKeep = !removedElements.includes(element);
          if (!shouldKeep) {
            element.removeEventListener(event, handler);
          }
          return shouldKeep;
        });

        // Update tracked elements
        removedElements.forEach((el) => {
          this.tourState.trackedElements?.delete(el);
        });
      }

      // Add styling and listeners to new elements
      if (newElements.length > 0) {
        guidedTourService.reapplyStylingToElements(newElements, step, this.config);

        // Re-attach listeners to new elements
        const newListeners = guidedTourService.reattachListenersToElements(
          newElements,
          step,
          this.config,
          this.tourState,
          () => this.clearNextOnListeners(),
          () => this.hideTour(),
          () => this.moveNext(),
        );

        this.nextOnListeners.push(...newListeners);

        // Update tracked elements
        newElements.forEach((el) => {
          this.tourState.trackedElements?.add(el);
        });

        // Update overlay to include new elements if they should be highlighted
        if (this.tourState.activeElement && this.tourState.overlayElement) {
          guidedTourService.updateOverlay(this.tourState.activeElement, this.config, this.tourState, step);
        }
      }
    },
    handleKeydown(event: Event): void {
      const keyEvent = event as KeyboardEvent;
      if (!this.tourState.isInitialized) return;

      switch (keyEvent.key) {
        case "Escape":
          if (this.config.allowClose) {
            this.showExitConfirmation();
          }
          break;
        case "ArrowRight":
          keyEvent.preventDefault();
          this.moveNext();
          break;
        case "ArrowLeft":
          keyEvent.preventDefault();
          this.movePrevious();
          break;
      }
    },
    // Tour visibility
    hideTour(): void {
      this.showPopover = false;
      if (this.tourState.overlayElement) {
        this.tourState.overlayElement.style.display = "none";
      }
      // cleanup active elements
      guidedTourService.cleanupActiveElements();
    },
    showTour(): void {
      if (this.tourState.overlayElement) {
        this.tourState.overlayElement.style.display = "block";
      }
    },
    refresh(): void {
      if (!this.tourState.isInitialized || this.tourState.activeIndex === undefined) return;
      this.highlightStep(this.tourState.activeIndex);
    },
    // Modal management
    showExitConfirmation(): void {
      this.showPopover = false;
      this.showExitConfirmModal = true;
    },
    hideExitConfirmation(): void {
      this.showExitConfirmModal = false;
      if (this.tourState.isInitialized && this.tourState.activeElement) {
        // Verify the element is still in the DOM before showing the popover
        if (!document.contains(this.tourState.activeElement)) {
          // Element no longer in DOM, clean up gracefully
          this.destroy();
          return;
        }
        this.showPopover = true;
      }
    },
    confirmExit(): void {
      // Use project tour exit handler if project tour is active
      if (this.isProjectTourActive && this.projectTourExitCallback) {
        this.confirmProjectTourExit();
        return;
      }
      // Use department tour exit handler if department tour is active
      if (this.isDepartmentTourActive && this.departmentTourExitCallback) {
        this.confirmDepartmentTourExit();
        return;
      }
      this.hideExitConfirmation();
      this.destroy();
    },
    closeCompletionModal(): void {
      this.showCompletionModal = false;
      // Execute completion callback for project tours (e.g., navigate back to GPU quota step)
      if (this.projectTourCompletionCallback) {
        const callback = this.projectTourCompletionCallback;
        this.projectTourCompletionCallback = null;
        callback();
      }
      // Execute completion callback for department tours (e.g., navigate back to GPU quota step)
      if (this.departmentTourCompletionCallback) {
        const callback = this.departmentTourCompletionCallback;
        this.departmentTourCompletionCallback = null;
        callback();
      }
    },
    openCompletionModal(): void {
      this.showCompletionModal = true;
    },
    // Cleanup
    destroy(): void {
      const activeElement = this.tourState.activeElement;
      const activeStep = this.tourState.activeStep;

      const onDestroyed = this.config.onDestroyed;
      if (onDestroyed && activeElement && activeStep) {
        onDestroyed(activeElement, activeStep, this.context);
      }

      this.reset();
    },
    // Tour instance factory
    createTourInstance(): GuidedTour {
      return {
        isActive: () => this.tourState.isInitialized,
        drive: (stepIndex = 0) => this.drive(stepIndex),
        destroy: () => this.destroy(),
        refresh: () => this.refresh(),
        hide: () => this.hideTour(),
        show: () => this.showTour(),
        setConfig: (newConfig: TourConfig) => this.setConfig(newConfig),
        setSteps: (steps: TourStep[]) => this.setSteps(steps),
        getConfig: () => this.config,
        getState: () => this.tourState,
        moveNext: () => this.moveNext(),
        movePrevious: () => this.movePrevious(),
        moveTo: (index: number) => this.moveTo(index),
        getActiveIndex: () => this.tourState.activeIndex,
        getActiveStep: () => this.tourState.activeStep,
        getActiveElement: () => this.tourState.activeElement,
        isFirstStep: () => this.isFirstStep,
        isLastStep: () => this.isLastStep,
        hasNextStep: () => this.hasNext,
        hasPreviousStep: () => this.hasPrev,
        highlight: (step: TourStep) => {
          if (!this.tourState.isInitialized) {
            this.initialize();
          }
          if (this.tourState.activeIndex !== undefined && this.tourState.activeStep === step) {
            this.highlightStep(this.tourState.activeIndex);
          } else {
            this.highlightStep(0);
          }
        },
      };
    },
    openWorkloadsGuidedTour(): void {
      this.setConfig(DEFAULT_CONFIG);
      this.setSteps(getWorkloadsGuidedTourSteps(this.workloadTourTemplateName));
      this.initialize();
      this.drive(0);
    },
    openOverviewGuidedTour(): void {
      this.isOverviewTourActive = true;
      this.setConfig(OVERVIEW_TOUR_CONFIG);
      this.setSteps(OVERVIEW_GUIDED_TOUR_STEPS);
      this.initialize();
      this.drive(0);
    },
    /**
     * Opens the project guided tour with dynamic step configuration
     * @param options Configuration options for the tour
     * @param onExitCallback Optional callback to execute when the tour is exited (e.g., navigate to scheduling rules)
     * @param onStepCallback Optional callback to execute when navigating between tour steps (e.g., navigate wizard)
     * @param onCompletionCallback Optional callback to execute when the tour is completed (e.g., navigate to GPU quota)
     * @param onPrepCallback Optional callback to execute to prepare DOM before showing a step
     * @param onCreateCallback Optional callback to trigger "Create Project & Continue" from tour
     */
    openProjectGuidedTour(
      options: ProjectTourOptions,
      onExitCallback?: () => void,
      onStepCallback?: (stepName: string) => void,
      onCompletionCallback?: () => void,
      onPrepCallback?: (stepName: string) => void,
      onCreateCallback?: () => Promise<boolean>,
    ): void {
      // Cancel any existing tour (including department tour when implemented)
      if (this.isActive) {
        this.destroy();
      }

      this.isProjectTourActive = true;
      this.projectTourExitCallback = onExitCallback ?? null;
      this.projectTourStepCallback = onStepCallback ?? null;
      this.projectTourPrepCallback = onPrepCallback ?? null;
      this.projectTourCreateCallback = onCreateCallback ?? null;
      this.projectTourCompletionCallback = onCompletionCallback ?? null;
      this.projectCreatedSuccessfully = false;

      const steps = getProjectGuidedTourSteps(options);
      this.setConfig(PROJECT_TOUR_CONFIG);
      this.setSteps(steps);
      this.initialize();
      this.drive(0);
    },
    /**
     * Navigate to a specific wizard step via the step callback (used by tour steps)
     * @param stepName The name of the wizard step to navigate to (e.g., 'over-quota')
     */
    navigateProjectWizardStep(stepName: string): void {
      if (this.projectTourStepCallback) {
        this.projectTourStepCallback(stepName);
      }
    },
    /**
     * Prepare DOM for a specific tour step via the prep callback (used by tour steps)
     * @param stepName The name of the step to prepare (e.g., 'gpu-quota', 'over-quota')
     */
    async prepareProjectWizardStep(stepName: string): Promise<void> {
      if (this.projectTourPrepCallback) {
        await this.projectTourPrepCallback(stepName);
      }
    },
    /**
     * Trigger project creation via the create callback (used by tour step 0)
     * Returns true if project created successfully, false otherwise
     */
    async triggerProjectCreate(): Promise<boolean> {
      if (this.projectTourCreateCallback) {
        const success = await this.projectTourCreateCallback();
        if (success) {
          this.projectCreatedSuccessfully = true;
        }
        return success;
      }
      return false;
    },
    /**
     * Mark project as created successfully (used by nextOn eventCondition)
     */
    markProjectCreated(): void {
      this.projectCreatedSuccessfully = true;
    },
    /**
     * Confirms exit from the tour and executes the exit callback if set
     * Overrides the default confirmExit behavior for project tour
     */
    confirmProjectTourExit(): void {
      const callback = this.projectTourExitCallback;
      this.hideExitConfirmation();
      this.destroy();
      // Execute the callback after destroy to navigate to the correct step
      if (callback) {
        callback();
      }
    },
    /**
     * Opens the department guided tour with dynamic step configuration
     * @param options Configuration options for the tour
     * @param onExitCallback Optional callback to execute when the tour is exited (e.g., navigate to scheduling rules)
     * @param onStepCallback Optional callback to execute when navigating between tour steps (e.g., navigate wizard)
     * @param onCompletionCallback Optional callback to execute when the tour is completed (e.g., navigate to GPU quota)
     * @param onPrepCallback Optional callback to execute to prepare DOM before showing a step
     * @param onCreateCallback Optional callback to trigger "Create Department & Continue" from tour
     */
    openDepartmentGuidedTour(
      options: DepartmentTourOptions,
      onExitCallback?: () => void,
      onStepCallback?: (stepName: string) => void,
      onCompletionCallback?: () => void,
      onPrepCallback?: (stepName: string) => void,
      onCreateCallback?: () => Promise<boolean>,
    ): void {
      // Cancel any existing tour (including project tour)
      if (this.isActive) {
        this.destroy();
      }

      this.isDepartmentTourActive = true;
      this.departmentTourExitCallback = onExitCallback ?? null;
      this.departmentTourStepCallback = onStepCallback ?? null;
      this.departmentTourPrepCallback = onPrepCallback ?? null;
      this.departmentTourCreateCallback = onCreateCallback ?? null;
      this.departmentTourCompletionCallback = onCompletionCallback ?? null;
      this.departmentCreatedSuccessfully = false;

      const steps = getDepartmentGuidedTourSteps(options);
      this.setConfig(DEPARTMENT_TOUR_CONFIG);
      this.setSteps(steps);
      this.initialize();
      this.drive(0);
    },
    /**
     * Navigate to a specific wizard step via the step callback (used by department tour steps)
     * @param stepName The name of the wizard step to navigate to (e.g., 'over-quota')
     */
    navigateDepartmentWizardStep(stepName: string): void {
      if (this.departmentTourStepCallback) {
        this.departmentTourStepCallback(stepName);
      }
    },
    /**
     * Prepare DOM for a specific department tour step via the prep callback (used by tour steps)
     * @param stepName The name of the step to prepare (e.g., 'gpu-quota', 'over-quota')
     */
    async prepareDepartmentWizardStep(stepName: string): Promise<void> {
      if (this.departmentTourPrepCallback) {
        await this.departmentTourPrepCallback(stepName);
      }
    },
    /**
     * Trigger department creation via the create callback (used by tour step 0)
     * Returns true if department created successfully, false otherwise
     */
    async triggerDepartmentCreate(): Promise<boolean> {
      if (this.departmentTourCreateCallback) {
        const success = await this.departmentTourCreateCallback();
        if (success) {
          this.departmentCreatedSuccessfully = true;
        }
        return success;
      }
      return false;
    },
    /**
     * Mark department as created successfully (used by nextOn eventCondition)
     */
    markDepartmentCreated(): void {
      this.departmentCreatedSuccessfully = true;
    },
    /**
     * Confirms exit from the department tour and executes the exit callback if set
     * Overrides the default confirmExit behavior for department tour
     */
    confirmDepartmentTourExit(): void {
      const callback = this.departmentTourExitCallback;
      this.hideExitConfirmation();
      this.destroy();
      // Execute the callback after destroy to navigate to the correct step
      if (callback) {
        callback();
      }
    },
    async setWorkloadTourAllowed(): Promise<void> {
      const permissionStore = usePermissionStore();
      const settingStore = useSettingStore();
      const hasCreateWorkloadPermission = permissionStore.hasPermission(ResourceType.Workloads, Action.Create);
      const hasReadProjectPermission = permissionStore.hasPermission(ResourceType.Project, Action.Read);
      const hasReadTemplatesPermission = permissionStore.hasPermission(ResourceType.Templates, Action.Read);

      if (!hasCreateWorkloadPermission || !hasReadProjectPermission || !hasReadTemplatesPermission) {
        this.workloadTourAllowed = false;
        return;
      }

      const flexSubmissionSupported =
        useClusterStore().isAllClustersSupportSpecificVersion(MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION) &&
        settingStore.isHybridSubmissionEnabled;

      // check that flex templates enabled
      const flexTemplatesEnabled = settingStore.isFeatureEnabled(SettingKeys.EnableTemplatesV2);

      if (!flexSubmissionSupported || !flexTemplatesEnabled) {
        this.workloadTourAllowed = false;
        return;
      }

      try {
        // Check that one of the jupyter lab template presets exists under system scope.
        // Path-based clusters use "jupyter-lab-one-gpu", host-based use "jupyter-lab-1-gpu".
        // Prefer the host-based name (the newer default) so host-based clusters pick the
        // right variant even when the tenant also has a legacy path-based preset seeded.
        // If the template was updated it would no longer be system scope (as the backend changes them to tenant scope).
        const jupyterTemplateNames = ["jupyter-lab-1-gpu", "jupyter-lab-one-gpu"];
        const workspaceSystemTemplates = await workloadTemplateService.workspaceTemplatesList(
          undefined,
          undefined,
          undefined,
          undefined,
          ["scopeType==system"],
        );

        const templateNamesInResponse = new Set(workspaceSystemTemplates?.entries?.map((t) => t.meta.name) ?? []);
        const preferredTemplateName = jupyterTemplateNames.find((name) => templateNamesInResponse.has(name));
        const jupyterTemplate = preferredTemplateName
          ? workspaceSystemTemplates?.entries?.find((t) => t.meta.name === preferredTemplateName)
          : undefined;

        if (!jupyterTemplate) {
          console.warn("guided workloads tour required template not found");
          this.workloadTourAllowed = false;
          return;
        }

        this.workloadTourTemplateName = jupyterTemplate.meta.name;

        const projectsCount = await orgUnitService.countProjects({});
        if (projectsCount <= 0) {
          console.warn("guided workloads tour requires at least one project");
          this.workloadTourAllowed = false;
          return;
        }

        this.workloadTourAllowed = true;
      } catch (error) {
        console.warn("Error checking guided workloads tour conditions", error);
        this.workloadTourAllowed = false;
      }
    },
  },
});
