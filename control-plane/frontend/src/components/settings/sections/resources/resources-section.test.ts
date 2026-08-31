/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import ResourcesSection from "./resources-section.vue";
import type { IResourceSectionSettings } from "@/models/setting.model";

vi.mock("@/stores/node-pool.store");

const createWrapper = (resourceSettings: IResourceSectionSettings) => {
  return mount(ResourcesSection, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      stubs: {
        RunaiExpansionItem: {
          template: "<div><slot /></div>",
        },
        ToggleSetting: {
          name: "ToggleSetting",
          props: {
            setting: { type: Object, default: () => ({}) },
            disabled: { type: Boolean, default: false },
            tooltipMessage: { type: String, default: "" },
            resourceType: { type: String },
          },
          template: "<div></div>",
        },
      },
    },
    props: { resourceSettings },
  });
};

describe("ResourcesSection - sectionSubHeader counts", () => {
  it("counts all enabled settings correctly", () => {
    const wrapper = createWrapper({
      projectsOverDepartmentQuota: { key: "projectsOverDepartmentQuota", value: "true" },
      cpuOverQuota: { key: "cpuOverQuota", value: "true" },
      overQuotaPriority: { key: "overQuotaPriority", value: "true" },
      gpuOptimization: { key: "gpuOptimization", value: "true" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (4) / Disabled (0)");
  });

  it("counts all disabled settings correctly", () => {
    const wrapper = createWrapper({
      projectsOverDepartmentQuota: { key: "projectsOverDepartmentQuota", value: "false" },
      cpuOverQuota: { key: "cpuOverQuota", value: "false" },
      overQuotaPriority: { key: "overQuotaPriority", value: "false" },
      gpuOptimization: { key: "gpuOptimization", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (4)");
  });

  it("counts mixed enabled and disabled settings correctly", () => {
    const wrapper = createWrapper({
      projectsOverDepartmentQuota: { key: "projectsOverDepartmentQuota", value: "true" },
      cpuOverQuota: { key: "cpuOverQuota", value: "false" },
      overQuotaPriority: { key: "overQuotaPriority", value: "true" },
      gpuOptimization: { key: "gpuOptimization", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (2) / Disabled (2)");
  });

  it("excludes null/undefined settings from the count", () => {
    const wrapper = createWrapper({
      projectsOverDepartmentQuota: { key: "projectsOverDepartmentQuota", value: "true" },
      cpuOverQuota: undefined,
      overQuotaPriority: undefined,
      gpuOptimization: { key: "gpuOptimization", value: "false" },
    });
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (1) / Disabled (1)");
  });

  it("shows Enabled (0) / Disabled (0) when no settings are present", () => {
    const wrapper = createWrapper({});
    expect(wrapper.vm.sectionSubHeader).toBe("Enabled (0) / Disabled (0)");
  });
});

describe("ResourcesSection - permission wiring", () => {
  it("does not override the toggle resourceType (falls back to toggle-setting's 'Settings' default)", () => {
    const wrapper = createWrapper({
      projectsOverDepartmentQuota: { key: "projectsOverDepartmentQuota", value: "true" },
      cpuOverQuota: { key: "cpuOverQuota", value: "false" },
    });

    const toggles = wrapper.findAllComponents({ name: "ToggleSetting" });
    expect(toggles.length).toBeGreaterThan(0);
    expect(toggles.every((t) => t.props("resourceType") === undefined)).toBe(true);
  });
});
