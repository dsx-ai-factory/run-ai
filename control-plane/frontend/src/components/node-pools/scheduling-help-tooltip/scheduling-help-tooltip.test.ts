/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { QTooltip, Quasar } from "quasar";
import SchedulingHelpTooltip from "@/components/node-pools/scheduling-help-tooltip/scheduling-help-tooltip.vue";

afterEach(() => vi.useRealTimers());
describe("scheduling help", () => {
  it("keeps documentation available while focus moves from the trigger into the popup", async () => {
    vi.useFakeTimers();
    const wrapper = mount(SchedulingHelpTooltip, {
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
  it("leaves plain help to native tooltip events and forwards the weight popup offset", () => {
    const wrapper = mount(SchedulingHelpTooltip, {
      props: { tooltipText: "Weight", offset: [120, 0] },
      global: { plugins: [Quasar] },
    });
    const popup = wrapper.findComponent(QTooltip);
    expect(popup.props("noParentEvent")).toBe(false);
    expect(popup.props("offset")).toEqual([120, 0]);
    expect(popup.props("anchor")).toBe("center right");
    expect(popup.props("self")).toBe("center left");
    wrapper.unmount();
  });
});
