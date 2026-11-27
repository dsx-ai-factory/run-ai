/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import SchedulingSectionTitle from "./scheduling-section-title.vue";

describe("scheduling section help presentation", () => {
  it("preserves scheduling popup presentation for ordinary question-mark help", () => {
    const wrapper = mount(SchedulingSectionTitle, {
      props: { title: "Historical window", tooltipText: "Window help" },
      global: { plugins: [Quasar] },
    });
    const tooltip = wrapper.findComponent(RunaiTooltip);
    expect(tooltip.props("popupClass")).toContain("scheduling-help-popup");
    expect(tooltip.props("anchor")).toBe("center right");
    expect(tooltip.props("self")).toBe("center left");
    expect(wrapper.text()).toContain("Historical window");
    wrapper.unmount();
  });
});
