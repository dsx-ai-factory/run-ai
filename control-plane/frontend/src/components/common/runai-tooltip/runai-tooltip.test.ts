/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { QTooltip, Quasar } from "quasar";
import { afterEach, describe, expect, it, vi } from "vitest";
import Tooltip from "./runai-tooltip.vue";

afterEach(() => vi.useRealTimers());
describe("interactive tooltip opt-in", () => {
  it("keeps existing consumers on Quasar's default parent events", () => {
    const wrapper = mount(Tooltip, { props: { tooltipText: "Help" }, global: { plugins: [Quasar] } });
    expect(wrapper.findComponent(QTooltip).props("noParentEvent")).toBe(false);
    wrapper.unmount();
  });
  it("allows crossing from the trigger into tooltip content and closes after leaving", async () => {
    vi.useFakeTimers();
    const wrapper = mount(Tooltip, {
      props: { tooltipText: "Documentation", interactive: true },
      global: { plugins: [Quasar] },
    });
    const tooltip = wrapper.findComponent(QTooltip);
    await wrapper.trigger("mouseenter");
    expect(tooltip.props("modelValue")).toBe(true);
    await wrapper.trigger("mouseleave");
    await vi.advanceTimersByTimeAsync(100);
    tooltip.vm.$emit("mouseenter");
    await vi.advanceTimersByTimeAsync(300);
    expect(tooltip.props("modelValue")).toBe(true);
    tooltip.vm.$emit("mouseleave");
    await vi.advanceTimersByTimeAsync(250);
    expect(tooltip.props("modelValue")).toBe(false);
    wrapper.unmount();
  });
});
