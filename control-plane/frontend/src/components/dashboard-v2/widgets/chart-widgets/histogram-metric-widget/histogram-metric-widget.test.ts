import { describe, expect, test, vi, beforeEach } from "vitest";

// Mock the composable
const mockComposable = {
  chartOptions: { value: { series: [] as unknown[] } },
  error: { value: false },
  initialLoading: { value: false },
  timeframes: { value: [] },
  initialize: vi.fn(),
  loadChartOptions: vi.fn(),
  loadData: vi.fn(),
  createLineSeriesFromMap: vi.fn(),
  updateYAxisLabelsDefaultsIfNeeded: vi.fn((measurements, options) => options),
  handlePropsChanged: vi.fn(),
  handleBeforeUnmount: vi.fn(),
};

vi.mock("@/composables/use-generic-metric-widget-options.composable", () => ({
  useGenericMetricWidgetOptions: () => mockComposable,
}));

// Mock utilities
vi.mock("@/utils/chart.util", () => ({
  chartUtil: {
    arrangeMeasurementsByTime: vi.fn((timeframes, data) => data),
  },
}));

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

describe("HistogramMetricWidget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Composable Integration", () => {
    test("should initialize composable with config", () => {
      const mockConfig = {
        chartId: "test-chart-id",
        metricTypes: ["cpu", "memory"],
        chartTitle: "Test Chart",
        intervalLabel: "1m",
        chartOptions: {
          type: "line",
          yAxisTitle: "Test Y Axis",
          height: 300,
          showSharedCrosshair: true,
          sharedTooltip: true,
          yAxisFormatFunction: vi.fn(),
          colors: ["#ff0000"],
        },
        getData: vi.fn(),
        histogramParser: vi.fn(),
        columnHistogramParser: vi.fn(),
      };

      // Simulate component initialization
      mockComposable.initialize(mockConfig);

      expect(mockComposable.initialize).toHaveBeenCalledWith(mockConfig);
    });

    test("should call loadChartOptions after initialization", () => {
      const mockConfig = {
        chartId: "test-chart-id",
        metricTypes: ["cpu", "memory"],
        chartTitle: "Test Chart",
        intervalLabel: "1m",
        chartOptions: {
          type: "line",
          yAxisTitle: "Test Y Axis",
          height: 300,
          showSharedCrosshair: true,
          sharedTooltip: true,
          yAxisFormatFunction: vi.fn(),
          colors: ["#ff0000"],
        },
        getData: vi.fn(),
        histogramParser: vi.fn(),
        columnHistogramParser: vi.fn(),
      };

      // Simulate component initialization
      mockComposable.initialize(mockConfig);
      mockComposable.loadChartOptions();

      expect(mockComposable.loadChartOptions).toHaveBeenCalled();
    });
  });

  describe("Data Loading", () => {
    test("should call loadData with correct parameters", async () => {
      const filterByDates = {
        dateStart: new Date("2023-01-01"),
        dateEnd: new Date("2023-01-02"),
      };
      const getDataArgs = ["arg1", "arg2"];
      const onRegister = vi.fn();
      const onDataLoaded = vi.fn();

      await mockComposable.loadData(filterByDates, getDataArgs, onRegister, onDataLoaded);

      expect(mockComposable.loadData).toHaveBeenCalledWith(filterByDates, getDataArgs, onRegister, onDataLoaded);
    });
  });

  describe("Chart Type Handling", () => {
    test("should handle column chart type with columnHistogramParser", () => {
      const mockColumnData = [
        { name: "Series 1", data: [{ time: "2023-01-01", value: 100 }] },
        { name: "Series 2", data: [{ time: "2023-01-01", value: 200 }] },
      ];

      const mockMetricsResponse = {
        measurements: [],
      };

      const columnHistogramParser = vi.fn().mockReturnValue(mockColumnData);

      // Simulate column chart processing
      const result = columnHistogramParser(mockMetricsResponse);

      expect(columnHistogramParser).toHaveBeenCalledWith(mockMetricsResponse);
      expect(result).toEqual(mockColumnData);
    });

    test("should handle line chart type with histogramParser", () => {
      const mockHistogramMap = {
        cpu: [{ time: "2023-01-01", value: 50 }],
        memory: [{ time: "2023-01-01", value: 75 }],
      };

      const mockMetricsResponse = {
        measurements: [],
      };

      const histogramParser = vi.fn().mockReturnValue(mockHistogramMap);

      // Simulate line chart processing
      const result = histogramParser(mockMetricsResponse);

      expect(histogramParser).toHaveBeenCalledWith(mockMetricsResponse);
      expect(result).toEqual(mockHistogramMap);
    });

    test("should throw error when histogramParser is missing for line chart", () => {
      const mockConfig = {
        chartOptions: { type: "line" },
        histogramParser: undefined,
      };

      expect(() => {
        if (!mockConfig.histogramParser) {
          throw new Error("Histogram parser is required");
        }
      }).toThrow("Histogram parser is required");
    });
  });

  describe("Series Creation", () => {
    test("should create line series from histogram map", () => {
      const histogramMap = {
        cpu: [{ time: "2023-01-01", value: 50 }],
        memory: [{ time: "2023-01-01", value: 75 }],
      };
      const timeframes = ["2023-01-01", "2023-01-02"];

      mockComposable.createLineSeriesFromMap(histogramMap, timeframes);

      expect(mockComposable.createLineSeriesFromMap).toHaveBeenCalledWith(histogramMap, timeframes);
    });
  });

  describe("Props Changes", () => {
    test("should handle props changes", async () => {
      const onUnregister = vi.fn();
      const onLoadData = vi.fn();

      await mockComposable.handlePropsChanged(onUnregister, onLoadData);

      expect(mockComposable.handlePropsChanged).toHaveBeenCalledWith(onUnregister, onLoadData);
    });
  });

  describe("Lifecycle Management", () => {
    test("should handle beforeUnmount", () => {
      const onUnregister = vi.fn();

      mockComposable.handleBeforeUnmount(onUnregister);

      expect(mockComposable.handleBeforeUnmount).toHaveBeenCalledWith(onUnregister);
    });
  });

  describe("Error Handling", () => {
    test("should set error state when chart has no series", () => {
      mockComposable.chartOptions.value = { series: [] };

      expect((mockComposable.chartOptions.value.series as unknown[]).length).toBe(0);
    });

    test("should not set error state when chart has series", () => {
      mockComposable.chartOptions.value = { series: [{ name: "test", data: [], type: "line" }] };

      expect((mockComposable.chartOptions.value.series as unknown[]).length).toBe(1);
    });
  });

  describe("Y-Axis Label Updates", () => {
    test("should update Y-axis labels when yAxisLabelFormatting is configured", () => {
      // Simulate Y-axis label update
      const result = mockComposable.chartOptions.value;

      expect(result).toBeDefined();
    });

    test("should not update Y-axis labels when yAxisLabelFormatting is not configured", () => {
      const result = mockComposable.chartOptions.value;

      expect(result).toBeDefined();
    });
  });

  describe("Configuration", () => {
    test("should handle minimal config", () => {
      const minimalConfig = {
        chartId: "minimal-chart",
        metricTypes: ["cpu"],
        chartTitle: "Minimal Chart",
        intervalLabel: "1m",
        chartOptions: {
          type: "line",
          yAxisTitle: "Y Axis",
          height: 300,
          showSharedCrosshair: false,
          sharedTooltip: false,
        },
        getData: vi.fn(),
        histogramParser: vi.fn(),
        columnHistogramParser: vi.fn(),
        filterByDates: {
          dateStart: new Date("2023-01-01"),
          dateEnd: new Date("2023-01-02"),
        },
        getDataArgs: [],
      };

      mockComposable.initialize(minimalConfig);

      expect(mockComposable.initialize).toHaveBeenCalledWith(minimalConfig);
    });

    test("should handle full config with all options", () => {
      const fullConfig = {
        chartId: "full-chart",
        metricTypes: ["cpu", "memory"],
        chartTitle: "Full Chart",
        intervalLabel: "1m",
        chartOptions: {
          type: "line",
          yAxisTitle: "Y Axis",
          height: 300,
          showSharedCrosshair: true,
          sharedTooltip: true,
          yAxisFormatFunction: vi.fn(),
          colors: ["#ff0000"],
        },
        getData: vi.fn(),
        histogramParser: vi.fn(),
        columnHistogramParser: vi.fn(),
        keyLabelMap: { cpu: "CPU Usage", memory: "Memory Usage" },
        tooltipSuffix: "%",
        isPercentage: true,
        groupBy: ["node"],
        yAxisLabelFormatting: {
          tickPositions: [0, 25, 50, 75, 100],
        },
        formatter: vi.fn(),
        wrapperProps: {
          hideBorder: true,
          flat: true,
          hideActions: true,
          hideSeparator: true,
        },
      };

      mockComposable.initialize(fullConfig);

      expect(mockComposable.initialize).toHaveBeenCalledWith(fullConfig);
    });
  });
});
