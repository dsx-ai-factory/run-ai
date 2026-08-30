import { mount } from "@vue/test-utils";
import { describe, expect, test, vi } from "vitest";

import RunaiResourcesChart from "./runai-resources-chart.vue";

describe("runai-resources-chart", () => {
  test("positions its tooltip outside the chart on the available side", () => {
    const wrapper = mount(RunaiResourcesChart, {
      props: {
        overQuotaValue: 2,
        quotaValue: 10,
        allocatedValue: 12,
      },
      global: {
        stubs: { VChart: { template: '<div class="chart-stub" />' } },
      },
    });
    const chartElement = wrapper.get('[aid="runai-resources-chart"]').element as HTMLElement;
    vi.spyOn(chartElement, "getBoundingClientRect").mockReturnValue({
      left: 100,
      right: 324,
      top: 100,
      bottom: 205,
      width: 224,
      height: 105,
      x: 100,
      y: 100,
      toJSON: () => ({}),
    });
    vi.stubGlobal("innerWidth", 1200);
    vi.stubGlobal("innerHeight", 900);
    const tooltip = wrapper.vm.echartsOptions.tooltip as {
      position: (
        point: [number, number],
        params: unknown,
        dom: HTMLElement,
        rect: unknown,
        size: { contentSize: [number, number]; viewSize: [number, number] },
      ) => [number, number];
    };

    expect(
      tooltip.position([0, 0], [], document.createElement("div"), null, {
        contentSize: [80, 48],
        viewSize: [224, 105],
      }),
    ).toEqual([157, 28.5]);
  });
});
