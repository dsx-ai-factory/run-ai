/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import ChartZoomControls from "./chart-zoom-controls.vue";

describe("ChartZoomControls", () => {
  function mountComponent(): VueWrapper {
    return mount(ChartZoomControls, {
      global: { plugins: [Quasar] },
    });
  }

  it("uses Phosphor icons for zoom in and zoom out", () => {
    const wrapper = mountComponent();
    const buttons = wrapper.findAllComponents({ name: "QBtn" });

    expect(buttons[0].props("icon")).toBe("ph-bold ph-magnifying-glass-plus");
    expect(buttons[1].props("icon")).toBe("ph-bold ph-magnifying-glass-minus");
  });

  it("positions both tooltips to the left of their buttons", () => {
    const wrapper = mountComponent();
    const tooltips = wrapper.findAllComponents(RunaiTooltipWrapper);

    expect(tooltips).toHaveLength(2);
    tooltips.forEach((tooltip) => {
      expect(tooltip.props("anchor")).toBe("center left");
      expect(tooltip.props("self")).toBe("center right");
    });
  });

  it("preserves zoom events and disabled states", async () => {
    const wrapper = mountComponent();
    const buttons = wrapper.findAllComponents({ name: "QBtn" });

    expect(buttons[0].props("disable")).toBe(false);
    expect(buttons[1].props("disable")).toBe(true);

    await buttons[0].trigger("click");

    expect(wrapper.emitted("zoom-change")).toEqual([[{ start: 10, end: 90 }]]);
    expect(buttons[1].props("disable")).toBe(false);

    await buttons[1].trigger("click");

    expect(wrapper.emitted("zoom-change")?.[1]).toEqual([{ start: 0, end: 100 }]);
    expect(buttons[1].props("disable")).toBe(true);
  });

  it("disables zoom in at the maximum zoom level", async () => {
    const wrapper = mountComponent();
    const zoomInButton = wrapper.findAllComponents({ name: "QBtn" })[0];

    await zoomInButton.trigger("click");
    await zoomInButton.trigger("click");
    await zoomInButton.trigger("click");

    expect(zoomInButton.props("disable")).toBe(true);
    expect(wrapper.emitted("zoom-change")?.at(-1)).toEqual([{ start: 25, end: 75 }]);
  });

  it.each(["+", "="])("zooms in when the %s key is pressed", async (key) => {
    const wrapper = mountComponent();
    const zoomInButton = wrapper.findAllComponents({ name: "QBtn" })[0];

    await zoomInButton.trigger("keydown", { key });

    expect(wrapper.emitted("zoom-change")).toEqual([[{ start: 10, end: 90 }]]);
  });

  it("zooms out when the minus key is pressed", async () => {
    const wrapper = mountComponent();
    const buttons = wrapper.findAllComponents({ name: "QBtn" });

    await buttons[0].trigger("keydown", { key: "+" });
    await buttons[1].trigger("keydown", { key: "-" });

    expect(wrapper.emitted("zoom-change")?.at(-1)).toEqual([{ start: 0, end: 100 }]);
  });
});
