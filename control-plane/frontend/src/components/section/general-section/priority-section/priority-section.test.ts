/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, test, vi, beforeEach, afterEach } from "vitest";

import { PrioritySection } from ".";
import { RunaiSelect } from "@/components/common/runai-select";
import { prioritySectionMock } from "./priority-section.mock";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

const {
  MOCK_PRIORITY_NAMES,
  mockWorkloadPriorities,
  mockPolicyRules,
  mockDefaultFactoryProps,
  mockProject,
  mockWorkloadPriorityDataPending,
} = prioritySectionMock;

describe("priority-section.vue", () => {
  let prioritySection: VueWrapper<InstanceType<typeof PrioritySection>>;

  function factory(props = {}): VueWrapper<InstanceType<typeof PrioritySection>> {
    const defaultProps = {
      ...mockDefaultFactoryProps,
      ...props,
    };

    const wrapper = mount(PrioritySection, {
      global: {
        plugins: [Quasar],
      },
      props: defaultProps,
    });

    wrapper.vm.$q.notify = vi.fn();
    return wrapper;
  }

  beforeEach(() => {
    prioritySection = factory();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe("Priority Options & Computation", () => {
    test("Should compute workloadPrioritiesToDisplay correctly", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      // Should be sorted by value
      expect(displayedPriorities.length).toBeGreaterThan(0);

      // Check sorting
      for (let i = 0; i < displayedPriorities.length - 1; i++) {
        const currentPriority = mockWorkloadPriorities.find((p) => p.name === displayedPriorities[i].value);
        const nextPriority = mockWorkloadPriorities.find((p) => p.name === displayedPriorities[i + 1].value);
        if (currentPriority && nextPriority) {
          expect(currentPriority.value).toBeLessThanOrEqual(nextPriority.value);
        }
      }
    });

    test("Should compute selectedPriorityOption correctly", () => {
      const selectedOption = prioritySection.vm.selectedPriorityOption;
      expect(selectedOption).toBeDefined();
      expect(selectedOption?.value).toBe(MOCK_PRIORITY_NAMES.MEDIUM_LOW);
    });

    test("Should include correct labels from policy rules", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      // Find options that should have custom labels from demo policy rules
      const highPriorityOption = displayedPriorities.find((opt) => opt.value === "high");
      const lowPriorityOption = displayedPriorities.find((opt) => opt.value === "low");

      expect(highPriorityOption?.label).toBe("Train Priority");
      expect(lowPriorityOption?.label).toBe("Build Priority");
    });

    test("Should add tooltip for non-preemptible priorities that are not disabled", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      const nonPreemptibleOptions = displayedPriorities.filter((opt) => {
        const priority = mockWorkloadPriorities.find((p) => p.name === opt.value);
        const policyOption = prioritySection.vm.policyRules?.options?.find((option) => option.value === priority?.name);
        const isDisabled = prioritySection.vm.policyRules?.options ? !policyOption : false;
        return priority && priority.value >= 100 && !isDisabled;
      });

      nonPreemptibleOptions.forEach((option) => {
        expect(option.toolTip).toBe(prioritySection.vm.nonPreemptibleTooltip);
      });
    });
  });

  describe("Select Component Functionality", () => {
    test("Should emit update:modelValue when priority changes", async () => {
      const selectComponent = prioritySection.findComponent(RunaiSelect);
      expect(selectComponent.exists()).toBeTruthy();

      const newOption = { label: "Test Priority", value: "test-priority" };
      await selectComponent.vm.$emit("update:model-value", newOption);

      expect(prioritySection.emitted("update:model-value")).toBeTruthy();
      const emittedValues = prioritySection.emitted("update:model-value") as string[][];
      expect(emittedValues[0][0]).toBe("test-priority");
    });

    test("Should pass correct props to RunaiSelect", () => {
      const selectComponent = prioritySection.findComponent(RunaiSelect);

      expect(selectComponent.props("options")).toEqual(prioritySection.vm.workloadPrioritiesToDisplay);
      expect(selectComponent.props("policyRules")).toEqual(prioritySection.vm.selectGenericPolicyRules);
      expect(selectComponent.props("disable")).toBe(prioritySection.vm.isSectionDisabled);
      expect(selectComponent.props("label")).toBe("Priority");
    });
  });

  describe("Policy Rules and Restrictions", () => {
    test("Should disable section when canEdit is false", async () => {
      await prioritySection.setProps({
        policyRules: mockPolicyRules.nonEditable,
      });

      expect(prioritySection.vm.isSectionDisabled).toBeTruthy();

      const selectComponent = prioritySection.findComponent(RunaiSelect);
      expect(selectComponent.props("disable")).toBeTruthy();
    });

    test("Should disable section when disabled prop is true", async () => {
      await prioritySection.setProps({
        disabled: true,
      });

      expect(prioritySection.vm.isSectionDisabled).toBeTruthy();

      const selectComponent = prioritySection.findComponent(RunaiSelect);
      expect(selectComponent.props("disable")).toBeTruthy();
    });
    test("Should use displayed names from policy rules", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      const highOption = displayedPriorities.find((opt) => opt.value === "high");
      const veryHighOption = displayedPriorities.find((opt) => opt.value === "very-high");

      expect(highOption?.label).toBe("Train Priority");
      expect(veryHighOption?.label).toBe("Train Priority High");
    });

    test("Should fallback to priority name when no displayed value", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      // Find priorities that don't have displayed values in policy rules
      const mediumLowOption = displayedPriorities.find((opt) => opt.value === "medium-low");
      const mediumOption = displayedPriorities.find((opt) => opt.value === "medium");
      const mediumHighOption = displayedPriorities.find((opt) => opt.value === "medium-high");

      // These priorities should fallback to their name as label since they have no displayed value
      expect(mediumLowOption?.label).toBe("medium-low");
      expect(mediumOption?.label).toBe("medium");
      expect(mediumHighOption?.label).toBe("medium-high");

      // Verify that all options have some label (either displayed or fallback)
      expect(displayedPriorities.every((opt) => opt.label !== undefined && opt.label !== "")).toBeTruthy();
    });
  });

  describe("Component Props", () => {
    test("Should handle modelValue prop correctly", async () => {
      await prioritySection.setProps({
        modelValue: MOCK_PRIORITY_NAMES.HIGH,
      });

      const selectedOption = prioritySection.vm.selectedPriorityOption;
      expect(selectedOption?.value).toBe(MOCK_PRIORITY_NAMES.HIGH);
    });

    test("Should handle null workloadPriorities", async () => {
      await prioritySection.setProps({
        workloadPriorityData: {
          ...mockWorkloadPriorityDataPending,
          workloadPriorities: undefined,
        },
      });

      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;
      expect(displayedPriorities).toEqual([]);
    });

    test("Should respect disabled prop", async () => {
      await prioritySection.setProps({
        disabled: true,
      });

      expect(prioritySection.vm.isSectionDisabled).toBeTruthy();
    });
  });

  describe("Tooltip Configuration", () => {
    test("Should have non-preemptible tooltip in data", () => {
      expect(prioritySection.vm.nonPreemptibleTooltip).toBe(
        "This priority makes the workload non-preemptible, which means it will use the project's available GPU quota and won't be interrupted once it starts running.",
      );
    });

    test("Should apply tooltip to non-preemptible priorities", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;

      displayedPriorities.forEach((option) => {
        const priority = mockWorkloadPriorities.find((p) => p.name === option.value);
        if (priority && priority.value >= 100) {
          expect(option.toolTip).toBe(prioritySection.vm.nonPreemptibleTooltip);
        } else {
          expect(option.toolTip).toBeUndefined();
        }
      });
    });
  });

  describe("Select Configuration", () => {
    test("Should configure select with correct properties", () => {
      const selectComponent = prioritySection.findComponent(RunaiSelect);

      expect(selectComponent.props("modelValue")).toEqual(prioritySection.vm.selectedPriorityOption);
      expect(selectComponent.props("options")).toEqual(prioritySection.vm.workloadPrioritiesToDisplay);
      expect(selectComponent.props("policyRules")).toEqual(prioritySection.vm.selectGenericPolicyRules);
      expect(selectComponent.props("disable")).toBe(prioritySection.vm.isSectionDisabled);
      expect(selectComponent.props("label")).toBe("Priority");
    });
  });

  describe("exceedsAvailableQuota Service Method", () => {
    test("Should return false for priority values less than 100", () => {
      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 50,
        currProject: mockProject,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 5,
      });
      expect(result).toBe(false);
    });

    test("Should return false when currProject is missing", () => {
      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: undefined,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 5,
      });
      expect(result).toBe(false);
    });

    test("Should return false when selectedNodePoolsResources is empty", () => {
      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: mockProject,
        selectedNodePoolsResources: [],
        totalGpuRequest: 5,
      });
      expect(result).toBe(false);
    });

    test("Should return false when nodePoolQuotaStatuses is missing", () => {
      const projectWithoutStatus = {
        ...mockProject,
        status: {
          ...mockProject.status,
          nodePoolQuotaStatuses: undefined,
        },
      };

      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: projectWithoutStatus,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 5,
      });
      expect(result).toBe(false);
    });

    test("Should return false when quota is sufficient (not pending)", () => {
      // Using mockWorkloadPriorityData: quota(15) >= gpuRequest(2) + allocatedNonPreemptible(10)
      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: mockProject,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 2,
      });
      expect(result).toBe(false);
    });

    test("Should return true when quota is insufficient (will be pending)", () => {
      // Using mockWorkloadPriorityDataPending: quota(15) < gpuRequest(10) + allocatedNonPreemptible(10)
      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: mockProject,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 10,
      });
      expect(result).toBe(true);
    });

    test("Should return false when at least one node pool has sufficient quota", () => {
      // default: 15 quota, 10 allocated non-preemptible -> 4 + 10 = 14 < 15 (OK)
      // gpu-pool: 5 quota, 3 allocated non-preemptible -> 4 + 3 = 7 > 5 (exceeds)
      // Since default can accommodate, should return false
      const multiNodePoolProject = {
        ...mockProject,
        resources: [
          ...mockProject.resources,
          {
            nodePool: {
              id: "100001",
              name: "gpu-pool",
            },
            gpu: {
              deserved: 5, // Low quota
              limit: -1,
              overQuotaWeight: 5,
            },
            priority: "MediumLow",
          },
        ],
        status: {
          ...mockProject.status,
          nodePoolQuotaStatuses: [
            ...(mockProject.status?.nodePoolQuotaStatuses || []),
            {
              allocated: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              allocatedNonPreemptible: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              requested: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              nodePoolName: "gpu-pool",
              nodePoolId: "100001",
            },
          ],
        },
      };

      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: multiNodePoolProject,
        selectedNodePoolsResources: ["default", "gpu-pool"],
        totalGpuRequest: 4,
      });
      expect(result).toBe(false);
    });

    test("Should return true only when ALL node pools exceed quota", () => {
      // default: 15 quota, 10 allocated non-preemptible -> 6 + 10 = 16 > 15 (exceeds)
      // gpu-pool: 5 quota, 3 allocated non-preemptible -> 6 + 3 = 9 > 5 (exceeds)
      // Both exceed, should return true
      const multiNodePoolProject = {
        ...mockProject,
        resources: [
          ...mockProject.resources,
          {
            nodePool: {
              id: "100001",
              name: "gpu-pool",
            },
            gpu: {
              deserved: 5,
              limit: -1,
              overQuotaWeight: 5,
            },
            priority: "MediumLow",
          },
        ],
        status: {
          ...mockProject.status,
          nodePoolQuotaStatuses: [
            ...(mockProject.status?.nodePoolQuotaStatuses || []),
            {
              allocated: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              allocatedNonPreemptible: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              requested: {
                gpu: 3,
                cpu: 500,
                memory: 512,
              },
              nodePoolName: "gpu-pool",
              nodePoolId: "100001",
            },
          ],
        },
      };

      const result = workloadService.exceedsAvailableQuota({
        priorityValue: 150,
        currProject: multiNodePoolProject,
        selectedNodePoolsResources: ["default", "gpu-pool"],
        totalGpuRequest: 6,
      });
      expect(result).toBe(true);
    });
  });

  describe("Warning Tooltip Functionality", () => {
    test("Should show warning tooltip for pending non-preemptible workload", async () => {
      await prioritySection.setProps({
        workloadPriorityData: mockWorkloadPriorityDataPending,
      });

      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;
      const veryHighOption = displayedPriorities.find((opt) => opt.value === MOCK_PRIORITY_NAMES.VERY_HIGH);

      expect(veryHighOption?.warningTooltip).toBe(prioritySection.vm.quotaWarningText);
      expect(veryHighOption?.toolTip).toBeUndefined();
    });

    test("Should show regular tooltip for non-preemptible but not pending workload", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;
      const veryHighOption = displayedPriorities.find((opt) => opt.value === MOCK_PRIORITY_NAMES.VERY_HIGH);

      expect(veryHighOption?.toolTip).toBe(prioritySection.vm.nonPreemptibleTooltip);
      expect(veryHighOption?.warningTooltip).toBeUndefined();
    });

    test("Should not show any tooltip for preemptible priorities", () => {
      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;
      const lowOption = displayedPriorities.find((opt) => opt.value === MOCK_PRIORITY_NAMES.LOW);

      expect(lowOption?.toolTip).toBeUndefined();
      expect(lowOption?.warningTooltip).toBeUndefined();
    });

    test("Should not show tooltip for disabled options", async () => {
      const restrictivePolicyRules = {
        sourceOfRule: null,
        required: false,
        canEdit: true,
        options: [
          {
            value: MOCK_PRIORITY_NAMES.LOW,
            displayed: "Build Priority",
          },
          {
            value: MOCK_PRIORITY_NAMES.MEDIUM,
            displayed: "Medium Priority",
          },
        ],
      };

      await prioritySection.setProps({
        policyRules: restrictivePolicyRules,
        workloadPriorityData: {
          ...mockWorkloadPriorityDataPending,
          workloadPriorities: mockWorkloadPriorities,
          supportPriorityClass: true,
        },
      });

      const displayedPriorities = prioritySection.vm.workloadPrioritiesToDisplay;
      const veryHighOption = displayedPriorities.find((opt) => opt.value === MOCK_PRIORITY_NAMES.VERY_HIGH);

      // Should not have tooltip because it's disabled
      expect(veryHighOption?.toolTip).toBeUndefined();
      expect(veryHighOption?.warningTooltip).toBeUndefined();
    });
  });

  describe("determineTooltipValues Method", () => {
    test("Should return warning tooltip when shouldShowPendingTooltip is true", () => {
      const result = prioritySection.vm.determineTooltipValues(true, false);
      expect(result.toolTip).toBeUndefined();
      expect(result.warningTooltip).toBe(prioritySection.vm.quotaWarningText);
    });

    test("Should return regular tooltip when shouldShowNonPreemptibleTooltip is true", () => {
      const result = prioritySection.vm.determineTooltipValues(false, true);
      expect(result.toolTip).toBe(prioritySection.vm.nonPreemptibleTooltip);
      expect(result.warningTooltip).toBeUndefined();
    });
  });
});
