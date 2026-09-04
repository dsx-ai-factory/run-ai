/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import ItemIconWithCounter from "./item-icon-with-counter.vue";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { beforeEach, describe, expect, it } from "vitest";
import { Quasar } from "quasar";

describe("item-icon-with-counter.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof ItemIconWithCounter>>;

  const defaultTestIcons = [{ icon: "custom-tool", name: "test" }];
  const factory = (props = {}) => {
    return mount(ItemIconWithCounter, {
      global: { plugins: [Quasar] },
      props: { icons: defaultTestIcons, ...props },
    });
  };

  describe("item-icon-with-counter.vue", () => {
    beforeEach(() => {
      wrapper = factory();
    });

    it("renders the first icon correctly", () => {
      const svgIcon = wrapper.findComponent(RunaiSvgIcon);
      expect(svgIcon.exists()).toBe(true);
      expect(svgIcon.props("name")).toBe("custom-tool");
    });

    it("renders Phosphor icons through Quasar instead of treating them as SVG filenames", () => {
      wrapper = factory({ icons: [{ icon: "ph-bold ph-hard-drives", name: "compute" }] });

      expect(wrapper.findComponent(RunaiSvgIcon).exists()).toBe(false);
      expect(wrapper.findComponent({ name: "QIcon" }).props("name")).toBe("ph-bold ph-hard-drives");
    });

    it("does not show additionals counter when there is only one icon", () => {
      const additionalsCounter = wrapper.find(".additionals-counter");
      expect(additionalsCounter.exists()).toBe(false);
    });
  });

  it("shows additionals counter when there are multiple icons", () => {
    const icons = [
      { icon: "custom-tool", name: "test1" },
      { icon: "custom-tool", name: "test2" },
    ];
    wrapper = factory({ icons });
    const additionalsCounter = wrapper.find(".additionals-counter");
    expect(additionalsCounter.exists()).toBe(true);
    expect(additionalsCounter.text()).toBe("+1");
  });

  it("correctly calculates additionals counter", () => {
    const icons = [
      { icon: "custom-tool", name: "test1" },
      { icon: "custom-tool", name: "test2" },
      { icon: "custom-tool", name: "test3" },
    ];
    wrapper = factory({ icons });
    const additionalsCounter = wrapper.find(".additionals-counter");
    expect(additionalsCounter.exists()).toBe(true);
    expect(additionalsCounter.text()).toBe("+2");
  });
});
