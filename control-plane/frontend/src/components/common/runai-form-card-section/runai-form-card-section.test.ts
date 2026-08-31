/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";

import RunaiFormCardSection, { START_FROM_SCRATCH_ID } from "./runai-form-card-section.vue";

import type { ICardListItem, TCardCmpName } from "@/components/common/runai-card-list";
import { cardsList } from "./mock.data";
import { beforeEach, describe, expect, test, vi } from "vitest";

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
describe("runai-form-card-section.vue", () => {
  let runaiFormCardSectionCmp: VueWrapper<InstanceType<typeof RunaiFormCardSection>>;

  const fromScratchCard = {
    id: START_FROM_SCRATCH_ID,
    cardName: "create-new-card" as TCardCmpName,
    data: {
      meta: {
        id: START_FROM_SCRATCH_ID,
      },
    },
  };

  const defaultProps = {
    loading: false as boolean,
    emptyMessage: "Looks like you don't have any compute resources yet..." as string,
    mainMessage: "Select the resources needed to run your" as string,
    entityName: "compute resource" as string,
    cardsList: cardsList as Array<ICardListItem>,
    selectedCardsIds: [] as Array<string>,
    iconName: "compute-resource-gray" as string,
    searchName: "compute resources" as string,
    multiSelect: false,
    allowActions: true,
  };

  describe("empty state", () => {
    test("the component will render an empty state when there is no items in the list", async () => {
      runaiFormCardSectionCmp = mount(RunaiFormCardSection, {
        global: {
          plugins: [Quasar, pinia],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          ...defaultProps,
          cardsList: [],
        },
        provide: {
          policy: {},
          isGPUOverProvisioningEnabled: false,
        },
      });

      expect(runaiFormCardSectionCmp.find(".empty-box").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".empty-message").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".create-new-btn").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.html()).contains(defaultProps.emptyMessage);
    });

    test("the component will not render the empty state when there is loading", async () => {
      runaiFormCardSectionCmp = mount(RunaiFormCardSection, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          ...defaultProps,
          cardsList: [],
          loading: true,
        },
        provide: {
          policy: {},
          isGPUOverProvisioningEnabled: false,
        },
      });
      expect(runaiFormCardSectionCmp.find(".empty-box").exists()).toBe(false);
      expect(runaiFormCardSectionCmp.find(".empty-message").exists()).toBe(false);
      expect(runaiFormCardSectionCmp.find(".create-new-btn").exists()).toBe(false);
    });
  });

  describe("full list state", () => {
    beforeEach(() => {
      runaiFormCardSectionCmp = mount(RunaiFormCardSection, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          ...defaultProps,
        },
        provide: {
          policy: {},
          isGPUOverProvisioningEnabled: false,
        },
      });
    });

    test("the component will render the list component as cards when the list is not empty ", async () => {
      expect(runaiFormCardSectionCmp.find(".header").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".subheader").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".actions-container").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".search-input").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".add-new-btn").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.find(".cards-list-container").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.findAll(".cards-list-container li").length).toBe(6);
      expect(runaiFormCardSectionCmp.find(".runai-general-search").exists()).toBe(true);
      expect(runaiFormCardSectionCmp.html()).contains("new " + defaultProps.entityName);
      expect(runaiFormCardSectionCmp.html()).contains(defaultProps.mainMessage);
    });
  });

  describe("search state", () => {
    beforeEach(() => {
      runaiFormCardSectionCmp = mount(RunaiFormCardSection, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          ...defaultProps,
          loading: false,
        },
        provide: {
          policy: {},
          isGPUOverProvisioningEnabled: false,
        },
      });
    });

    test("the component will render the empty search state when there is no any match", async () => {
      await runaiFormCardSectionCmp.setData({ localSearchTerm: "some search" });
      expect(runaiFormCardSectionCmp.findAll(".cards-list-container li").length).toBe(0);
      expect(runaiFormCardSectionCmp.html()).contains("No compute resources match your filters");
    });

    test("the component will render the filtered list when there is at list one match", async () => {
      await runaiFormCardSectionCmp.setData({ localSearchTerm: "half" });
      expect(runaiFormCardSectionCmp.findAll(".cards-list-container li").length).toBe(1);
    });

    test("the list will include 'start from scarch' card FIRST if it exists in the cards' array and no searchTerm exists", async () => {
      await runaiFormCardSectionCmp.setData({ localSearchTerm: "" });
      await runaiFormCardSectionCmp.setProps({
        cardsList: [fromScratchCard, ...cardsList],
      });
      expect(runaiFormCardSectionCmp.vm.list[0].id).toBe("-1");
    });

    test("the list will NOT include 'start from scarch' card if it exists in the cards' array and searchTerm exists", async () => {
      await runaiFormCardSectionCmp.setData({ localSearchTerm: "something" });
      await runaiFormCardSectionCmp.setProps({
        cardsList: [fromScratchCard, ...cardsList],
      });

      expect(runaiFormCardSectionCmp.vm.list.find((c) => c.id === "-1")).toBe(undefined);
    });
  });

  describe("server-side pagination", () => {
    beforeEach(() => {
      runaiFormCardSectionCmp = mount(RunaiFormCardSection, {
        global: {
          plugins: [Quasar],
          directives: {
            clickOutside: vi.fn(),
          },
        },
        props: {
          ...defaultProps,
          loading: false,
          listQuery: { search: "test", sortBy: null },
          totalCount: 10,
          allowActions: true,
        },
        provide: {
          policy: {},
          isGPUOverProvisioningEnabled: false,
        },
      });
    });

    test("should emit query-changed event when search term changes", async () => {
      await runaiFormCardSectionCmp
        .findComponent({ name: "runai-general-search" })
        .vm.$emit("update-search-term", "test");

      const emitResult = runaiFormCardSectionCmp.emitted("query-changed");
      expect(emitResult).toBeTruthy();
      expect(emitResult?.[0][0]).toEqual({
        sortBy: null,
        search: "test",
      });
    });

    test("should emit get-next-batch event when pagination reaches end of loaded data", async () => {
      await runaiFormCardSectionCmp.findComponent({ name: "runai-card-list" }).vm.$emit("get-next-batch", 12);

      const emitResult = runaiFormCardSectionCmp.emitted("get-next-batch");
      expect(emitResult).toBeTruthy();
      expect(emitResult?.[0][0]).toBe(12);
    });

    test("should use listQuery 'search' instead of local search term when provided", async () => {
      await runaiFormCardSectionCmp.setProps({
        listQuery: { search: "test", sortBy: null },
      });

      expect(runaiFormCardSectionCmp.vm.relevantSearchTerm).toBe("test");
    });

    test("should use listQuery sortBy instead of local sort option when provided", async () => {
      await runaiFormCardSectionCmp.setProps({
        listQuery: { sortBy: "name", search: null },
      });

      expect(runaiFormCardSectionCmp.vm.relevantSortOption?.value).toBe("name");
    });
  });
});
