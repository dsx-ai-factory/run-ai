/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import RunaiDropdown from "./runai-dropdown.vue";
import { Quasar } from "quasar";

interface Option {
  id: number;
  name: string;
}

interface IRunaiDropDownProps {
  options: Option[];
  rowKey: string;
  autoClose: boolean;
  isDisabled: boolean;
  disabledTooltip: string;
}

interface SlotProps {
  name: string;
}

interface Slots {
  btn?: string; // or a VNode if you're using Vue 3
  option?: (props: SlotProps) => string; // or a function returning VNode if you're using Vue 3
}

const getMountedComponent = (
  props: IRunaiDropDownProps,
  slots: Slots = {},
): VueWrapper<InstanceType<typeof RunaiDropdown>> => {
  return mount(RunaiDropdown, {
    props,
    slots: {
      btn: "<span>Dropdown Button</span>",
      option: ({ name }) => `<span>${name}</span>`,
      ...slots,
    },
    attachTo: document.body, // Attach the component to the body to handle teleport
    global: {
      plugins: [Quasar],
      stubs: {
        teleport: true,
        qMenu: {
          template: '<div class="q-menu"><slot></slot></div>', // Simple stub with essential structure
        },
        qTooltip: { template: '<div class="q-tooltip"><slot></slot></div>' }, // Simple stub with essential structure
      },
    },
  });
};

describe("RunaiDropdown", () => {
  const options: Option[] = [
    { id: 1, name: "Option 1" },
    { id: 2, name: "Option 2" },
  ];

  it("renders options correctly", async () => {
    const wrapper = getMountedComponent({
      options,
      rowKey: "id",
      autoClose: true,
      isDisabled: false,
      disabledTooltip: "Disabled",
    });

    expect(wrapper.find(".runai-dropdown").exists()).toBe(true);
    await wrapper.find(".runai-dropdown").trigger("click");
    const optionElements = wrapper.findAll(".runai-dropdown-option");
    expect(optionElements.length).toBe(options.length);
    expect(optionElements[0].text()).toBe("<span>Option 1</span>");
    expect(optionElements[1].text()).toBe("<span>Option 2</span>");
  });

  it("emits on-selected event when option is clicked", async () => {
    const wrapper = getMountedComponent({
      options,
      rowKey: "id",
      autoClose: true,
      isDisabled: false,
      disabledTooltip: "Disabled",
    });

    await wrapper.find(".runai-dropdown").trigger("click");
    await wrapper.findAll(".runai-dropdown-option")[0].trigger("click");
    expect(wrapper.emitted("on-selected")).toBeTruthy();
    expect(wrapper.emitted("on-selected")?.[0]).toEqual([{ id: 1, name: "Option 1" }]);
  });

  it("displays tooltip when disabled", async () => {
    const wrapper = getMountedComponent({
      options,
      rowKey: "id",
      autoClose: true,
      isDisabled: true,
      disabledTooltip: "TooltipText",
    });
    expect(wrapper.find(".q-tooltip").exists()).toBe(true);
    expect(wrapper.find(".q-tooltip").text()).toBe(wrapper.vm.$props.disabledTooltip);
  });
});
