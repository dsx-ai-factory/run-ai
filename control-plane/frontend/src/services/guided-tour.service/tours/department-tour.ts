import type { TourStep } from "@/components/common/guided-tour-popup/guided-tour.types";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { guidedTourService } from "../guided-tour.service";
import { waitForCondition } from "../guided-tour.utils";
import { EStepNames } from "@/models/org-unit-forms.model";

/**
 * Configuration options for generating department guided tour steps
 */
export interface DepartmentTourOptions {
  /** Whether the department has multiple node pools (not just default) */
  hasMultipleNodePools: boolean;
  /** Whether CPU resources quota is enabled */
  isCPUResourcesQuotaEnabled: boolean;
  /** Whether over quota priority (weights) is enabled */
  isOverQuotaPriorityEnabled: boolean;
}

/**
 * Generate department guided tour steps based on configuration options
 * Steps are dynamically built based on node pool count and feature flags
 */
export const getDepartmentGuidedTourSteps = (options: DepartmentTourOptions): TourStep[] => {
  const steps: TourStep[] = [];

  // Step 0: Scope & Name step - form is interactive, no overlay, popup above action buttons
  steps.push({
    element: ".wizard-content-footer .q-card__section",
    hideOverlay: true,
    disableActiveInteraction: false,
    beforeShowStep: {
      waitForElement: ".wizard-content-footer .q-card__section",
    },
    popover: {
      description:
        "Set your department's scope and name.<br />Create the department and close to proceed with the system defaults or continue to configure custom resources, scheduling, and access rules.",
      side: "top",
      align: "center",
      offset: [0, 2],
      showProgress: true,
      showButtons: ["close"],
    },
    // Also advance when clicking "Create Department & Continue" button directly (waits for success)
    nextOn: {
      event: "click",
      selector: "[aid='wizard-primary-btn']",
      // Wait for department creation to succeed (store flag set by component on success)
      eventCondition: async () => {
        return await waitForCondition(() => useGuidedTourStore().departmentCreatedSuccessfully, {
          timeout: 10000,
          interval: 100,
          description: "department creation success",
        });
      },
    },
  });

  // Step 1: Node pool tabs (only if multiple node pools exist)
  if (options.hasMultipleNodePools) {
    steps.push({
      element: ".node-pool-tabs-container",
      disableActiveInteraction: true,
      beforeShowStep: {
        waitForElement: ".node-pool-tabs-container",
        setup: async () => {
          // Wait for component to mount and Vue to bind refs
          await guidedTourService.waitForElement(".gpu-quota-card", undefined, 3000);
          await new Promise((resolve) => setTimeout(resolve, 300));
          useGuidedTourStore().prepareDepartmentWizardStep(EStepNames.GpuQuota);
        },
      },
      popover: {
        description:
          "Set up node pools and their resources. Add, remove, and arrange them in the order workloads should use them.",
        side: "bottom",
        align: "start",
        offset: [0, 10],
        showProgress: true,
      },
    });
  }

  // Step 2: GPU quota input + node pool overview
  steps.push({
    element: ".gpu-quota-card .quota-input-container",
    disableActiveInteraction: true,
    beforeShowStep: {
      waitForElement: ".gpu-quota-card .quota-input-container",
      setup: async () => {
        // Wait for component to mount and Vue to bind refs
        await guidedTourService.waitForElement(".gpu-quota-card", undefined, 3000);
        await new Promise((resolve) => setTimeout(resolve, 300));
        useGuidedTourStore().prepareDepartmentWizardStep(EStepNames.GpuQuota);
      },
    },
    popover: {
      description:
        "Set your department's resources. View and adjust other departments' quotas to free up and reallocate resources.",
      side: "bottom",
      align: "center",
      offset: [0, 10],
      showProgress: true,
    },
  });

  // Step 3: Over-quota limit input (in GPU over-quota step)
  steps.push({
    element: "[aid='over-quota-toggle'] .q-toggle__inner",
    disableActiveInteraction: true,
    beforeShowStep: {
      setup: async () => {
        useGuidedTourStore().navigateDepartmentWizardStep(EStepNames.OverQuota);
        await new Promise((resolve) => setTimeout(resolve, 200));
        useGuidedTourStore().prepareDepartmentWizardStep(EStepNames.OverQuota);
        await new Promise((resolve) => setTimeout(resolve, 300));
      },
      waitForElement: "[aid='over-quota-toggle'] .q-toggle__inner",
    },
    popover: {
      description: "Decide whether your department can go over its quota, and if so, set its over-quota limit.",
      side: "bottom",
      align: "center",
      offset: [0, 10],
      showProgress: true,
    },
  });

  // Step 4: Rank/priority description text
  steps.push({
    element: "[aid='rank-description-text']",
    disableActiveInteraction: true,
    beforeShowStep: {
      waitForElement: "[aid='rank-description-text']",
    },
    popover: {
      description:
        "Rank your department. Higher-ranked departments will receive additional resources as they become available, and their workloads will be prioritized.",
      side: "top",
      align: "start",
      offset: [0, 10],
      showProgress: true,
    },
  });

  // Step 5: Weight configuration - target the current department row in the rank table
  const step5Description = options.isOverQuotaPriorityEnabled
    ? "Set your department's weight. This determines what portion of spare resources departments with the same ranking receive during a given period."
    : "Weight is derived from a department's assigned quota and is a secondary factor in allocating spare resources. It determines what portion departments with the same ranking receive during a given period.";

  steps.push({
    element: ".rank-org-unit-rank-table .runai-table-row.selected .cell-trigger",
    disableActiveInteraction: true,
    beforeShowStep: {
      waitForElement: ".rank-org-unit-rank-table .runai-table-row.selected",
    },
    popover: {
      description: step5Description,
      side: "top",
      align: "center",
      offset: [0, 10],
      showProgress: true,
    },
  });

  // Step 6: Continue button with dynamic message based on CPU quota
  const step6Description = options.isCPUResourcesQuotaEnabled
    ? "Next you'll be able to configure CPU quota, scheduling and access rules or close the form."
    : "Next you'll be able to configure scheduling and access rules or close the form.";

  steps.push({
    element: ".q-btn[aid='wizard-primary-btn']",
    disableActiveInteraction: true,
    beforeShowStep: {
      waitForElement: ".q-btn[aid='wizard-primary-btn']",
    },
    popover: {
      description: step6Description,
      side: "top",
      align: "center",
      offset: [0, 25],
      showProgress: true,
      doneBtnText: "Done",
      onNextClick: () => {
        const store = useGuidedTourStore();
        // Save the completion callback before destroy clears it
        const completionCallback = store.departmentTourCompletionCallback;
        // Kill the tour
        store.destroy();
        // Restore callback for completion modal
        store.departmentTourCompletionCallback = completionCallback;
        // Open completion modal
        store.openCompletionModal();
      },
    },
  });

  return steps;
};
