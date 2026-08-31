/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import RunaiCardList from "./runai-card-list.vue";
import { mount } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { computeResourcesMockedData } from "./runai-card-list.mock";
import type { TTestEmit } from "@/models/unit-test.model";
import type { ICardListItem } from "./runai-card-list.model";
import type { ComputeAsset } from "@/swagger-models/assets-service-client";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("runai-card-list.vue", () => {
  let runaiCardListCmp: VueWrapper<InstanceType<typeof RunaiCardList>>;

  const props: Partial<typeof RunaiCardList.props> = {
    list: computeResourcesMockedData.list as Array<ICardListItem>,
    loading: true as boolean,
    multiSelect: false as boolean,
    selectedItems: [] as Array<string>,
  };

  const factory = (props: Partial<typeof RunaiCardList.props>) => {
    runaiCardListCmp = mount(RunaiCardList, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        selectedItems: [],
        list: [],
        ...props,
      },
      provide: {
        policy: {},
        isGPUOverProvisioningEnabled: false,
      },
    });
  };

  beforeEach(() => {
    factory(props);
  });

  test("component is loading properly", async () => {
    expect(runaiCardListCmp.classes()).toContain("runai-card-list");
  });

  test("the component will render the list of skeleton cards when data is loading", async () => {
    // note that in the props before each test loading is set to true by default
    expect(runaiCardListCmp.find("[aid=skeleton-list]").exists()).toBe(true);
    expect(runaiCardListCmp.find("[aid=list-content]").exists()).toBe(false);
  });

  test("the component will hide the list of skeleton cards and render the real cards list loading is false", async () => {
    await runaiCardListCmp.setProps({ loading: false });
    expect(runaiCardListCmp.find("[aid=skeleton-list]").exists()).toBe(false);
    expect(runaiCardListCmp.find("[aid=list-content]").exists()).toBe(true);
  });

  test("the component will hide empty-data-section when the list array is not empty", async () => {
    await runaiCardListCmp.setProps({ loading: false });
    expect(runaiCardListCmp.find(".empty-data-section").exists()).toBe(false);
  });

  test("the component will emit the updated selected entities correctly when a new selection is made AND multiSelect prop is set to false", async () => {
    const firstCard = computeResourcesMockedData.list[0];
    const secondCard = computeResourcesMockedData.list[1];
    await runaiCardListCmp.setProps({ loading: false, multiSelect: false, selectedItems: [firstCard.id] });
    await runaiCardListCmp.find(`[aid=${(secondCard.data as ComputeAsset).meta.name}-card]`).trigger("click");
    const expectedEmit: Array<string> = [secondCard.id];
    const emitResult: TTestEmit<string[]> = runaiCardListCmp.emitted("changed") || [];
    expect(emitResult[0][0]).toEqual(expectedEmit);
  });

  test("the component will emit the updated selected entities correctly when a new selection is made AND multiSelect prop is set to true", async () => {
    const firstCard = computeResourcesMockedData.list[0];
    const secondCard = computeResourcesMockedData.list[1];
    await runaiCardListCmp.setProps({ loading: false, multiSelect: true, selectedItems: [firstCard.id] });
    expect(await runaiCardListCmp.find(".selected-container").exists()).toBe(true);

    await runaiCardListCmp.find(`[aid=${(secondCard.data as ComputeAsset).meta.name}-card]`).trigger("click");
    const expectedEmit: Array<string> = [firstCard.id, secondCard.id];
    const emitResult: TTestEmit<string[]> = runaiCardListCmp.emitted("changed") || [];
    expect(emitResult[0][0]).toEqual(expectedEmit);
  });

  test("the component will render the correct amount of selected entities when multiSelect prop is set to true", async () => {
    const firstCard = computeResourcesMockedData.list[0];
    const secondCard = computeResourcesMockedData.list[1];
    await runaiCardListCmp.setProps({
      loading: false,
      multiSelect: true,
      selectedItems: [firstCard.id, secondCard.id],
    });

    expect(runaiCardListCmp.find(".selected-container").text()).toContain(`Selected ${2} / ${props.list.length}`);
  });

  test("the component will not render the pagination buttons when the list prop array length is lesser than 6", async () => {
    // note: default list size in the props of each test is 5
    await runaiCardListCmp.setProps({ loading: false });
    expect(runaiCardListCmp.find("[aid=pagination-btns-container]").exists()).toBe(false);
  });

  test("the component will render the pagination buttons when the list prop array length is larger than 6", async () => {
    await runaiCardListCmp.setProps({ loading: false, list: computeResourcesMockedData.largeList });
    expect(runaiCardListCmp.find("[aid=pagination-btns-container]").exists()).toBe(true);
  });

  test("clicking pagination button will set the 'slide' variable in component data to the correct index", async () => {
    await runaiCardListCmp.setProps({ loading: false, list: computeResourcesMockedData.largeList });
    const idx = 1;
    await runaiCardListCmp.findAll("[aid=pagination-btns-container] button")[idx].trigger("click");
    expect(runaiCardListCmp.vm.slide).toBe(idx);
  });

  describe("server-side pagination", () => {
    beforeEach(() => {
      factory({
        loading: false,
        list: computeResourcesMockedData.list,
        totalCount: 10,
        listQuery: { page: 1, filterBy: "", sortBy: "", search: "" },
      });
    });

    test("should show pagination when totalCount is greater than 6", async () => {
      await runaiCardListCmp.setProps({
        loading: false,
        list: computeResourcesMockedData.list,
        totalCount: 10,
        listQuery: { page: 1, filterBy: "", sortBy: "", search: "" },
      });
      expect(runaiCardListCmp.find("[aid=pagination-btns-container]").exists()).toBe(true);
    });

    test("should emit get-next-batch event when reaching last loaded page", async () => {
      await runaiCardListCmp.setProps({
        loading: false,
        list: computeResourcesMockedData.list,
        totalCount: 10,
        listQuery: { page: 1, filterBy: "", sortBy: "", search: "" },
      });

      // Click on page 2
      await runaiCardListCmp.findAll("[aid=pagination-btns-container] button")[1].trigger("click");

      const emitResult = runaiCardListCmp.emitted("get-next-batch");
      expect(emitResult).toBeTruthy();
      expect(emitResult?.[0][0]).toBe(12); // Should request 12 items (2 pages worth)
    });

    test("should reset pagination when listQuery changes", async () => {
      await runaiCardListCmp.setProps({
        loading: false,
        list: computeResourcesMockedData.list,
        totalCount: 10,
        listQuery: { page: 1, filterBy: "", sortBy: "", search: "" },
      });

      // Change to page 2
      await runaiCardListCmp.findAll("[aid=pagination-btns-container] button")[1].trigger("click");
      expect(runaiCardListCmp.vm.page).toBe(2);

      // Change listQuery
      await runaiCardListCmp.setProps({
        listQuery: { page: 1, filterBy: "test", sortBy: "", search: "" },
      });

      // Should reset to page 1
      expect(runaiCardListCmp.vm.page).toBe(1);
      expect(runaiCardListCmp.vm.slide).toBe(0);
    });

    test("should calculate max pages based on totalCount when provided", async () => {
      await runaiCardListCmp.setProps({
        loading: false,
        list: computeResourcesMockedData.list,
        totalCount: 20,
        listQuery: { page: 1, filterBy: "", sortBy: "", search: "" },
      });

      expect(runaiCardListCmp.vm.maxAmountOfTotalPages).toBe(4); // 20 items / 6 per page = 4 pages
    });
  });
});
