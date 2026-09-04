/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import ChartDrilldownBreadcrumbs from "./chart-drilldown-breadcrumbs.vue";

describe("chart-drilldown-breadcrumbs", () => {
  it("renders previous levels as buttons and the current level as text", () => {
    const wrapper = mount(ChartDrilldownBreadcrumbs, {
      props: { items: ["All node pools", "department-a", "project-a"] },
    });

    expect(wrapper.findAll("button").map((button) => button.text())).toEqual(["All node pools", "department-a"]);
    expect(wrapper.get('[aria-current="page"]').text()).toBe("project-a");
    expect(wrapper.get('[role="navigation"]').attributes("aria-label")).toBe("Chart drilldown");
  });

  it("emits the selected breadcrumb index", async () => {
    const wrapper = mount(ChartDrilldownBreadcrumbs, {
      props: { items: ["All node pools", "department-a", "project-a"] },
    });

    await wrapper.findAll("button")[1].trigger("click");

    expect(wrapper.emitted("navigate")).toEqual([[1]]);
  });

  it("preserves widget-specific positioning and hides an empty path", async () => {
    const wrapper = mount(ChartDrilldownBreadcrumbs, {
      props: { items: ["All node pools", "department-a"], top: 18, left: 72, separatorGap: 10 },
    });

    expect(wrapper.attributes("style")).toContain("top: 18px");
    expect(wrapper.attributes("style")).toContain("left: 72px");
    expect(wrapper.get(".breadcrumb-separator").attributes("style")).toContain("10px");

    await wrapper.setProps({ items: [] });

    expect(wrapper.attributes("style")).toContain("display: none");
  });
});
