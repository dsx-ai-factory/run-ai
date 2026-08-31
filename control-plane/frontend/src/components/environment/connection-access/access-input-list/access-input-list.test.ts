/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import AccessInputList from "./access-input-list.vue";

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { TTestEmit } from "@/models/unit-test.model";

import type { IAccessInput } from "./access-input-list.model";
import { describe, beforeEach, afterEach, vi, it, expect } from "vitest";

describe("access-input-list.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof AccessInputList>>;

  const defaultProps = {
    items: [],
    inputLabel: "New Item",
    addItemBtnLabel: "Add Item",
    validationRules: [],
  };

  beforeEach(() => {
    wrapper = mount(AccessInputList, {
      global: {
        plugins: [Quasar],
      },
      props: { ...defaultProps },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("initializes with one empty item if items prop is empty", () => {
    expect(wrapper.vm.itemsData).toHaveLength(1);
    expect(wrapper.vm.itemsData[0].text).toBe("");
    expect(wrapper.vm.itemsData[0].label).toBe("New Item");
    expect(wrapper.vm.itemsData[0].removable).toBe(false);
  });

  it("adds a new item when addNewItem is called", async () => {
    // the first item should not be removable
    expect(wrapper.vm.itemsData[0].removable).toBe(false);
    await wrapper.vm.addNewItem();
    expect(wrapper.vm.itemsData).toHaveLength(2);
    expect(wrapper.vm.itemsData[1].text).toBe("");
    expect(wrapper.vm.itemsData[1].label).toBe("New Item");
    expect(wrapper.vm.itemsData[1].removable).toBe(true);

    // after adding a new item, the first item should be removable
    expect(wrapper.vm.itemsData[0].removable).toBe(true);
  });

  it("removes an item when removeItem is called", async () => {
    await wrapper.vm.addNewItem();
    expect(wrapper.vm.itemsData).toHaveLength(2);
    expect(wrapper.vm.itemsData[0].removable).toBe(true);
    expect(wrapper.vm.itemsData[1].removable).toBe(true);
    wrapper.vm.removeItem(wrapper.vm.itemsData[0].id);
    await wrapper.vm.$nextTick();
    expect(wrapper.vm.itemsData).toHaveLength(1);
    expect(wrapper.vm.itemsData[0].removable).toBe(false);
  });

  it("emits changed event when item is added", async () => {
    await wrapper.vm.addNewItem();
    await wrapper.vm.$nextTick();
    const emitResult: TTestEmit<IAccessInput[]> = wrapper.emitted("changed") || [];
    expect(emitResult[0][0].length).toBe(2);
    expect(emitResult[0][0]).toEqual([
      { text: "", label: "New Item" },
      { text: "", label: "New Item" },
    ]);
  });

  it("emits changed event when item is removed", async () => {
    await wrapper.vm.addNewItem();
    let emitResult: TTestEmit<IAccessInput[]> = wrapper.emitted("changed") || [];
    expect(emitResult[0][0].length).toBe(2);
    wrapper.vm.removeItem(wrapper.vm.itemsData[0].id);
    emitResult = wrapper.emitted("changed") || [];
    expect(emitResult[1][0].length).toBe(1);
    expect(emitResult[1][0]).toEqual([{ text: "", label: "New Item" }]);
  });

  it("focuses on the last input when a new item is added", async () => {
    const focusSpy = vi.spyOn(wrapper.vm, "focusInput");
    await wrapper.vm.addNewItem();
    expect(focusSpy).toHaveBeenCalled();
  });

  it("updates item text correctly when input value is changed", async () => {
    const input = wrapper.findComponent({ ref: "inputRefs" });
    await (input as VueWrapper).vm.$emit("update:model-value", "Test Value");
    expect(wrapper.vm.itemsData[0].text).toBe("Test Value");
  });

  it("calls initList when inputLabel prop is changed", async () => {
    const focusSpy = vi.spyOn(wrapper.vm, "initList");
    // the first time we call to the initList is when the component is mounted
    // so we still didn't spied it.
    expect(focusSpy).toHaveBeenCalledTimes(0);
    wrapper.setProps({ inputLabel: "New Item Label" });
    await wrapper.vm.$nextTick();
    expect(focusSpy).toHaveBeenCalledTimes(1);
  });

  describe("non-removable items", () => {
    const defaultProps = {
      items: [
        { text: "Item 1", label: "User", removable: false },
        { text: "Item 2", label: "User", removable: true },
        { text: "Item 3", label: "User", removable: false },
      ],
      inputLabel: "New Item",
      addItemBtnLabel: "Add Item",
      validationRules: [],
    };

    beforeEach(() => {
      wrapper = mount(AccessInputList, {
        global: {
          plugins: [Quasar],
        },
        props: { ...defaultProps },
      });
    });

    it("does not show remove button for non-removable items", async () => {
      const removeButton = wrapper.findAll('[aid="remove-item-button"]');
      expect(removeButton.length).toBe(1);
    });
  });

  describe("policyRules: disable and canEdit", () => {
    const defaultProps = {
      items: [{ text: "Item 1", label: "User", removable: false }],
      inputLabel: "New Item",
      addItemBtnLabel: "Add Item",
      validationRules: [],
    };

    it("disables add button and shows tooltip when canEdit is false", async () => {
      const policyRules = { canEdit: false };
      const wrapper = mount(AccessInputList, {
        global: { plugins: [Quasar] },
        props: { ...defaultProps, policyRules },
      });
      await wrapper.vm.$nextTick();
      const addBtn = wrapper.find('[aid="add-item-button"]');
      expect(addBtn.attributes("disabled")).toBeDefined();
      expect(addBtn.attributes("aria-disabled")).toBe("true");

      // Tooltip should exist
      const tooltip = wrapper.find(".add-item-button-container").findComponent({ name: "QTooltip" });
      expect(tooltip.exists()).toBe(true);
      expect(tooltip.vm.$slots.default()[0].children).toBe(wrapper.vm.cantBeModifiedPolicy);
    });

    it("enables add button when canEdit is true or undefined", async () => {
      const wrapperTrue = mount(AccessInputList, {
        global: { plugins: [Quasar] },
        props: { ...defaultProps, policyRules: { canEdit: true } },
      });
      await wrapperTrue.vm.$nextTick();
      const addBtnTrue = wrapperTrue.find('[aid="add-item-button"]');
      expect(addBtnTrue.attributes("disabled")).toBeUndefined();

      const wrapperUndef = mount(AccessInputList, {
        global: { plugins: [Quasar] },
        props: { ...defaultProps },
      });
      const addBtnUndef = wrapperUndef.find('[aid="add-item-button"]');
      expect(addBtnUndef.attributes("disabled")).toBeUndefined();
    });

    it("passes policyRules prop to policy-string-field", async () => {
      const policyRules = { canEdit: false };
      const wrapper = mount(AccessInputList, {
        global: { plugins: [Quasar] },
        props: { ...defaultProps, policyRules },
      });

      await wrapper.vm.$nextTick();
      const field = wrapper.findComponent({ name: "policy-string-field" });
      expect(field.props("policyRules")).toEqual(policyRules);
    });
  });
});
