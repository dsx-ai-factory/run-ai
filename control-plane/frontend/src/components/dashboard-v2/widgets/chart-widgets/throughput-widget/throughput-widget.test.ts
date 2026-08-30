import { vi, describe, afterEach, it, expect } from "vitest";
import ThroughputWidget from "./throughput-widget.vue";

import { flushPromises, shallowMount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

// utils
import type { IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";

// models
import { InferenceWorkloadMetricType, type MetricsResponse } from "@/swagger-models/workloads-client";
interface SeriesLineOptions {
  id?: string;
  name?: string;
  data?: unknown[];
}
function getSeries(chartOptions: unknown): SeriesLineOptions[] {
  const series = (chartOptions as { series?: unknown } | null)?.series;
  return (Array.isArray(series) ? series : series ? [series] : []) as SeriesLineOptions[];
}

vi.mock("@/services/control-plane/inference.service/inference.service");

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

interface IThroughputWidgetProps {
  workloadId: string;
  filterByDates: IRangeDates;
}

describe("throughput-widget.vue", () => {
  const workloadId = "1234";
  const filterByDates: IRangeDates = {
    dateStart: new Date("2023-01-01"),
    dateEnd: new Date("2023-12-31"),
    label: "Default Label",
  };

  function getMountedComponent(props: IThroughputWidgetProps): VueWrapper<InstanceType<typeof ThroughputWidget>> {
    return shallowMount(ThroughputWidget, {
      global: {
        plugins: [Quasar],
      },
      props: props,
    });
  }

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    vi.spyOn(inferenceService, "getMetrics");
    const stopIntervalSpy = vi.spyOn(intervalUtil, "stopInterval");
    const startIntervalSpy = vi.spyOn(intervalUtil, "startInterval");
    const wrapper = getMountedComponent({ workloadId, filterByDates });

    expect(wrapper.exists()).toBe(true);
    expect(inferenceService.getMetrics).toHaveBeenCalled();
    expect(intervalUtil.stopInterval).toHaveBeenCalled();
    expect(stopIntervalSpy).toHaveBeenCalledTimes(1);
    await flushPromises();
    expect(startIntervalSpy).toHaveBeenCalledTimes(1);
  });

  it("loads chart options on creation", () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    expect(wrapper.vm.chartOptions).not.toBeNull();
  });

  it("calls loadChartData method on prop changes", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });

    const loadChartDataSpy = vi.spyOn(wrapper.vm, "loadChartData");
    wrapper.setProps({
      workloadId: "new-workload-id",
    });
    await wrapper.vm.$nextTick();
    expect(loadChartDataSpy).toHaveBeenCalled();
  });

  it("emits register event after loading chart data", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    await wrapper.vm.loadChartData();
    await wrapper.vm.$nextTick();
    expect(wrapper.emitted().register).toBeTruthy();
  });

  it("sets error to true on loadChartData failure", async () => {
    vi.mocked(inferenceService.getMetrics).mockRejectedValue(new Error("API error"));
    const wrapper = getMountedComponent({ workloadId, filterByDates });

    const errorLog = console.error;
    console.error = vi.fn(); // I want to prevent the console.error from being printed
    await wrapper.vm.loadChartData();
    expect(wrapper.vm.error).toBe(true);
    console.error = errorLog;
  });

  it("correctly adds series details", () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    const series = getSeries(wrapper.vm.chartOptions);
    expect(series).toHaveLength(2);
    expect((series[0] as SeriesLineOptions)?.id).toBe("connected-tooltip-anchor");
    const seriesData: SeriesLineOptions | undefined = series[1] as SeriesLineOptions;
    expect(seriesData?.name).toBe("Total throughput");
  });

  it("correctly handles beforeUnmount lifecycle hook", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });

    intervalUtil.stopInterval(EIntervalLabels.ThroughputWidget);
    const wrapper = getMountedComponent({ workloadId, filterByDates });

    const stopIntervalSpy = vi.spyOn(intervalUtil, "stopInterval");
    wrapper.unmount();
    await wrapper.vm.$nextTick();
    expect(stopIntervalSpy).toHaveBeenCalledWith(EIntervalLabels.ThroughputWidget);
    expect(wrapper.emitted().unregister).toBeTruthy();
  });

  it("should add series details if chartOptions is initialized", () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });

    // Set chartOptions to null to simulate uninitialized state
    wrapper.vm.chartOptions = {
      ...wrapper.vm.chartOptions,
      series: undefined,
    };

    expect(wrapper.vm.chartOptions?.series).toBeUndefined();

    // Call the method
    wrapper.vm.addSeriesDetails();

    expect(wrapper.vm.chartOptions?.series).not.toBeUndefined();
  });

  it("should not add series details if chartOptions is not initialized", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    await flushPromises();
    // Set chartOptions to null to simulate uninitialized state
    wrapper.vm.chartOptions = null;

    expect(wrapper.vm.chartOptions).toBeNull();

    // Call the method
    wrapper.vm.addSeriesDetails();

    expect(wrapper.vm.chartOptions).toBeNull();
  });

  it("should not set series if chartOptions or chartOptions.series is not defined", async () => {
    vi.mocked(inferenceService.getMetrics).mockResolvedValue({ measurements: [] });
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    await flushPromises();

    // Set chartOptions to null and call the method
    wrapper.vm.chartOptions = null;
    wrapper.vm.setSeries();
    // Assert that method returns early
    expect(wrapper.vm.chartOptions).toBeNull();

    // Set chartOptions.series to null and call the method
    wrapper.vm.chartOptions = { series: undefined };
    wrapper.vm.setSeries();
    // Assert that method returns early
    expect(wrapper.vm.chartOptions.series).toBeUndefined();
  });

  it("should set the data property of the first series in chartOptions.series", async () => {
    // Mock the getMeasurements method
    const mockMeasurements: MetricsResponse = {
      measurements: [
        {
          type: InferenceWorkloadMetricType.Throughput,
          values: [{ timestamp: "2023-01-01T00:00:00Z", value: "0" }],
        },
      ],
    };

    vi.mocked(inferenceService.getMetrics).mockResolvedValue(mockMeasurements);
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    wrapper.vm.timeframes = ["2023-01-01", "2023-01-02"];

    await flushPromises();

    expect(wrapper.vm.measurementMap).toEqual({
      [InferenceWorkloadMetricType.Throughput]: mockMeasurements.measurements[0].values,
    });

    const series = getSeries(wrapper.vm.chartOptions);
    const seriesData: SeriesLineOptions | undefined = series[1] as SeriesLineOptions;
    expect(seriesData?.data?.length).toBe(2);
  });

  it("should not set the data of the first series if measurement type does not exist in chartOptions.series", async () => {
    // Mock the getMeasurements method
    const mockMeasurements: MetricsResponse = {
      measurements: [
        {
          type: InferenceWorkloadMetricType.Latency, // to simulate a different metric type
          values: [{ timestamp: "2023-01-01T00:00:00Z", value: "0" }],
        },
      ],
    };

    vi.mocked(inferenceService.getMetrics).mockResolvedValue(mockMeasurements);
    const wrapper = getMountedComponent({ workloadId, filterByDates });
    wrapper.vm.timeframes = ["2023-01-01", "2023-01-02"];

    await flushPromises();

    expect(wrapper.vm.measurementMap).toEqual({
      // measurementMap should not contains Throughput metric type
      [InferenceWorkloadMetricType.Latency]: mockMeasurements.measurements[0].values,
    });

    // Assert that data is set correctly
    const series = getSeries(wrapper.vm.chartOptions);
    const seriesData: SeriesLineOptions | undefined = series[1] as SeriesLineOptions;
    expect(seriesData?.data?.length).toBe(0);
  });
});
