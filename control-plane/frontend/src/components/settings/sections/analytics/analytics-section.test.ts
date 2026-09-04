/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import AnalyticsSection from "./analytics-section.vue";
import type { IAnalyticsSectionSettings } from "@/models/setting.model";

const createWrapper = (analyticsSectionSettings: IAnalyticsSectionSettings) => {
  return mount(AnalyticsSection, {
    global: {
      plugins: [Quasar],
      stubs: {
        RunaiExpansionItem: {
          template: "<div><slot /></div>",
        },
        ToggleSetting: {
          name: "ToggleSetting",
          props: {
            setting: { type: Object, default: () => ({}) },
            resourceType: { type: String },
          },
          template: "<div></div>",
        },
      },
    },
    props: { analyticsSectionSettings },
  });
};

describe("AnalyticsSection - sectionSubHeader counts", () => {
  it("shows Enabled (1) / Disabled (0) when advancedGpuMetrics is enabled", () => {
    const wrapper = createWrapper({
      advancedGpuMetrics: { key: "advancedGpuMetrics", value: "true" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (0)");
  });

  it("shows Enabled (0) / Disabled (1) when advancedGpuMetrics is disabled", () => {
    const wrapper = createWrapper({
      advancedGpuMetrics: { key: "advancedGpuMetrics", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (1)");
  });

  it("shows Enabled (0) / Disabled (0) when advancedGpuMetrics is not present", () => {
    const wrapper = createWrapper({});
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (0)");
  });
});

describe("AnalyticsSection - permission wiring", () => {
  it("does not override the toggle resourceType (falls back to toggle-setting's 'Settings' default)", () => {
    const wrapper = createWrapper({
      advancedGpuMetrics: { key: "advancedGpuMetrics", value: "true" },
    });

    const toggle = wrapper.findComponent({ name: "ToggleSetting" });
    expect(toggle.exists()).toBe(true);
    expect(toggle.props("resourceType")).toBeUndefined();
  });
});
