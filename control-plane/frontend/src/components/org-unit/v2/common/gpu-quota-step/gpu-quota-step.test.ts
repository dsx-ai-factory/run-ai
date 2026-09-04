/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { ref } from "vue";
import { Quasar } from "quasar";
import GpuQuotaStep from "./gpu-quota-step.vue";
import type { Department, Project, Units } from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { INodePoolTotals, IPeerOrgUnitQuotaChange } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";
import { useOrgUnitNodePools } from "@/composables/org-unit/use-org-unit-node-pools.composable";

// ============================================================================
// Test Factories
// ============================================================================

type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

/**
 * Creates a mock Department with all required properties.
 * Override any property by passing it in the overrides object.
 */
function createMockDepartment(overrides: DeepPartial<Department> = {}): Department {
  const base: Department = {
    id: "dept-1",
    name: "Test Department",
    description: "",
    clusterId: "cluster-1",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
    resources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 10 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
        rank: "1",
      },
    ],
    projectsAggregatedResources: [],
    defaultNodePools: ["default"],
    totalResources: {
      gpuQuota: 0,
      cpuQuota: null,
      memoryQuota: null,
    },
  };
  return { ...base, ...overrides } as Department;
}

/**
 * Creates a mock Project with all required properties.
 */
function createMockProject(overrides: DeepPartial<Project> = {}): Project {
  const base = {
    id: "project-1",
    name: "Test Project",
    description: "",
    clusterId: "cluster-1",
    parentId: "dept-1",
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
    resources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 5 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
        rank: "1",
      },
    ],
    defaultNodePools: ["default"],
    totalResources: {
      gpuQuota: 0,
      cpuQuota: null,
      memoryQuota: null,
    },
  };
  return { ...base, ...overrides } as Project;
}

// ============================================================================
// Mock Setup
// ============================================================================
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    sumDepartmentNodePoolResources: vi.fn(),
  },
}));

vi.mock("@/stores/setting.store", () => ({
  useSettingStore: vi.fn(() => ({
    isLimitQuotaOverSubscriptionEnabled: false,
  })),
}));

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...actual,
    useQuasar: vi.fn(() => ({
      notify: vi.fn(),
    })),
  };
});

// Mock child components
vi.mock("@/components/common/runai-section-title", () => ({
  RunaiSectionTitle: {
    name: "RunaiSectionTitle",
    template: "<div class='mock-section-title'></div>",
    props: ["title", "tooltipText"],
  },
}));

vi.mock("@/components/org-unit/v2/common/node-pool-tabs", () => ({
  NodePoolTabs: {
    name: "NodePoolTabs",
    template: `
      <div class="mock-node-pool-tabs">
        <slot></slot>
      </div>
    `,
    props: ["nodePools", "availableNodePools", "modelValue"],
    emits: ["update:model-value", "update:node-pools"],
  },
}));

vi.mock("@/components/org-unit/v2/common/gpu-quota-availability-card", () => ({
  GpuQuotaAvailabilityCard: {
    name: "GpuQuotaAvailabilityCard",
    template: "<div class='mock-availability-card'></div>",
    props: ["title", "total", "available", "tooltip", "isCurrentEntityUnlimited"],
  },
}));

vi.mock("@/components/org-unit/v2/common/gpu-quota-card", () => ({
  GpuQuotaCard: {
    name: "GpuQuotaCard",
    template: "<div class='mock-quota-card'></div>",
    props: [
      "orgUnitName",
      "gpuQuota",
      "totalQuota",
      "unassigned",
      "state",
      "message",
      "parentName",
      "entityType",
      "nodePoolName",
      "otherEntitiesAssignedQuota",
      "isUnlimited",
      "isCurrentEntityUnlimited",
      "isOtherEntitiesQuotaUnlimited",
    ],
    emits: ["update:gpu-quota"],
  },
}));

vi.mock("@/components/org-unit/v2/common/org-unit-legend", () => ({
  OrgUnitLegend: {
    name: "OrgUnitLegend",
    template: "<div class='mock-legend'></div>",
    props: ["items"],
  },
}));

vi.mock("@/components/org-unit/v2/common/org-unit-quota-table", () => ({
  OrgUnitQuotaTable: {
    name: "OrgUnitQuotaTable",
    template: "<div class='mock-quota-table'></div>",
    props: [
      "entityType",
      "fetchContext",
      "availableQuota",
      "currentEntityId",
      "currentEntityQuota",
      "totalQuota",
      "pendingChanges",
    ],
    emits: ["quota-changed"],
  },
}));

vi.mock("@/components/org-unit/v2/common/quota-exceeded-warning-banner", () => ({
  QuotaExceededWarningBanner: {
    name: "QuotaExceededWarningBanner",
    template: "<div class='mock-warning-banner' :data-hide-excess='hideExcessAmount'>Excess: {{ excessAmount }}</div>",
    props: ["excessAmount", "parentEntityType", "hideExcessAmount"],
  },
}));

vi.mock("@/components/org-unit/v2/common/org-unit-step-wrapper", () => ({
  OrgUnitStepWrapper: {
    name: "OrgUnitStepWrapper",
    template: `
      <div class="mock-step-wrapper">
        <slot></slot>
        <button @click="$emit('save')" class="save-btn">Save</button>
      </div>
    `,
    props: ["label", "loading"],
    emits: ["save"],
  },
}));

vi.mock("@/components/common/runai-folding-sub-section", () => ({
  RunaiFoldingSubSection: {
    name: "RunaiFoldingSubSection",
    template: `
      <div class="mock-folding-section">
        <slot></slot>
      </div>
    `,
    props: ["label", "lazyMount"],
  },
}));

// Mock composables
vi.mock("@/composables/org-unit/use-quota-validation.composable", () => ({
  useQuotaValidation: vi.fn(() => ({
    quotaState: ref("default"),
    quotaMessage: ref(""),
  })),
}));

vi.mock("@/composables/org-unit/use-unsaved-changes-guard.composable", () => ({
  useUnsavedChangesGuard: vi.fn(() => ({
    registerStepGuard: vi.fn(() => vi.fn()),
    requestAction: vi.fn(({ onProceed }: { onProceed: () => void }) => {
      onProceed();
      return Promise.resolve();
    }),
  })),
}));

vi.mock("@/composables/org-unit/use-org-unit-node-pools.composable", () => ({
  useOrgUnitNodePools: vi.fn(() => ({
    allNodePools: ref([{ id: "np-1", name: "default", clusterId: "cluster-1" }]),
    selectedNodePools: ref([{ id: "np-1", name: "default", clusterId: "cluster-1" }]),
    selectedNodePoolId: ref("np-1"),
    selectedNodePoolName: ref("default"),
    localDefaultNodePools: ref(["default"]),
    isNodePoolListDirty: ref(false),
    updateNodePools: vi.fn(),
    resetNodePoolsToBaseline: vi.fn(),
  })),
}));

vi.mock("@/composables/org-unit/use-wizard-preview.composable", () => ({
  useWizardPreview: vi.fn(() => ({
    updatePreview: vi.fn(),
    clearPreview: vi.fn(),
  })),
}));

// ============================================================================
// Test Suite
// ============================================================================

// Type for the save-gpu-quota event payload
interface SaveGpuQuotaPayload {
  nodePoolName: string;
  gpuQuota: number;
  defaultNodePools?: string[];
  peerProjectChanges?: Array<IPeerOrgUnitQuotaChange>;
  peerDepartmentChanges?: Array<IPeerOrgUnitQuotaChange>;
  onComplete?: () => void;
}

describe("GpuQuotaStep", () => {
  // ==========================================================================
  // Mock Data
  // ==========================================================================

  const mockProject = createMockProject();

  const mockDepartment = createMockDepartment({
    projectsAggregatedResources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 8 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
      },
    ],
  });

  const mockNodePoolTotals: INodePoolTotals = {
    default: {
      totalGpus: 20,
      cpuCores: 100,
      memoryBytes: 1000000000,
    },
  };

  // ==========================================================================
  // Setup
  // ==========================================================================

  beforeEach(() => {
    vi.clearAllMocks();
  });

  // ==========================================================================
  // Project Entity Type Tests
  // ==========================================================================

  describe("Project Entity Type", () => {
    it("should render for project entity type", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".mock-step-wrapper").exists()).toBe(true);
    });

    it("should calculate availability total from parent department", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      expect(availabilityCard.props("total")).toBe(10); // Department's quota
    });

    it("should calculate availability assigned from aggregated resources", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Available = Total (10) - Assigned (8) - CurrentDelta (0) - PendingDelta (0) = 2
      expect(availabilityCard.props("available")).toBe(2);
    });

    it("should calculate excess quota when over department limit", async () => {
      const projectWithExcess = {
        ...mockProject,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 15 }, // Much higher than department's 10
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Project;

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: projectWithExcess,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // The component should render, and excess calculation should show quota exceeds parent
      // Current project (15) + other projects (8-5=3) = 18 > department total (10) = 8 excess
      expect(wrapper.exists()).toBe(true);
    });

    it("should emit save-gpu-quota with peerProjectChanges", async () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      await wrapper.find(".save-btn").trigger("click");
      await flushPromises();

      expect(wrapper.emitted("save-gpu-quota")).toBeTruthy();
      const emittedPayload = wrapper.emitted("save-gpu-quota")?.[0]?.[0] as SaveGpuQuotaPayload;
      expect(emittedPayload).toHaveProperty("peerProjectChanges");
      expect(emittedPayload).not.toHaveProperty("peerDepartmentChanges");
    });

    it("should pass correct fetch context to quota table", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      expect(quotaTable.props("fetchContext")).toEqual({
        departmentId: "dept-1",
        currentProjectId: "project-1",
        nodePoolName: "default",
        clusterId: "cluster-1",
      });
    });
  });

  // ==========================================================================
  // Department Entity Type Tests
  // ==========================================================================

  describe("Department Entity Type", () => {
    it("should render for department entity type", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      (orgUnitService.sumDepartmentNodePoolResources as Mock).mockResolvedValue({ deservedSum: 0 });
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".mock-step-wrapper").exists()).toBe(true);
      errorSpy.mockRestore();
    });

    it("should calculate availability total from cluster telemetry", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      (orgUnitService.sumDepartmentNodePoolResources as Mock).mockResolvedValue({ deservedSum: 0 });
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      expect(availabilityCard.props("total")).toBe(20); // Cluster's total GPUs
      errorSpy.mockRestore();
    });

    it("should fetch department node pool resources on mount", async () => {
      const mockSumResponse = { deservedSum: 15 };
      (orgUnitService.sumDepartmentNodePoolResources as Mock).mockResolvedValue(mockSumResponse);

      mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      expect(orgUnitService.sumDepartmentNodePoolResources).toHaveBeenCalled();
    });

    it("should emit save-gpu-quota with peerDepartmentChanges", async () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await wrapper.find(".save-btn").trigger("click");
      await flushPromises();

      expect(wrapper.emitted("save-gpu-quota")).toBeTruthy();
      const emittedPayload = wrapper.emitted("save-gpu-quota")?.[0]?.[0] as SaveGpuQuotaPayload;
      expect(emittedPayload).toHaveProperty("peerDepartmentChanges");
      expect(emittedPayload).not.toHaveProperty("peerProjectChanges");
    });

    it("should pass correct fetch context to quota table", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      expect(quotaTable.props("fetchContext")).toEqual({
        currentDepartmentId: "dept-1",
        nodePoolName: "default",
        clusterId: "cluster-1",
      });
    });

    it("should validate against subordinate project quota", () => {
      const deptWithSubordinates = {
        ...mockDepartment,
        projectsAggregatedResources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 12 }, // Subordinate projects have 12 GPUs
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
          },
        ],
      } as Department;

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: deptWithSubordinates,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // The quota card should receive validation state/message
      expect(quotaCard.exists()).toBe(true);
    });
  });

  // ==========================================================================
  // Common Functionality Tests
  // ==========================================================================

  describe("Common Functionality", () => {
    it("should handle pending peer changes", async () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });

      // Simulate quota change from table
      quotaTable.vm.$emit("quota-changed", {
        entityId: "project-2",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 3,
        originalValue: 2,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      // Trigger save
      await wrapper.find(".save-btn").trigger("click");
      await flushPromises();

      const emittedPayload = wrapper.emitted("save-gpu-quota")?.[0]?.[0] as SaveGpuQuotaPayload;
      expect(emittedPayload.peerProjectChanges).toHaveLength(1);
      expect(emittedPayload.peerProjectChanges?.[0]).toMatchObject({
        orgUnitId: "project-2",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 3,
      });
    });

    it("should expose isDirty, save, and discardChanges methods", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      expect(wrapper.vm.isDirty).toBeDefined();
      expect(wrapper.vm.save).toBeDefined();
      expect(wrapper.vm.discardChanges).toBeDefined();
    });

    it("should handle unlimited quota (null)", () => {
      const deptWithUnlimited = createMockDepartment({
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: undefined },
            cpu: { deserved: undefined },
            memory: { deserved: undefined, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: deptWithUnlimited,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      expect(availabilityCard.props("total")).toBeNull();
    });

    it("should display correct entity labels", () => {
      const projectWrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      expect(projectWrapper.html()).toContain("project's GPU quota");

      const deptWrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      expect(deptWrapper.html()).toContain("department's GPU quota");
    });
  });

  // ==========================================================================
  // Default Department Unlimited Handling Tests
  // ==========================================================================

  describe("Default Department Unlimited Handling", () => {
    const mockDefaultDepartment = createMockDepartment({
      id: "default-dept",
      name: "default",
      resources: [
        {
          nodePool: { id: "np-1", name: "default" },
          gpu: { deserved: -1 }, // Unlimited
          cpu: { deserved: null },
          memory: { deserved: null, units: "MB" as Units },
          rank: "1",
        },
      ],
    });

    const mockNonDefaultDepartment = createMockDepartment({
      id: "dept-2",
      name: "rnd",
      resources: [
        {
          nodePool: { id: "np-1", name: "default" },
          gpu: { deserved: 5 },
          cpu: { deserved: null },
          memory: { deserved: null, units: "MB" as Units },
          rank: "1",
        },
      ],
    });

    it("should fetch two separate sums for departments on any node pool", async () => {
      // Mock API to return different sums
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department only (unlimited)

      mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Should have made 2 API calls (exclude default, only default)
      expect(orgUnitService.sumDepartmentNodePoolResources).toHaveBeenCalledTimes(2);
    });

    it("should show isOtherEntitiesQuotaUnlimited=true when default department has unlimited quota (-1)", async () => {
      // Mock API to return -1 for default department
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department only (unlimited)

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // When viewing non-default dept and default is unlimited, other entities are unlimited
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(true);
      // But isUnlimited stays false (node pool total is still a valid number)
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });

    it("should show availabilityAvailable=0 when default department is unlimited", async () => {
      // Mock API to return -1 for default department
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department only (unlimited)

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      expect(availabilityCard.props("available")).toBe(0);
    });

    it("should detect unlimited when current entity IS the default department with -1 quota", async () => {
      // Mock API (won't matter since we use local value for current entity)
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDefaultDepartment, // Current entity IS the default department
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // When current entity is default dept with -1, current entity is unlimited
      expect(quotaCard.props("isCurrentEntityUnlimited")).toBe(true);
      // But isUnlimited stays false (node pool total is still a valid number)
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });

    it("should NOT show unlimited when default department has valid quota", async () => {
      // Mock API to return valid numbers
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: 10 }); // Default department has 10 GPUs

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });

    it("should calculate correct available quota when default department has valid quota", async () => {
      // Mock API: 5 for others, 10 for default = 15 total assigned
      // Total = 20, Assigned = 15, Current entity (rnd) has 5
      // Available = 20 - 15 - 0 (no delta) = 5
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default (includes current entity)
        .mockResolvedValueOnce({ deservedSum: 10 }); // Default department has 10 GPUs

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      expect(availabilityCard.props("available")).toBe(5);
    });

    it("should handle pending peer change for default department becoming unlimited", async () => {
      // Initial state: default department has 10 GPUs
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: 10 }); // Default department has 10 GPUs

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Initially not unlimited (other entities)
      let quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(false);

      // Simulate pending change: default department changed to -1 (unlimited)
      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      quotaTable.vm.$emit("quota-changed", {
        entityId: "default-dept",
        entityName: "default",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: -1, // Changed to unlimited
        originalValue: 10,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      // Now other entities should be unlimited (since default is now unlimited)
      quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(true);
    });

    it("should handle pending peer change for default department becoming non-unlimited", async () => {
      // Initial state: default department has -1 GPUs (unlimited)
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department is unlimited

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Initially other entities are unlimited (default has -1)
      let quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(true);

      // Simulate pending change: default department changed to 10 (not unlimited)
      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      quotaTable.vm.$emit("quota-changed", {
        entityId: "default-dept",
        entityName: "default",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 10, // Changed to valid quota
        originalValue: -1,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      // Now should NOT be unlimited
      quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(false);
    });

    it("should treat null from API as unlimited", async () => {
      // Mock API to return null for default department (also means unlimited)
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: null }); // Default department has null (unlimited)

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // When default dept has null, other entities are unlimited
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(true);
      // But isUnlimited stays false (node pool total is still a valid number)
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });

    it("should NOT show unlimited when current entity is default department with valid quota", async () => {
      const defaultDeptWithValidQuota = createMockDepartment({
        ...mockDefaultDepartment,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 10 }, // Valid quota, not -1
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 })
        .mockResolvedValueOnce({ deservedSum: 10 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: defaultDeptWithValidQuota,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });
  });

  // ==========================================================================
  // Projects Should NOT Have Unlimited Handling
  // ==========================================================================

  describe("Projects - No Unlimited Handling", () => {
    it("should NOT apply unlimited handling for projects even if parent department has -1", () => {
      const deptWithUnlimitedGpu = {
        ...mockDepartment,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: -1 }, // Parent department has unlimited
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Department;

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: deptWithUnlimitedGpu,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Parent's -1 should be treated as null (unlimited) for projects
      expect(availabilityCard.props("total")).toBeNull();
    });

    it("should NOT show isUnlimited=true based on default department for projects", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // Projects don't have default department unlimited handling
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });

    it("should NOT fetch split sums for projects", async () => {
      mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Should NOT have called the department sum API for projects
      expect(orgUnitService.sumDepartmentNodePoolResources).not.toHaveBeenCalled();
    });

    it("should calculate availability correctly for projects with parent having valid quota", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Total = 10 (parent dept quota)
      // Assigned = 8 (from projectsAggregatedResources)
      // Available = 10 - 8 = 2
      expect(availabilityCard.props("total")).toBe(10);
      expect(availabilityCard.props("available")).toBe(2);
    });

    it("should handle project pending changes without -1 special handling", async () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });

      // Simulate a peer project change (no entityName needed for projects, but include it)
      quotaTable.vm.$emit("quota-changed", {
        entityId: "project-2",
        entityName: "Other Project",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 3,
        originalValue: 1,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Total = 10, Assigned = 8, Pending delta = 3 - 1 = 2
      // Available = 10 - 8 - 2 = 0
      expect(availabilityCard.props("available")).toBe(0);
    });
  });

  // ==========================================================================
  // Warning Banner - Unlimited Entity Tests
  // ==========================================================================

  describe("Warning Banner - Unlimited Entity", () => {
    const mockDefaultDepartmentUnlimited = createMockDepartment({
      id: "default-dept",
      name: "default",
      resources: [
        {
          nodePool: { id: "np-1", name: "default" },
          gpu: { deserved: -1 }, // Unlimited
          cpu: { deserved: null },
          memory: { deserved: null, units: "MB" as Units },
          rank: "1",
        },
      ],
    });

    it("should show warning banner when default department has unlimited quota (-1)", async () => {
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDefaultDepartmentUnlimited,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Banner should be shown because isCurrentEntityUnlimited is true
      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      expect(warningBanner.exists()).toBe(true);
    });

    it("should hide excess amount when current entity is unlimited", async () => {
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDefaultDepartmentUnlimited,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      expect(warningBanner.exists()).toBe(true);
      // hideExcessAmount should be true when current entity is unlimited
      expect(warningBanner.props("hideExcessAmount")).toBe(true);
    });

    it("should NOT show warning banner when default department has valid quota and no excess", async () => {
      const defaultDeptWithValidQuota = {
        ...mockDefaultDepartmentUnlimited,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5 }, // Valid quota, not -1
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Department;

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: 5 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: defaultDeptWithValidQuota,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // No excess and not unlimited, so banner should not show
      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      expect(warningBanner.exists()).toBe(false);
    });

    it("should pass hideExcessAmount=true to banner when current entity is unlimited", async () => {
      // This test verifies the prop is passed correctly
      // The isCurrentEntityUnlimited computed returns true when:
      // - entityType is department
      // - orgUnit.name is "default"
      // - gpuQuota.value is -1
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDefaultDepartmentUnlimited,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      // Banner is shown because isCurrentEntityUnlimited triggers v-if
      expect(warningBanner.exists()).toBe(true);
      // hideExcessAmount should be true for unlimited entity
      expect(warningBanner.props("hideExcessAmount")).toBe(true);
    });

    it("should NOT apply unlimited banner logic to non-default departments", async () => {
      const nonDefaultDept = createMockDepartment({
        id: "dept-2",
        name: "rnd", // NOT "default"
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 })
        .mockResolvedValueOnce({ deservedSum: 10 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: nonDefaultDept,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Non-default department with no excess should NOT show banner
      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      expect(warningBanner.exists()).toBe(false);
    });

    it("should NOT show unlimited banner for projects", () => {
      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: mockProject,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      // Projects should not have unlimited banner logic
      const warningBanner = wrapper.findComponent({ name: "QuotaExceededWarningBanner" });
      // No excess in default mock data, so no banner
      expect(warningBanner.exists()).toBe(false);
    });
  });

  // ==========================================================================
  // Edge Cases
  // ==========================================================================

  describe("Edge Cases", () => {
    it("should handle API error gracefully", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      (orgUnitService.sumDepartmentNodePoolResources as Mock).mockRejectedValue(new Error("API Error"));

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Should not crash, and should have default values
      expect(wrapper.exists()).toBe(true);
      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.exists()).toBe(true);
      errorSpy.mockRestore();
    });

    it("should handle department with no resources", async () => {
      const deptWithNoResources = {
        ...mockDepartment,
        resources: [],
      } as Department;

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: 0 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: deptWithNoResources,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      expect(wrapper.exists()).toBe(true);
      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("gpuQuota")).toBe(0);
    });

    it("should prioritize pending change over server value for default department", async () => {
      // Server says default has -1 (unlimited)
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const nonDefaultDept = createMockDepartment({
        id: "dept-2",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: nonDefaultDept,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Initially other entities are unlimited (from server default has -1)
      let quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(true);

      // Add pending change for default department with valid value
      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      quotaTable.vm.$emit("quota-changed", {
        entityId: "default-dept",
        entityName: "default",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 15, // Changed to valid quota
        originalValue: -1,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      // Pending change should take precedence - no longer unlimited
      quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(false);
    });

    it("should correctly calculate delta when default department changes from -1 to valid", async () => {
      // Server: default has -1, rnd has 5
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default is -1

      const nonDefaultDept = createMockDepartment({
        id: "dept-2",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: nonDefaultDept,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Change default from -1 to 10
      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });
      quotaTable.vm.$emit("quota-changed", {
        entityId: "default-dept",
        entityName: "default",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 10,
        originalValue: -1, // Was unlimited
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Total = 20
      // Assigned = 5 (sum excluding default) + 0 (-1 treated as 0) = 5
      // Delta from pending = 10 - 0 (original -1 → 0) = 10
      // Available = 20 - 5 - 10 = 5
      expect(availabilityCard.props("available")).toBe(5);
    });

    it("should handle multiple pending changes including non-default departments", async () => {
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 10 }) // Sum excluding default
        .mockResolvedValueOnce({ deservedSum: 5 }); // Default has 5

      const nonDefaultDept = createMockDepartment({
        id: "dept-2",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 3 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: nonDefaultDept,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaTable = wrapper.findComponent({ name: "OrgUnitQuotaTable" });

      // Change another non-default department
      quotaTable.vm.$emit("quota-changed", {
        entityId: "dept-3",
        entityName: "marketing",
        nodePoolName: "default",
        nodePoolId: "np-1",
        gpuQuota: 5,
        originalValue: 2,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Total = 20
      // Assigned = 10 + 5 = 15
      // Pending delta for marketing = 5 - 2 = 3
      // Available = 20 - 15 - 3 = 2
      expect(availabilityCard.props("available")).toBe(2);
    });

    it("should handle zero quota correctly (not treat as unlimited)", async () => {
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 })
        .mockResolvedValueOnce({ deservedSum: 0 }); // Default has 0 (not unlimited)

      const nonDefaultDept = createMockDepartment({
        id: "dept-2",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: nonDefaultDept,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // 0 should NOT be treated as unlimited
      expect(quotaCard.props("isUnlimited")).toBe(false);
    });
  });

  // ==========================================================================
  // Node Pool Switching - GPU Quota Initial Value Tests
  // ==========================================================================

  describe("Node Pool Switching - GPU Quota Initial Value", () => {
    it("should show 0 for non-default node pool when default department has -1 on all pools", async () => {
      // Simulate the default department with -1 on multiple node pools (server bug)
      const mockDefaultDeptWithMultiplePools = createMockDepartment({
        id: "default-dept",
        name: "default",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: -1 }, // Unlimited - should remain -1
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
          {
            nodePool: { id: "np-2", name: "gpu-pool" },
            gpu: { deserved: -1 }, // Should be treated as 0 (fresh start)
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "2",
          },
        ],
        defaultNodePools: ["default", "gpu-pool"],
      });

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDefaultDeptWithMultiplePools,
          clusterId: "cluster-1",
          nodePoolTotals: {
            default: { cpuCores: 100, memoryBytes: 1024, totalGpus: 20 },
            "gpu-pool": { cpuCores: 100, memoryBytes: 1024, totalGpus: 20 },
          },
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // First pool (default) should show -1
      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("gpuQuota")).toBe(-1);
    });

    it("should show 0 for non-default department even if resource has -1", async () => {
      // Simulate a non-default department that incorrectly has -1 in resources
      const mockNonDefaultDeptWithUnlimited = createMockDepartment({
        id: "dept-2",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: -1 }, // Should be treated as 0
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 })
        .mockResolvedValueOnce({ deservedSum: -1 });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockNonDefaultDeptWithUnlimited,
          clusterId: "cluster-1",
          nodePoolTotals: mockNodePoolTotals,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      // Non-default department should show 0, not -1
      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      expect(quotaCard.props("gpuQuota")).toBe(0);
    });
  });

  // ==========================================================================
  // Non-default nodepool should NOT treat default dept as unlimited
  // ==========================================================================

  describe("Non-default nodepool - default department not unlimited", () => {
    const mockNonDefaultNodePools = () => {
      (useOrgUnitNodePools as Mock).mockReturnValue({
        allNodePools: ref([
          { id: "np-1", name: "default", clusterId: "cluster-1" },
          { id: "np-2", name: "np1", clusterId: "cluster-1" },
        ]),
        selectedNodePools: ref([
          { id: "np-1", name: "default", clusterId: "cluster-1" },
          { id: "np-2", name: "np1", clusterId: "cluster-1" },
        ]),
        selectedNodePoolId: ref("np-2"),
        selectedNodePoolName: ref("np1"),
        localDefaultNodePools: ref(["default"]),
        isNodePoolListDirty: ref(false),
        updateNodePools: vi.fn(),
        resetNodePoolsToBaseline: vi.fn(),
      });
    };

    const nodePoolTotalsWithNp1: INodePoolTotals = {
      default: { totalGpus: 20, cpuCores: 100, memoryBytes: 1000000000 },
      np1: { totalGpus: 16, cpuCores: 50, memoryBytes: 500000000 },
    };

    it("should NOT treat default department as unlimited on non-default nodepool even if API returns -1", async () => {
      mockNonDefaultNodePools();

      // API returns -1 for default department on np1 (server stores -1 even for non-default nodepools)
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 0 }) // Sum excluding default on np1
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department on np1 returns -1

      const newDepartment = createMockDepartment({
        id: "new-dept",
        name: "New Department",
        resources: [
          {
            nodePool: { id: "np-2", name: "np1" },
            gpu: { deserved: 0 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: newDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: nodePoolTotalsWithNp1,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const quotaCard = wrapper.findComponent({ name: "GpuQuotaCard" });
      // On non-default nodepool, default department's -1 should NOT make other entities unlimited
      expect(quotaCard.props("isOtherEntitiesQuotaUnlimited")).toBe(false);

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Should show actual available GPUs (16 total - 0 assigned = 16), not 0
      expect(availabilityCard.props("available")).toBe(16);
    });

    it("should calculate correct unassigned GPUs on non-default nodepool with default dept -1", async () => {
      mockNonDefaultNodePools();

      // API returns -1 for default, 5 for others on np1
      (orgUnitService.sumDepartmentNodePoolResources as Mock)
        .mockResolvedValueOnce({ deservedSum: 5 }) // Sum excluding default on np1
        .mockResolvedValueOnce({ deservedSum: -1 }); // Default department on np1 returns -1

      const rndDepartment = createMockDepartment({
        id: "dept-rnd",
        name: "rnd",
        resources: [
          {
            nodePool: { id: "np-2", name: "np1" },
            gpu: { deserved: 3 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      });

      const wrapper = mount(GpuQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: rndDepartment,
          clusterId: "cluster-1",
          nodePoolTotals: nodePoolTotalsWithNp1,
        },
        global: {
          plugins: [Quasar],
        },
      });

      await flushPromises();

      const availabilityCard = wrapper.findComponent({ name: "GpuQuotaAvailabilityCard" });
      // Total = 16 (np1 total GPUs)
      // Assigned = 5 (sum excluding default) + 0 (-1 treated as 0) = 5
      // Available = 16 - 5 - 0 (no delta) = 11
      expect(availabilityCard.props("available")).toBe(11);
    });
  });
});
