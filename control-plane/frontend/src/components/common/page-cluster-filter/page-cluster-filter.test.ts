/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { PageClusterFilter } from "@/components/common/page-cluster-filter/";
import { afterEach, beforeAll, describe, expect, test, vi } from "vitest";
import { Quasar } from "quasar";
import { unitTestUtil } from "@/utils/unit-test.util";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { useClusterStore } from "@/stores/cluster.store";
import { ALL_CLUSTER_FILTER_LABEL, ALL_CLUSTER_FILTER_VALUE } from "@/models/filter.model";
interface IClusterFilterProps {
  filterValue: string;
  hideAllOption?: boolean;
  applyGlobally?: boolean;
}
const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

describe("page-cluster-filter.vue", () => {
  let pageClusterFilterWrapper: VueWrapper<InstanceType<typeof PageClusterFilter>>;
  const clusterStore = useClusterStore(pinia);

  const mockClusters = [
    { name: "cluster-1", uuid: "1" },
    { name: "cluster-2", uuid: "2" },
    { name: "cluster-3", uuid: "3" },
  ];

  function mountComponent(
    props = { filterValue: "", hideAllOption: false, applyGlobally: false } as IClusterFilterProps,
  ) {
    pageClusterFilterWrapper = mount(PageClusterFilter, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  async function clickOnChip() {
    return pageClusterFilterWrapper.find(".runai-filter-chip").trigger("click");
  }
  beforeAll(() => {
    clusterStore.$patch({ clusters: mockClusters });
    clusterStore.$patch({ selectedCluster: mockClusters[0] });
  });

  afterEach(() => {
    pageClusterFilterWrapper.unmount();
  });

  test("component is loading properly", () => {
    mountComponent();
    expect(pageClusterFilterWrapper.html()).toContain("Cluster:");
    expect(pageClusterFilterWrapper.classes()).toContain("runai-filter-chip");
    expect(pageClusterFilterWrapper.find('[aid="cluster-filter-chip"]').exists()).toBe(true);
  });

  test("sets data correctly", () => {
    mountComponent();
    expect(pageClusterFilterWrapper.vm.options).toEqual([
      { label: ALL_CLUSTER_FILTER_LABEL, value: "" },
      { label: mockClusters[0].name, value: mockClusters[0].uuid },
      { label: mockClusters[1].name, value: mockClusters[1].uuid },
      { label: mockClusters[2].name, value: mockClusters[2].uuid },
    ]);
  });

  describe("renders values", () => {
    test("renders selected value correctly", () => {
      mountComponent({ filterValue: mockClusters[0].uuid });

      expect(pageClusterFilterWrapper.find(".filter-chip-title").text()).toBe("Cluster:");
      expect(pageClusterFilterWrapper.find(".filter-chip-value").text()).toBe(mockClusters[0].name);
    });
  });

  describe("emits", () => {
    test("emits 'update-filter' event on save", async () => {
      mountComponent({ filterValue: mockClusters[0].uuid });
      const newSelection = mockClusters[2];

      await clickOnChip();
      const popUpEl = unitTestUtil.getMenuElement();
      await popUpEl.find(`[aid=radio-option-${newSelection.name}]`).trigger("click");
      await popUpEl.find("[aid=apply-btn]").trigger("click");

      const expectedEmit = [[newSelection.uuid]];

      expect(pageClusterFilterWrapper.emitted("update-filter")).toEqual(expectedEmit);
    });
  });

  describe("options selection", () => {
    test("selected cluster appears first when 'All' option is hidden", () => {
      mountComponent({
        filterValue: mockClusters[1].uuid,
        hideAllOption: true,
        applyGlobally: false,
      });

      expect(pageClusterFilterWrapper.vm.options[0]).toEqual({
        label: mockClusters[1].name,
        value: mockClusters[1].uuid,
      });
    });

    test("selected cluster appears after 'All' option when 'All' option is visible", () => {
      mountComponent({ filterValue: mockClusters[1].uuid, hideAllOption: false, applyGlobally: false });
      expect(pageClusterFilterWrapper.vm.options[1]).toEqual({
        label: mockClusters[1].name,
        value: mockClusters[1].uuid,
      });
    });

    test("selected cluster is not moved if it is already in the correct position", () => {
      mountComponent({ filterValue: mockClusters[0].uuid, hideAllOption: false });
      expect(pageClusterFilterWrapper.vm.options[1]).toEqual({
        label: mockClusters[0].name,
        value: mockClusters[0].uuid,
      });
    });

    test("selected cluster is not moved if it is the first option and 'All' option is hidden", () => {
      mountComponent({ filterValue: mockClusters[0].uuid, hideAllOption: true });
      expect(pageClusterFilterWrapper.vm.options[0]).toEqual({
        label: mockClusters[0].name,
        value: mockClusters[0].uuid,
      });
    });

    test("renders 'All' option correctly when 'hideAllOption' is false", () => {
      mountComponent({ filterValue: "", hideAllOption: false });
      expect(pageClusterFilterWrapper.vm.options[0]).toEqual({
        label: ALL_CLUSTER_FILTER_LABEL,
        value: ALL_CLUSTER_FILTER_VALUE,
      });
    });

    test("does not render 'All' option when 'hideAllOption' is true", () => {
      mountComponent({ filterValue: "", hideAllOption: true, applyGlobally: false });
      expect(pageClusterFilterWrapper.vm.options).not.toContainEqual({
        label: ALL_CLUSTER_FILTER_LABEL,
        value: ALL_CLUSTER_FILTER_VALUE,
      });
    });

    test("should not update cluster store when 'applyGlobally' is false", async () => {
      mountComponent({ filterValue: "", hideAllOption: false, applyGlobally: false });
      const newSelection = mockClusters[2];

      await clickOnChip();
      const popUpEl = unitTestUtil.getMenuElement();
      await popUpEl.find(`[aid=radio-option-${newSelection.name}]`).trigger("click");
      await popUpEl.find("[aid=apply-btn]").trigger("click");

      expect(clusterStore.selectedCluster).toEqual(mockClusters[0]);
    });
  });
});
