/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";

//routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
//svc
import { workloadService } from "@/services/cluster/workload.service/workload.service";
//utils
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { unitTestUtil } from "@/utils/unit-test.util";
//cmps
import { WorkloadListModal } from "@/components/workload/workload-list-modal";
//mock
import { mockWorkloads } from "@/components/workload/workload-list-modal/workload-list-modal.mock";
//models
import { EWorkloadModalEntity, type IWorkloadListModalOptions } from "@/models/workload.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { workloadListModalColumns } from "@/table-models/workload.table-model";
import { EColumnFilterType, EFilterOperator, type IAdvancedFilterModel } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { nodePoolWorkloadListModalColumns } from "@/table-models/node-pool.table-model";

const clusterId = "34234-322-sdfs323";
vi.mock("@/services/cluster/workload.service/workload.service");
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("workload-list-modal component", () => {
  let wrapper: VueWrapper<InstanceType<typeof WorkloadListModal>>;

  const mockRouter = {
    push: vi.fn(),
  };
  const createComponent = (props: { modalOptions: IWorkloadListModalOptions }): void => {
    wrapper = mount(WorkloadListModal, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: props,
    });
    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };
  //mock api calls
  const mockGetWorkloadsCount = vi.mocked(workloadService.getWorkloadsCount);
  const mockGetWorkloads = vi.mocked(workloadService.getWorkloads);

  async function testRedirectToWorkloads(advancedFilters: IAdvancedFilterModel[]) {
    const saveAndNavigateToWorkloadsSpy = vi.spyOn(wrapper.vm, "saveAndNavigateToWorkloads");
    const setColumnAdvancedFilterSpy = vi.spyOn(filterService, "setColumnAdvancedFilter");
    const modalEl = unitTestUtil.getModalElement();
    await modalEl.find(".btn-link").trigger("click");

    expect(wrapper.vm.saveAndNavigateToWorkloads).toHaveBeenCalledTimes(1);
    expect(setColumnAdvancedFilterSpy).toHaveBeenCalledTimes(1);
    if (advancedFilters.length > 0) {
      expect(saveAndNavigateToWorkloadsSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          advancedFilters,
        }),
      );
    } else {
      expect(saveAndNavigateToWorkloadsSpy).not.toHaveBeenCalledWith(
        expect.objectContaining({
          advancedFilters: expect.anything(),
        }),
      );
    }

    expect(mockRouter.push).toHaveBeenCalledWith({
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
    });
  }
  describe("project entity", () => {
    beforeEach(async () => {
      mockGetWorkloadsCount.mockResolvedValue(mockWorkloads.length);
      mockGetWorkloads.mockResolvedValue(mockWorkloads);
      createComponent({
        modalOptions: {
          clusterId: clusterId,
          entityFilter: "projectId=1",
          entityName: "suricata-project",
          filterName: WorkloadSortFilterFields.ProjectName,
          entityType: EWorkloadModalEntity.Project,
          columns: workloadListModalColumns,
        },
      });
      await flushPromises();
    });

    afterEach(() => {
      wrapper.unmount();
      vi.clearAllMocks();
    });
    test("Modal load current entity data correctly", async () => {
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["projectId=1", "clusterId==34234-322-sdfs323"],
        undefined,
      );
    });

    test("Table Rendering Test", () => {
      expect(wrapper.vm.$data.filterBy.searchTerm).toBe("");
      expect(wrapper.findComponent(".runai-base-dialog-header").text()).eql(
        "Workloads Associated with Project suricata-project",
      );
    });

    test("Search Functionality Test", async () => {
      const modalEl = unitTestUtil.getModalElement();
      await modalEl.find("[aid=runai-list-modal-search]").setValue("pytorch");
      await flushPromises();
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["projectId=1", "clusterId==34234-322-sdfs323"],
        "pytorch",
      );
    });

    test("Table Data Loading Test", () => {
      const modalEl = unitTestUtil.getModalElement();
      expect(modalEl.find(".workloads-list-table").findAll(".runai-table-row").length).toEqual(2);
      modalEl
        .find(".workloads-list-table")
        .findAll('[aid="name"]')
        .forEach((el, index) => {
          expect(el.text()).eql(mockWorkloads[index].name);
        });
    });

    test("Redirect to Workloads Test", async () => {
      const modalEl = unitTestUtil.getModalElement();
      await modalEl.find(".btn-link").trigger("click");

      expect(mockRouter.push).toHaveBeenCalledTimes(1);
      expect(mockRouter.push).toHaveBeenCalledWith({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      });
    });

    test("Close Modal Test", async () => {
      const modalEl = unitTestUtil.getModalElement();
      await modalEl.find("[aid=close-workloads-modal-btn]").trigger("click");
      expect(wrapper.emitted("close")).toBeTruthy();
    });

    test("Redirect to workloads with project filter", async () => {
      await testRedirectToWorkloads([
        {
          label: "Project",
          name: WorkloadSortFilterFields.ProjectName,
          term: "suricata-project",
          type: EColumnFilterType.FreeText,
          value: EFilterOperator.Equals,
        },
      ]);
    });
  });

  describe("department entity", () => {
    beforeEach(async () => {
      mockGetWorkloadsCount.mockResolvedValue(mockWorkloads.length);
      mockGetWorkloads.mockResolvedValue(mockWorkloads);
      createComponent({
        modalOptions: {
          clusterId: "34234-322-sdfs323",
          entityFilter: "departmentId=1",
          entityName: "platform",
          filterName: WorkloadSortFilterFields.DepartmentName,
          entityType: EWorkloadModalEntity.Department,
          columns: workloadListModalColumns,
        },
      });
      await flushPromises();
    });

    afterEach(() => {
      wrapper.unmount();
      vi.clearAllMocks();
    });

    test("Modal load current entity data correctly", async () => {
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["departmentId=1", "clusterId==34234-322-sdfs323"],
        undefined,
      );
    });

    test("Table Rendering Test", () => {
      expect(wrapper.vm.$data.filterBy.searchTerm).toBe("");
      expect(wrapper.findComponent(".runai-base-dialog-header").text()).eql(
        "Workloads Associated with Department platform",
      );
    });

    test("Search Functionality Test", async () => {
      const modalEl = unitTestUtil.getModalElement();
      await modalEl.find("[aid=runai-list-modal-search]").setValue("tensorboard");
      await flushPromises();

      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["departmentId=1", "clusterId==34234-322-sdfs323"],
        "tensorboard",
      );
      await flushPromises();

      //clear input
      await modalEl.find("[aid=runai-list-modal-search-clear]").trigger("click");
      await flushPromises();
      //validate empty name filter
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["departmentId=1", "clusterId==34234-322-sdfs323"],
        undefined,
      );
    });

    test("Redirect to workloads with department filter", async () => {
      await testRedirectToWorkloads([
        {
          label: "Department",
          name: WorkloadSortFilterFields.DepartmentName,
          term: "platform",
          type: EColumnFilterType.FreeText,
          value: EFilterOperator.Equals,
        },
      ]);
    });
  });
  describe("nodepool entity", () => {
    beforeEach(async () => {
      mockGetWorkloadsCount.mockResolvedValue(mockWorkloads.length);
      mockGetWorkloads.mockResolvedValue(mockWorkloads);
      createComponent({
        modalOptions: {
          clusterId: "34234-322-sdfs323",
          entityFilter: "nodePool=1",
          entityName: "default",
          filterName: WorkloadSortFilterFields.Nodepool,
          entityType: EWorkloadModalEntity.Nodepool,
          columns: nodePoolWorkloadListModalColumns,
        },
      });
      await flushPromises();
    });

    afterEach(() => {
      wrapper.unmount();
      vi.clearAllMocks();
    });

    test("Modal load current entity data correctly", async () => {
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["nodePool=1", "clusterId==34234-322-sdfs323"],
        undefined,
      );
    });

    test("Table Rendering Test", () => {
      expect(wrapper.vm.$data.filterBy.searchTerm).toBe("");
      expect(wrapper.findComponent(".runai-base-dialog-header").text()).eql(
        "Workloads Associated with Nodepool default",
      );
    });

    test("Search Functionality Test", async () => {
      const modalEl = unitTestUtil.getModalElement();
      await modalEl.find("[aid=runai-list-modal-search]").setValue("tensorboard");
      await flushPromises();

      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["nodePool=1", "clusterId==34234-322-sdfs323"],
        undefined,
      );
      await flushPromises();

      //clear input
      await modalEl.find("[aid=runai-list-modal-search-clear]").trigger("click");
      await flushPromises();
      //validate empty name filter
      expect(mockGetWorkloads).toHaveBeenCalledWith(
        false,
        0,
        10,
        "asc",
        "name",
        ["nodePool=1", "clusterId==34234-322-sdfs323"],
        "tensorboard",
      );
    });

    test("Redirect to workloads", async () => {
      await testRedirectToWorkloads([
        {
          label: "Nodepool",
          name: WorkloadSortFilterFields.Nodepool,
          term: "default",
          type: EColumnFilterType.FreeText,
          value: EFilterOperator.Equals,
        },
      ]);
    });
  });
});
