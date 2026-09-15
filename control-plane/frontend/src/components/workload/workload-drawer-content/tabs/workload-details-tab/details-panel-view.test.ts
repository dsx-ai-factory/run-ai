/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import DetailsPanelView from "./details-panel-view.vue";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

const RUNAI_DETAILS_PANEL_STUB = {
  name: "RunaiDetailsPanel",
  template: "<div><slot name='controls-prepend' /></div>",
  props: ["sections"],
};

const SOME_SECTIONS: IDetailsPanelSectionDef[] = [
  { id: "general", title: "General information", customMatch: () => true, data: { groups: [] } },
  { id: "compute", title: "Compute", customMatch: () => true, data: { items: [] } },
];

// Pure-render tests: details-panel-view doesn't build section data itself — it just renders
// whatever `sections` model it's given, regardless of whether that model was built by
// workload-details-tab or element-details-tab. Section-building is covered by each of those
// components' own tests.
function mountView(sections: IDetailsPanelSectionDef[], overrides: Record<string, unknown> = {}) {
  return mount(DetailsPanelView, {
    props: { sections, ...overrides },
    global: {
      plugins: [[Quasar, { plugins: {} }]],
      stubs: {
        RunaiDetailsPanel: RUNAI_DETAILS_PANEL_STUB,
        RunaiSelect: true,
      },
    },
  });
}

describe("DetailsPanelView", () => {
  it("shows the loader when loading is true, regardless of sections", () => {
    const wrapper = mountView(SOME_SECTIONS, { loading: true });

    expect(wrapper.findComponent(RunaiDefaultLoader).exists()).toBe(true);
    expect(wrapper.findComponent(RunaiDrawerNoData).exists()).toBe(false);
    expect(wrapper.findComponent({ name: "RunaiDetailsPanel" }).exists()).toBe(false);
  });

  it("shows the no-data state when sections is empty", () => {
    const wrapper = mountView([]);

    const noData = wrapper.findComponent(RunaiDrawerNoData);
    expect(noData.exists()).toBe(true);
    expect(noData.props("messages")).toEqual(["There are no details yet."]);
    expect(wrapper.findComponent({ name: "RunaiDetailsPanel" }).exists()).toBe(false);
  });

  it("renders runai-details-panel instead of the no-data state when sections is non-empty", () => {
    const wrapper = mountView(SOME_SECTIONS);

    expect(wrapper.findComponent(RunaiDrawerNoData).exists()).toBe(false);
    expect(wrapper.findComponent({ name: "RunaiDetailsPanel" }).exists()).toBe(true);
  });

  it("passes the given sections through to runai-details-panel unchanged", () => {
    const wrapper = mountView(SOME_SECTIONS);
    expect(wrapper.findComponent({ name: "RunaiDetailsPanel" }).props("sections")).toEqual(SOME_SECTIONS);
  });

  it("hides the spec select when showSpecSelect is false", () => {
    const wrapper = mountView(SOME_SECTIONS, { showSpecSelect: false });
    expect(wrapper.findComponent({ name: "RunaiSelect" }).exists()).toBe(false);
  });

  it("shows the spec select when showSpecSelect is true", () => {
    const wrapper = mountView(SOME_SECTIONS, {
      showSpecSelect: true,
      specOptions: ["Master", "Worker"],
      selectedSpec: "Master",
    });
    expect(wrapper.findComponent({ name: "RunaiSelect" }).exists()).toBe(true);
  });

  it("emits update:selectedSpec when the select value changes", async () => {
    const wrapper = mountView(SOME_SECTIONS, {
      showSpecSelect: true,
      specOptions: ["Master", "Worker"],
      selectedSpec: "Master",
    });

    wrapper.findComponent({ name: "RunaiSelect" }).vm.$emit("update:model-value", "Worker");
    await wrapper.vm.$nextTick();

    expect(wrapper.emitted("update:selectedSpec")).toEqual([["Worker"]]);
  });
});
