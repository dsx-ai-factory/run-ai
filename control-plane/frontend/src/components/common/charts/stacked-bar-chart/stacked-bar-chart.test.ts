/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { nextTick } from "vue";
import type { IStackedBarSegment } from "./stacked-bar-chart.model";
import type { EChartsOption, BarSeriesOption } from "echarts";

import StackedBarChart from "./stacked-bar-chart.vue";

describe("StackedBarChart", () => {
  const defaultSegments: IStackedBarSegment[] = [
    { value: 30, color: "#FF0000", label: "Red" },
    { value: 50, color: "#00FF00", label: "Green" },
    { value: 20, color: "#0000FF", label: "Blue" },
  ];

  const mountComponent = async (
    props: Partial<{
      segments: IStackedBarSegment[];
      height: number;
      width: number;
      showLabels: boolean;
      minValue: number;
      maxValue: number;
      showScale: boolean;
      entityName: string;
      showPercentage: boolean;
      tooltipSummary: string;
      labelTextBorderWidth: number;
    }> = {},
  ) => {
    const wrapper = mount(StackedBarChart, {
      props: {
        segments: defaultSegments,
        ...props,
      },
    });

    await flushPromises();
    await nextTick();
    return wrapper;
  };

  const getChartOptions = (wrapper: Awaited<ReturnType<typeof mountComponent>>): EChartsOption => {
    return (wrapper.vm as unknown as Record<string, unknown>).chartOptions as EChartsOption;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("Rendering", () => {
    it("renders the component correctly", async () => {
      const wrapper = await mountComponent();
      expect(wrapper.exists()).toBe(true);
      expect(wrapper.find(".stacked-bar-chart").exists()).toBe(true);
      expect(wrapper.find(".chart-container").exists()).toBe(true);
    });

    it("renders the v-chart component", async () => {
      const wrapper = await mountComponent();
      const vchartComponent = wrapper.findComponent(".chart-container");
      expect(vchartComponent.exists()).toBe(true);
    });

    it("renders scale labels when showScale is true (default)", async () => {
      const wrapper = await mountComponent();
      expect(wrapper.find(".scale-labels").exists()).toBe(true);
      expect(wrapper.find(".scale-min").exists()).toBe(true);
      expect(wrapper.find(".scale-max").exists()).toBe(true);
    });

    it("hides scale labels when showScale is false (element still exists for layout)", async () => {
      const wrapper = await mountComponent({ showScale: false });
      expect(wrapper.find(".scale-labels").exists()).toBe(true);
      expect(wrapper.find(".scale-labels").classes()).toContain("scale-labels--hidden");
    });

    it("displays correct min and max values in scale labels", async () => {
      const wrapper = await mountComponent({ minValue: 10, maxValue: 200 });
      expect(wrapper.find(".scale-min").text()).toBe("10");
      expect(wrapper.find(".scale-max").text()).toBe("200");
    });

    it("displays default min (0) and max (100) values", async () => {
      const wrapper = await mountComponent();
      expect(wrapper.find(".scale-min").text()).toBe("0");
      expect(wrapper.find(".scale-max").text()).toBe("100");
    });
  });

  describe("Chart Configuration", () => {
    it("passes correct chart configuration", async () => {
      const wrapper = await mountComponent({ minValue: 0, maxValue: 150 });
      const options = getChartOptions(wrapper);

      expect(options.xAxis).toMatchObject({
        type: "value",
        show: false,
        min: 0,
        max: 150,
      });
    });

    it("builds series from segments", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      expect(series).toHaveLength(3);
      expect(series[0]).toMatchObject({
        type: "bar",
        name: "Red",
        data: [30],
        stack: "total",
        itemStyle: { color: "#FF0000" },
      });
      expect(series[1]).toMatchObject({
        type: "bar",
        name: "Green",
        data: [50],
        stack: "total",
        itemStyle: { color: "#00FF00" },
      });
      expect(series[0].clip).toBe(false);
      expect(series[0].labelLayout).toEqual(expect.any(Function));
    });

    it("keeps narrow edge-segment labels inside the chart bounds", async () => {
      const wrapper = await mountComponent();
      const chartElement = wrapper.find(".chart-container").element;
      Object.defineProperty(chartElement, "clientWidth", { configurable: true, value: 100 });
      const series = getChartOptions(wrapper).series as BarSeriesOption[];
      const labelLayout = series[0].labelLayout as (params: { labelRect: { x: number; width: number } }) => {
        dx?: number;
      };

      expect(labelLayout({ labelRect: { x: 90, width: 30 } }).dx).toBe(-20);
      expect(labelLayout({ labelRect: { x: -20, width: 30 } }).dx).toBe(20);
    });

    it("enables data labels when showLabels is true", async () => {
      const wrapper = await mountComponent({ showLabels: true });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      expect(series[0].label?.show).toBe(true);
    });

    it("uses the requested label outline width", async () => {
      const wrapper = await mountComponent({ labelTextBorderWidth: 2 });
      const series = getChartOptions(wrapper).series as BarSeriesOption[];

      expect(series[0].label?.textBorderWidth).toBe(2);
    });

    it("disables data labels when showLabels is false", async () => {
      const wrapper = await mountComponent({ showLabels: false });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      expect(series[0].label?.show).toBe(false);
    });

    it("removes all margins via grid", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);

      expect(options.grid).toMatchObject({
        top: 0,
        bottom: 0,
        left: 0,
        right: 0,
      });
    });

    it("disables legend", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);

      expect(options.legend).toMatchObject({ show: false });
    });

    it("configures stacking with stack property", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      series.forEach((s) => {
        expect(s.stack).toBe("total");
      });
    });
  });

  describe("Tooltip Formatting", () => {
    it("includes entity name in tooltip when provided", async () => {
      const wrapper = await mountComponent({ entityName: "Test Entity" });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("Test Entity");
    });

    it("does not include entity name section when not provided", async () => {
      const wrapper = await mountComponent({ entityName: "" });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).not.toContain("font-weight: 500; margin-bottom: 8px;");
    });

    it("includes all segments in tooltip", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("Red");
      expect(tooltipHtml).toContain("Green");
      expect(tooltipHtml).toContain("Blue");
      expect(tooltipHtml).toContain("30");
      expect(tooltipHtml).toContain("50");
      expect(tooltipHtml).toContain("20");
    });

    it("aligns the fixed tooltip marker with its text", async () => {
      const wrapper = await mountComponent();
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      expect((formatter as () => string)()).toContain("align-items: center");
      expect((formatter as () => string)()).toContain("flex:0 0 6px");
    });

    it("wraps long tooltip rows within a bounded width", async () => {
      const wrapper = await mountComponent({
        segments: [
          { value: 3, color: "#2167C9", label: "Weight" },
          { value: 10000, color: "#B2D8F6", label: "Total weight of other departments" },
        ],
      });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;
      const tooltipHtml = (formatter as () => string)();

      expect(tooltipHtml).toContain("max-width: 280px");
      expect(tooltipHtml).toContain("white-space: normal");
      expect(tooltipHtml).toContain("min-width: 0");
    });

    it("shows percentage in tooltip when showPercentage is true", async () => {
      const wrapper = await mountComponent({ showPercentage: true });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("30 (30%)");
      expect(tooltipHtml).toContain("50 (50%)");
      expect(tooltipHtml).toContain("20 (20%)");
    });

    it("does not show percentage when showPercentage is false", async () => {
      const wrapper = await mountComponent({ showPercentage: false });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).not.toContain("%");
    });

    it("uses 'Value' as default label when segment has no label", async () => {
      const segmentsWithoutLabels: IStackedBarSegment[] = [{ value: 50, color: "#FF0000" }];
      const wrapper = await mountComponent({ segments: segmentsWithoutLabels });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("Value");
    });

    it("includes tooltip summary when provided", async () => {
      const wrapper = await mountComponent({ tooltipSummary: "Department Engineering available quota: <b>100</b>" });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("Department Engineering available quota:");
      expect(tooltipHtml).toContain("100");
    });

    it("escapes plain fields and sanitizes the supported rich tooltip summary", async () => {
      const wrapper = await mountComponent({
        entityName: `<img src=x onerror="alert('entity')">`,
        segments: [
          {
            value: 3,
            color: "#2167C9",
            label: `<script>alert('label')</script>`,
            displayValue: `<svg onload="alert('value')">`,
          },
        ],
        tooltipSummary: `Quota: <b>100</b><img src=x onerror="alert('summary')">`,
      });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;
      const tooltipHtml = (formatter as () => string)();

      expect(tooltipHtml).toContain("&lt;img src=x onerror=&quot;");
      expect(tooltipHtml).toContain("&lt;script&gt;alert(&#39;label&#39;)&lt;/script&gt;");
      expect(tooltipHtml).toContain("&lt;svg onload=&quot;");
      expect(tooltipHtml).toContain("Quota: <b>100</b>");
      expect(tooltipHtml).not.toContain("<img");
      expect(tooltipHtml).not.toContain("<script");
      expect(tooltipHtml).not.toContain("<svg");
      expect(tooltipHtml).not.toContain('onerror="');
    });

    it("does not include tooltip summary when not provided", async () => {
      const wrapper = await mountComponent({ tooltipSummary: "" });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).not.toContain("margin-top: 4px");
    });
  });

  describe("Data Label Formatting", () => {
    it("formats data label with percentage when showPercentage is true", async () => {
      const wrapper = await mountComponent({ showPercentage: true });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      const formatter = series[0].label?.formatter as () => string;
      const result = formatter();
      expect(result).toBe("30 (30%)");
    });

    it("formats data label without percentage when showPercentage is false", async () => {
      const wrapper = await mountComponent({ showPercentage: false });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      const formatter = series[0].label?.formatter as () => string;
      const result = formatter();
      expect(result).toBe("30");
    });

    it("hides label for zero value (returns empty string)", async () => {
      const zeroSegments: IStackedBarSegment[] = [{ value: 0, color: "#FF0000", label: "Zero" }];
      const wrapper = await mountComponent({ segments: zeroSegments, showPercentage: true });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      const formatter = series[0].label?.formatter as () => string;
      const result = formatter();
      expect(result).toBe("");
    });
  });

  describe("Reactivity", () => {
    it("updates options when segments change", async () => {
      const wrapper = await mountComponent();
      const initialOptions = getChartOptions(wrapper);
      expect((initialOptions.series as BarSeriesOption[])[0].data).toEqual([30]);

      await wrapper.setProps({
        segments: [{ value: 100, color: "#000000", label: "New" }],
      });

      const updatedOptions = getChartOptions(wrapper);
      expect(updatedOptions.series as BarSeriesOption[]).toHaveLength(1);
      expect((updatedOptions.series as BarSeriesOption[])[0].data).toEqual([100]);
    });

    it("updates options when showLabels changes", async () => {
      const wrapper = await mountComponent({ showLabels: true });
      const initialOptions = getChartOptions(wrapper);
      expect((initialOptions.series as BarSeriesOption[])[0].label?.show).toBe(true);

      await wrapper.setProps({ showLabels: false });

      const updatedOptions = getChartOptions(wrapper);
      expect((updatedOptions.series as BarSeriesOption[])[0].label?.show).toBe(false);
    });

    it("updates options when maxValue changes", async () => {
      const wrapper = await mountComponent({ maxValue: 100 });
      const initialOptions = getChartOptions(wrapper);
      const xAxis = initialOptions.xAxis as { max: number };
      expect(xAxis.max).toBe(100);

      await wrapper.setProps({ maxValue: 200 });

      const updatedOptions = getChartOptions(wrapper);
      const updatedXAxis = updatedOptions.xAxis as { max: number };
      expect(updatedXAxis.max).toBe(200);
    });
  });

  describe("Edge Cases", () => {
    it("handles empty segments array", async () => {
      const wrapper = await mountComponent({ segments: [] });

      expect(wrapper.find(".stacked-bar-chart").exists()).toBe(true);

      const options = getChartOptions(wrapper);
      expect(options.series).toHaveLength(0);
    });

    it("handles single segment", async () => {
      const singleSegment: IStackedBarSegment[] = [{ value: 100, color: "#FF0000", label: "Single" }];
      const wrapper = await mountComponent({ segments: singleSegment });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      expect(series).toHaveLength(1);
      expect(series[0].data).toEqual([100]);
    });

    it("handles segments without labels", async () => {
      const noLabelSegments: IStackedBarSegment[] = [
        { value: 50, color: "#FF0000" },
        { value: 50, color: "#00FF00" },
      ];
      const wrapper = await mountComponent({ segments: noLabelSegments });
      const options = getChartOptions(wrapper);
      const series = options.series as BarSeriesOption[];

      expect(series[0].name).toBe("");
      expect(series[1].name).toBe("");
    });

    it("calculates correct percentages for non-100 totals", async () => {
      const customSegments: IStackedBarSegment[] = [
        { value: 25, color: "#FF0000", label: "Quarter" },
        { value: 75, color: "#00FF00", label: "ThreeQuarters" },
      ];
      const wrapper = await mountComponent({ segments: customSegments, showPercentage: true });
      const options = getChartOptions(wrapper);
      const formatter = options.tooltip && "formatter" in options.tooltip ? options.tooltip.formatter : undefined;

      const tooltipHtml = (formatter as () => string)();
      expect(tooltipHtml).toContain("25 (25%)");
      expect(tooltipHtml).toContain("75 (75%)");
    });
  });
});
