/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { h } from "vue";
import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe, expect, it } from "vitest";

import ChartWidgetWrapper from "./chart-widget-wrapper.vue";

function mountWrapper(
  props: Partial<InstanceType<typeof ChartWidgetWrapper>["$props"]> = {},
  slots: Record<string, unknown> = {},
): VueWrapper<InstanceType<typeof ChartWidgetWrapper>> {
  return mount(ChartWidgetWrapper, {
    global: {
      plugins: [Quasar],
      stubs: {
        RunaiActionsMenu: true,
        RunaiTooltipWrapper: true,
      },
    },
    props: {
      options: { title: "GPU utilization" },
      hideActions: true,
      ...props,
    },
    slots: slots as never,
  });
}

describe("chart-widget-wrapper", () => {
  it("uses the 24px overflow-menu icon when chart actions are visible", () => {
    const wrapper = mountWrapper({ hideActions: false });

    expect(wrapper.getComponent({ name: "RunaiActionsMenu" }).props("largeIcon")).toBe(true);
  });

  it("preserves the chart CSV download action", () => {
    const wrapper = mountWrapper({ hideActions: false });
    const [downloadAction] = wrapper.getComponent({ name: "RunaiActionsMenu" }).props("actionItems") as Array<{
      label: string;
      action: () => void;
    }>;

    expect(downloadAction.label).toBe("Download as CSV");

    downloadAction.action();

    expect(wrapper.emitted("export-csv")).toEqual([[]]);
  });

  it("uses a native heading to label one focusable semantic chart region", () => {
    const wrapper = mountWrapper({}, { default: '<svg class="chart-svg" tabindex="-1" />' });
    const heading = wrapper.get("h3.chart-widget-title");
    const region = wrapper.get(".chart-semantic-region");

    expect(heading.text()).toBe("GPU utilization");
    expect(region.attributes("role")).toBe("region");
    expect(region.attributes("aria-roledescription")).toBe("chart");
    expect(region.attributes("tabindex")).toBe("0");
    expect(region.attributes("aria-labelledby")).toBe(heading.attributes("id"));
    expect(region.attributes("aria-describedby")).toContain(`${heading.attributes("id")}-description`);
    expect(region.attributes("aria-describedby")).toContain(`${heading.attributes("id")}-summary`);
    expect(wrapper.findAll('[tabindex="0"]')).toHaveLength(1);
  });

  it("exposes the generated title id to custom headers", () => {
    const wrapper = mountWrapper(
      {},
      {
        header: ({ titleId }: { titleId: string }) => h("h3", { id: titleId }, "Custom chart title"),
      },
    );
    const heading = wrapper.get("h3");

    expect(wrapper.get(".chart-semantic-region").attributes("aria-labelledby")).toBe(heading.attributes("id"));
  });

  it("does not add a chart Tab stop to grid widgets", () => {
    const wrapper = mountWrapper({ isGrid: true });

    expect(wrapper.get(".chart-semantic-region").attributes("tabindex")).toBeUndefined();
  });
});
