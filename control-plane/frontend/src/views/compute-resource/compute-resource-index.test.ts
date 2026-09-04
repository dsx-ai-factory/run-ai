/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import computeResourceIndex from "./compute-resource-index.vue";
import type { DefineComponent } from "vue";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { computeResourceIndexMock } from "./compute-resource-index.mock";
import { setQuasarMock } from "@/utils/unit-test.util/unit-test.util";
import type { IFilterBy } from "@/models/filter.model";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import type { ComputeAsset, ComputeUpdateRequest } from "@/swagger-models/assets-service-client";
import { useRoute, type RouteMeta, type RouteLocation } from "vue-router";
import { deepCopy } from "@/utils/common.util";

vi.mock("@/services/infra/unleash.service/unleash.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
const computeStore = useComputeResourceStore(pinia);

computeStore.$patch({ computeResources: computeResourceIndexMock.computelist });
setActivePinia(pinia);

describe("compute-resource-index.vue", () => {
  let computeResourceIndexCmp: VueWrapper<InstanceType<typeof computeResourceIndex>>;

  const computeListSpy = vi.spyOn(computeStore, "loadComputeResources");

  function mountComponent(query = {}) {
    vi.mock("vue-router", () => ({ useRoute: vi.fn() }));
    vi.mocked(useRoute).mockReturnValue({
      name: "",
      path: "",
      params: {},
      meta: {} as RouteMeta,
      matched: [],
      fullPath: "",
      hash: "",
      redirectedFrom: {} as RouteLocation,
      query,
    });

    computeResourceIndexCmp = mount(computeResourceIndex, {
      global: {
        plugins: [Quasar, pinia],
        stubs: ["router-link"],
        directives: {
          clickOutside: vi.fn(),
          permission: vi.fn(),
        },
        mocks: {
          $route: {
            meta: {
              prevRoute: "",
            },
            query,
          },
          $router: [],
        },
      },
    });

    setQuasarMock(computeResourceIndexCmp);
  }

  beforeEach(() => {
    computeListSpy.mockResolvedValue(computeResourceIndexMock.computelist);
    mountComponent();
  });
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("component is loading properly", async () => {
    expect(computeResourceIndexCmp.classes()).toContain("compute-resource-index");
  });

  describe("component setup ('created' lifecycle logic))", () => {
    test("the component will invoke 'loadComputeResources' in the store with the filterBy saved in the storage if it exists", async () => {
      window.localStorage.setItem(
        "compute-resource-table-filters",
        JSON.stringify(computeResourceIndexMock.filterByFromStorage),
      );

      const expectedCalledWith = deepCopy(computeResourceIndexMock.filterByFromStorage);

      mountComponent();

      expect(computeStore.loadComputeResources).toHaveBeenCalledWith({ ...expectedCalledWith, clusterUuid: undefined });
    });
  });

  describe("handling page actions emits", () => {
    test("component will update the compute list whenever 'filter-changed' was emitted by runai-page-actions", async () => {
      const newFilter: IFilterBy = { ...computeResourceIndexMock.filterBy, rowsPerPage: 15, searchTerm: "test101" };
      computeResourceIndexCmp
        .findComponent<DefineComponent>(".runai-page-actions")
        .vm.$emit("filters-changed", newFilter);
      expect(computeResourceIndexCmp.vm.filterBy).toEqual(newFilter);
      expect(computeStore.loadComputeResources).toHaveBeenCalledWith(newFilter);
    });

    test("component will reset selected rows when 'selected-actions-close' is emitted by runai-page-actions", async () => {
      computeResourceIndexCmp.findComponent<DefineComponent>(".runai-page-actions").vm.$emit("selected-actions-close");
      expect(computeResourceIndexCmp.vm.selectedRows).toEqual([]);
      expect(computeStore.setSelectedComputeResource).toHaveBeenCalledWith(null);
    });

    test("component will redirect the user to the compute resource creation page when 'create-entity' is emitted by runai-page-actions", async () => {
      computeResourceIndexCmp.findComponent<DefineComponent>(".runai-page-actions").vm.$emit("create-entity");
      expect(computeResourceIndexCmp.vm.$router).toContainEqual({ name: "compute-resource-new", query: {} });
    });

    test("the component will open the rename modal when a compute resource is selected and and rename button is pressed", async () => {
      await computeResourceIndexCmp.setData({ selectedRows: [computeResourceIndexMock.computelist[0]] });
      await computeResourceIndexCmp.find("[aid=rename-compute-btn]").trigger("click");
      expect(computeResourceIndexCmp.vm.isRenameModalOpen).toBe(true);
    });

    test("the component will open the delete modal when a compute resource is selected and and delete button is pressed", async () => {
      await computeResourceIndexCmp.setData({ selectedRows: [computeResourceIndexMock.computelist[0]] });
      computeResourceIndexCmp.findComponent<DefineComponent>("[aid=delete-compute-btn]").vm.$emit("click"); // for some reason triggering click doesn't work in this case
      await computeResourceIndexCmp.vm.$nextTick();
      expect(computeResourceIndexCmp.vm.isDeleteModalOpen).toBe(true);
    });

    test("the component will redirect the user to the creation page when a compute resource is selected and and copy&edit button is pressed", async () => {
      await computeResourceIndexCmp.setData({ selectedRows: [computeResourceIndexMock.computelist[0]] });
      computeResourceIndexCmp.findComponent<DefineComponent>("[aid=copy-compute-btn]").vm.$emit("click"); // for some reason triggering click doesn't work in this case
      await computeResourceIndexCmp.vm.$nextTick();
      expect(computeResourceIndexCmp.vm.$router).toContainEqual({
        name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
        query: { fromCopyId: computeResourceIndexMock.computelist[0].meta.id },
      });
    });
  });

  describe("handling table emits", () => {
    test("the component will invoke the 'setSelectedComputeResource' action in the store when 'update:selected' is emitted by the table", async () => {
      const selectedCompute: ComputeAsset = computeResourceIndexMock.computelist[0];
      await computeResourceIndexCmp.setData({ selectedRows: [selectedCompute] });
      computeResourceIndexCmp.findComponent<DefineComponent>(".cr-table").vm.$emit("update:selected", [selectedCompute]);
      expect(computeStore.setSelectedComputeResource).toHaveBeenCalledWith(selectedCompute);
    });
  });

  describe("handling modal's emits", () => {
    test("the component will invoke the 'remove' action in the compute resources store with the selected compute data", async () => {
      const selectedCompute: ComputeAsset = computeResourceIndexMock.computelist[0];
      await computeResourceIndexCmp.setData({ selectedRows: [selectedCompute], isDeleteModalOpen: true });

      computeResourceIndexCmp.findComponent<DefineComponent>("[aid=delete-compute-modal]").vm.$emit("delete");
      expect(computeStore.remove).toHaveBeenCalledWith(selectedCompute);
    });

    test("the component will invoke the 'update' action in the compute resources store with the selected compute data", async () => {
      const selectedCompute: ComputeAsset = computeResourceIndexMock.computelist[2];
      await computeResourceIndexCmp.setData({ selectedRows: [selectedCompute], isRenameModalOpen: true });
      const newName = "new-cr-name";
      computeResourceIndexCmp
        .findComponent<DefineComponent>("[aid=rename-compute-modal]")
        .vm.$emit("rename", { name: newName });
      const expectedUpdateRequest: ComputeUpdateRequest = {
        meta: {
          name: newName,
        },
        spec: selectedCompute.spec,
      };
      expect(computeStore.update).toHaveBeenCalledWith(selectedCompute.meta.id, expectedUpdateRequest);
    });
  });
});
