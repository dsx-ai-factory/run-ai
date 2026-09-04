/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ref } from "vue";
import { Quasar } from "quasar";
import GpuOverQuotaStep from "./gpu-over-quota-step.vue";
import type { Department, Project, Units } from "@/swagger-models/org-unit-service-client";
import { EQuotaEntity } from "@/models/resource.model";

// ============================================================================
// Mock Setup
// ============================================================================

// Mock stores — value is mutated per test via mockIsOverQuotaPriorityEnabled
let mockIsOverQuotaPriorityEnabled = true;
vi.mock("@/stores/setting.store", () => ({
  useSettingStore: vi.fn(() => ({
    get isOverQuotaPriorityEnabled() {
      return mockIsOverQuotaPriorityEnabled;
    },
  })),
}));

// Mock composables
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
    isNodePoolListDirty: ref(false),
    updateNodePools: vi.fn(),
    localDefaultNodePools: ref(["default"]),
    selectedNodePoolName: ref("default"),
  })),
}));

vi.mock("@/composables/org-unit/use-wizard-preview.composable", () => ({
  useWizardPreview: vi.fn(() => ({
    state: ref({}),
    updatePreview: vi.fn(),
    clearPreview: vi.fn(),
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

vi.mock("@/components/org-unit/v2/common/org-unit-over-quota-settings", () => ({
  OrgUnitOverQuotaSettings: {
    name: "OrgUnitOverQuotaSettings",
    template: `
      <div class="mock-over-quota-settings">
        <slot></slot>
      </div>
    `,
    props: ["limit", "isOverQuotaEnabled", "deserved", "entityType"],
    emits: ["update:limit", "update:is-over-quota-enabled"],
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
    props: ["label"],
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

vi.mock("@/components/org-unit/v2/common/org-unit-legend", () => ({
  OrgUnitLegend: {
    name: "OrgUnitLegend",
    template: "<div class='mock-legend'></div>",
    props: ["items"],
  },
}));

vi.mock("@/components/org-unit/v2/common/peer-org-unit-rank-list", () => ({
  PeerOrgUnitRankList: {
    name: "PeerOrgUnitRankList",
    template: `<div class='mock-peer-rank-list' @click="$emit('pending-changes-update', true)"></div>`,
    props: ["entityType", "parentId", "nodePoolId", "clusterId", "currentOrgUnitId", "currentOrgUnitRank"],
    emits: ["pending-changes-update", "current-entity-weight-changed"],
    setup() {
      return {
        // Expose methods that the parent component expects
        getPendingOrgUnitChanges: () => new Map(),
        hasPendingChanges: () => false,
        clearPendingChanges: () => {},
        refetch: () => Promise.resolve(),
      };
    },
  },
}));

// ============================================================================
// Test Suite
// ============================================================================

// Type for the save-over-quota event payload
interface SaveOverQuotaPayload {
  nodePoolName: string;
  limit: number | null;
  overQuotaWeight?: number;
  peerOrgUnitChanges?: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank?: string;
    originalRank?: string;
    newWeight?: number;
    originalWeight?: number | null;
  }>;
  currentProjectNewRank?: string;
  currentProjectNewWeight?: number;
  currentDepartmentNewRank?: string;
  currentDepartmentNewWeight?: number;
  onComplete?: () => void;
}

describe("GpuOverQuotaStep", () => {
  // ==========================================================================
  // Mock Data
  // ==========================================================================

  const mockProject = {
    id: "project-1",
    name: "Test Project",
    clusterId: "cluster-1",
    parentId: "dept-1",
    resources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 5, limit: 10, overQuotaWeight: 2 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
        rank: "1",
      },
    ],
    defaultNodePools: ["default"],
  } as Project;

  const mockDepartment = {
    id: "dept-1",
    name: "Test Department",
    clusterId: "cluster-1",
    resources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 20, limit: 30, overQuotaWeight: 3 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
        rank: "1",
      },
    ],
    projectsAggregatedResources: [
      {
        nodePool: { id: "np-1", name: "default" },
        gpu: { deserved: 15 },
        cpu: { deserved: null },
        memory: { deserved: null, units: "MB" as Units },
      },
    ],
    defaultNodePools: ["default"],
  } as Department;

  // ==========================================================================
  // Setup
  // ==========================================================================

  beforeEach(() => {
    vi.clearAllMocks();
    mockIsOverQuotaPriorityEnabled = true;
  });

  // ==========================================================================
  // Project Entity Type Tests
  // ==========================================================================

  describe("Project Entity Type", () => {
    it("should render for project entity type", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

    it("should load limit and weight from project resources", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      expect(overQuotaSettings.props("limit")).toBe(10);
      expect(overQuotaSettings.props("deserved")).toBe(5);
    });

    it("should pass entity type to over quota settings for projects", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      expect(overQuotaSettings.props("entityType")).toBe(EQuotaEntity.project);
    });

    it("should emit save-over-quota with project-specific fields", async () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      expect(wrapper.emitted("save-over-quota")).toBeTruthy();
      const emittedEvents = wrapper.emitted("save-over-quota");
      expect(emittedEvents).toBeDefined();
      if (emittedEvents && emittedEvents[0]) {
        const emittedPayload = emittedEvents[0][0] as SaveOverQuotaPayload;
        expect(emittedPayload).toHaveProperty("currentProjectNewRank");
        expect(emittedPayload).toHaveProperty("currentProjectNewWeight");
        expect(emittedPayload).not.toHaveProperty("currentDepartmentNewRank");
        expect(emittedPayload).not.toHaveProperty("currentDepartmentNewWeight");
      }
    });

    it("should pass department ID as parent ID to peer rank list", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      const peerList = wrapper.findComponent({ name: "PeerOrgUnitRankList" });
      expect(peerList.props("parentId")).toBe("dept-1");
      expect(peerList.props("entityType")).toBe("project");
    });

    it("should handle null limit (unlimited over-quota)", () => {
      const projectWithUnlimited = {
        ...mockProject,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5, limit: null, overQuotaWeight: 2 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Project;

      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: projectWithUnlimited,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      expect(overQuotaSettings.props("limit")).toBe(null);
      expect(overQuotaSettings.props("isOverQuotaEnabled")).toBe(true);
    });
  });

  // ==========================================================================
  // Department Entity Type Tests
  // ==========================================================================

  describe("Department Entity Type", () => {
    it("should render for department entity type", () => {
      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".mock-step-wrapper").exists()).toBe(true);
    });

    it("should load limit and weight from department resources", () => {
      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      expect(overQuotaSettings.props("limit")).toBe(30);
      expect(overQuotaSettings.props("deserved")).toBe(20);
    });

    it("should pass entity type to over quota settings for departments", () => {
      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      expect(overQuotaSettings.props("entityType")).toBe(EQuotaEntity.department);
    });

    it("should emit save-over-quota with department-specific fields", async () => {
      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      await wrapper.find(".save-btn").trigger("click");
      await flushPromises();

      expect(wrapper.emitted("save-over-quota")).toBeTruthy();
      const emittedEvents = wrapper.emitted("save-over-quota");
      expect(emittedEvents).toBeDefined();
      if (emittedEvents && emittedEvents[0]) {
        const emittedPayload = emittedEvents[0][0] as SaveOverQuotaPayload;
        expect(emittedPayload).toHaveProperty("currentDepartmentNewRank");
        expect(emittedPayload).toHaveProperty("currentDepartmentNewWeight");
        expect(emittedPayload).not.toHaveProperty("currentProjectNewRank");
        expect(emittedPayload).not.toHaveProperty("currentProjectNewWeight");
      }
    });

    it("should pass cluster ID as parent ID to peer rank list", () => {
      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.department,
          orgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const peerList = wrapper.findComponent({ name: "PeerOrgUnitRankList" });
      expect(peerList.props("parentId")).toBe("cluster-1");
      expect(peerList.props("entityType")).toBe("department");
    });
  });

  // ==========================================================================
  // Common Functionality Tests
  // ==========================================================================

  describe("Common Functionality", () => {
    it("should expose isDirty, save, and discardChanges methods", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

    it("should include node pool name in save payload", async () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      const emittedEvents = wrapper.emitted("save-over-quota");
      expect(emittedEvents).toBeDefined();
      if (emittedEvents && emittedEvents[0]) {
        const emittedPayload = emittedEvents[0][0] as SaveOverQuotaPayload;
        expect(emittedPayload.nodePoolName).toBe("default");
        expect(emittedPayload.limit).toBe(10);
        expect(emittedPayload.overQuotaWeight).toBe(2);
      }
    });

    it("should handle over-quota enabled state correctly", () => {
      const projectWithOverQuota = {
        ...mockProject,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5, limit: 10, overQuotaWeight: 2 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Project;

      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: projectWithOverQuota,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      // limit (10) > deserved (5) should enable over-quota
      expect(overQuotaSettings.props("isOverQuotaEnabled")).toBe(true);
    });

    it("should handle over-quota disabled state correctly", () => {
      const projectWithoutOverQuota = {
        ...mockProject,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5, limit: 5, overQuotaWeight: 0 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Project;

      const wrapper = mount(GpuOverQuotaStep, {
        props: {
          entityType: EQuotaEntity.project,
          orgUnit: projectWithoutOverQuota,
          parentOrgUnit: mockDepartment,
          clusterId: "cluster-1",
        },
        global: {
          plugins: [Quasar],
        },
      });

      const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
      // limit (5) === deserved (5) should disable over-quota
      expect(overQuotaSettings.props("isOverQuotaEnabled")).toBe(false);
    });

    it("should pass current org unit rank to peer list", () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      const peerList = wrapper.findComponent({ name: "PeerOrgUnitRankList" });
      expect(peerList.props("currentOrgUnitRank")).toBe("1");
      expect(peerList.props("currentOrgUnitId")).toBe("project-1");
    });

    it("should mark as dirty when peer rank list has pending changes", async () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      // Initially not dirty
      expect(wrapper.vm.isDirty).toBe(false);

      // Simulate peer rank list emitting pending-changes-update
      const peerList = wrapper.findComponent({ name: "PeerOrgUnitRankList" });
      await peerList.vm.$emit("pending-changes-update", true);
      await flushPromises();

      // Should now be dirty
      expect(wrapper.vm.isDirty).toBe(true);
    });

    // backend rejects overQuotaWeight: 0 when project.quota.manual_over_quota is off
    describe("Over-quota feature flag handling", () => {
      const projectWithLegacyZeroWeight = {
        ...mockProject,
        resources: [
          {
            nodePool: { id: "np-1", name: "default" },
            gpu: { deserved: 5, limit: null, overQuotaWeight: 0 },
            cpu: { deserved: null },
            memory: { deserved: null, units: "MB" as Units },
            rank: "1",
          },
        ],
      } as Project;

      it("omits overQuotaWeight when feature is off and loaded value is 0 (migration scenario)", async () => {
        mockIsOverQuotaPriorityEnabled = false;

        const wrapper = mount(GpuOverQuotaStep, {
          props: {
            entityType: EQuotaEntity.project,
            orgUnit: projectWithLegacyZeroWeight,
            parentOrgUnit: mockDepartment,
            clusterId: "cluster-1",
          },
          global: { plugins: [Quasar] },
        });

        await wrapper.find(".save-btn").trigger("click");
        await flushPromises();

        const events = wrapper.emitted("save-over-quota");
        expect(events).toBeDefined();
        const payload = events?.[0]?.[0] as SaveOverQuotaPayload;
        expect(payload.overQuotaWeight).toBeUndefined();
      });

      it("omits overQuotaWeight when feature is off and user toggles over-quota off", async () => {
        mockIsOverQuotaPriorityEnabled = false;

        const wrapper = mount(GpuOverQuotaStep, {
          props: {
            entityType: EQuotaEntity.project,
            orgUnit: mockProject,
            parentOrgUnit: mockDepartment,
            clusterId: "cluster-1",
          },
          global: { plugins: [Quasar] },
        });

        const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
        await overQuotaSettings.vm.$emit("update:is-over-quota-enabled", false);
        await flushPromises();

        await wrapper.find(".save-btn").trigger("click");
        await flushPromises();

        const events = wrapper.emitted("save-over-quota");
        const payload = events?.[0]?.[0] as SaveOverQuotaPayload;
        expect(payload.overQuotaWeight).toBeUndefined();
      });

      it("emits overQuotaWeight=0 when feature is on and user toggles over-quota off", async () => {
        mockIsOverQuotaPriorityEnabled = true;

        const wrapper = mount(GpuOverQuotaStep, {
          props: {
            entityType: EQuotaEntity.project,
            orgUnit: mockProject,
            parentOrgUnit: mockDepartment,
            clusterId: "cluster-1",
          },
          global: { plugins: [Quasar] },
        });

        const overQuotaSettings = wrapper.findComponent({ name: "OrgUnitOverQuotaSettings" });
        await overQuotaSettings.vm.$emit("update:is-over-quota-enabled", false);
        await flushPromises();

        await wrapper.find(".save-btn").trigger("click");
        await flushPromises();

        const events = wrapper.emitted("save-over-quota");
        const payload = events?.[0]?.[0] as SaveOverQuotaPayload;
        expect(payload.overQuotaWeight).toBe(0);
      });
    });

    it("should reset pending changes state when discarding changes", async () => {
      const wrapper = mount(GpuOverQuotaStep, {
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

      // Set pending changes via event
      const peerList = wrapper.findComponent({ name: "PeerOrgUnitRankList" });
      await peerList.vm.$emit("pending-changes-update", true);
      await flushPromises();

      expect(wrapper.vm.isDirty).toBe(true);

      // Discard changes
      wrapper.vm.discardChanges();
      await flushPromises();

      // Should no longer be dirty
      expect(wrapper.vm.isDirty).toBe(false);
    });
  });
});
