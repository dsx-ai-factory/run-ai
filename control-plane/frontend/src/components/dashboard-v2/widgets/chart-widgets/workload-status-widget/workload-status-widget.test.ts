/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { WorkloadStatusWidget } from "@/components/dashboard-v2/widgets/chart-widgets/workload-status-widget/index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import {
  mockDataWithMissingPhases,
  WorkloadStatusWidgetMock,
  WorkloadStatusWidgetMockFilterByNodepool,
} from "@/components/dashboard-v2/widgets/chart-widgets/workload-status-widget/workload-status-widget.mock";
import { Phase, type TelemetryResponse, WorkloadTelemetryType } from "@/swagger-models/workloads-client";
import { EWorkloadTelemetryGroupBy, type EWorkloadTelemetryPointMeta } from "@/models/workload.model";
import type { EChartsOption, BarSeriesOption } from "echarts";

interface SeriesOptionsType {
  name?: string;
  data?: unknown[];
  color?: string;
}

interface WorkloadStatusWidgetInstance {
  convertSeriesToECharts: (
    series: SeriesOptionsType[],
    isDrilldown?: boolean,
  ) => { categories: string[]; series: BarSeriesOption[] };
  buildChartOption: (categories: string[], series: BarSeriesOption[], isDrilldown?: boolean) => EChartsOption;
  drillDownTo: (meta: EWorkloadTelemetryPointMeta) => Promise<void>;
  getDrillDownSeries: (meta: EWorkloadTelemetryPointMeta) => Promise<SeriesOptionsType[]>;
  sortSeriesByPhase: (series: SeriesOptionsType[], isDrilldown: boolean) => SeriesOptionsType[];
  chartOption: EChartsOption | null;
  drillBreadcrumbs: string[];
  drillStack: unknown[];
}

vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof WorkloadStatusWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "34234-322-sdfs323";

describe("WorkloadStatusWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }) => {
    wrapper = mount(WorkloadStatusWidget, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        ...props,
        clusterId,
        hasDepartments: true,
      },
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    vi.clearAllMocks();
  });

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  it("should render WorkloadStatusWidget component", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    await flushPromises();
    expect(wrapper.find('[aid="workloads-by-status"]').text()).eql("Workloads by status");
  });

  it("should render WorkloadStatusWidget component with filter", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(WorkloadStatusWidgetMockFilterByNodepool),
    );
    createComponent({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
      WorkloadTelemetryType.WorkloadsCount,
      clusterId,
      "node-pool-2",
      undefined,
      undefined,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.Phase],
    );
    expect(wrapper.find(".widget-link-btn").text()).eql("All workloads");
  });

  it("should link to projects table when clicked", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "workload-index" });
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
    await flushPromises();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsTelemetry)
      .mockRejectedValueOnce({ response: { status: 500 } })
      .mockResolvedValueOnce(WorkloadStatusWidgetMock);

    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(true);
    expect(wrapper.find(".error-section").text()).eql("Data is not available right now");

    // validate that error message is not displayed after successful fetch
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(false);
    errorSpy.mockRestore();
  });

  it("should include zero-value phases when some phases are missing", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(mockDataWithMissingPhases as TelemetryResponse),
    );
    createComponent();
    await flushPromises();

    expect(wrapper.vm.isChartEmpty).toBe(false);
    expect(wrapper.vm.error).toBe(false);
    const series = wrapper.vm.chartOption?.series as Array<{ name: string; data: Array<{ value: number }> }> | undefined;
    expect(series?.[0].name).eql(Phase.Failed);
    expect(series?.[1].name).eql(Phase.Pending);
    expect(series?.[2].name).eql(Phase.Running);

    expect(series?.[0]?.data.length).eql(1);
    expect(series?.[0]?.data[0].value).eql(0);

    expect(series?.[1]?.data.length).eql(1);
    expect(series?.[1]?.data[0].value).eql(0);

    expect(series?.[2]?.data.length).eql(1);
    expect(series?.[2]?.data[0].value).eql(5);
  });

  it("should keep breadcrumbs hidden until drilldown data is ready", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadStatusWidgetInstance;
    vm.chartOption = { series: [] };

    let resolveDrilldown: (series: SeriesOptionsType[]) => void = () => undefined;
    vi.spyOn(vm, "getDrillDownSeries").mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveDrilldown = resolve;
        }),
    );

    const meta = { entityId: clusterId, groupByKey: EWorkloadTelemetryGroupBy.ClusterId };
    const drilldownPromise = vm.drillDownTo(meta);

    expect(vm.drillBreadcrumbs).toEqual([]);
    expect(vm.drillStack).toHaveLength(0);

    resolveDrilldown([{ name: Phase.Running, data: [{ name: "project-a", y: 2 }] }]);
    await drilldownPromise;

    expect(vm.drillBreadcrumbs).toHaveLength(2);
    expect(vm.drillBreadcrumbs[1]).toBe("Departments");
    expect(vm.drillStack).toHaveLength(1);
  });

  it("should preserve prod phase order and use full-size drilldown geometry", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadStatusWidgetMock));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadStatusWidgetInstance;
    const sortedDrilldown = vm.sortSeriesByPhase(
      [
        { name: Phase.Pending, data: [] },
        { name: Phase.Failed, data: [] },
        { name: Phase.Running, data: [] },
      ],
      true,
    );

    expect(sortedDrilldown.map((series) => series.name)).toEqual([Phase.Running, Phase.Failed, Phase.Pending]);

    const intermediate = vm.convertSeriesToECharts(
      [
        {
          name: Phase.Running,
          data: [
            {
              name: "department-a",
              y: 2,
              drilldown: true,
              meta: { entityId: "department-a", groupByKey: EWorkloadTelemetryGroupBy.DepartmentId },
            },
          ],
        },
      ],
      true,
    );
    const intermediateOption = vm.buildChartOption(intermediate.categories, intermediate.series, true);

    expect(intermediateOption.grid).toMatchObject({ top: 52, bottom: 86, left: 56, right: 12 });
    expect(intermediateOption.tooltip).toMatchObject({
      trigger: "axis",
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect(intermediateOption.xAxis).toMatchObject({
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect(intermediateOption.yAxis).toMatchObject({
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect((intermediateOption.series as BarSeriesOption[])[0]).toMatchObject({
      barMaxWidth: 30,
      barGap: "25%",
      barCategoryGap: "40%",
      legendHoverLink: true,
      emphasis: { focus: "series" },
      blur: {
        itemStyle: { opacity: 0.2 },
        label: { opacity: 0.2 },
      },
    });
    expect(intermediateOption.xAxis).toMatchObject({ axisLabel: { show: false } });

    const finalLevel = vm.convertSeriesToECharts([{ name: Phase.Running, data: [{ name: "project-a", y: 2 }] }], true);
    const finalOption = vm.buildChartOption(finalLevel.categories, finalLevel.series, true);

    expect(finalOption.xAxis).toMatchObject({
      data: ["project-a"],
      axisLabel: { show: true },
    });

    const longProjectFinalLevel = vm.convertSeriesToECharts(
      [
        {
          name: Phase.Running,
          data: [
            { name: "project-a", y: 2 },
            { name: "chartseed-20260512-vision", y: 1 },
            { name: "chartseed-20260512-burst", y: 1 },
            { name: "chartseed-20260512-platform", y: 1 },
          ],
        },
      ],
      true,
    );
    const longProjectFinalOption = vm.buildChartOption(
      longProjectFinalLevel.categories,
      longProjectFinalLevel.series,
      true,
    );

    expect(longProjectFinalOption.grid).toMatchObject({ top: 52, bottom: 86, left: 56, right: 12 });
    expect(longProjectFinalOption.xAxis).toMatchObject({
      axisLabel: {
        rotate: 0,
        overflow: "truncate",
        ellipsis: "…",
      },
    });
  });
});
