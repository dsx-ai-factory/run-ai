/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, it, expect } from "vitest";
import { HybridSectionWrapper } from ".";
import { Quasar } from "quasar";

describe("hybrid-section-wrapper.vue", () => {
  const factory = (props = {}, slots = {}) => {
    return mount(HybridSectionWrapper, {
      global: {
        plugins: [Quasar],
      },
      props: {
        selectedName: "selected test name",
        label: "Test Name",
        subheader: "Test Subheader",
        ...props,
      },
      slots,
    });
  };

  it("renders subheader slot content", () => {
    const wrapper = factory(
      {},
      {
        subheader: '<div class="subheader-content">Subheader Content</div>',
      },
    );
    expect(wrapper.find(".subheader-content").exists()).toBe(true);
  });

  it("renders content-header section", () => {
    const wrapper = factory();
    expect(wrapper.find('[aid="header-content"]').exists()).toBe(true);
  });

  it("emits load-section-templates event when button is clicked", async () => {
    const wrapper = factory();
    await wrapper.find("[aid='load-existing-templates']").trigger("click");
    expect(wrapper.emitted("load-section-templates")).toBeTruthy();
  });

  it("renders selectedName when provided", () => {
    const wrapper = factory();
    expect(wrapper.find("[aid='reset-btn']").text()).toContain("selected test name");
  });

  it("emits reset event when reset button is clicked", async () => {
    const wrapper = factory();
    await wrapper.find("[aid='reset-btn']").trigger("click");
    expect(wrapper.emitted("reset")).toBeTruthy();
  });

  it("emits clear event when clear button is clicked", async () => {
    const wrapper = factory();
    await wrapper.find("[aid='clear-btn']").trigger("click");
    expect(wrapper.emitted("clear")).toBeTruthy();
  });

  it("renders default slot content", () => {
    const wrapper = factory({}, { default: '<div class="default-slot-content">Default Slot Content</div>' });

    expect(wrapper.find(".default-slot-content").exists()).toBe(true);
  });

  it("applies props correctly", () => {
    const wrapper = mount(HybridSectionWrapper, {
      global: {
        plugins: [Quasar],
      },
      props: {
        label: "Test Label",
        subheader: "Test Subheader",
        defaultOpened: true,
        sectionDisabled: true,
        hideExpendIcon: true,
        showSubHeader: true,
        disableClosing: true,
        disableOpening: true,
        sectionInvalid: true,
        loading: true,
      },
    });
    expect(wrapper.props()).toMatchObject({
      label: "Test Label",
      subheader: "Test Subheader",
      defaultOpened: true,
      sectionDisabled: true,
      hideExpendIcon: true,
      showSubHeader: true,
      disableClosing: true,
      disableOpening: true,
      sectionInvalid: true,
      loading: true,
    });
  });
});
