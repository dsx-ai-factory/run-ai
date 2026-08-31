/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import RunaiSubExpansionTextItem from "./runai-sub-expansion-text-item.vue";

interface RunaiSubExpansionTextItemProps {
  label?: string;
  defaultOpened?: boolean;
  aid?: string;
}

interface RunaiSubExpansionTextItemSlots {
  header?: string;
  body?: string;
}

describe("runai-sub-expansion-text-item.vue", () => {
  function factory(
    props: RunaiSubExpansionTextItemProps,
    slots?: RunaiSubExpansionTextItemSlots,
  ): VueWrapper<InstanceType<typeof RunaiSubExpansionTextItem>> {
    return mount(RunaiSubExpansionTextItem, {
      global: {
        plugins: [Quasar],
      },
      props,
      slots: {
        ...slots,
      },
    });
  }

  it("renders the label in the header slot", () => {
    const label = "Test Label";
    const header = "<div>Custom Header Slot</div>";
    const wrapper = factory({ label }, { header });

    expect(wrapper.html()).toContain("Custom Header Slot");
  });

  it("renders default slot content in the body slot", () => {
    const label = "Test Label";
    const body = "<div>Custom Body Slot</div>";
    const wrapper = factory({ label }, { body });

    expect(wrapper.html()).toContain("Custom Body Slot");
  });

  it("toggles expansion state when header is clicked", async () => {
    const label = "Test Label";
    const body = "<div>Custom Body Slot</div>";
    const wrapper = factory({ label, defaultOpened: false }, { body });

    const header = wrapper.find(".runai-sub-expansion-header");

    expect(wrapper.vm.isOpen).toBe(false);
    expect(wrapper.html()).toContain("Custom Body Slot");

    await header.trigger("click");
    expect(wrapper.vm.isOpen).toBe(true);

    await header.trigger("click");
    expect(wrapper.vm.isOpen).toBe(false);
  });

  it("respects the defaultOpened prop", () => {
    const label = "Test Label";
    const wrapper = factory({ label, defaultOpened: true });
    expect(wrapper.vm.isOpen).toBe(true);
  });

  it("passes the aid prop to the header div", () => {
    const aid = "test-aid";
    const label = "Test Label";
    const wrapper = factory({ label, aid });

    const header = wrapper.find(".runai-sub-expansion-header");
    expect(header.attributes("aid")).toBe(aid);
  });
});
