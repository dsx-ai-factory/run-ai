import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { beforeEach, afterEach, vi, describe, it, expect } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import ResourceAllocationTimeRangeWidget from "./resource-allocation-time-range-widget.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { mockClusterMetricsValues } from "./resource-allocation-time-range-widget.mock";

// models
import type { MetricsResponse } from "@/swagger-models/cluster-service-client";
import type { IRangeDates } from "@/utils/date.util";
import { useClusterStore } from "@/stores/cluster.store";
import type { TMeasurementsTimeAndValue } from "@/models/chart.model";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

function getSeries(chartOptions: unknown): Record<string, unknown>[] {
  const series = (chartOptions as { series?: unknown } | null)?.series;
  return (Array.isArray(series) ? series : series ? [series] : []) as Record<string, unknown>[];
}

const pinia = createTestingPinia({ createSpy: vi.fn });

setActivePinia(pinia);

describe("resource-allocation-time-range-widget.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mountComponent(
    props: { clusterId: string; filterByDates: IRangeDates },
    mockResolvedValue?: MetricsResponse,
    mockRejectedValue?: string,
  ): VueWrapper<InstanceType<typeof ResourceAllocationTimeRangeWidget>> {
    if (mockResolvedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValues);
    } else if (mockRejectedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockRejectedValue(mockRejectedValue);
    }

    useClusterStore().isClusterVersionSufficient = () => true;

    return mount(ResourceAllocationTimeRangeWidget, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        clusterId: props.clusterId,
        filterByDates: props.filterByDates,
      },
    });
  }

  it("renders chart wrapper with loading state", async () => {
    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );

    expect(wrapper.find(".chart-widget-wrapper").exists()).toBe(true);

    await flushPromises();

    expect(clusterService.getClusterMetrics).toHaveBeenCalled();
    expect(wrapper.find(".error-section").exists()).toBe(false);
    expect(wrapper.find(".resource-allocation-time-range-highchart").exists()).toBe(true);
  });

  it("does not consume page scrolling while area zoom is inactive", () => {
    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000);
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );
    const dataZoom = Array.isArray(wrapper.vm.chartOptions?.dataZoom)
      ? wrapper.vm.chartOptions.dataZoom[0]
      : wrapper.vm.chartOptions?.dataZoom;

    expect(dataZoom).toMatchObject({
      type: "inside",
      disabled: true,
      zoomOnMouseWheel: false,
      moveOnMouseWheel: false,
      moveOnMouseMove: false,
      preventDefaultMouseMove: false,
    });
  });

  it("renders chart wrapper with error state", async () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      undefined,
      "failed to load data",
    );

    expect(wrapper.find(".chart-widget-wrapper").exists()).toBe(true);

    await flushPromises();

    expect(clusterService.getClusterMetrics).toHaveBeenCalled();
    expect(consoleError).toHaveBeenCalledWith("failed to load data");
    expect(wrapper.find(".error-section").exists()).toBe(true);
    expect(wrapper.find(".resource-allocation-highchart").exists()).toBe(false);
  });

  it("renders series correctly", async () => {
    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const numberOfSamples = wrapper.vm.timeframes.length;
    const series = getSeries(wrapper.vm.chartOptions);

    const tooltipAnchor = series[0];
    expect(tooltipAnchor.id).toBe("connected-tooltip-anchor");
    expect((tooltipAnchor.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUComputeTotal = series[1];
    if (!gPUComputeTotal) {
      return;
    }
    expect(gPUComputeTotal.name).toBe("GPU devices (total)");
    expect(gPUComputeTotal.type).toBe("line");
    expect((gPUComputeTotal.lineStyle as Record<string, unknown>)?.type).toEqual([4, 4]);
    expect((gPUComputeTotal.lineStyle as Record<string, unknown>)?.color).toBe("#2CAFFE");
    expect((gPUComputeTotal.itemStyle as Record<string, unknown>)?.color).toBe("#2CAFFE");
    expect((gPUComputeTotal.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUComputeUnallocated = series[2];
    if (!gPUComputeUnallocated) {
      return;
    }
    expect(gPUComputeUnallocated.name).toBe("GPU devices (unallocated)");
    expect(gPUComputeUnallocated.type).toBe("line");
    expect((gPUComputeUnallocated.lineStyle as Record<string, unknown>)?.color).toBe("#FE6A35");
    expect((gPUComputeUnallocated.itemStyle as Record<string, unknown>)?.color).toBe("#FE6A35");
    expect((gPUComputeUnallocated.areaStyle as Record<string, unknown>)?.color).toBe("#FE6A35");
    expect((gPUComputeUnallocated.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUComputeAllocated = series[3];
    if (!gPUComputeAllocated) {
      return;
    }
    expect(gPUComputeAllocated.name).toBe("GPU devices (allocated)");
    expect(gPUComputeAllocated.type).toBe("line");
    expect((gPUComputeAllocated.lineStyle as Record<string, unknown>)?.color).toBe("#00E272");
    expect((gPUComputeAllocated.itemStyle as Record<string, unknown>)?.color).toBe("#00E272");
    expect((gPUComputeAllocated.areaStyle as Record<string, unknown>)?.color).toBe("#00E272");
    expect((gPUComputeAllocated.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUQuota = series[4];
    if (!gPUQuota) {
      return;
    }
    expect(gPUQuota.name).toBe("GPU quota (All departments)");
    expect(gPUQuota.type).toBe("line");
    expect((gPUQuota.lineStyle as Record<string, unknown>)?.type).toEqual([4, 4]);
    expect((gPUQuota.lineStyle as Record<string, unknown>)?.color).toBe("#544FC5");
    expect((gPUQuota.itemStyle as Record<string, unknown>)?.color).toBe("#544FC5");
    expect((gPUQuota.data as unknown[])?.length).toBe(numberOfSamples);
  });

  it("calculates the difference between total and allocated GPU values", async () => {
    const totalGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 10],
      [1710881258622, 20],
      [1710881258622, 30],
    ];
    const allocatedGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 5],
      [1710881258622, 15],
      [1710881258622, 25],
    ];

    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const differenceArray = wrapper.vm.calculateUnallocatedValues(totalGPU, allocatedGPU);

    expect(differenceArray).toEqual([
      [1710881258622, 5],
      [1710881258622, 5],
      [1710881258622, 5],
    ]);
  });

  it("should not calculates the difference for a null value", async () => {
    const totalGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 10],
      [1710881258622, null],
      [1710881258622, 30],
    ];
    const allocatedGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 5],
      [1710881258622, 15],
      [1710881258622, 25],
    ];

    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const differenceArray = wrapper.vm.calculateUnallocatedValues(totalGPU, allocatedGPU);

    expect(differenceArray).toEqual([
      [1710881258622, 5],
      [1710881258622, null],
      [1710881258622, 5],
    ]);
  });

  it("throws an error if the input arrays have different lengths", async () => {
    const totalGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 10],
      [1710881258622, 20],
    ];
    const allocatedGPU: Array<TMeasurementsTimeAndValue> = [
      [1710881258622, 5],
      [1710881258622, 15],
      [1710881258622, 25],
    ];

    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 24 * 60 * 60 * 1000); // 24 hours ago
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: "" },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    expect(() => wrapper.vm.calculateUnallocatedValues(totalGPU, allocatedGPU)).toThrowError(
      "Arrays must have the same length.",
    );
  });
});
