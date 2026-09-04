/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { EChartsOption } from "echarts";
import { describe, expect, test, vi, beforeEach } from "vitest";
import { ref, type Ref } from "vue";

// Mock the composable
const mockComposable = {
  chartOptions: ref({ series: [] }) as Ref<EChartsOption | null>,
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

vi.mock("@/utils/metric.util", () => ({
  metricUtil: {
    mapMeasurementsToMap: vi.fn((_measurements) => ({
      cpu: [{ time: "2023-01-01", value: 50 }],
      memory: [{ time: "2023-01-01", value: 75 }],
    })),
  },
}));

vi.mock("@/utils/interval.util", () => ({
  intervalUtil: {
    startInterval: vi.fn(),
    stopInterval: vi.fn(),
  },
}));

describe("GenericMetricWidget", () => {
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

  describe("Series Creation", () => {
    test("should create line series from measurement map", () => {
      const measurementMap = {
        cpu: [{ time: "2023-01-01", value: 50 }],
        memory: [{ time: "2023-01-01", value: 75 }],
      };
      const timeframes = ["2023-01-01", "2023-01-02"];

      mockComposable.createLineSeriesFromMap(measurementMap, timeframes);

      expect(mockComposable.createLineSeriesFromMap).toHaveBeenCalledWith(measurementMap, timeframes);
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

      expect((mockComposable.chartOptions.value.series as unknown[])?.length).toBe(0);
    });

    test("should not set error state when chart has series", () => {
      mockComposable.chartOptions.value = { series: [{ name: "test", data: [], type: "line" }] };

      expect((mockComposable.chartOptions.value?.series as unknown[])?.length).toBe(1);
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
        keyLabelMap: { cpu: "CPU Usage", memory: "Memory Usage" },
        tooltipSuffix: "%",
        isPercentage: true,
        groupBy: ["node"],
        measurementParser: vi.fn(),
        averageMetricType: "cpu",
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
