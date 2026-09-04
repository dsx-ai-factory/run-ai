/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { Quasar } from "quasar";
import OrgUnitQuotaTable from "./org-unit-quota-table.vue";
import type { Props } from "./org-unit-quota-table.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { ProjectNodePoolResources } from "@/swagger-models/org-unit-service-client";
import { ECustomCell } from "@/models/table.model";
import { EQuotaEntity, RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC } from "@/models/resource.model";
import type {
  QuotaTableFetchContext,
  IQuotaTableRow,
  ICustomCellGpuQuotaColConfig,
} from "@/table-models/org-unit.table-model";
import type { ITableColumn } from "@/models/table.model";

// Mock dependencies
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: {
    getProjectNodePoolResources: vi.fn(),
    countProjectsNodePoolResources: vi.fn(),
    getDepartmentNodePoolResources: vi.fn(),
    countDepartmentNodePoolResources: vi.fn(),
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

vi.mock("@/components/common/runai-table", () => ({
  RunaiTable: {
    name: "RunaiTable",
    template: `
      <div class="mock-runai-table">
        <slot name="custom-cell-gpu-quota-col" :row="{}" :column="{}"></slot>
        <slot v-if="rows.length === 0 && !loading" name="no-data"></slot>
      </div>
    `,
    props: {
      columns: { type: Array, default: () => [] },
      rows: { type: Array, default: () => [] },
      loading: { type: Boolean, default: false },
      filterBy: { type: Object, default: () => ({}) },
      getRowKey: { type: Function, default: () => "" },
      rowsPerPageOptions: { type: Array, default: () => [] },
      disableSelection: { type: Boolean, default: false },
    },
    emits: ["update-filters"],
  },
}));

vi.mock("@/components/common/runai-table/runai-custom-table-columns", () => ({
  OrgUnitGpuEditableQuotaCol: {
    name: "OrgUnitGpuEditableQuotaCol",
    template: "<div class='mock-gpu-quota-col'></div>",
    props: ["row", "column"],
  },
  OrgUnitQuotaDistributionCol: {
    name: "OrgUnitQuotaDistributionCol",
    template: "<div class='mock-quota-distribution-col'></div>",
    props: ["row", "column"],
  },
}));

vi.mock("@/components/common/runai-general-search", () => ({
  RunaiGeneralSearch: {
    name: "RunaiGeneralSearch",
    template: "<div class='mock-general-search'></div>",
    props: ["searchTerm", "showSearchText"],
    emits: ["update-search-term"],
  },
}));

vi.mock("@/components/common/runai-table/runai-table-no-data", () => ({
  RunaiTableNoData: {
    name: "RunaiTableNoData",
    template: "<div class='mock-table-no-data'>{{ customMessage }}</div>",
    props: {
      entityName: { type: String, default: "" },
      filterBy: { type: Object, default: () => ({}) },
      showCreateBtn: { type: Boolean, default: true },
      showFilterIconAndButton: { type: Boolean, default: true },
      customMessage: { type: String, default: "" },
    },
  },
}));

describe("OrgUnitQuotaTable", () => {
  const mockProjectNodePoolResources: ProjectNodePoolResources[] = [
    {
      id: "resource-1",
      orgUnitId: "project-1",
      orgUnitName: "Project Alpha",
      parentId: "dept-1",
      clusterId: "cluster-1",
      nodePoolId: "pool-1",
      nodePoolName: "default",
      rank: "1",
      resources: {
        gpu: { deserved: 10, limit: null, overQuotaWeight: null },
      },
    },
    {
      id: "resource-2",
      orgUnitId: "project-2",
      orgUnitName: "Project Beta",
      parentId: "dept-1",
      clusterId: "cluster-1",
      nodePoolId: "pool-1",
      nodePoolName: "default",
      rank: "1",
      resources: {
        gpu: { deserved: 20, limit: null, overQuotaWeight: null },
      },
    },
    {
      id: "resource-current",
      orgUnitId: "current-project",
      orgUnitName: "Current Project",
      parentId: "dept-1",
      clusterId: "cluster-1",
      nodePoolId: "pool-1",
      nodePoolName: "default",
      rank: "1",
      resources: {
        gpu: { deserved: 30, limit: null, overQuotaWeight: null },
      },
    },
  ];

  const defaultProps: Props = {
    entityType: EQuotaEntity.project,
    fetchContext: {
      departmentId: "dept-1",
      nodePoolName: "default",
      currentProjectId: "current-project",
      clusterId: "cluster-1",
    } as QuotaTableFetchContext,
    availableQuota: 50,
    currentEntityId: "current-project",
    currentEntityQuota: 8, // Main project's quota
    totalQuota: 100,
    serverTotalAssigned: 38, // Sum of all projects' quotas from API (10 + 20 + 8 for current)
  };

  const mountComponent = async (props: Partial<Props> = {}) => {
    const mockNotify = vi.fn();
    const wrapper = mount(OrgUnitQuotaTable, {
      global: {
        plugins: [
          [
            Quasar,
            {
              plugins: {},
            },
          ],
        ],
        stubs: {
          QInput: true,
        },
        mocks: {
          $q: {
            notify: mockNotify,
          },
        },
      },
      props: {
        ...defaultProps,
        ...props,
      },
    });

    await flushPromises();
    return wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
      nodePoolResources: mockProjectNodePoolResources,
    });
    (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(3);
  });

  describe("Rendering", () => {
    it("renders the component correctly", async () => {
      const wrapper = await mountComponent();
      expect(wrapper.exists()).toBe(true);
    });

    it("renders RunaiGeneralSearch component", async () => {
      const wrapper = await mountComponent();
      const search = wrapper.findComponent({ name: "RunaiGeneralSearch" });
      expect(search.exists()).toBe(true);
    });

    it("renders RunaiTable component", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      expect(table.exists()).toBe(true);
    });

    it("passes correct props to RunaiTable", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });

      expect(table.props("disableSelection")).toBe(true);
      expect(table.props("rowsPerPageOptions")).toEqual([50]);
    });
  });

  describe("Data Fetching", () => {
    it("fetches projects on mount", async () => {
      await mountComponent();
      expect(orgUnitService.getProjectNodePoolResources).toHaveBeenCalled();
      expect(orgUnitService.countProjectsNodePoolResources).toHaveBeenCalled();
    });

    it("filters out current project from results", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows") as IQuotaTableRow[];

      expect(rows).toHaveLength(2);
      expect(rows.find((r: IQuotaTableRow) => r.id === "current-project")).toBeUndefined();
    });

    it("processes entities correctly with GPU quota", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows[0]).toMatchObject({
        gpuQuota: 10,
        id: "project-1",
        name: "Project Alpha",
      });

      expect(rows[1]).toMatchObject({
        id: "project-2",
        name: "Project Beta",
        gpuQuota: 20,
      });
    });

    it("calculates quota distribution correctly", async () => {
      const wrapper = await mountComponent({
        totalQuota: 100,
      });
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      // Project with 10 GPUs out of 100 = 10%
      expect(rows[0].quotaDistribution).toBe(10);
      // Project with 20 GPUs out of 100 = 20%
      expect(rows[1].quotaDistribution).toBe(20);
    });

    it("handles zero totalQuota gracefully", async () => {
      const wrapper = await mountComponent({
        totalQuota: 0,
      });
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows[0].quotaDistribution).toBe(0);
      expect(rows[1].quotaDistribution).toBe(0);
    });

    it("handles null totalQuota gracefully", async () => {
      const wrapper = await mountComponent({
        totalQuota: null,
      });
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows[0].quotaDistribution).toBe(0);
      expect(rows[1].quotaDistribution).toBe(0);
    });

    it("adjusts count by subtracting current entity", async () => {
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(5);
      const wrapper = await mountComponent();
      await flushPromises();

      const table = wrapper.findComponent({ name: "RunaiTable" });
      const filterBy = table.props("filterBy");

      // 5 total - 1 (current) = 4
      expect(filterBy.rowsNumber).toBe(4);
    });

    it("handles fetch errors gracefully", async () => {
      const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
      (orgUnitService.getProjectNodePoolResources as Mock).mockRejectedValue(new Error("Network error"));

      const wrapper = await mountComponent();
      await flushPromises();

      const table = wrapper.findComponent({ name: "RunaiTable" });
      expect(table.props("loading")).toBe(false);

      consoleError.mockRestore();
    });
  });

  describe("Search Functionality", () => {
    it("passes searchTerm to RunaiGeneralSearch", async () => {
      const wrapper = await mountComponent();
      const search = wrapper.findComponent({ name: "RunaiGeneralSearch" });

      expect(search.props("searchTerm")).toBe("");
      expect(search.props("showSearchText")).toBe(false);
    });

    it("updates search term and refetches data", async () => {
      const wrapper = await mountComponent();
      const search = wrapper.findComponent({ name: "RunaiGeneralSearch" });

      vi.clearAllMocks();
      await search.vm.$emit("update-search-term", "Alpha");
      await flushPromises();

      expect(orgUnitService.getProjectNodePoolResources).toHaveBeenCalled();
      const call = (orgUnitService.getProjectNodePoolResources as Mock).mock.calls[0];
      expect(call[5]).toBe("Alpha"); // search parameter
    });

    it("clears search and refetches", async () => {
      const wrapper = await mountComponent();
      const search = wrapper.findComponent({ name: "RunaiGeneralSearch" });

      // Set search first
      await search.vm.$emit("update-search-term", "Alpha");
      await flushPromises();

      // Verify search was applied
      let calls = (orgUnitService.getProjectNodePoolResources as Mock).mock.calls;
      expect(calls[1][5]).toBe("Alpha"); // Second call (after mount)

      // Clear search
      await search.vm.$emit("update-search-term", "");
      await flushPromises();

      // Verify empty search was applied (could be empty string or undefined)
      calls = (orgUnitService.getProjectNodePoolResources as Mock).mock.calls;
      const searchParam = calls[2][5];
      expect(searchParam === "" || searchParam === undefined).toBe(true);
    });
  });

  describe("Filter Updates", () => {
    it("handles filter updates from table", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });

      vi.clearAllMocks();

      await table.vm.$emit("update-filters", {
        page: 2,
        rowsPerPage: 10,
        sortBy: "name",
      });
      await flushPromises();

      expect(orgUnitService.getProjectNodePoolResources).toHaveBeenCalled();
    });
  });

  describe("Column Configuration", () => {
    it("injects custom cell config for GPU quota column", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const columns = table.props("columns") as ITableColumn[];

      // Debug: log all columns
      const gpuQuotaColumn = columns.find((col: ITableColumn) => col.field === "gpuQuota");

      expect(gpuQuotaColumn).toBeDefined();
      expect(gpuQuotaColumn?.customCellConfig).toBeDefined();

      const config = gpuQuotaColumn?.customCellConfig?.[ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL] as
        | ICustomCellGpuQuotaColConfig
        | undefined;
      expect(config).toBeDefined();
      expect(config).toMatchObject({
        availableQuota: 50,
        totalQuota: 100,
        nodePoolName: "default",
      });
    });

    it("provides onSave callback in GPU quota config", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const columns = table.props("columns") as ITableColumn[];

      const gpuQuotaColumn = columns.find((col: ITableColumn) => col.field === "gpuQuota");
      const config = gpuQuotaColumn?.customCellConfig?.[ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL] as
        | ICustomCellGpuQuotaColConfig
        | undefined;

      expect(config?.onSave).toBeInstanceOf(Function);
    });

    it("emits quota-changed event when onSave is called", async () => {
      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const columns = table.props("columns") as ITableColumn[];

      const gpuQuotaColumn = columns.find((col: ITableColumn) => col.field === "gpuQuota");
      const config = gpuQuotaColumn?.customCellConfig?.[ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL] as
        | ICustomCellGpuQuotaColConfig
        | undefined;

      expect(config).toBeDefined();
      config?.onSave({
        projectId: "project-1",
        entityName: "project-1",
        nodePoolName: "default",
        nodePoolId: "pool-1",
        gpuQuota: 15,
        originalValue: 10,
        originalLimit: null,
        originalOverQuotaWeight: null,
      });

      expect(wrapper.emitted("quota-changed")).toBeTruthy();
      expect(wrapper.emitted("quota-changed")?.[0]).toEqual([
        {
          entityId: "project-1",
          entityName: "project-1",
          gpuQuota: 15,
          nodePoolId: "pool-1",
          nodePoolName: "default",
          originalValue: 10,
          originalLimit: null,
          originalOverQuotaWeight: null,
        },
      ]);
    });
  });

  describe("Entity Type Handling", () => {
    it("uses project columns for project entity type", async () => {
      const wrapper = await mountComponent({
        entityType: EQuotaEntity.project,
      });
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const columns = table.props("columns");

      expect(columns).toBeDefined();
      expect(Array.isArray(columns)).toBe(true);
    });

    it("uses department columns for department entity type", async () => {
      (orgUnitService.getDepartmentNodePoolResources as Mock).mockResolvedValue({ nodePoolResources: [] });
      const wrapper = await mountComponent({
        entityType: EQuotaEntity.department,
      });

      const table = wrapper.findComponent({ name: "RunaiTable" });
      const columns = table.props("columns");

      expect(columns).toBeDefined();
      expect(Array.isArray(columns)).toBe(true);
    });
  });

  describe("Loading State", () => {
    it("shows loading state while fetching", async () => {
      let resolveProjects: (value: { nodePoolResources: ProjectNodePoolResources[] }) => void = () => {};
      const projectsPromise = new Promise<{ nodePoolResources: ProjectNodePoolResources[] }>((resolve) => {
        resolveProjects = resolve;
      });

      (orgUnitService.getProjectNodePoolResources as Mock).mockReturnValue(projectsPromise);
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(3);

      const wrapper = mount(OrgUnitQuotaTable, {
        global: {
          plugins: [Quasar],
        },
        props: defaultProps,
      });

      // Wait for component to mount
      await flushPromises();

      const table = wrapper.findComponent({ name: "RunaiTable" });
      expect(table.props("loading")).toBe(true);

      // Resolve the promise
      resolveProjects({ nodePoolResources: mockProjectNodePoolResources });
      await flushPromises();

      expect(table.props("loading")).toBe(false);
    });
  });

  describe("Edge Cases", () => {
    it("handles projects without resources", async () => {
      const projectsWithoutResources = {
        nodePoolResources: [],
      };
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue(projectsWithoutResources);

      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows[0]).toBeUndefined();
    });

    it("handles projects without status (N/A for new API)", async () => {
      // The new API doesn't return status, so this test is no longer applicable
      // Just verify empty response works
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows).toHaveLength(0);
    });

    it("handles empty project list", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows).toHaveLength(0);
    });

    it("handles missing node pool in resources (filtered by API)", async () => {
      // The new API filters by nodePoolName, so this scenario shouldn't happen
      // But we can test with empty response
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({ nodePoolResources: [] });

      const wrapper = await mountComponent({
        fetchContext: {
          departmentId: "dept-1",
          nodePoolName: "default",
          currentProjectId: "current-project",
          clusterId: "cluster-1",
        } as QuotaTableFetchContext,
      });
      const table = wrapper.findComponent({ name: "RunaiTable" });
      const rows = table.props("rows");

      expect(rows).toHaveLength(0);
    });
  });

  describe("Pagination", () => {
    it("applies correct pagination filters", async () => {
      await mountComponent();

      expect(orgUnitService.getProjectNodePoolResources).toHaveBeenCalled();
      const call = (orgUnitService.getProjectNodePoolResources as Mock).mock.calls[0];

      // Check offset and limit parameters
      expect(call[3]).toBe(0); // offset
      expect(call[4]).toBe(50); // limit (default rowsPerPage)
    });
  });

  describe("Empty States", () => {
    beforeEach(() => {
      vi.clearAllMocks();
    });

    it("shows zero state message when no projects exist", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(1); // 1 - 1 (current) = 0

      const wrapper = await mountComponent();
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.exists()).toBe(true);
      expect(noDataComponent.props("customMessage")).toBe("No other projects yet");
    });

    it("shows filtered state message when search has no results", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(1);

      const wrapper = await mountComponent();
      await flushPromises();

      // Simulate search
      const searchComponent = wrapper.findComponent({ name: "RunaiGeneralSearch" });
      searchComponent.vm.$emit("update-search-term", "nonexistent");
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.exists()).toBe(true);
      expect(noDataComponent.props("customMessage")).toBe("No projects match your filters");
    });

    it("hides create button in empty state", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("showCreateBtn")).toBe(false);
    });

    it("hides filter icon and button in empty state", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("showFilterIconAndButton")).toBe(false);
    });

    it("passes filterBy to RunaiTableNoData", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("filterBy")).toBeDefined();
      expect(noDataComponent.props("filterBy")).toHaveProperty("searchTerm");
    });

    it("updates empty state message dynamically when search is cleared", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      await flushPromises();

      // Start with search
      const searchComponent = wrapper.findComponent({ name: "RunaiGeneralSearch" });
      searchComponent.vm.$emit("update-search-term", "search term");
      await flushPromises();

      let noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("customMessage")).toBe("No projects match your filters");

      // Clear search
      searchComponent.vm.$emit("update-search-term", "");
      await flushPromises();

      noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("customMessage")).toBe("No other projects yet");
    });

    it("uses correct entity name for pluralization", async () => {
      (orgUnitService.getProjectNodePoolResources as Mock).mockResolvedValue({
        nodePoolResources: [],
      });
      (orgUnitService.countProjectsNodePoolResources as Mock).mockResolvedValue(0);

      const wrapper = await mountComponent();
      await flushPromises();

      const noDataComponent = wrapper.findComponent({ name: "RunaiTableNoData" });
      expect(noDataComponent.props("entityName")).toBe("project");
    });
  });

  describe("Total Assigned Quota Calculation", () => {
    it("calculates totalAssignedQuota as serverTotal + currentEntityQuota", async () => {
      // serverTotalAssigned = 30 (other entities only, excluding current)
      // currentEntityQuota = 8 (current entity's quota from main input)
      // Expected totalAssignedQuota = 30 + 8 = 38
      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
      });

      expect(wrapper.vm.totalAssignedQuota).toBe(38);
    });

    it("updates totalAssignedQuota when currentEntityQuota changes", async () => {
      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
      });

      expect(wrapper.vm.totalAssignedQuota).toBe(38);

      // Simulate user changing main entity quota from 8 to 15
      await wrapper.setProps({ currentEntityQuota: 15 });
      await flushPromises();

      // totalAssignedQuota should now be 30 + 15 = 45
      expect(wrapper.vm.totalAssignedQuota).toBe(45);
    });

    it("includes pending peer changes in totalAssignedQuota", async () => {
      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      // Simulate changing project-1 quota from 10 to 15 (delta = +5)
      pendingChanges.set("project-1", {
        nodePoolName: "default",
        newValue: 15,
        originalValue: 10,
      });

      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
        pendingChanges,
      });

      // totalAssignedQuota = 30 (server) + 8 (current) + 5 (delta) = 43
      expect(wrapper.vm.totalAssignedQuota).toBe(43);
    });

    it("handles multiple pending changes correctly", async () => {
      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      // project-1: 10 -> 15 (delta = +5)
      pendingChanges.set("project-1", {
        nodePoolName: "default",
        newValue: 15,
        originalValue: 10,
      });
      // project-2: 20 -> 10 (delta = -10)
      pendingChanges.set("project-2", {
        nodePoolName: "default",
        newValue: 10,
        originalValue: 20,
      });

      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
        pendingChanges,
      });

      // totalAssignedQuota = 30 (server) + 8 (current) + 5 - 10 (deltas) = 33
      expect(wrapper.vm.totalAssignedQuota).toBe(33);
    });

    it("only considers pending changes for the current node pool", async () => {
      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      // Change on same node pool (should be counted)
      pendingChanges.set("project-1", {
        nodePoolName: "default",
        newValue: 15,
        originalValue: 10,
      });
      // Change on different node pool (should be ignored)
      pendingChanges.set("project-2", {
        nodePoolName: "other-pool",
        newValue: 100,
        originalValue: 20,
      });

      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
        pendingChanges,
      });

      // Only project-1's delta counts: 30 + 8 + 5 = 43
      expect(wrapper.vm.totalAssignedQuota).toBe(43);
    });

    it("handles null values in quota calculations", async () => {
      const wrapper = await mountComponent({
        serverTotalAssigned: null,
        currentEntityQuota: null,
      });

      // null is treated as 0
      expect(wrapper.vm.totalAssignedQuota).toBe(0);
    });

    it("handles null values in pending changes", async () => {
      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      pendingChanges.set("project-1", {
        nodePoolName: "default",
        newValue: 15,
        originalValue: null, // was null, now 15 (delta = +15)
      });

      const wrapper = await mountComponent({
        serverTotalAssigned: 30,
        currentEntityQuota: 8,
        pendingChanges,
      });

      // 30 + 8 + 15 = 53
      expect(wrapper.vm.totalAssignedQuota).toBe(53);
    });

    it("treats unlimited (-1) originalValue as 0 for department pending changes", async () => {
      (orgUnitService.getDepartmentNodePoolResources as Mock).mockResolvedValue({ nodePoolResources: [] });
      (orgUnitService.countDepartmentNodePoolResources as Mock).mockResolvedValue(0);

      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      // Default department changed from unlimited to 77
      pendingChanges.set("default-dept", {
        nodePoolName: "default",
        newValue: 77,
        originalValue: RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC,
      });

      const wrapper = await mountComponent({
        entityType: EQuotaEntity.department,
        fetchContext: {
          currentDepartmentId: "dept-2",
          nodePoolName: "default",
          clusterId: "cluster-1",
        } as QuotaTableFetchContext,
        serverTotalAssigned: 0,
        currentEntityQuota: 3,
        pendingChanges,
      });

      // delta should be 77 - 0 = 77 (not 77 - (-1) = 78)
      // totalAssignedQuota = 0 + 3 + 77 = 80
      expect(wrapper.vm.totalAssignedQuota).toBe(80);
    });

    it("treats unlimited (-1) newValue as 0 for department pending changes", async () => {
      (orgUnitService.getDepartmentNodePoolResources as Mock).mockResolvedValue({ nodePoolResources: [] });
      (orgUnitService.countDepartmentNodePoolResources as Mock).mockResolvedValue(0);

      const pendingChanges = new Map<
        string,
        { nodePoolName: string; newValue: number | null; originalValue: number | null }
      >();
      // Department changed from 77 to unlimited
      pendingChanges.set("default-dept", {
        nodePoolName: "default",
        newValue: RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC,
        originalValue: 77,
      });

      const wrapper = await mountComponent({
        entityType: EQuotaEntity.department,
        fetchContext: {
          currentDepartmentId: "dept-2",
          nodePoolName: "default",
          clusterId: "cluster-1",
        } as QuotaTableFetchContext,
        serverTotalAssigned: 77,
        currentEntityQuota: 3,
        pendingChanges,
      });

      // delta should be 0 - 77 = -77 (not (-1) - 77 = -78)
      // totalAssignedQuota = 77 + 3 + (-77) = 3
      expect(wrapper.vm.totalAssignedQuota).toBe(3);
    });
  });
});
