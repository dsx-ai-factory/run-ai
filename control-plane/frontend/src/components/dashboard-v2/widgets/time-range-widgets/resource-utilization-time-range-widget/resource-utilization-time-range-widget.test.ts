import { beforeEach, afterEach, describe, expect, test, vi, it } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import ResourceUtilizationTimeRangeWidget from "./resource-utilization-time-range-widget.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { mockClusterMetricsValues } from "./resource-utilization-time-range-widget.mock";

// models
import type { IRangeDates } from "@/utils/date.util";
import type { MetricsResponse, GroupedMeasurementResponseValuesInner } from "@/swagger-models/cluster-service-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

function getSeries(chartOptions: unknown): Record<string, unknown>[] {
  const series = (chartOptions as { series?: unknown } | null)?.series;
  return (Array.isArray(series) ? series : series ? [series] : []) as Record<string, unknown>[];
}

describe("resource-utilization-time-range-widget.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mountComponent(
    props: { clusterId: string; filterByDates: IRangeDates; confineTooltip?: boolean },
    mockResolvedValue?: MetricsResponse,
    mockRejectedValue?: string,
  ): VueWrapper<InstanceType<typeof ResourceUtilizationTimeRangeWidget>> {
    if (mockResolvedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValues);
    } else if (mockRejectedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockRejectedValue(mockRejectedValue);
    }

    return mount(ResourceUtilizationTimeRangeWidget, {
      global: {
        plugins: [Quasar],
      },
      props: {
        clusterId: props.clusterId,
        filterByDates: props.filterByDates,
        confineTooltip: props.confineTooltip,
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
    expect(wrapper.find(".chart-container").exists()).toBe(true);
  });

  it("renders chart wrapper with error state", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
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
    expect(wrapper.find(".error-section").exists()).toBe(true);
    expect(wrapper.find(".chart-container").exists()).toBe(false);
    errorSpy.mockRestore();
  });

  it("renders series correctly", async () => {
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: {
          dateStart: new Date("2024-03-19T09:39:00.346Z"),
          dateEnd: new Date("2024-03-19T12:39:44.658Z"),
          label: "",
        },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const numberOfSamples = wrapper.vm.timeframes.length;
    const series = getSeries(wrapper.vm.chartOptions);

    const tooltipAnchor = series[0];
    expect(tooltipAnchor.id).toBe("connected-tooltip-anchor");
    expect((tooltipAnchor.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUComputeUtilization = series[1];
    if (!gPUComputeUtilization) {
      return;
    }
    expect(gPUComputeUtilization.name).toBe("GPU compute utilization");
    expect(gPUComputeUtilization.type).toBe("line");
    expect((gPUComputeUtilization.data as unknown[])?.length).toBe(numberOfSamples);

    const gPUMemoryUtilization = series[2];
    if (!gPUMemoryUtilization) {
      return;
    }
    expect(gPUMemoryUtilization.name).toBe("GPU memory utilization");
    expect(gPUMemoryUtilization.type).toBe("line");
    expect((gPUMemoryUtilization.data as unknown[])?.length).toBe(numberOfSamples);

    const cPUComputeUtilization = series[3];
    if (!cPUComputeUtilization) {
      return;
    }
    expect(cPUComputeUtilization.name).toBe("CPU compute utilization");
    expect(cPUComputeUtilization.type).toBe("line");
    expect((cPUComputeUtilization.data as unknown[])?.length).toBe(numberOfSamples);

    const cPUMemoryUtilization = series[4];
    if (!cPUMemoryUtilization) {
      return;
    }
    expect(cPUMemoryUtilization.name).toBe("CPU memory utilization");
    expect(cPUMemoryUtilization.type).toBe("line");
    expect((cPUMemoryUtilization.data as unknown[])?.length).toBe(numberOfSamples);
  });

  it("confines the tooltip when rendered inside a scrolling drawer", () => {
    const dateEnd = new Date();
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart: new Date(dateEnd.getTime() - 60 * 60 * 1000), dateEnd, label: "" },
        confineTooltip: true,
      },
      mockClusterMetricsValues,
    );

    const tooltip = wrapper.vm.chartOptions?.tooltip as { appendToBody?: boolean; confine?: boolean };
    expect(tooltip.appendToBody).toBe(false);
    expect(tooltip.confine).toBe(true);
  });

  it("returns 0 when input array is empty", async () => {
    const preDefindDates = "Custom";
    const dateStart = new Date(2023, 2, 3, 12, 30, 45, 678); // March 3rd, 2023, 12:30:45.678
    const dateEnd = new Date(2024, 2, 3, 12, 30, 45, 678); // March 3rd, 2024, 12:30:45.678
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: preDefindDates },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const numbers: Array<number> = [];
    const average = wrapper.vm.calculateAverage(numbers);
    expect(average).toBe(0);
  });

  it("calculates the average of positive numbers", async () => {
    const preDefindDates = "Custom";
    const dateStart = new Date(2023, 2, 3, 12, 30, 45, 678); // March 3rd, 2023, 12:30:45.678
    const dateEnd = new Date(2024, 2, 3, 12, 30, 45, 678); // March 3rd, 2024, 12:30:45.678
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: preDefindDates },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const numbers: Array<number> = [1, 2, 3, 4, 5];
    const average = wrapper.vm.calculateAverage(numbers);
    expect(average).toBe(3);
  });

  it("return average value with percentage with only 2 numbers after the dot", async () => {
    vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValues);
    const preDefindDates = "Custom";
    const dateStart = new Date(2023, 2, 3, 12, 30, 45, 678); // March 3rd, 2023, 12:30:45.678
    const dateEnd = new Date(2024, 2, 3, 12, 30, 45, 678); // March 3rd, 2024, 12:30:45.678
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: preDefindDates },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const average = wrapper.vm.getAverageMetric("GPU_MEMORY_UTILIZATION");
    expect(average.includes("%")).toBe(true);
    expect(average.replace("%", "").split(".")[1].length).toBe(2);
  });

  test("should return average metric after filtering by min and max", async () => {
    vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValues);

    const preDefindDates = "Custom";
    const dateStart = new Date(2023, 2, 3, 12, 30, 45, 678); // March 3rd, 2023, 12:30:45.678
    const dateEnd = new Date(2024, 2, 3, 12, 30, 45, 678); // March 3rd, 2024, 12:30:45.678
    const wrapper = mountComponent(
      {
        clusterId: "someClusterId",
        filterByDates: { dateStart, dateEnd, label: preDefindDates },
      },
      mockClusterMetricsValues,
    );

    await flushPromises();

    const fullAverage = wrapper.vm.getAverageMetric("GPU_UTILIZATION");
    expect(fullAverage).toBe("55.22%");

    const min = new Date("2024-03-01T00:05:30.000Z").getTime();
    const max = new Date("2024-03-01T00:06:30.000Z").getTime();
    const gpuUtilization = mockClusterMetricsValues.measurements[0];
    if (!gpuUtilization.values) {
      return;
    }
    const measurements: Array<GroupedMeasurementResponseValuesInner> = gpuUtilization.values.filter(
      (measurement: GroupedMeasurementResponseValuesInner) => {
        if (!measurement.timestamp) return false;
        const currentTime = new Date(measurement.timestamp).getTime();
        return currentTime >= min && currentTime <= max;
      },
    );

    const sum = measurements.reduce((sum: number, measurement: GroupedMeasurementResponseValuesInner) => {
      sum += Number(measurement.value);
      return sum;
    }, 0);
    const avg = `${(sum / measurements.length).toFixed(2)}%`;
    const filteredAverage = wrapper.vm.getAverageMetric("GPU_UTILIZATION", min, max);
    expect(filteredAverage).toBe(avg);
  });
  it("should refresh data at the specified interval", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.useFakeTimers();

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
    const loadChartDataSpy = vi.spyOn(wrapper.vm, "loadChartData");
    const getEndDateSpy = vi.spyOn(wrapper.vm, "getEndDate");

    vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockClusterMetricsValues);

    await flushPromises();
    expect(clusterService.getClusterMetrics).toHaveBeenCalledTimes(1);

    vi.advanceTimersByTime(60000); // 1 minute

    await flushPromises();
    expect(clusterService.getClusterMetrics).toHaveBeenCalledTimes(2);
    expect(loadChartDataSpy).toHaveBeenCalledWith(true);
    expect(getEndDateSpy).toHaveBeenCalled();
    vi.useRealTimers();
    errorSpy.mockRestore();
  });

  it("should render the avg utilization summary section correctly", async () => {
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

    const avgUtilizationSection = wrapper.find(".avg-utilization-summary");
    expect(avgUtilizationSection.exists()).toBe(true);

    const avgUtilizationItems = avgUtilizationSection.findAll(".avg-item");
    expect(avgUtilizationItems.length).toBe(4);
    wrapper.vm.averageMetrics.forEach((metric: { name: string; value: string }, index: number) => {
      const avgUtilizationItem = avgUtilizationItems[index];
      expect(avgUtilizationItem.find(".avg-title").text()).toBe(metric.name);
      expect(avgUtilizationItem.find(".avg-value").text()).toBe(metric.value);
    });
  });
});
