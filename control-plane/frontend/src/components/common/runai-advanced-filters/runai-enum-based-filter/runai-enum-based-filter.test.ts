/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, test } from "vitest";
import type { IEnumBasedFilterModel } from "@/models/filter.model";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { RunaiEnumBasedFilter } from "@/components/common/runai-advanced-filters/runai-enum-based-filter/index";
import { unitTestUtil } from "@/utils/unit-test.util";
import { EColumnFilterType } from "@/models/filter.model";

describe("RunaiFreeTextFilter", () => {
  let wrapper: VueWrapper<InstanceType<typeof RunaiEnumBasedFilter>>;
  const mountComponent = async (filter: IEnumBasedFilterModel): Promise<void> => {
    wrapper = mount(RunaiEnumBasedFilter, {
      global: {
        plugins: [Quasar],
      },
      props: {
        filter,
      },
    });
    await wrapper.vm.$nextTick();
  };

  afterEach(() => {
    wrapper.unmount();
  });

  it("should renders the options correctly", async () => {
    // Example of props for testing
    const filter: IEnumBasedFilterModel = {
      selectOptions: [
        { label: "Creating", value: "Creating" },
        { label: "Initializing", value: "Initializing" },
        { label: "Resuming", value: "Resuming" },
        { label: "Pending", value: "Pending" },
        { label: "Deleting", value: "Deleting" },
        { label: "Running", value: "Running" },
      ],
      selectedValues: ["Deleting", "Pending"],
      label: "Status",
      name: "phase",
      type: EColumnFilterType.EnumBased,
    };

    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();

    expect(wrapper.exists()).toBe(true);
    expect(menuElement.find(".checkbox-container").exists()).toBe(true);
    const allCheckboxes = menuElement.findAll(".q-checkbox");
    expect(allCheckboxes.length).toBe(filter.selectOptions.length);
    filter.selectOptions.forEach((option, index) => {
      expect(allCheckboxes[index].text()).toBe(option.label);
    });
  });

  it("should emits save event with updated filter on save", async () => {
    const filter: IEnumBasedFilterModel = {
      selectOptions: [
        { label: "Creating", value: "Creating" },
        { label: "Initializing", value: "Initializing" },
        { label: "Resuming", value: "Resuming" },
        { label: "Pending", value: "Pending" },
        { label: "Deleting", value: "Deleting" },
        { label: "Running", value: "Running" },
      ],
      selectedValues: ["Deleting", "Pending"],
      label: "Status",
      name: "phase",
      type: EColumnFilterType.EnumBased,
    };

    await mountComponent(filter);

    await wrapper.setData({ selectedValues: ["Resuming"] });

    wrapper.vm.saveFilter();

    expect(wrapper.emitted("save")).toBeTruthy();

    expect((wrapper.emitted("save")?.[0]?.[0] as Record<string, unknown>).selectedValues).toEqual(["Resuming"]);
  });

  test("should emits 'close'", async () => {
    const filter: IEnumBasedFilterModel = {
      selectOptions: [
        { label: "Creating", value: "Creating" },
        { label: "Initializing", value: "Initializing" },
        { label: "Resuming", value: "Resuming" },
        { label: "Pending", value: "Pending" },
        { label: "Deleting", value: "Deleting" },
        { label: "Running", value: "Running" },
      ],
      selectedValues: ["Deleting", "Pending"],
      label: "Status",
      name: "phase",
      type: EColumnFilterType.EnumBased,
    };
    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();
    const closeBtn = menuElement.find("[aid=close-filter-btn]");
    await closeBtn.trigger("click");

    expect(wrapper.emitted("close")).toBeTruthy();
  });

  it("should render the options correctly for radio buttons", async () => {
    const filter: IEnumBasedFilterModel = {
      selectOptions: [
        { label: "Creating", value: true },
        { label: "Initializing", value: false },
      ],
      selectedValues: [true],
      label: "Status",
      name: "phase",
      type: EColumnFilterType.EnumBased,
      inputType: "radio",
    };

    await mountComponent(filter);
    const menuElement = unitTestUtil.getMenuElement();

    expect(wrapper.exists()).toBe(true);
    expect(menuElement.find(".checkbox-container").exists()).toBe(true);
    const allRadios = menuElement.findAll(".q-radio");
    expect(allRadios.length).toBe(filter.selectOptions.length);
    filter.selectOptions.forEach((option, index) => {
      expect(allRadios[index].text()).toBe(option.label);
    });
  });

  it("should emit save event with updated filter on save for radio buttons", async () => {
    const filter: IEnumBasedFilterModel = {
      selectOptions: [
        { label: "Creating", value: true },
        { label: "Initializing", value: false },
      ],
      selectedValues: [true],
      label: "Status",
      name: "phase",
      type: EColumnFilterType.EnumBased,
      inputType: "radio",
    };

    await mountComponent(filter);

    await wrapper.setData({ selectedOption: false });

    wrapper.vm.saveFilter();

    expect(wrapper.emitted("save")).toBeTruthy();

    expect((wrapper.emitted("save")?.[0]?.[0] as Record<string, unknown>).selectedValues).toEqual([false]);
  });
});
