/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import OverviewIndex from "./overview-index.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { EWidgetComponent } from "@/models/chart.model";
import { storageUtil } from "@/utils/storage.util";

// Mock services
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepools: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("@/utils/storage.util", () => ({
  storageUtil: {
    get: vi.fn().mockReturnValue(null),
    save: vi.fn(),
  },
}));

// Mock child components to avoid deep rendering
vi.mock("@/components/dashboard-v2/dashboard-page-filter/", () => ({
  DashboardPageFilter: {
    name: "DashboardPageFilter",
    template: '<div class="mock-dashboard-page-filter"></div>',
    props: ["filterOptions", "asyncFilterOptionsPromiseMap", "filterBy"],
  },
}));

vi.mock("@/components/dashboard-v2/widgets/common/widget-layout/", () => ({
  WidgetLayout: {
    name: "WidgetLayout",
    template: '<div class="mock-widget-layout"></div>',
    props: [
      "widgetsLayout",
      "clusterId",
      "nodePoolName",
      "filterByDates",
      "filterBy",
      "hasDepartments",
      "hasReadProjectsPermission",
      "hasReadDepartmentsPermission",
    ],
  },
}));

const mockRouter = {
  push: vi.fn(),
};

let pinia: TestingPinia;

function mountComponent(permissionOverrides?: {
  hasReadNodesPermission?: boolean;
  hasReadClustersPermission?: boolean;
  hasReadNodePoolsPermission?: boolean;
  hasReadWorkloadsPermission?: boolean;
  hasReadDepartmentsPermission?: boolean;
  hasReadProjectsPermission?: boolean;
}) {
  pinia = createTestingPinia({ createSpy: vi.fn });

  const permissionStore = usePermissionStore(pinia);
  const clusterStore = useClusterStore(pinia);
  const appStore = useAppStore(pinia);
  const authStore = useAuthStore(pinia);

  vi.mocked(permissionStore.hasReadNodesPermission).mockReturnValue(permissionOverrides?.hasReadNodesPermission ?? true);
  vi.mocked(permissionStore.hasReadClustersPermission).mockReturnValue(
    permissionOverrides?.hasReadClustersPermission ?? true,
  );
  vi.mocked(permissionStore.hasReadNodePoolsPermission).mockReturnValue(
    permissionOverrides?.hasReadNodePoolsPermission ?? true,
  );
  vi.mocked(permissionStore.hasReadWorkloadsPermission).mockReturnValue(
    permissionOverrides?.hasReadWorkloadsPermission ?? true,
  );
  vi.mocked(permissionStore.hasReadDepartmentsPermission).mockReturnValue(
    permissionOverrides?.hasReadDepartmentsPermission ?? true,
  );
  vi.mocked(permissionStore.hasReadProjectsPermission).mockReturnValue(
    permissionOverrides?.hasReadProjectsPermission ?? true,
  );

  Object.defineProperty(clusterStore, "selectedClusterId", {
    get: () => "test-cluster-id",
    configurable: true,
  });

  // Setup auth store
  Object.defineProperty(authStore, "orgUnits", {
    get: () => [],
    configurable: true,
  });

  // Setup app store
  Object.defineProperty(appStore, "isWhiteLabel", {
    get: () => false,
    configurable: true,
  });
  Object.defineProperty(appStore, "isOnboardingModalClosed", {
    get: () => true,
    configurable: true,
  });

  return mount(OverviewIndex, {
    global: {
      plugins: [Quasar, pinia],
      mocks: {
        $route: {},
        $router: mockRouter,
      },
      directives: {
        clickOutside: vi.fn(),
        permission: vi.fn(),
      },
      stubs: {
        QPage: { template: "<div><slot /></div>" },
      },
    },
  });
}

describe("OverviewIndex.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("component rendering", () => {
    it("should render correctly", async () => {
      const wrapper = mountComponent();
      await flushPromises();
      expect(wrapper.exists()).toBe(true);
    });
  });

  describe("widget filtering based on permissions", () => {
    describe("READ Nodes permission", () => {
      it("should include node widgets when user has READ Nodes permission", async () => {
        const wrapper = mountComponent({ hasReadNodesPermission: true });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.ReadyNodesWidget);
        expect(allWidgets).toContain(EWidgetComponent.ReadyGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.NodePoolFreeResourcesWidget);
      });

      it("should exclude node widgets when user lacks READ Nodes permission", async () => {
        const wrapper = mountComponent({ hasReadNodesPermission: false });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.ReadyGpuWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.NodePoolFreeResourcesWidget);
        // Note: IdleAllocatedGpuWidget is NOT excluded here because it requires READ Nodes OR READ Workloads,
        // and hasReadWorkloadsPermission defaults to true
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });

      it("should keep other widgets when user lacks READ Nodes permission", async () => {
        const wrapper = mountComponent({ hasReadNodesPermission: false });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        // These widgets should still be present
        expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.RunningWorkloads);
        expect(allWidgets).toContain(EWidgetComponent.PendingWorkloads);
        expect(allWidgets).toContain(EWidgetComponent.FailedWorkloads);
      });
    });

    describe("READ Clusters permission", () => {
      it("should include cluster widgets when user has READ Clusters permission", async () => {
        const wrapper = mountComponent({ hasReadClustersPermission: true });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
      });

      it("should exclude overtime widgets when user lacks cluster and org-unit permissions", async () => {
        const wrapper = mountComponent({
          hasReadClustersPermission: false,
          hasReadDepartmentsPermission: false,
          hasReadProjectsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
      });

      it("should include overtime widgets when user has READ Departments but not READ Clusters", async () => {
        const wrapper = mountComponent({
          hasReadClustersPermission: false,
          hasReadDepartmentsPermission: true,
          hasReadProjectsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
      });

      it("should include overtime widgets when user has READ Projects but not READ Clusters", async () => {
        const wrapper = mountComponent({
          hasReadClustersPermission: false,
          hasReadDepartmentsPermission: false,
          hasReadProjectsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
      });

      it("should include cluster widgets when user has READ NodePools permission and nodepool filter is active", async () => {
        // Setup storage mock to return nodepool filter
        vi.mocked(storageUtil.get).mockImplementation((key) => {
          if (key === "overview-filters") {
            return {
              clusterUuid: "test-cluster-id",
              columnFilters: [{ field: "nodepool", term: "default" }],
            };
          }
          return null;
        });

        const wrapper = mountComponent({
          hasReadClustersPermission: false,
          hasReadNodePoolsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);

        // Reset the mock
        vi.mocked(storageUtil.get).mockReturnValue(null);
      });

      it("should exclude overtime widgets when user has READ NodePools but no nodepool filter and no org-unit permissions", async () => {
        const wrapper = mountComponent({
          hasReadClustersPermission: false,
          hasReadNodePoolsPermission: true,
          hasReadDepartmentsPermission: false,
          hasReadProjectsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
      });

      it("should keep non-cluster widgets when user lacks READ Clusters permission", async () => {
        const wrapper = mountComponent({ hasReadClustersPermission: false });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        // These widgets should still be present
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadStatus);
        expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);
      });
    });

    describe("combined permission scenarios", () => {
      it("should exclude both node and overtime widgets when user lacks nodes, clusters, and org-unit permissions", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: false,
          hasReadClustersPermission: false,
          hasReadDepartmentsPermission: false,
          hasReadProjectsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        // Node widgets should be hidden
        expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.ReadyGpuWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.NodePoolFreeResourcesWidget);

        // Overtime widgets should be hidden
        expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);

        // Widgets requiring READ Nodes OR READ Workloads should still be visible (has READ Workloads by default)
        expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);
      });

      it("should remove empty rows from layout", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: false,
          hasReadClustersPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];

        // Verify no empty rows exist
        widgetsLayout.forEach((row) => {
          expect(row.length).toBeGreaterThan(0);
        });
      });
    });

    describe("READ Workloads permission", () => {
      it("should include workload widgets when user has READ Workloads permission", async () => {
        const wrapper = mountComponent({ hasReadWorkloadsPermission: true });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.IdleGpuWorkloadWidget);
        expect(allWidgets).toContain(EWidgetComponent.RunningWorkloads);
        expect(allWidgets).toContain(EWidgetComponent.PendingWorkloads);
        expect(allWidgets).toContain(EWidgetComponent.FailedWorkloads);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadStatus);
        expect(allWidgets).toContain(EWidgetComponent.WorkloadPendingTimeDistributionWidget);
      });

      it("should exclude workload widgets when user lacks READ Workloads permission", async () => {
        const wrapper = mountComponent({ hasReadWorkloadsPermission: false });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.IdleGpuWorkloadWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
        expect(allWidgets).not.toContain(EWidgetComponent.PendingWorkloads);
        expect(allWidgets).not.toContain(EWidgetComponent.FailedWorkloads);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByTypeWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadStatus);
        expect(allWidgets).not.toContain(EWidgetComponent.WorkloadPendingTimeDistributionWidget);
        // Note: IdleAllocatedGpuWidget is NOT excluded here because it requires READ Nodes OR READ Workloads,
        // and hasReadNodesPermission defaults to true
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });
    });

    describe("READ Departments permission", () => {
      it("should include department widgets when user has READ Departments permission", async () => {
        const wrapper = mountComponent({ hasReadDepartmentsPermission: true });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.GpuAllocationRatioWidget);
      });

      it("should exclude department widgets when user lacks READ Departments permission", async () => {
        const wrapper = mountComponent({ hasReadDepartmentsPermission: false });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);
      });
    });

    describe("OrgUnitRankingWidget permission (READ Projects OR READ Departments)", () => {
      it("should include OrgUnitRankingWidget when user has READ Projects permission", async () => {
        const wrapper = mountComponent({
          hasReadProjectsPermission: true,
          hasReadDepartmentsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);
      });

      it("should include OrgUnitRankingWidget when user has READ Departments permission", async () => {
        const wrapper = mountComponent({
          hasReadProjectsPermission: false,
          hasReadDepartmentsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);
      });

      it("should include OrgUnitRankingWidget when user has both permissions", async () => {
        const wrapper = mountComponent({
          hasReadProjectsPermission: true,
          hasReadDepartmentsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);
      });

      it("should exclude OrgUnitRankingWidget when user has neither permission", async () => {
        const wrapper = mountComponent({
          hasReadProjectsPermission: false,
          hasReadDepartmentsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.OrgUnitRankingWidget);
      });
    });

    describe("Widgets requiring READ Nodes OR READ Workloads permission", () => {
      it("should include nodesOrWorkloads widgets when user has READ Nodes permission", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: true,
          hasReadWorkloadsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });

      it("should include nodesOrWorkloads widgets when user has READ Workloads permission", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: false,
          hasReadWorkloadsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });

      it("should include nodesOrWorkloads widgets when user has both permissions", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: true,
          hasReadWorkloadsPermission: true,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });

      it("should exclude nodesOrWorkloads widgets when user has neither permission", async () => {
        const wrapper = mountComponent({
          hasReadNodesPermission: false,
          hasReadWorkloadsPermission: false,
        });
        await flushPromises();

        const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
        const allWidgets = widgetsLayout.flat();

        expect(allWidgets).not.toContain(EWidgetComponent.AllocatedGpuWidget);
        expect(allWidgets).not.toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      });
    });
  });

  describe("filter options based on permissions", () => {
    it("should include node pool filter when user has READ Node Pools permission", async () => {
      const wrapper = mountComponent({ hasReadNodePoolsPermission: true });
      await flushPromises();

      const filterOptions = wrapper.vm.filterOptions;
      expect(filterOptions).toHaveLength(1);
      expect(filterOptions[0].field).toBe("nodepool");
      expect(filterOptions[0].name).toBe("Node pool");
    });

    it("should exclude node pool filter when user lacks READ Node Pools permission", async () => {
      const wrapper = mountComponent({ hasReadNodePoolsPermission: false });
      await flushPromises();

      const filterOptions = wrapper.vm.filterOptions;
      expect(filterOptions).toHaveLength(0);
    });
  });

  describe("minimal permission scenarios", () => {
    it("should show only node widgets when user has only READ Nodes permission", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: true,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Node widgets should be visible
      expect(allWidgets).toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).toContain(EWidgetComponent.ReadyGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.NodePoolFreeResourcesWidget);

      // Widgets requiring READ Nodes OR READ Workloads should be visible (has READ Nodes)
      expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);

      // Workload-only widgets should be hidden (no READ Workloads)
      expect(allWidgets).not.toContain(EWidgetComponent.IdleGpuWorkloadWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.OrgUnitRankingWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
    });

    it("should show only workload widgets when user has only READ Workloads permission", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: true,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Workload widgets should be visible
      expect(allWidgets).toContain(EWidgetComponent.IdleGpuWorkloadWidget);
      expect(allWidgets).toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.PendingWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.FailedWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadStatus);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadPendingTimeDistributionWidget);

      // Widgets requiring READ Nodes OR READ Workloads should be visible (has READ Workloads)
      expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);

      // Node widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyGpuWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.NodePoolFreeResourcesWidget);

      // Other widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.OrgUnitRankingWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
    });

    it("should show only department widgets when user has only READ Departments permission", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: true,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Department widgets should be visible
      expect(allWidgets).toContain(EWidgetComponent.GpuAllocationRatioWidget);

      // OrgUnitRankingWidget should be visible (has READ Departments)
      expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);

      // Other widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).not.toContain(EWidgetComponent.AllocatedGpuWidget);
    });

    it("should show only OrgUnitRankingWidget when user has only READ Projects permission", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: true,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // OrgUnitRankingWidget should be visible (has READ Projects)
      expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);

      // GpuAllocationRatioWidget should be hidden (requires READ Departments)
      expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);

      // Other widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).not.toContain(EWidgetComponent.AllocatedGpuWidget);
    });

    it("should show only cluster widgets when user has only READ Clusters permission", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: true,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Cluster widgets should be visible
      expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);

      // Other widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).not.toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.OrgUnitRankingWidget);
    });
  });

  describe("edge cases", () => {
    it("should show no widgets when user has no permissions", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // All widgets should be hidden
      expect(allWidgets).toHaveLength(0);
      expect(widgetsLayout).toHaveLength(0);
    });

    it("should show all widgets when user has all permissions", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: true,
        hasReadClustersPermission: true,
        hasReadNodePoolsPermission: true,
        hasReadWorkloadsPermission: true,
        hasReadDepartmentsPermission: true,
        hasReadProjectsPermission: true,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // All widget types should be visible
      expect(allWidgets).toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).toContain(EWidgetComponent.ReadyGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.NodePoolFreeResourcesWidget);
      expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.PendingWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.FailedWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadStatus);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadPendingTimeDistributionWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleGpuWorkloadWidget);
      expect(allWidgets).toContain(EWidgetComponent.GpuAllocationRatioWidget);
      expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);
      expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
    });

    it("should exclude overtime widgets when user lacks cluster, nodepool, and org-unit permissions", async () => {
      const wrapper = mountComponent({
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
    });

    it("should not show overtime widgets when user has READ NodePools but filter is empty string and no org-unit permissions", async () => {
      vi.mocked(storageUtil.get).mockImplementation((key) => {
        if (key === "overview-filters") {
          return {
            clusterUuid: "test-cluster-id",
            columnFilters: [{ field: "nodepool", term: "" }],
          };
        }
        return null;
      });

      const wrapper = mountComponent({
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: true,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Empty string is falsy and no org-unit access, so overtime widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).not.toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);

      vi.mocked(storageUtil.get).mockReturnValue(null);
    });

    it("should show cluster widgets when user has READ Clusters regardless of READ NodePools", async () => {
      const wrapper = mountComponent({
        hasReadClustersPermission: true,
        hasReadNodePoolsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);
    });
  });

  describe("IdleAllocatedGpuWidget specific behavior (READ Nodes OR READ Workloads)", () => {
    it("should show IdleAllocatedGpuWidget when user has READ Nodes but not READ Workloads", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: true,
        hasReadWorkloadsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // IdleAllocatedGpuWidget requires READ Nodes OR READ Workloads
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
    });

    it("should show IdleAllocatedGpuWidget when user has READ Workloads but not READ Nodes", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadWorkloadsPermission: true,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);
    });

    it("should hide IdleAllocatedGpuWidget when user has neither READ Nodes nor READ Workloads", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadWorkloadsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).not.toContain(EWidgetComponent.IdleAllocatedGpuWidget);
    });
  });

  describe("IdleGpuWorkloadWidget specific behavior (READ Workloads only)", () => {
    it("should hide IdleGpuWorkloadWidget when user has READ Nodes but not READ Workloads", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: true,
        hasReadWorkloadsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // IdleGpuWorkloadWidget requires READ Workloads
      expect(allWidgets).not.toContain(EWidgetComponent.IdleGpuWorkloadWidget);
    });

    it("should show IdleGpuWorkloadWidget when user has READ Workloads but not READ Nodes", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadWorkloadsPermission: true,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).toContain(EWidgetComponent.IdleGpuWorkloadWidget);
    });

    it("should hide IdleGpuWorkloadWidget when user has neither READ Nodes nor READ Workloads", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadWorkloadsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      expect(allWidgets).not.toContain(EWidgetComponent.IdleGpuWorkloadWidget);
    });
  });

  describe("complex multi-permission scenarios", () => {
    it("should correctly handle READ Nodes + READ Departments combination", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: true,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: true,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Node widgets
      expect(allWidgets).toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).toContain(EWidgetComponent.ReadyGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.NodePoolFreeResourcesWidget);

      // READ Nodes OR READ Workloads widgets (via READ Nodes)
      expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);

      // Department widgets
      expect(allWidgets).toContain(EWidgetComponent.GpuAllocationRatioWidget);
      expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);

      // Workload-only widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.IdleGpuWorkloadWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);
    });

    it("should correctly handle READ Workloads + READ Projects combination", async () => {
      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: false,
        hasReadNodePoolsPermission: false,
        hasReadWorkloadsPermission: true,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: true,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Workload widgets
      expect(allWidgets).toContain(EWidgetComponent.IdleGpuWorkloadWidget);
      expect(allWidgets).toContain(EWidgetComponent.RunningWorkloads);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByTypeWidget);

      // READ Nodes OR READ Workloads widgets (via READ Workloads)
      expect(allWidgets).toContain(EWidgetComponent.AllocatedGpuWidget);
      expect(allWidgets).toContain(EWidgetComponent.IdleAllocatedGpuWidget);

      // OrgUnitRankingWidget (via READ Projects)
      expect(allWidgets).toContain(EWidgetComponent.OrgUnitRankingWidget);

      // Node widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);

      // Department-only widget should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.GpuAllocationRatioWidget);
    });

    it("should correctly handle READ Clusters + READ NodePools with filter", async () => {
      vi.mocked(storageUtil.get).mockImplementation((key) => {
        if (key === "overview-filters") {
          return {
            clusterUuid: "test-cluster-id",
            columnFilters: [{ field: "nodepool", term: "gpu-pool" }],
          };
        }
        return null;
      });

      const wrapper = mountComponent({
        hasReadNodesPermission: false,
        hasReadClustersPermission: true,
        hasReadNodePoolsPermission: true,
        hasReadWorkloadsPermission: false,
        hasReadDepartmentsPermission: false,
        hasReadProjectsPermission: false,
      });
      await flushPromises();

      const widgetsLayout = wrapper.vm.widgetsLayout as EWidgetComponent[][];
      const allWidgets = widgetsLayout.flat();

      // Cluster widgets should be visible
      expect(allWidgets).toContain(EWidgetComponent.OverviewTimeRangeWidgets);
      expect(allWidgets).toContain(EWidgetComponent.WorkloadByCategoryTimeRangeWidget);

      // Other widgets should be hidden
      expect(allWidgets).not.toContain(EWidgetComponent.ReadyNodesWidget);
      expect(allWidgets).not.toContain(EWidgetComponent.RunningWorkloads);

      vi.mocked(storageUtil.get).mockReturnValue(null);
    });
  });
});
