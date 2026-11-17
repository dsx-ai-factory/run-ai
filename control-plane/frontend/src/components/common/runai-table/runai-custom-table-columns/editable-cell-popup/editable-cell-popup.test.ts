/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { QBtn, QMenu, Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import EditableCellPopup from "./editable-cell-popup.vue";

describe("editable cell popup", () => {
  it("preserves the existing opened-after-show lifecycle by default", async () => {
    const wrapper = mount(EditableCellPopup, { props: { title: "Weight" }, global: { plugins: [Quasar] } });
    const menu = wrapper.findComponent(QMenu);
    await wrapper.get(".cell-trigger").trigger("click");
    expect(wrapper.emitted("opened")).toBeUndefined();
    menu.vm.$emit("show");
    expect(wrapper.emitted("opened")).toHaveLength(1);
    wrapper.unmount();
  });
  it("exposes before-show separately without changing the opened lifecycle", async () => {
    const wrapper = mount(EditableCellPopup, {
      props: { title: "Weight" },
      global: { plugins: [Quasar] },
    });
    await wrapper.get(".cell-trigger").trigger("click");
    expect(wrapper.emitted("before-show")).toHaveLength(1);
    expect(wrapper.emitted("opened")).toBeUndefined();
    wrapper.findComponent(QMenu).vm.$emit("show");
    expect(wrapper.emitted("opened")).toHaveLength(1);
    wrapper.unmount();
  });
  it("disables Apply during invalid or loading input and enables it when submission becomes valid", async () => {
    const wrapper = mount(EditableCellPopup, {
      props: { title: "Weight", isSubmitDisabled: true },
      global: {
        plugins: [Quasar],
        stubs: { QMenu: { template: "<div><slot /></div>" } },
      },
    });
    const apply = wrapper.findAllComponents(QBtn).filter((button) => button.props("label") === "APPLY")[0];
    expect(apply.attributes("disabled")).toBeDefined();
    await apply.trigger("click");
    expect(wrapper.emitted("apply")).toBeUndefined();

    await wrapper.setProps({ isSubmitDisabled: false });
    expect(apply.attributes("disabled")).toBeUndefined();
    await apply.trigger("click");
    expect(wrapper.emitted("apply")).toHaveLength(1);
    wrapper.unmount();
  });
});
