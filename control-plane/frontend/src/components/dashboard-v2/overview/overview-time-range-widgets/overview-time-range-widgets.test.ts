/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, afterEach, describe, expect, vi, it } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import { OverviewTimeRangeWidgets } from "@/components/dashboard-v2/overview/overview-time-range-widgets/index";
import ResourceUtilizationTimeRangeWidget from "@/components/dashboard-v2/widgets/time-range-widgets/resource-utilization-time-range-widget/resource-utilization-time-range-widget.vue";
import ResourceAllocationTimeRangeWidget from "@/components/dashboard-v2/widgets/time-range-widgets/resource-allocation-time-range-widget/resource-allocation-time-range-widget.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { mockClusterMetricsValuesEnriched } from "@/components/dashboard-v2/widgets/time-range-widgets/resource-utilization-time-range-widget/resource-utilization-time-range-widget.mock";
import { mockWorkloadByCategoryGroupedMetrics } from "@/components/dashboard-v2/widgets/time-range-widgets/workload-by-category-time-range-widget/workload-by-category-time-range-widget.mock";
import { useAuthStore } from "@/stores/auth.store";
import { ScopeType } from "@/swagger-models/org-unit-service-client";

// models
import type { IRangeDates } from "@/utils/date.util";
import type { TestingPinia } from "@pinia/testing";
import { createTestingPinia } from "@pinia/testing";
import { EOrgUnitMetricSourceType } from "@/models/org-unit.model";

vi.mock("@/services/control-plane/cluster.service/cluster.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

interface IProps {
  clusterId: string;
  filterByDates?: IRangeDates;
  nodePoolName?: string;
  hasReadClustersPermission?: boolean;
  hasReadDepartmentsPermission?: boolean;
  hasReadProjectsPermission?: boolean;
}

const defaultDates = (): IRangeDates => {
  const dateEnd = new Date();
  const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000);
  return { dateStart, dateEnd, label: "" };
};

const mockDepartments = [
  { id: "dept-1", name: "Engineering", type: ScopeType.Department, parentId: "cluster-1" },
  { id: "dept-2", name: "Research", type: ScopeType.Department, parentId: "cluster-1" },
];

const mockProjects = [
  { id: "proj-1", name: "ML-Training", type: ScopeType.Project, parentId: "dept-1" },
  { id: "proj-2", name: "Inference", type: ScopeType.Project, parentId: "dept-1" },
];

describe("OverviewTimeRangeWidgets.vue", () => {
  let pinia: TestingPinia;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mountComponent(
    props: IProps,
    orgUnits: Array<{ id: string; name: string; type: ScopeType; parentId?: string }> = [],
    mockRejectedValue?: string,
  ): VueWrapper<InstanceType<typeof OverviewTimeRangeWidgets>> {
    pinia = createTestingPinia({ createSpy: vi.fn });

    if (mockRejectedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockRejectedValue(mockRejectedValue);
    } else {
      if (props.nodePoolName) {
        vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue(mockClusterMetricsValuesEnriched);
      } else {
        vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValuesEnriched);
      }
    }

    vi.mocked(orgUnitService.getDepartmentMetrics).mockResolvedValue(mockClusterMetricsValuesEnriched as never);
    vi.mocked(orgUnitService.getProjectMetrics).mockResolvedValue(mockClusterMetricsValuesEnriched as never);
    vi.mocked(clusterService.getClusterGroupedMetrics).mockResolvedValue(mockWorkloadByCategoryGroupedMetrics);

    const authStore = useAuthStore(pinia);
    Object.defineProperty(authStore, "orgUnits", {
      get: () => orgUnits,
      configurable: true,
    });

    return mount(OverviewTimeRangeWidgets, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        clusterId: props.clusterId,
        filterByDates: props.filterByDates,
        nodePoolName: props.nodePoolName,
        hasReadClustersPermission: props.hasReadClustersPermission ?? false,
        hasReadDepartmentsPermission: props.hasReadDepartmentsPermission ?? false,
        hasReadProjectsPermission: props.hasReadProjectsPermission ?? false,
      },
    });
  }

  describe("cluster mode (hasReadClustersPermission)", () => {
    it("renders child component with loading state", async () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadClustersPermission: true,
      });

      expect(wrapper.findComponent(ResourceUtilizationTimeRangeWidget).exists()).toBe(true);

      await flushPromises();

      expect(clusterService.getClusterMetrics).toHaveBeenCalled();
      expect(wrapper.find(".error-section").exists()).toBe(false);
    });

    it("renders child component with error state", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadClustersPermission: true,
        },
        [],
        "failed to load data",
      );

      expect(wrapper.findComponent(ResourceUtilizationTimeRangeWidget).exists()).toBe(true);

      await flushPromises();

      expect(clusterService.getClusterMetrics).toHaveBeenCalled();
      expect(wrapper.find(".error-section").exists()).toBe(true);
      expect(wrapper.find(".chart-container").exists()).toBe(false);
    });

    it("passes props correctly to child component", async () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        nodePoolName: "default",
        hasReadClustersPermission: true,
      });

      await flushPromises();

      const childComponent = wrapper.findComponent(ResourceUtilizationTimeRangeWidget);
      expect(childComponent.props().clusterId).toBe("someClusterId");
      expect(childComponent.props().filterByDates.dateStart).toBeInstanceOf(Date);
      expect(childComponent.props().filterByDates.dateEnd).toBeInstanceOf(Date);
      expect(childComponent.props().nodePoolName).toBe("default");
    });

    it("handles prop changes correctly", async () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadClustersPermission: true,
      });

      await flushPromises();

      const newDateEnd = new Date();
      const newDateStart = new Date(newDateEnd.getTime() - 48 * 60 * 60 * 1000);
      await wrapper.setProps({
        clusterId: "newClusterId",
        filterByDates: { dateStart: newDateStart, dateEnd: newDateEnd, label: "newLabel" },
      });

      const childComponent = wrapper.findComponent(ResourceUtilizationTimeRangeWidget);
      expect(childComponent.props().clusterId).toBe("newClusterId");
      expect(childComponent.props().filterByDates.dateStart).toBeInstanceOf(Date);
      expect(childComponent.props().filterByDates.dateEnd).toBeInstanceOf(Date);
    });

    it("does not show org-unit filters in cluster mode", () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadClustersPermission: true,
      });

      expect(wrapper.find("[aid='overtime-department-filter']").exists()).toBe(false);
      expect(wrapper.find("[aid='overtime-project-filter']").exists()).toBe(false);
    });

    it("does not show empty state in cluster mode", () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadClustersPermission: true,
      });

      expect(wrapper.find(".empty-state-message").exists()).toBe(false);
    });
  });

  describe("org-unit mode (department/project permissions)", () => {
    it("shows department filter when hasReadDepartmentsPermission is true", () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
        },
        mockDepartments,
      );

      expect(wrapper.find("[aid='overtime-department-filter']").exists()).toBe(true);
    });

    it("auto-selects first department and shows widgets", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
        },
        mockDepartments,
      );

      await flushPromises();

      expect(wrapper.vm.selectedDepartment).toBe("dept-1");
      expect(wrapper.findComponent(ResourceUtilizationTimeRangeWidget).exists()).toBe(true);
      expect(wrapper.findComponent(ResourceAllocationTimeRangeWidget).exists()).toBe(true);
    });

    it("passes orgUnitMetricSource to child widgets when department is selected", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
        },
        mockDepartments,
      );

      await flushPromises();

      const utilizationWidget = wrapper.findComponent(ResourceUtilizationTimeRangeWidget);
      expect(utilizationWidget.props().orgUnitMetricSource).toEqual({
        type: EOrgUnitMetricSourceType.Department,
        id: "dept-1",
      });
    });

    it("auto-selects first project when hasReadProjectsPermission is true", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadProjectsPermission: true,
        },
        mockProjects,
      );

      await flushPromises();

      expect(wrapper.vm.selectedProject).toBe("proj-1");
    });

    it("passes project metric source when project is selected", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadProjectsPermission: true,
        },
        mockProjects,
      );

      await flushPromises();

      const utilizationWidget = wrapper.findComponent(ResourceUtilizationTimeRangeWidget);
      expect(utilizationWidget.props().orgUnitMetricSource).toEqual({
        type: EOrgUnitMetricSourceType.Project,
        id: "proj-1",
      });
    });

    it("shows empty state when no department is selected", () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadDepartmentsPermission: true,
      });

      expect(wrapper.find(".empty-state-message").exists()).toBe(true);
      expect(wrapper.find(".empty-state-message").text()).toBe("Select a department to view overtime metrics");
    });

    it("shows project-specific empty state for project-only users", () => {
      const wrapper = mountComponent({
        clusterId: "someClusterId",
        filterByDates: defaultDates(),
        hasReadProjectsPermission: true,
      });

      expect(wrapper.find(".empty-state-message").exists()).toBe(true);
      expect(wrapper.find(".empty-state-message").text()).toBe("Select a project to view overtime metrics");
    });

    it("hides project filter when no projects available", () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
          hasReadProjectsPermission: true,
        },
        mockDepartments, // no projects in org units
      );

      expect(wrapper.find("[aid='overtime-project-filter']").exists()).toBe(false);
    });

    it("resets project selection when department changes", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
          hasReadProjectsPermission: true,
        },
        [...mockDepartments, ...mockProjects],
      );

      await flushPromises();

      // Auto-selected first department and project
      expect(wrapper.vm.selectedDepartment).toBe("dept-1");
      expect(wrapper.vm.selectedProject).toBe("proj-1");

      // Change department (component emits the ID directly)
      wrapper.vm.handleDepartmentChange("dept-2");
      expect(wrapper.vm.selectedDepartment).toBe("dept-2");
      expect(wrapper.vm.selectedProject).toBeNull();
    });

    it("project metric source takes priority over department", async () => {
      const wrapper = mountComponent(
        {
          clusterId: "someClusterId",
          filterByDates: defaultDates(),
          hasReadDepartmentsPermission: true,
          hasReadProjectsPermission: true,
        },
        [...mockDepartments, ...mockProjects],
      );

      await flushPromises();

      // Both dept and project selected — project wins
      expect(wrapper.vm.orgUnitMetricSource).toEqual({
        type: EOrgUnitMetricSourceType.Project,
        id: "proj-1",
      });
    });
  });
});
