/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TourStep } from "@/components/common/guided-tour-popup/guided-tour.types";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { guidedTourService } from "./guided-tour.service";

// Re-export project and department tours
export { getProjectGuidedTourSteps, type ProjectTourOptions } from "./tours";
export { getDepartmentGuidedTourSteps, type DepartmentTourOptions } from "./tours";

// Example of how to configure scroll locking for a guided tour
// You can disable scroll locking for specific tours if needed:
// const tourConfig = { lockScrolling: false };
// store.setConfig(tourConfig);

// Example of how to configure scroll behavior for a guided tour
// You can customize how elements are positioned when scrolled into view:
// const tourConfig = {
//   scrollBehavior: "center", // Options: "center" (default), "start", "end", "nearest", "smart"
//   smoothScroll: true        // Enable smooth scrolling for better UX
// };
// store.setConfig(tourConfig);

// Example of how to use the waitForCondition utility for delayed conditions:
// eventCondition: async (event, element) => {
//   return await guidedTourService.waitForCondition(
//     () => /* your condition here */,
//     10000, // timeout in ms (optional, default: 10000)
//     100    // check interval in ms (optional, default: 100)
//   );
// };

// Example of how to use timeout properties for delayed conditions:
// nextOn: {
//   event: "click",
//   selector: ".my-button",
//   timeout: 10000,        // 10 seconds timeout
//   checkInterval: 100,    // Check every 100ms
//   eventCondition: async (event, element) => {
//     // Your condition logic here
//     return someCondition === true;
//   },
// };
//
// Note: The timeout system automatically cleans up previous loops when:
// - A new eventCondition is triggered (prevents multiple concurrent loops)
// - The tour advances to the next step
// - The tour is destroyed or cleaned up

// DYNAMIC ELEMENT HANDLING
// ========================
// The guided tour now automatically handles dynamic content changes:
//
// 1. AUTOMATIC MONITORING: When a step uses nextOn with selectors, the tour
//    automatically monitors the DOM for changes using MutationObserver
//
// 2. ELEMENT RE-DETECTION: When elements matching the selectors appear or
//    disappear (e.g., due to filtering, pagination, or dynamic loading),
//    the tour automatically detects these changes
//
// 3. LISTENER RE-ATTACHMENT: Event listeners are automatically re-attached
//    to newly appeared elements that match the selectors
//
// 4. STYLING RE-APPLICATION: Guided tour styling (highlighting, accessibility
//    attributes) is automatically re-applied to newly appeared elements
//
// 5. STALE CLEANUP: Event listeners and styling are automatically removed
//    from elements that are no longer in the DOM or no longer match selectors
//
// This means your tour will continue to work even when:
// - Users filter a gallery and elements disappear/reappear
// - Content is loaded dynamically via AJAX/fetch
// - Elements are re-rendered by frameworks like Vue/React
// - Pagination changes the visible elements
//
// Example of a step that benefits from dynamic handling:
// {
//   element: '.gallery-item',
//   nextOn: {
//     event: 'click',
//     selector: '.gallery-item:not(.filtered)', // Only clickable when not filtered
//     attachToAllElements: true,                // Monitor all matching elements
//     eventCondition: async (event, element) => {
//       // This will work even if the element was just added to the DOM
//       return !element.classList.contains('loading');
//     }
//   }
// }

export function getWorkloadsGuidedTourSteps(templateName: string): TourStep[] {
  const templateCardSelector = `[aid="${templateName}-card"]`;

  return [
    // Step 1: Click create workload dropdown
    {
      element: '[aid="primary-cta-btn"]',
      beforeShowStep: {
        waitForElement: '[aid="primary-cta-btn"]:not([disabled])',
      },
      popover: {
        title:
          '<span class="text-weight-regular">Welcome to NVIDIA Run:ai! <br/> Let\'s get started by submitting your first Jupyter Notebook.</span>',
        description: '<span class="text-bold">Click + NEW WORKLOAD to begin.</span>',
        side: "right",
        offset: [15, 0], // Custom offset: 15px to the right
      },
      nextOn: { event: "click" },
    },
    // Step 2: Select workspace option from dropdown
    {
      element: '[aid="Workspace-btn"]',
      popover: {
        title: "Select Workspace",
        description: "A workspace is an interactive workload for development and experimentation.",
        side: "right",
        offset: [10, 10], // Custom offset: [x, y] - 20px above the element
      },
      nextOn: {
        event: "click",
        hideOnAction: true,
      },
    },
    // Step 3: Select project (after routing to workspace-new) - Only if multiple projects exist
    {
      element: '[aid="project-section"] .cards-list-container',
      accessibleElements: ["[aid='project-section'] .pagination-container"],
      beforeShowStep: {
        waitForElement: '[aid="project-section"] .runai-card-list .q-card.card-content',
      },
      popover: {
        title: "Select a project",
        description:
          "Projects are created by your administrator for the team or initiative you're part of. Your workloads will run under the project you select.",
        side: "right",
        offset: [20, 20], // Custom offset: [x, y] - 20px above the element
      },
      nextOn: {
        event: "click",
        selector: '[aid="project-section"] .runai-card-list .q-card.card-content',
        attachToAllElements: true,
        hideOnAction: true,
      },
    },
    // Step 4: Select Jupyter Lab template
    {
      element: ".template-section .cards-list-container",
      accessibleElements: ["[aid='template-section-v2'] .pagination-container"],
      beforeShowStep: {
        waitForElement: templateCardSelector,
        setup: async () => {
          // Wait for search setup to complete first
          await new Promise<void>((resolve) => {
            setTimeout(() => {
              const templatesGallerySearchBtn = document.querySelector('[aid="-search-card-input"]') as HTMLElement;
              if (!templatesGallerySearchBtn) {
                console.warn("Templates search input not found");
                resolve();
                return;
              }

              // Click to focus the search input
              templatesGallerySearchBtn.click();

              // Find the input element
              const input = templatesGallerySearchBtn.querySelector("input") as HTMLInputElement;
              if (!input) {
                console.warn("Input element not found within search container");
                resolve();
                return;
              }

              // Set the value
              input.value = templateName;

              // Trigger event input event so the filter would initiate a new list
              input.dispatchEvent(new Event("input", { bubbles: true, cancelable: true }));

              // Focus the input to ensure it's active
              input.focus();

              resolve();
            }, 100);
          });

          // Now wait for the element to appear after filtering is complete
          await guidedTourService.waitForElement(templateCardSelector);
        },
      },
      popover: {
        description: `Select the <span class="text-bold">${templateName} </span>template to launch a ready-to-use notebook.`,
        side: "right",
        align: "center",
        offset: [20, 20],
      },
      nextOn: {
        event: "click",
        selector: templateCardSelector,
      },
    },
    // Step 5: Give workspace a name
    {
      element: ".quick-submission-button .q-btn:first-of-type",
      accessibleElements: ["[aid='ws-name-section']"],
      beforeShowStep: {
        waitForElement: ".quick-submission-button",
      },
      popover: {
        description:
          'Give your workspace a <span class="text-bold">name</span> and <span class="text-bold">Click CREATE WORKSPACE</span>',
        side: "right",
        offset: [10, 20], // Custom offset: [x, y] - 20px above the element
      },
      nextOn: {
        event: "click",
        selector: ".quick-submission-button .q-btn:first-of-type",
        hideOnAction: true,
        timeout: 10000, // 10 seconds timeout
        checkInterval: 100, // Check every 100ms
        eventCondition: async () => {
          // check that we are now in the workloads index page
          return document.querySelector(".workload-index") !== null;
        },
      },
    },
    //step 7: select workload
    {
      element: ".runai-table-row .q-checkbox",
      beforeShowStep: {
        waitForElement: ".runai-table-row",
      },
      popover: {
        description: 'Select the <span class="text-bold">workload</span> you just created',
        side: "top",
        offset: [5, 20], // Custom offset: [x, y] - 20px above the element
      },
      nextOn: {
        event: "click",
        selector: ".runai-table-row .q-checkbox",
        hideOnAction: true,
        eventCondition: async (event, element) => {
          // Small delay to let the checkbox state update
          await new Promise((resolve) => setTimeout(resolve, 100));
          const checkboxInner = element.querySelector(".q-checkbox__inner");
          return (
            checkboxInner?.classList.contains("q-checkbox__inner--truthy") ||
            element.getAttribute("aria-checked") === "true"
          );
        },
      },
    },
    // Step 8: Final step - connect workload tool
    {
      element: '[aid="connect-workload-btn"]',
      popover: {
        description:
          'Once the workload\'s status changes to <span class="text-primary">Running</span>,<br/> click <span class="text-bold">Connect</span> to start using your Jupyter Notebook',
        side: "right",
        offset: [5, 35], // Custom offset: 10px left, 25px below
      },
      nextOn: {
        event: "click",
        onAdvance: () => {
          // Kill the tour
          useGuidedTourStore().destroy();
          // Open completion modal immediately
          useGuidedTourStore().openCompletionModal();
        },
      },
    },
  ];
}

/*** OVERVIEW_GUIDED_TOUR_STEPS ***/
export const OVERVIEW_GUIDED_TOUR_STEPS: TourStep[] = [
  {
    element: ".runai-page-filter",
    popover: {
      description:
        "Let's explore the dashboard and help you get the most of it. <br/> <br/> Filter to monitor and investigate by node pool. ",
      side: "right",
      align: "end",
      offset: [20, 20],
      showProgress: true,
    },
  },
  {
    element: "[aid='workload-group-by-select']",
    popover: {
      description: "Customize the view by selecting different segments or sorting options.",
      side: "right",
      offset: [80, 15],
      showProgress: true,
    },
  },
  {
    element: '[aid="workload-by-type-drilldown-label"]',
    beforeShowStep: {
      setup: async () => {
        // Wait for Vue to react to activeIndex change and render the mock skeleton
        await new Promise((resolve) => setTimeout(resolve, 150));
      },
      waitForElement: '[aid="workload-by-type-drilldown-label"]',
    },
    popover: {
      description: "Drill down to view detailed department and project data.",
      side: "right",
      showProgress: true,
    },
  },
  {
    element: "[aid='show-the-last-filter-chip']",
    popover: {
      description: "View averages over different time frames.",
      side: "right",
      align: "center",
      offset: [10, 10],
      showProgress: true,
    },
  },
  {
    element: "[aid='ranking-redirect-btn']",
    popover: {
      description: "For a full breakdown, visit the dedicated page and view additional metrics and details.",
      side: "right",
      align: "center",
      offset: [10, 10],
      showProgress: true,
    },
  },
  {
    element: "[aid='overtime-widget-date-picker']",
    popover: {
      description: "Change the date range to view data trends over time.",
      side: "right",
      align: "center",
      offset: [10, 10],
      showProgress: true,
      onNextClick: () => {
        // Kill the tour
        useGuidedTourStore().destroy();
        // Open completion modal immediately
        useGuidedTourStore().openCompletionModal();
      },
    },
  },
];
