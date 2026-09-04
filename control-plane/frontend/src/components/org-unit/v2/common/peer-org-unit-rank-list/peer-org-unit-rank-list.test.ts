/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { Quasar } from "quasar";
import PeerOrgUnitRankList from "./peer-org-unit-rank-list.vue";
import { EQuotaEntity } from "@/models/resource.model";
import type { ProjectNodePoolResources, DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import { EEntityPriorityValue } from "@/models/org-unit.model";

// ============================================================================
// Mock Setup
// ============================================================================

// Mock org unit service
const mockGetProjectNodePoolResources = vi.fn();
const mockGetDepartmentNodePoolResources = vi.fn();
const mockCountProjectsNodePoolResources = vi.fn();
const mockCountDepartmentNodePoolResources = vi.fn();
const mockSumProjectNodePoolResources = vi.fn();
const mockSumDepartmentNodePoolResources = vi.fn();

vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getProjectNodePoolResources: (...args: unknown[]) => mockGetProjectNodePoolResources(...args),
    getDepartmentNodePoolResources: (...args: unknown[]) => mockGetDepartmentNodePoolResources(...args),
    countProjectsNodePoolResources: (...args: unknown[]) => mockCountProjectsNodePoolResources(...args),
    countDepartmentNodePoolResources: (...args: unknown[]) => mockCountDepartmentNodePoolResources(...args),
    sumProjectNodePoolResources: (...args: unknown[]) => mockSumProjectNodePoolResources(...args),
    sumDepartmentNodePoolResources: (...args: unknown[]) => mockSumDepartmentNodePoolResources(...args),
  },
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

// Mock stores
vi.mock("@/stores/setting.store", () => ({
  useSettingStore: vi.fn(() => ({
    isOverQuotaPriorityEnabled: true,
  })),
}));

// Mock child components
vi.mock("@/components/common/runai-page-actions", () => ({
  RunaiPageActions: {
    name: "RunaiPageActions",
    template: `<div class="mock-page-actions"><slot /><slot name="selected-rows-actions" /></div>`,
    props: ["columns", "filters", "selectedRowsAmount"],
    emits: ["filters-changed", "selected-actions-close"],
  },
}));

vi.mock("@/components/common/runai-page-actions/runai-action-button", () => ({
  RunaiActionButton: {
    name: "RunaiActionButton",
    template: `<button class="mock-action-button"></button>`,
    props: ["btnAction", "disable"],
  },
}));

vi.mock("@/components/common/runai-table/runai-table-no-data", () => ({
  RunaiTableNoData: {
    name: "RunaiTableNoData",
    template: `<div class="mock-no-data"></div>`,
    props: ["entityName", "filterBy", "showCreateBtn", "showFilterIconAndButton"],
  },
}));

vi.mock("./rank-group-accordion", () => ({
  RankGroupAccordion: {
    name: "RankGroupAccordion",
    template: `
      <div class="mock-rank-accordion">
        <div v-for="row in rows" :key="row.rank" class="rank-row" :data-rank="row.rank">
          <slot name="expanded-content" :row="row" />
        </div>
      </div>
    `,
    props: ["rows", "entityType", "expandedRank", "currentOrgUnitRank"],
    emits: ["rank-toggled"],
  },
}));

vi.mock("./rank-org-unit-table", () => ({
  RankOrgUnitTable: {
    name: "RankOrgUnitTable",
    template: `<div class="mock-rank-table" :data-org-units="JSON.stringify(orgUnits.map(u => u.orgUnitId))"></div>`,
    props: [
      "orgUnits",
      "entityType",
      "loading",
      "nodePoolId",
      "pendingWeightChanges",
      "totalWeightInRank",
      "isOverQuotaPriorityEnabled",
      "currentOrgUnitId",
      "parentFilterBy",
      "selected",
    ],
    emits: ["weight-changed", "update-filters", "update:selected"],
  },
}));

vi.mock("./move-to-rank-modal", () => ({
  MoveToRankModal: {
    name: "MoveToRankModal",
    template: `<div class="mock-move-modal"></div>`,
    props: ["selectedCount", "entityType"],
    emits: ["close", "move"],
  },
}));

// ============================================================================
// Test Helpers
// ============================================================================

const createMockProjectNodePoolResource = (
  id: string,
  name: string,
  rank: string,
  weight = 50,
): ProjectNodePoolResources => ({
  id: `resource-${id}`,
  orgUnitId: id,
  orgUnitName: name,
  nodePoolId: "np-1",
  nodePoolName: "default",
  parentId: "dept-1",
  clusterId: "cluster-1",
  rank,
  resources: {
    gpu: { deserved: 10, limit: 20, overQuotaWeight: weight },
  },
});

const createMockDepartmentNodePoolResource = (
  id: string,
  name: string,
  rank: string,
  weight = 50,
): DepartmentNodePoolResources => ({
  id: `resource-${id}`,
  orgUnitId: id,
  orgUnitName: name,
  nodePoolId: "np-1",
  nodePoolName: "default",
  clusterId: "cluster-1",
  rank,
  resources: {
    gpu: { deserved: 10, limit: 20, overQuotaWeight: weight },
  },
});

const defaultProps = {
  entityType: EQuotaEntity.project,
  nodePoolId: "np-1",
  clusterId: "cluster-1",
  currentOrgUnitId: "current-project",
  currentOrgUnitRank: EEntityPriorityValue.Medium,
};

const mountComponent = async (props = {}) => {
  const wrapper = mount(PeerOrgUnitRankList, {
    props: { ...defaultProps, ...props },
    global: {
      plugins: [Quasar],
      stubs: {
        QCard: { template: "<div class='q-card'><slot /></div>" },
        QLinearProgress: { template: "<div class='q-progress' />" },
      },
    },
  });

  await flushPromises();
  return wrapper;
};

// ============================================================================
// Tests
// ============================================================================

describe("PeerOrgUnitRankList", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock implementations
    mockCountProjectsNodePoolResources.mockResolvedValue(0);
    mockCountDepartmentNodePoolResources.mockResolvedValue(0);
    mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });
    mockGetDepartmentNodePoolResources.mockResolvedValue({ nodePoolResources: [] });
    mockSumProjectNodePoolResources.mockResolvedValue({ totalWeight: 100 });
    mockSumDepartmentNodePoolResources.mockResolvedValue({ totalWeight: 100 });
  });

  describe("Current Entity Pinning", () => {
    it("fetches current entity separately when loading a rank containing it (project)", async () => {
      // Setup: 3 projects in Medium rank, current entity is one of them
      const currentEntity = createMockProjectNodePoolResource("current-project", "Current Project", "Medium", 100);
      const otherEntity1 = createMockProjectNodePoolResource("other-1", "Other 1", "Medium", 80);
      const otherEntity2 = createMockProjectNodePoolResource("other-2", "Other 2", "Medium", 60);

      // Mock count to show Medium rank has 3 entities
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (rankFilter?.includes("Medium")) {
          return Promise.resolve(3);
        }
        return Promise.resolve(0);
      });

      // Track API calls to verify separate fetches
      const apiCalls: { filters: string[]; offset?: number; limit?: number }[] = [];

      mockGetProjectNodePoolResources.mockImplementation(
        (filters: string[], _sortBy: string, _sortOrder: string, offset?: number, limit?: number) => {
          apiCalls.push({ filters, offset, limit });

          // Check if this is the current entity fetch (filter contains orgUnitId=current-project)
          const isCurrentEntityFetch = filters?.some(
            (f: string) => f.includes("orgUnitId") && f.includes("current-project") && !f.includes("!="),
          );

          // Check if this is the paginated fetch (filter contains orgUnitId!=current-project)
          const isPaginatedFetch = filters?.some((f: string) => f.includes("orgUnitId") && f.includes("!="));

          if (isCurrentEntityFetch) {
            return Promise.resolve({ nodePoolResources: [currentEntity] });
          }
          if (isPaginatedFetch) {
            return Promise.resolve({ nodePoolResources: [otherEntity1, otherEntity2] });
          }

          // Default - return all
          return Promise.resolve({ nodePoolResources: [currentEntity, otherEntity1, otherEntity2] });
        },
      );

      const wrapper = await mountComponent();

      // Clear API calls from mount phase
      apiCalls.length = 0;

      // Simulate expanding the Medium rank
      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
        expandedRankOrgUnits: ProjectNodePoolResources[];
      };

      vm.expandedRank = EEntityPriorityValue.Medium;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.Medium);
      await flushPromises();

      // Verify two separate API calls were made for the Medium rank
      const mediumRankCalls = apiCalls.filter((call) => call.filters?.some((f: string) => f.includes("Medium")));

      expect(mediumRankCalls.length).toBe(2);

      // Verify one call fetches only the current entity
      const currentEntityCall = mediumRankCalls.find((call) =>
        call.filters?.some((f: string) => f.includes("orgUnitId") && f.includes("current-project") && !f.includes("!=")),
      );
      expect(currentEntityCall).toBeDefined();
      expect(currentEntityCall?.limit).toBe(1);

      // Verify one call excludes the current entity
      const paginatedCall = mediumRankCalls.find((call) =>
        call.filters?.some((f: string) => f.includes("orgUnitId") && f.includes("!=")),
      );
      expect(paginatedCall).toBeDefined();
    });

    it("places current entity first in the results (pinned at top)", async () => {
      const currentEntity = createMockProjectNodePoolResource("current-project", "Current Project", "Medium", 100);
      const otherEntity1 = createMockProjectNodePoolResource("other-1", "Other 1", "Medium", 80);
      const otherEntity2 = createMockProjectNodePoolResource("other-2", "Other 2", "Medium", 60);

      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        if (filters?.some((f: string) => f.includes("Medium"))) {
          return Promise.resolve(3);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockImplementation((filters: string[]) => {
        const isCurrentEntityFetch = filters?.some(
          (f: string) => f.includes("orgUnitId") && f.includes("current-project") && !f.includes("!="),
        );
        const isPaginatedFetch = filters?.some((f: string) => f.includes("orgUnitId") && f.includes("!="));

        if (isCurrentEntityFetch) {
          return Promise.resolve({ nodePoolResources: [currentEntity] });
        }
        if (isPaginatedFetch) {
          // Return other entities sorted by weight (other-1 first because higher weight)
          return Promise.resolve({ nodePoolResources: [otherEntity1, otherEntity2] });
        }
        return Promise.resolve({ nodePoolResources: [] });
      });

      const wrapper = await mountComponent();

      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
        expandedRankOrgUnits: ProjectNodePoolResources[];
        getOrgUnitsForRank: (rank: string) => ProjectNodePoolResources[];
      };

      vm.expandedRank = EEntityPriorityValue.Medium;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.Medium);
      await flushPromises();

      // Verify current entity is first in expandedRankOrgUnits
      expect(vm.expandedRankOrgUnits[0].orgUnitId).toBe("current-project");
      expect(vm.expandedRankOrgUnits[1].orgUnitId).toBe("other-1");
      expect(vm.expandedRankOrgUnits[2].orgUnitId).toBe("other-2");

      // Verify getOrgUnitsForRank returns same order
      const orgUnitsForRank = vm.getOrgUnitsForRank(EEntityPriorityValue.Medium);
      expect(orgUnitsForRank[0].orgUnitId).toBe("current-project");
    });

    it("does not make separate fetch when current entity is in a different rank", async () => {
      const otherEntity1 = createMockProjectNodePoolResource("other-1", "Other 1", "High", 80);
      const otherEntity2 = createMockProjectNodePoolResource("other-2", "Other 2", "High", 60);

      // Current entity is in Medium rank, but we're loading High rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        if (filters?.some((f: string) => f.includes("High"))) {
          return Promise.resolve(2);
        }
        if (filters?.some((f: string) => f.includes("Medium"))) {
          return Promise.resolve(1); // Current entity is here
        }
        return Promise.resolve(0);
      });

      const apiCalls: { filters: string[] }[] = [];

      mockGetProjectNodePoolResources.mockImplementation((filters: string[]) => {
        apiCalls.push({ filters });
        return Promise.resolve({ nodePoolResources: [otherEntity1, otherEntity2] });
      });

      const wrapper = await mountComponent({
        currentOrgUnitRank: EEntityPriorityValue.Medium, // Current entity is in Medium
      });

      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
      };

      // Expand High rank (different from current entity's rank)
      vm.expandedRank = EEntityPriorityValue.High;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.High);
      await flushPromises();

      // Should only make ONE API call (no separate current entity fetch)
      const highRankCalls = apiCalls.filter((call) => call.filters?.some((f: string) => f.includes("High")));
      expect(highRankCalls.length).toBe(1);

      // Verify no filter for excluding current entity
      expect(highRankCalls[0].filters?.some((f: string) => f.includes("!="))).toBe(false);
    });

    it("works correctly for departments", async () => {
      const currentEntity = createMockDepartmentNodePoolResource("current-dept", "Current Dept", "Medium", 100);
      const otherEntity = createMockDepartmentNodePoolResource("other-dept", "Other Dept", "Medium", 80);

      mockCountDepartmentNodePoolResources.mockImplementation((filters: string[]) => {
        if (filters?.some((f: string) => f.includes("Medium"))) {
          return Promise.resolve(2);
        }
        return Promise.resolve(0);
      });

      mockGetDepartmentNodePoolResources.mockImplementation((filters: string[]) => {
        const isCurrentEntityFetch = filters?.some(
          (f: string) => f.includes("orgUnitId") && f.includes("current-dept") && !f.includes("!="),
        );
        const isPaginatedFetch = filters?.some((f: string) => f.includes("orgUnitId") && f.includes("!="));

        if (isCurrentEntityFetch) {
          return Promise.resolve({ nodePoolResources: [currentEntity] });
        }
        if (isPaginatedFetch) {
          return Promise.resolve({ nodePoolResources: [otherEntity] });
        }
        return Promise.resolve({ nodePoolResources: [] });
      });

      const wrapper = await mountComponent({
        entityType: EQuotaEntity.department,
        currentOrgUnitId: "current-dept",
        currentOrgUnitRank: EEntityPriorityValue.Medium,
      });

      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
        expandedRankOrgUnits: DepartmentNodePoolResources[];
      };

      vm.expandedRank = EEntityPriorityValue.Medium;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.Medium);
      await flushPromises();

      // Verify current entity is first
      expect(vm.expandedRankOrgUnits[0].orgUnitId).toBe("current-dept");
      expect(vm.expandedRankOrgUnits[1].orgUnitId).toBe("other-dept");
    });

    it("handles case when current entity is not found (search filter)", async () => {
      // Current entity doesn't match search filter
      const otherEntity = createMockProjectNodePoolResource("other-1", "Other 1", "Medium", 80);

      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        if (filters?.some((f: string) => f.includes("Medium"))) {
          return Promise.resolve(1);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockImplementation((filters: string[]) => {
        const isCurrentEntityFetch = filters?.some(
          (f: string) => f.includes("orgUnitId") && f.includes("current-project") && !f.includes("!="),
        );

        if (isCurrentEntityFetch) {
          // Current entity not found (e.g., search filter excludes it)
          return Promise.resolve({ nodePoolResources: [] });
        }
        return Promise.resolve({ nodePoolResources: [otherEntity] });
      });

      const wrapper = await mountComponent();

      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
        expandedRankOrgUnits: ProjectNodePoolResources[];
      };

      vm.expandedRank = EEntityPriorityValue.Medium;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.Medium);
      await flushPromises();

      // Should only have the other entity (current entity not found)
      expect(vm.expandedRankOrgUnits.length).toBe(1);
      expect(vm.expandedRankOrgUnits[0].orgUnitId).toBe("other-1");
    });

    it("caches current entity for pending changes tracking", async () => {
      const currentEntity = createMockProjectNodePoolResource("current-project", "Current Project", "Medium", 100);
      const otherEntity = createMockProjectNodePoolResource("other-1", "Other 1", "Medium", 80);

      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        if (filters?.some((f: string) => f.includes("Medium"))) {
          return Promise.resolve(2);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockImplementation((filters: string[]) => {
        const isCurrentEntityFetch = filters?.some(
          (f: string) => f.includes("orgUnitId") && f.includes("current-project") && !f.includes("!="),
        );
        if (isCurrentEntityFetch) {
          return Promise.resolve({ nodePoolResources: [currentEntity] });
        }
        return Promise.resolve({ nodePoolResources: [otherEntity] });
      });

      const wrapper = await mountComponent();

      const vm = wrapper.vm as unknown as {
        loadOrgUnitsForRank: (rank: string) => Promise<void>;
        expandedRank: string | null;
        allLoadedOrgUnits: Map<string, ProjectNodePoolResources>;
      };

      vm.expandedRank = EEntityPriorityValue.Medium;
      await vm.loadOrgUnitsForRank(EEntityPriorityValue.Medium);
      await flushPromises();

      // Verify current entity is cached
      expect(vm.allLoadedOrgUnits.has("current-project")).toBe(true);
      expect(vm.allLoadedOrgUnits.get("current-project")?.orgUnitName).toBe("Current Project");
    });
  });

  describe("Pending Changes Event Emission", () => {
    it("emits pending-changes-update when pending changes go from empty to non-empty", async () => {
      const wrapper = await mountComponent();

      const vm = wrapper.vm as unknown as {
        pendingOrgUnitChanges: Map<string, { newRank?: string }>;
      };

      // Add a pending change
      vm.pendingOrgUnitChanges.set("org-1", { newRank: "High" });
      await flushPromises();

      // Verify event was emitted with true
      const emitted = wrapper.emitted("pending-changes-update");
      expect(emitted).toBeTruthy();
      expect(emitted?.[0]).toEqual([true]);
    });

    it("emits pending-changes-update when pending changes go from non-empty to empty", async () => {
      const wrapper = await mountComponent();

      const vm = wrapper.vm as unknown as {
        pendingOrgUnitChanges: Map<string, { newRank?: string }>;
      };

      // Add then remove a pending change
      vm.pendingOrgUnitChanges.set("org-1", { newRank: "High" });
      await flushPromises();

      vm.pendingOrgUnitChanges.delete("org-1");
      await flushPromises();

      // Verify events were emitted: first true, then false
      const emitted = wrapper.emitted("pending-changes-update");
      expect(emitted).toBeTruthy();
      expect(emitted?.length).toBe(2);
      expect(emitted?.[0]).toEqual([true]);
      expect(emitted?.[1]).toEqual([false]);
    });
  });

  describe("Pagination Count with Pending Changes", () => {
    it("returns adjusted count in getFilterByForRank when entities are moved between ranks", async () => {
      // Setup: 3 projects in MediumLow rank, 2 in Lowest rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (rankFilter?.includes("MediumLow")) {
          return Promise.resolve(3);
        }
        if (rankFilter?.includes("Lowest")) {
          return Promise.resolve(2);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        currentOrgUnitId: undefined, // No current entity to simplify test
        currentOrgUnitRank: undefined,
      });

      const vm = wrapper.vm as unknown as {
        getFilterByForRank: (rank: string) => { rowsNumber: number };
        pendingOrgUnitChanges: Map<string, { originalRank?: string; newRank?: string }>;
      };

      // Before any pending changes, counts should match server values
      expect(vm.getFilterByForRank(EEntityPriorityValue.MediumLow).rowsNumber).toBe(3);
      expect(vm.getFilterByForRank(EEntityPriorityValue.Lowest).rowsNumber).toBe(2);

      // Simulate moving a project from MediumLow to Lowest
      vm.pendingOrgUnitChanges.set("project-1", {
        originalRank: EEntityPriorityValue.MediumLow,
        newRank: EEntityPriorityValue.Lowest,
      });

      // After pending change, counts should be adjusted
      expect(vm.getFilterByForRank(EEntityPriorityValue.MediumLow).rowsNumber).toBe(2); // 3 - 1 = 2
      expect(vm.getFilterByForRank(EEntityPriorityValue.Lowest).rowsNumber).toBe(3); // 2 + 1 = 3
    });

    it("returns adjusted count when multiple entities are moved", async () => {
      // Setup: 5 projects in Medium rank, 1 in Low rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (
          rankFilter?.includes("Medium") &&
          !rankFilter?.includes("MediumLow") &&
          !rankFilter?.includes("MediumHigh")
        ) {
          return Promise.resolve(5);
        }
        if (rankFilter?.includes("Low") && !rankFilter?.includes("MediumLow") && !rankFilter?.includes("VeryLow")) {
          return Promise.resolve(1);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        currentOrgUnitId: undefined,
        currentOrgUnitRank: undefined,
      });

      const vm = wrapper.vm as unknown as {
        getFilterByForRank: (rank: string) => { rowsNumber: number };
        pendingOrgUnitChanges: Map<string, { originalRank?: string; newRank?: string }>;
      };

      // Move 2 projects from Medium to Low
      vm.pendingOrgUnitChanges.set("project-1", {
        originalRank: EEntityPriorityValue.Medium,
        newRank: EEntityPriorityValue.Low,
      });
      vm.pendingOrgUnitChanges.set("project-2", {
        originalRank: EEntityPriorityValue.Medium,
        newRank: EEntityPriorityValue.Low,
      });

      // Verify adjusted counts
      expect(vm.getFilterByForRank(EEntityPriorityValue.Medium).rowsNumber).toBe(3); // 5 - 2 = 3
      expect(vm.getFilterByForRank(EEntityPriorityValue.Low).rowsNumber).toBe(3); // 1 + 2 = 3
    });

    it("correctly handles moving entity back to original rank (reverts count)", async () => {
      // Setup: 2 projects in High rank, 2 in VeryHigh rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (rankFilter?.includes("High") && !rankFilter?.includes("VeryHigh") && !rankFilter?.includes("MediumHigh")) {
          return Promise.resolve(2);
        }
        if (rankFilter?.includes("VeryHigh")) {
          return Promise.resolve(2);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        currentOrgUnitId: undefined,
        currentOrgUnitRank: undefined,
      });

      const vm = wrapper.vm as unknown as {
        getFilterByForRank: (rank: string) => { rowsNumber: number };
        pendingOrgUnitChanges: Map<string, { originalRank?: string; newRank?: string }>;
      };

      // Move project from High to VeryHigh
      vm.pendingOrgUnitChanges.set("project-1", {
        originalRank: EEntityPriorityValue.High,
        newRank: EEntityPriorityValue.VeryHigh,
      });

      expect(vm.getFilterByForRank(EEntityPriorityValue.High).rowsNumber).toBe(1);
      expect(vm.getFilterByForRank(EEntityPriorityValue.VeryHigh).rowsNumber).toBe(3);

      // Clear the pending change (simulating undo or save)
      vm.pendingOrgUnitChanges.delete("project-1");

      // Counts should revert to original
      expect(vm.getFilterByForRank(EEntityPriorityValue.High).rowsNumber).toBe(2);
      expect(vm.getFilterByForRank(EEntityPriorityValue.VeryHigh).rowsNumber).toBe(2);
    });

    it("does not adjust count when pending change has no rank change (only weight)", async () => {
      // Setup: 3 projects in Medium rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (
          rankFilter?.includes("Medium") &&
          !rankFilter?.includes("MediumLow") &&
          !rankFilter?.includes("MediumHigh")
        ) {
          return Promise.resolve(3);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        currentOrgUnitId: undefined,
        currentOrgUnitRank: undefined,
      });

      const vm = wrapper.vm as unknown as {
        getFilterByForRank: (rank: string) => { rowsNumber: number };
        pendingOrgUnitChanges: Map<string, { originalRank?: string; newRank?: string; newWeight?: number }>;
      };

      // Add a weight-only change (no rank change)
      vm.pendingOrgUnitChanges.set("project-1", {
        originalRank: EEntityPriorityValue.Medium,
        newWeight: 75,
        // newRank is not set - entity stays in same rank
      });

      // Count should remain unchanged
      expect(vm.getFilterByForRank(EEntityPriorityValue.Medium).rowsNumber).toBe(3);
    });

    it("rankGroups uses same adjusted counts as getFilterByForRank", async () => {
      // Setup: 3 projects in MediumLow rank, 2 in Lowest rank
      mockCountProjectsNodePoolResources.mockImplementation((filters: string[]) => {
        const rankFilter = filters?.find((f: string) => f.includes("rank"));
        if (rankFilter?.includes("MediumLow")) {
          return Promise.resolve(3);
        }
        if (rankFilter?.includes("Lowest")) {
          return Promise.resolve(2);
        }
        return Promise.resolve(0);
      });

      mockGetProjectNodePoolResources.mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        currentOrgUnitId: undefined,
        currentOrgUnitRank: undefined,
      });

      const vm = wrapper.vm as unknown as {
        getFilterByForRank: (rank: string) => { rowsNumber: number };
        rankGroups: Array<{ rank: string; count: number }>;
        pendingOrgUnitChanges: Map<string, { originalRank?: string; newRank?: string }>;
      };

      // Move a project from MediumLow to Lowest
      vm.pendingOrgUnitChanges.set("project-1", {
        originalRank: EEntityPriorityValue.MediumLow,
        newRank: EEntityPriorityValue.Lowest,
      });

      // Verify rankGroups and getFilterByForRank return consistent counts
      const mediumLowGroup = vm.rankGroups.find((g) => g.rank === EEntityPriorityValue.MediumLow);
      const lowestGroup = vm.rankGroups.find((g) => g.rank === EEntityPriorityValue.Lowest);

      expect(mediumLowGroup?.count).toBe(vm.getFilterByForRank(EEntityPriorityValue.MediumLow).rowsNumber);
      expect(lowestGroup?.count).toBe(vm.getFilterByForRank(EEntityPriorityValue.Lowest).rowsNumber);
    });
  });
});
