/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { QIcon, Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import RunaiBaseDialog from "./runai-base-dialog.vue";

describe("base dialog close control", () => {
  it.each([
    ["existing", undefined],
    ["custom", "img:/department-close.svg"],
  ])("keeps the %s icon and dismisses the dialog", async (_name, closeIcon) => {
    const wrapper = mount(RunaiBaseDialog, {
      props: { modelValue: true, closeIcon },
      global: { plugins: [Quasar], stubs: { QDialog: { template: "<div><slot /></div>" } } },
    });
    try {
      const button = wrapper.get(".runai-base-dialog-close-btn");
      const icon = wrapper.findComponent(QIcon);
      if (closeIcon) expect(icon.get("img").attributes("src")).toBe("/department-close.svg");
      else expect(icon.classes()).toEqual(expect.arrayContaining(["ph-bold", "ph-x"]));
      await button.trigger("click");
      expect(wrapper.emitted("close")).toHaveLength(1);
    } finally {
      wrapper.unmount();
    }
  });
});
