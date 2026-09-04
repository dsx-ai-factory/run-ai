/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { Quasar } from "quasar";
import MnnvlSection, { mnnvlOptions } from "./mnnvl-section.vue";
import { MnnvlMode } from "@/swagger-models/workloads-manager-client";

const mountComponent = (props = {}) => {
  return shallowMount(MnnvlSection, {
    global: {
      plugins: [Quasar],
      stubs: {
        RunaiSection: { template: "<div><slot /></div>" },
        RunaiTooltip: true,
        QOptionGroup: { template: "<div />" },
      },
    },
    props: {
      modelValue: null,
      ...props,
    },
  });
};

describe("MnnvlSection", () => {
  describe("mnnvlOptions", () => {
    it("has exactly two options: None and Required", () => {
      expect(mnnvlOptions).toHaveLength(2);
      expect(mnnvlOptions[0].value).toBe(MnnvlMode.None);
      expect(mnnvlOptions[1].value).toBe(MnnvlMode.Required);
    });

    it("labels options correctly", () => {
      expect(mnnvlOptions[0].label).toBe("Not required");
      expect(mnnvlOptions[1].label).toBe("Required");
    });

    it("each option has a tooltip", () => {
      expect(mnnvlOptions[0].toolTip).toBeTruthy();
      expect(mnnvlOptions[1].toolTip).toBeTruthy();
    });
  });

  describe("rendering", () => {
    it("renders without errors", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.None });
      expect(wrapper.exists()).toBe(true);
    });

    it("renders info link pointing to GB200 docs", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.None });
      const link = wrapper.find("a");
      expect(link.exists()).toBe(true);
      expect(link.attributes("target")).toBe("_blank");
    });

    it("info link text references GB200 NVL72", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.None });
      expect(wrapper.find("a").text()).toContain("GB200 NVL72");
    });
  });

  describe("updateMnnvl", () => {
    it("emits update:model-value with the selected value", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.None });
      wrapper.vm.updateMnnvl(MnnvlMode.Required);
      expect(wrapper.emitted("update:model-value")).toEqual([[MnnvlMode.Required]]);
    });

    it("emits MnnvlMode.None when switching back to Not required", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.Required });
      wrapper.vm.updateMnnvl(MnnvlMode.None);
      expect(wrapper.emitted("update:model-value")).toEqual([[MnnvlMode.None]]);
    });
  });

  describe("props", () => {
    it("accepts modelValue as null", () => {
      const wrapper = mountComponent({ modelValue: null });
      expect(wrapper.props("modelValue")).toBeNull();
    });

    it("accepts modelValue as Required", () => {
      const wrapper = mountComponent({ modelValue: MnnvlMode.Required });
      expect(wrapper.props("modelValue")).toBe(MnnvlMode.Required);
    });

    it("disabled defaults to false", () => {
      const wrapper = mountComponent({ modelValue: null });
      expect(wrapper.props("disabled")).toBe(false);
    });

    it("accepts disabled prop as true", () => {
      const wrapper = mountComponent({ modelValue: null, disabled: true });
      expect(wrapper.props("disabled")).toBe(true);
    });
  });
});
