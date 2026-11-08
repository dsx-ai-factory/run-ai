/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import Editor from "./department-resource-editor.vue";
const popup = {
  template: '<td><slot name="input" /></td>',
  props: ["isSubmitDisabled"],
  emits: ["apply", "reset", "opened"],
};
describe("department resource editor", () => {
  it("only applies valid values, and cancel resets without mutation", async () => {
    const wrapper = mount(Editor, {
      props: { value: 2, kind: "weight", id: "a" },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    const input = wrapper.get("input");
    await input.setValue("5");
    expect(wrapper.emitted("apply")).toBeUndefined();
    wrapper.findComponent(popup).vm.$emit("reset");
    await wrapper.vm.$nextTick();
    expect((input.element as HTMLInputElement).value).toBe("2");
    await input.setValue("-1");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toBeUndefined();
    await input.setValue("0");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")?.[0]).toEqual([0]);
  });
  it("applies finite quota with saved unlimited maximum allocation", async () => {
    const wrapper = mount(Editor, {
      props: { value: 2, kind: "quota", id: "a", limit: -1, context: { childQuotaSum: 0 } },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    await wrapper.get("input").setValue("5");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")?.[0]).toEqual([5]);
    wrapper.unmount();
  });
});
