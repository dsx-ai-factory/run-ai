/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineComponent, h, nextTick, ref, type Ref } from "vue";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useAdaptiveCategoricalChartLayout } from "./use-adaptive-categorical-chart-layout.composable";

class ResizeObserverMock {
  observe = vi.fn();
  disconnect = vi.fn();
}

describe("useAdaptiveCategoricalChartLayout", () => {
  const originalRequestAnimationFrame = globalThis.requestAnimationFrame;
  const originalCancelAnimationFrame = globalThis.cancelAnimationFrame;
  const originalResizeObserver = globalThis.ResizeObserver;

  afterEach(() => {
    vi.restoreAllMocks();
    vi.stubGlobal("requestAnimationFrame", originalRequestAnimationFrame);
    vi.stubGlobal("cancelAnimationFrame", originalCancelAnimationFrame);
    vi.stubGlobal("ResizeObserver", originalResizeObserver);
  });

  it("applies responsive axis-label and grouped-bar geometry after mounting", async () => {
    vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
      callback(0);
      return 1;
    });
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.stubGlobal("ResizeObserver", ResizeObserverMock);

    const setOption = vi.fn();
    const chart = {
      getWidth: () => 500,
      getHeight: () => 240,
      getOption: () => ({ xAxis: [{ data: ["one", "two"] }], series: [{ id: "running" }, { id: "pending" }] }),
      setOption,
    };
    const chartRef = ref({ chart });
    let getCategoryAxisLabelOptions: ((categoryCount: number) => { width: number }) | undefined;

    const wrapper = mount(
      defineComponent({
        setup() {
          const containerRef = ref<HTMLElement | null>(null);
          ({ getCategoryAxisLabelOptions } = useAdaptiveCategoricalChartLayout({
            chartRef,
            containerRef,
            horizontalGridPadding: 68,
          }));
          return () => h("div", { ref: containerRef });
        },
      }),
    );
    await nextTick();

    expect(setOption).toHaveBeenCalledTimes(1);
    expect(setOption.mock.calls[0][0]).not.toHaveProperty("grid");
    expect(setOption.mock.calls[0][0]).not.toHaveProperty("legend");
    expect(setOption.mock.calls[0][0].series).toEqual([
      { id: "running", barWidth: 30, barMaxWidth: 30, barGap: "116%", barCategoryGap: "40%" },
      { id: "pending", barWidth: 30, barMaxWidth: 30, barGap: "116%", barCategoryGap: "40%" },
    ]);
    expect(getCategoryAxisLabelOptions?.(2).width).toBe(208);

    wrapper.unmount();
  });

  it("sizes category labels from their available plot slots and preserves chart-owned label options", async () => {
    vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
      callback(0);
      return 1;
    });
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.stubGlobal("ResizeObserver", ResizeObserverMock);

    const setOption = vi.fn();
    const chart = {
      getWidth: () => 500,
      getHeight: () => 240,
      getOption: () => ({ xAxis: [{ data: ["short", "a-project-with-an-extraordinarily-long-name"] }] }),
      setOption,
    };
    const chartRef = ref({ chart });
    let layout: Ref<{ externalCategoryLabelBottom: number }> | undefined;

    const wrapper = mount(
      defineComponent({
        setup() {
          const containerRef = ref<HTMLElement | null>(null);
          ({ layout } = useAdaptiveCategoricalChartLayout({
            chartRef,
            containerRef,
            horizontalGridPadding: 68,
          }));
          return () => h("div", { ref: containerRef });
        },
      }),
    );
    await nextTick();

    expect(layout?.value.externalCategoryLabelBottom).toBe(48);
    expect(setOption).toHaveBeenCalledWith(
      expect.objectContaining({
        xAxis: {
          axisLabel: expect.objectContaining({
            width: 208,
            overflow: "truncate",
            ellipsis: "…",
          }),
        },
      }),
      { lazyUpdate: true },
    );

    const axisLabel = setOption.mock.calls[0][0].xAxis.axisLabel;
    expect(axisLabel).not.toEqual(
      expect.objectContaining({
        formatter: expect.anything(),
        hideOverlap: expect.anything(),
        rotate: expect.anything(),
      }),
    );

    wrapper.unmount();
  });

  it("uses Highcharts-like group and point padding for wide and crowded grouped bars", () => {
    const chart = {
      getWidth: () => 500,
      getOption: () => ({ xAxis: [{ data: [] }], series: [] }),
      setOption: vi.fn(),
    };
    const chartRef = ref({ chart });
    const containerRef = ref<HTMLElement | null>(null);
    const { getGroupedBarLayout } = useAdaptiveCategoricalChartLayout({
      chartRef,
      containerRef,
      horizontalGridPadding: 68,
    });

    expect(getGroupedBarLayout(1, 2)).toEqual({
      barWidth: 30,
      barMaxWidth: 30,
      barGap: "332%",
      barCategoryGap: "40%",
    });
    expect(getGroupedBarLayout(10, 3)).toEqual({
      barWidth: 6.91,
      barMaxWidth: 30,
      barGap: "25%",
      barCategoryGap: "40%",
    });
  });

  it("lets ECharts perform the first layout when dimensions are not measurable yet", () => {
    const chartRef = ref(null);
    const containerRef = ref<HTMLElement | null>(null);
    const { getGroupedBarLayout } = useAdaptiveCategoricalChartLayout({
      chartRef,
      containerRef,
      horizontalGridPadding: 68,
    });

    expect(getGroupedBarLayout(1, 2)).toEqual({
      barMaxWidth: 30,
      barGap: "25%",
      barCategoryGap: "40%",
    });
  });
});
