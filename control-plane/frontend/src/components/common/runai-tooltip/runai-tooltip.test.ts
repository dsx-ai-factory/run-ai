/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { QTooltip, Quasar } from "quasar";
import RunaiTooltip from "./runai-tooltip.vue";

afterEach(() => vi.useRealTimers());
describe("shared help tooltip", () => {
  it("preserves native tooltip events and existing defaults", () => {
    const wrapper = mount(RunaiTooltip, { props: { tooltipText: "Help" }, global: { plugins: [Quasar] } });
    const popup = wrapper.findComponent(QTooltip);
    expect(popup.props()).toMatchObject({
      noParentEvent: false,
      maxWidth: "250px",
      anchor: "top middle",
      self: "bottom middle",
    });
    popup.vm.$emit("show");
    popup.vm.$emit("hide");
    expect(wrapper.emitted("revealed")).toHaveLength(1);
    expect(wrapper.emitted("hidden")).toHaveLength(1);
    wrapper.unmount();
  });

  it("allows explicit anchor, self and offset values", () => {
    const wrapper = mount(RunaiTooltip, {
      props: { tooltipText: "Help", anchor: "center right", self: "center left", offset: [120, 0] },
      global: { plugins: [Quasar] },
    });
    expect(wrapper.findComponent(QTooltip).props()).toMatchObject({
      anchor: "center right",
      self: "center left",
      offset: [120, 0],
    });
    wrapper.unmount();
  });

  it("keeps documentation available while focus transfers from the trigger into the popup", async () => {
    vi.useFakeTimers();
    const wrapper = mount(RunaiTooltip, {
      props: { tooltipText: "Documentation", interactive: true },
      slots: { content: '<a href="https://docs.run.ai">Documentation</a>' },
      global: { plugins: [Quasar] },
    });
    const popup = wrapper.findComponent(QTooltip);
    await wrapper.trigger("focusin");
    expect(popup.props("modelValue")).toBe(true);
    expect(popup.props("noParentEvent")).toBe(true);
    await wrapper.trigger("focusout");
    popup.vm.$emit("focusin");
    await vi.advanceTimersByTimeAsync(250);
    expect(popup.props("modelValue")).toBe(true);
    popup.vm.$emit("focusout");
    await vi.advanceTimersByTimeAsync(250);
    expect(popup.props("modelValue")).toBe(false);
    wrapper.unmount();
  });

  it("keeps documentation open across a pointer gap and closes on Escape", async () => {
    vi.useFakeTimers();
    const wrapper = mount(RunaiTooltip, {
      props: { tooltipText: "Documentation", interactive: true },
      global: { plugins: [Quasar] },
    });
    const popup = wrapper.findComponent(QTooltip);
    await wrapper.trigger("mouseenter");
    expect(popup.props("modelValue")).toBe(true);
    await wrapper.trigger("mouseleave");
    await vi.advanceTimersByTimeAsync(100);
    popup.vm.$emit("mouseenter");
    await vi.advanceTimersByTimeAsync(250);
    expect(popup.props("modelValue")).toBe(true);
    await wrapper.trigger("keydown", { key: "Escape" });
    expect(popup.props("modelValue")).toBe(false);
    wrapper.unmount();
  });
});
