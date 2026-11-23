/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { EResourceEditorKind } from "@/models/node-pool-scheduling.model";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { QBtn, QInput, QMenu, Quasar } from "quasar";
import Editor from "@/components/node-pools/department-resource-editor/department-resource-editor.vue";
const popup = {
  template: '<td><slot name="input" /></td>',
  props: ["isSubmitDisabled"],
  emits: ["apply", "reset", "before-show"],
};
describe("resource editor opening transition", () => {
  it("does not overwrite input entered while the popup is opening", async () => {
    const wrapper = mount(Editor, {
      props: { id: "a", kind: EResourceEditorKind.Weight, value: 2 },
      global: { plugins: [Quasar] },
      attachTo: document.body,
    });
    try {
      await wrapper.get(".cell-trigger").trigger("click");
      await flushPromises();
      const input = wrapper.findComponent(QInput).get("input");
      await input.setValue("7");
      wrapper.findComponent(QMenu).vm.$emit("show");
      await wrapper.vm.$nextTick();
      expect((input.element as HTMLInputElement).value).toBe("7");
    } finally {
      wrapper.unmount();
    }
  });
});
describe("department resource editor", () => {
  it("keeps scheduling cell validation feedback local while blocking invalid Apply", async () => {
    const wrapper = mount(Editor, {
      props: {
        id: "a",
        kind: EResourceEditorKind.Quota,
        value: 6,
        context: { childQuotaSum: 10, enforceChildQuota: false },
      },
      global: { plugins: [Quasar], stubs: { QMenu: { template: "<div><slot /></div>" } } },
    });
    await wrapper.get(".cell-trigger").trigger("click");
    wrapper.findComponent(QMenu).vm.$emit("before-show");
    await wrapper.vm.$nextTick();
    expect(wrapper.classes()).toContain("resource-cell--warning");
    const apply = wrapper.findAllComponents(QBtn).filter((button) => button.props("label") === "APPLY")[0];
    expect(apply.attributes("disabled")).toBeUndefined();
    await wrapper.get("input").setValue("-2");
    expect(wrapper.classes()).toContain("resource-cell--error");
    expect(wrapper.classes()).not.toContain("resource-cell--warning");
    expect(wrapper.text()).toContain("Enter a value equal to or above 0");
    expect(apply.attributes("disabled")).toBeDefined();
    await apply.trigger("click");
    expect(wrapper.emitted("apply")).toBeUndefined();
    wrapper.unmount();
  });

  it("applies a numeric zero maximum and reserves Unlimited for an empty input", async () => {
    const wrapper = mount(Editor, {
      props: { id: "a", kind: EResourceEditorKind.Limit, value: -1, quota: 0 },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    await wrapper.get("input").setValue("0");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[0]]);
    await wrapper.get("input").setValue("");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[0], [-1]]);
    wrapper.unmount();
  });
  it("rejects letters, exponents and decimal entry in the whole-number weight input", () => {
    const wrapper = mount(Editor, {
      props: { id: "a", kind: EResourceEditorKind.Weight, value: 2 },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    const input = wrapper.get("input").element;
    for (const key of ["a", "e", "E", "+", "-", "."]) {
      const event = new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true });
      input.dispatchEvent(event);
      expect(event.defaultPrevented, key).toBe(true);
    }
    const navigation = new KeyboardEvent("keydown", { key: "ArrowLeft", bubbles: true, cancelable: true });
    input.dispatchEvent(navigation);
    expect(navigation.defaultPrevented).toBe(false);
    wrapper.unmount();
  });
  it("shows the maximum-allocation input label as well as its popup title", () => {
    const wrapper = mount(Editor, {
      props: { id: "a", kind: EResourceEditorKind.Limit, value: -1 },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(wrapper.get(".q-field__label").text()).toBe("Max GPU allocation");
    wrapper.unmount();
  });
  it("uses a supplied non-preemptible usage floor in quota recommendations", () => {
    const wrapper = mount(Editor, {
      props: {
        value: 12,
        kind: EResourceEditorKind.Quota,
        id: "a",
        context: { childQuotaSum: 3, nonPreemptibleUsage: 10 },
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(wrapper.get(".quota-hint").text()).toContain("from 10");
    wrapper.unmount();
  });
  it("recommends and warns against the capacity remaining after other finite quotas", async () => {
    const wrapper = mount(Editor, {
      props: {
        value: 6,
        kind: EResourceEditorKind.Quota,
        id: "a",
        context: { childQuotaSum: 0, capacity: 144, totalQuota: 30, finiteQuotaSum: 30, unlimitedQuotaCount: 0 },
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(wrapper.get(".quota-hint").text()).toContain("to 120");
    await wrapper.get("input").setValue("130");
    expect(wrapper.get("p.text-warning").text()).toContain("up to 120 GPUs");
    wrapper.unmount();
  });
  it("does not invent a remaining maximum when other quotas are unlimited or aggregates are incomplete", async () => {
    const wrapper = mount(Editor, {
      props: {
        value: 6,
        kind: EResourceEditorKind.Quota,
        id: "a",
        context: { childQuotaSum: 0, capacity: 144, finiteQuotaSum: 30, unlimitedQuotaCount: 1 },
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(wrapper.get(".quota-hint").text()).not.toContain(" to ");
    await wrapper.get("input").setValue("200");
    expect(wrapper.get("p.text-warning").text()).not.toContain("up to");
    wrapper.unmount();
    const incomplete = mount(Editor, {
      props: { value: 6, kind: EResourceEditorKind.Quota, id: "a", context: { childQuotaSum: 0, capacity: 144 } },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(incomplete.get(".quota-hint").text()).not.toContain(" to ");
    incomplete.unmount();
  });
  it("reclaims the edited department's own unlimited sentinel without counting it as another unlimited quota", async () => {
    const wrapper = mount(Editor, {
      props: {
        value: -1,
        kind: EResourceEditorKind.Quota,
        id: "default",
        context: { childQuotaSum: 0, capacity: 144, finiteQuotaSum: 24, unlimitedQuotaCount: 1 },
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    await wrapper.get("input").setValue("6");
    expect(wrapper.get(".quota-hint").text()).toContain("to 120");
    wrapper.unmount();
  });
  it("shows only blocking errors for an invalid quota, restoring advisories for a valid value", async () => {
    const wrapper = mount(Editor, {
      props: {
        value: 6,
        kind: EResourceEditorKind.Quota,
        id: "a",
        context: { childQuotaSum: 10, enforceChildQuota: false },
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    const menu = wrapper.findComponent(popup);
    await wrapper.get("input").setValue("-2");
    expect(wrapper.text()).toContain("Enter a value equal to or above 0");
    expect(wrapper.find(".text-warning").exists()).toBe(false);
    menu.vm.$emit("apply");
    expect(wrapper.emitted("apply")).toBeUndefined();
    await wrapper.get("input").setValue("6");
    expect(wrapper.find(".text-warning").exists()).toBe(true);
    expect(wrapper.findComponent(QInput).props("color")).toBe("warning");
    menu.vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[6]]);
    wrapper.unmount();
  });
  it("renders the saved unlimited quota sentinel as Unlimited", () => {
    const wrapper = mount(Editor, {
      props: { value: -1, kind: EResourceEditorKind.Quota, id: "default" },
      global: { plugins: [Quasar] },
    });
    expect(wrapper.get('[role="button"]').text()).toBe("Unlimited");
    wrapper.unmount();
  });
  it("does not interpret incomplete native numeric input as Unlimited", async () => {
    const wrapper = mount(Editor, {
      props: { id: "a", kind: EResourceEditorKind.Limit, value: 12 },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    const input = wrapper.get("input");
    await input.setValue("");
    Object.defineProperty(input.element, "validity", { configurable: true, value: { badInput: true } });
    await input.trigger("input");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toBeUndefined();
    Object.defineProperty(input.element, "validity", { configurable: true, value: { badInput: false } });
    await input.trigger("input");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[-1]]);
    wrapper.unmount();
  });
  it("blocks a maximum below quota and preserves Unlimited without project-quota reads", async () => {
    const wrapper = mount(Editor, {
      props: { value: -1, kind: EResourceEditorKind.Limit, quota: 5, id: "a" },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    const menu = wrapper.findComponent(popup);
    menu.vm.$emit("before-show");
    await wrapper.vm.$nextTick();
    expect((wrapper.get("input").element as HTMLInputElement).value).toBe("");
    menu.vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[-1]]);
    await wrapper.get("input").setValue("4");
    menu.vm.$emit("apply");
    expect(wrapper.emitted("apply")).toHaveLength(1);
    await wrapper.get("input").setValue("5.5");
    menu.vm.$emit("apply");
    expect(wrapper.emitted("apply")?.[1]).toEqual([5.5]);
    wrapper.unmount();
  });
  it("enforces the Figma weight range and separates percentage from the input", async () => {
    const wrapper = mount(Editor, {
      props: {
        value: 2,
        kind: EResourceEditorKind.Weight,
        id: "a",
        previewWeightPercentage: (weight: number) => weight * 5,
      },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    expect(wrapper.get(".weight-percentage").text()).toBe("(10%)");
    expect(wrapper.find(".q-field__suffix").exists()).toBe(false);
    for (const value of ["", "-2", "0.5", "21", "2223"]) {
      await wrapper.get("input").setValue(value);
      expect(wrapper.text()).toContain("Enter a whole number from 0 to 20");
      expect(wrapper.find(".weight-percentage").exists()).toBe(false);
      wrapper.findComponent(popup).vm.$emit("apply");
    }
    expect(wrapper.emitted("apply")).toBeUndefined();
    await wrapper.get("input").setValue("20");
    expect(wrapper.get(".weight-percentage").text()).toBe("(100%)");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")).toEqual([[20]]);
    wrapper.unmount();
  });
  it("only applies valid values, and cancel resets without mutation", async () => {
    const wrapper = mount(Editor, {
      props: { value: 2, kind: EResourceEditorKind.Weight, id: "a" },
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
      props: { value: 2, kind: EResourceEditorKind.Quota, id: "a", limit: -1, context: { childQuotaSum: 0 } },
      global: { plugins: [Quasar], stubs: { EditableCellPopup: popup } },
    });
    await wrapper.get("input").setValue("5");
    wrapper.findComponent(popup).vm.$emit("apply");
    expect(wrapper.emitted("apply")?.[0]).toEqual([5]);
    wrapper.unmount();
  });
});
