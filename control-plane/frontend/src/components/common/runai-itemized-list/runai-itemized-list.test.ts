/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { shallowMount, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import RunaiItemizedList from "./runai-itemized-list.vue";
import type { TTestEmit } from "@/models/unit-test.model";
import type { IItemizedListItem } from "./runai-itemized-list.model";

describe("runai-itemized-list.vue", () => {
  let runaiItemizedListCmp: VueWrapper<InstanceType<typeof RunaiItemizedList>>;

  describe("initial tests", () => {
    beforeEach(() => {
      runaiItemizedListCmp = shallowMount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          title: "Set the items with variables",
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });
    });

    it("should have a title", () => {
      expect(runaiItemizedListCmp.html()).toContain("Set the items with variables");
    });

    it("should not render the input boxes when there are no items", () => {
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(0);
    });
  });

  describe("add new item btn", () => {
    it("should have an enable add new item btn", () => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          title: "Set the items with variables",
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });

      const addItemButton = runaiItemizedListCmp.find(".add-item-button");
      expect(addItemButton.exists()).toBe(true);
      expect(addItemButton.attributes("disabled")).toBeUndefined();
    });

    it("should have a disable add new item btn", () => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          policyRules: { instances: { canAdd: false } },
          title: "Set the items with variables",
          addItemBtn: "ENVIRONMENT VARIABLE",
          itemsList: [{ name: "var1", value: "val1" }],
        },
      });

      const addItemButton = runaiItemizedListCmp.find(".add-item-button");
      expect(addItemButton.exists()).toBe(true);
      expect(addItemButton.attributes("disabled")).toBeDefined();
    });

    it("should add a new empty row to the list", async () => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          title: "Set the items with variables",
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });

      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(0);
      const addItemButton = runaiItemizedListCmp.find(".add-item-button");
      await addItemButton.trigger("click");
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(1);
    });
  });

  describe("list container", () => {
    const itemsList = [
      { name: "var1", value: "" },
      { name: "var2", value: "" },
      { name: "var3", value: "", exclude: true },
      { name: "var4", value: "" },
      { name: "var5", value: "" },
      { name: "var6", value: "" },
    ];

    beforeEach(() => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          policyRules: { instances: { canAdd: false, locked: ["var5", "var6"] } },
          title: "Set the items with variables",
          itemsList: itemsList,
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });
    });

    it("should render the locked items at the top of the list", () => {
      const allNameItems = runaiItemizedListCmp.findAll("[aid=itemized-name-input]");
      expect(allNameItems[0].attributes("disabled")).toBeDefined();
      expect(allNameItems[1].attributes("disabled")).toBeDefined();
      expect(allNameItems[2].attributes("disabled")).not.toBeDefined();

      const allValueItems = runaiItemizedListCmp.findAll("[aid=itemized-value-input]");
      expect(allValueItems[0].attributes("disabled")).toBeDefined();
      expect(allValueItems[1].attributes("disabled")).toBeDefined();
      expect(allValueItems[2].attributes("disabled")).not.toBeDefined();
    });

    it("should not render the excluded items", () => {
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(itemsList.length - 1);
    });

    it("should add new item to the list", async () => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          title: "Set the items with variables",
          itemsList: [{ name: "var1", value: "" }],
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });

      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(1);
      const addItemButton = runaiItemizedListCmp.find(".add-item-button");
      await addItemButton.trigger("click");
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(2);
    });
  });

  describe("emit chenged events", () => {
    const itemsList = [
      { name: "var1", value: "" },
      { name: "var3", value: "", exclude: true },
    ];

    beforeEach(() => {
      runaiItemizedListCmp = mount(RunaiItemizedList, {
        global: {
          plugins: [Quasar],
        },
        props: {
          title: "Set the items with variables",
          itemsList: itemsList,
          addItemBtn: "ENVIRONMENT VARIABLE",
        },
      });
    });

    it("should emit 'changed' event with the full updated array", async () => {
      const newName = "new-name1";
      await runaiItemizedListCmp.find("[aid=itemized-name-input]").setValue(newName);
      const changedEvent: Array<Array<Array<IItemizedListItem>>> | undefined = runaiItemizedListCmp.emitted("changed");
      const eventData: Array<IItemizedListItem> | undefined = changedEvent ? changedEvent[0][0] : changedEvent;
      const updatedName = eventData?.[0]?.name;
      expect(updatedName).toBe(newName);
    });

    it("should emit 'changed' event with the updated array when remove button clicked", async () => {
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(1);
      await runaiItemizedListCmp.find("[aid=remove-itemized-button]").trigger("click");
      const changedEvent: TTestEmit<Array<IItemizedListItem>> | undefined = runaiItemizedListCmp.emitted("changed");
      const eventData: Array<IItemizedListItem> | undefined = changedEvent ? changedEvent[0][0] : changedEvent;
      expect(runaiItemizedListCmp.find(".list-container").element.children.length).toBe(0);
      expect(eventData?.length).toBe(itemsList.length - 1);
    });

    it("should emit 'changed' event with the updated array when add new button clicked", async () => {
      await runaiItemizedListCmp.find("[aid=add-item-button]").trigger("click");
      const changedEvent: TTestEmit<Array<IItemizedListItem>> | undefined = runaiItemizedListCmp.emitted("changed");
      const eventData: Array<IItemizedListItem> | undefined = changedEvent ? changedEvent[0][0] : changedEvent;
      expect(eventData?.length).toBe(itemsList.length + 1);
    });
  });
});
