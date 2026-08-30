import { describe, it, expect } from "vitest";
import { metricUtil } from "./metric.util";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

import {
  type MetricsResponse,
  type GroupedMeasurementResponseValuesInner,
  MetricsType,
} from "@/swagger-models/cluster-service-client";

describe("metric.util", () => {
  describe("mapMeasurementsToMap", () => {
    it("returns an empty object if measurements array is empty in metrics response", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({});
    });
    it("it should handle return empty array in case of measurement.value is null ", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          {
            type: "TOTAL_GPU",
            values: null,
          },
          {
            type: "ALLOCATED_GPU",
            values: null,
          },
          {
            type: "GPU_QUOTA",
            values: null,
          },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({ ALLOCATED_GPU: [], GPU_QUOTA: [], TOTAL_GPU: [] });
    });

    // Test case for non-empty measurements array in metrics response
    it("returns a map with measurements mapped by type", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          { type: MetricsType.GpuUtilization, values: [{ value: "25", timestamp: "1646522400000" }] },
          { type: MetricsType.CpuUtilization, values: [{ value: "50", timestamp: "1646522400000" }] },
          { type: MetricsType.GpuQuota, values: [{ value: "1013", timestamp: "1646522400000" }] },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [{ value: "25", timestamp: "1646522400000" }],
        [MetricsType.CpuUtilization]: [{ value: "50", timestamp: "1646522400000" }],
        [MetricsType.GpuQuota]: [{ value: "1013", timestamp: "1646522400000" }],
      });
    });

    it("merges values from duplicate metric types (e.g. nodepool label change creates two Prometheus series)", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          {
            type: MetricsType.GpuUtilization,
            values: [{ value: "25", timestamp: "1646522400000" }],
          },
          {
            type: MetricsType.GpuUtilization,
            values: [{ value: "30", timestamp: "1646526000000" }],
          },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [
          { value: "25", timestamp: "1646522400000" },
          { value: "30", timestamp: "1646526000000" },
        ],
      });
    });

    it("caps percentage metric values above 100 at 100", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          {
            type: MetricsType.GpuUtilization,
            values: [
              { value: "150", timestamp: "1646522400000" },
              { value: "1500000", timestamp: "1646526000000" },
            ],
          },
          {
            type: MetricsType.GpuMemoryUtilization,
            values: [{ value: "120.5", timestamp: "1646522400000" }],
          },
          {
            type: MetricsType.CpuUtilization,
            values: [{ value: "101", timestamp: "1646522400000" }],
          },
          {
            type: MetricsType.CpuMemoryUtilization,
            values: [{ value: "200", timestamp: "1646522400000" }],
          },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [
          { value: "100", timestamp: "1646522400000" },
          { value: "100", timestamp: "1646526000000" },
        ],
        [MetricsType.GpuMemoryUtilization]: [{ value: "100", timestamp: "1646522400000" }],
        [MetricsType.CpuUtilization]: [{ value: "100", timestamp: "1646522400000" }],
        [MetricsType.CpuMemoryUtilization]: [{ value: "100", timestamp: "1646522400000" }],
      });
    });

    it("leaves percentage metric values at or below 100 untouched", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          {
            type: MetricsType.GpuUtilization,
            values: [
              { value: "0", timestamp: "1646522400000" },
              { value: "42.5", timestamp: "1646526000000" },
              { value: "100", timestamp: "1646529600000" },
            ],
          },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [
          { value: "0", timestamp: "1646522400000" },
          { value: "42.5", timestamp: "1646526000000" },
          { value: "100", timestamp: "1646529600000" },
        ],
      });
    });

    it("does not cap non-percentage metric types (e.g. counts and quotas)", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          { type: MetricsType.GpuQuota, values: [{ value: "1013", timestamp: "1646522400000" }] },
          { type: "ALLOCATED_GPU", values: [{ value: "500", timestamp: "1646522400000" }] },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuQuota]: [{ value: "1013", timestamp: "1646522400000" }],
        ALLOCATED_GPU: [{ value: "500", timestamp: "1646522400000" }],
      });
    });

    it("caps DCGM activity-ratio per-GPU metrics and NIM cache percent at 100", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          { type: "GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU", values: [{ value: "150", timestamp: "1646522400000" }] },
          { type: "GPU_SM_ACTIVITY_PER_GPU", values: [{ value: "175", timestamp: "1646522400000" }] },
          { type: "GPU_SM_OCCUPANCY_PER_GPU", values: [{ value: "200", timestamp: "1646522400000" }] },
          { type: "GPU_TENSOR_ACTIVITY_PER_GPU", values: [{ value: "120", timestamp: "1646522400000" }] },
          { type: "GPU_FP64_ENGINE_ACTIVITY_PER_GPU", values: [{ value: "300", timestamp: "1646522400000" }] },
          { type: "GPU_FP32_ENGINE_ACTIVITY_PER_GPU", values: [{ value: "999", timestamp: "1646522400000" }] },
          { type: "GPU_FP16_ENGINE_ACTIVITY_PER_GPU", values: [{ value: "101", timestamp: "1646522400000" }] },
          {
            type: "GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU",
            values: [{ value: "500", timestamp: "1646522400000" }],
          },
          { type: "NIM_GPU_CACHE_USAGE_PERC", values: [{ value: "250", timestamp: "1646522400000" }] },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      const cappedAt100: Array<GroupedMeasurementResponseValuesInner> = [{ value: "100", timestamp: "1646522400000" }];
      expect(result).toEqual({
        GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU: cappedAt100,
        GPU_SM_ACTIVITY_PER_GPU: cappedAt100,
        GPU_SM_OCCUPANCY_PER_GPU: cappedAt100,
        GPU_TENSOR_ACTIVITY_PER_GPU: cappedAt100,
        GPU_FP64_ENGINE_ACTIVITY_PER_GPU: cappedAt100,
        GPU_FP32_ENGINE_ACTIVITY_PER_GPU: cappedAt100,
        GPU_FP16_ENGINE_ACTIVITY_PER_GPU: cappedAt100,
        GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU: cappedAt100,
        NIM_GPU_CACHE_USAGE_PERC: cappedAt100,
      });
    });

    it("preserves non-numeric values for percentage metric types", () => {
      const metricsResponse: MetricsResponse = {
        measurements: [
          {
            type: MetricsType.GpuUtilization,
            values: [
              { value: "NaN", timestamp: "1646522400000" },
              { value: "not-a-number", timestamp: "1646526000000" },
              { value: "200", timestamp: "1646529600000" },
            ],
          },
        ],
      };
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [
          { value: "NaN", timestamp: "1646522400000" },
          { value: "not-a-number", timestamp: "1646526000000" },
          { value: "100", timestamp: "1646529600000" },
        ],
      });
    });

    it("handles measurement.values being undefined", () => {
      const metricsResponse = {
        measurements: [{ type: MetricsType.GpuUtilization }, { type: MetricsType.GpuQuota }],
      } as unknown as MetricsResponse;
      const result = metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(result).toEqual({
        [MetricsType.GpuUtilization]: [],
        [MetricsType.GpuQuota]: [],
      });
    });

    it("does not mutate the input measurement values when capping", () => {
      const originalValues = [
        { value: "150", timestamp: "1646522400000" },
        { value: "50", timestamp: "1646526000000" },
      ];
      const snapshot = originalValues.map((v) => ({ ...v }));
      const metricsResponse: MetricsResponse = {
        measurements: [{ type: MetricsType.GpuUtilization, values: originalValues }],
      };
      metricUtil.mapMeasurementsToMap(metricsResponse);
      expect(originalValues).toEqual(snapshot);
    });
  });
  describe("getMeasurementValue", () => {
    it("returns an empty array if measurements array is empty", () => {
      const measurements: GroupedMeasurementResponseValuesInner[] = [];
      const result = metricUtil.getMeasurementValue(measurements);
      expect(result).toEqual([]);
    });

    it("returns array of numbers from the value", () => {
      const measurements: GroupedMeasurementResponseValuesInner[] = [
        {
          value: "0",
          timestamp: "2024-02-29T14:49:01Z",
        },
        {
          value: "2",
          timestamp: "2024-02-29T17:13:01Z",
        },
      ];
      const result = metricUtil.getMeasurementValue(measurements);
      expect(result.length).toBe(measurements.length);
      expect(result[0]).toEqual(0);
      expect(result[1]).toEqual(2);
    });
  });

  describe("getMeasurementTimestamp", () => {
    it("returns an empty array if measurements array is empty", () => {
      const measurements: GroupedMeasurementResponseValuesInner[] = [];
      const result = metricUtil.getMeasurementTimestamp(measurements);
      expect(result).toEqual([]);
    });

    it("returns formatted date array from the timestamp", () => {
      const measurements: GroupedMeasurementResponseValuesInner[] = [
        {
          value: "0",
          timestamp: "2024-02-29T14:49:01Z",
        },
        {
          value: "0",
          timestamp: "",
        },
      ];
      const result = metricUtil.getMeasurementTimestamp(measurements);
      expect(result.length).toBe(measurements.length);
      expect(result[0]).toEqual(
        dateUtil.dateFormat(new Date(measurements[0].timestamp as string), DATE_FORMATS.SHORT_DATE_TIME),
      );
      expect(result[1]).toEqual("");
    });
  });

  describe("insertConstantValueIntoTimeframe", () => {
    it("should insert constant value into each timeframe starting from specified time", () => {
      const constantValue = 5;
      const timeframes = [0, 1, 2, 3, 4, 5];
      const startTime = 2; // Insert constant value starting from time 2

      const result = metricUtil.insertConstantValueIntoTimeframe(constantValue, timeframes, startTime);

      expect(result).toEqual([
        [0, null],
        [1, null],
        [2, 5],
        [3, 5],
        [4, 5],
        [5, 5],
      ]);
    });

    it("should insert constant value into each timeframe starting from beginning if no start time provided", () => {
      const constantValue = 10;
      const timeframes = [0, 1, 2, 3, 4, 5];

      const result = metricUtil.insertConstantValueIntoTimeframe(constantValue, timeframes);

      expect(result).toEqual([
        [0, 10],
        [1, 10],
        [2, 10],
        [3, 10],
        [4, 10],
        [5, 10],
      ]);
    });

    it("should return empty array if timeframes array is empty", () => {
      const constantValue = 10;
      const timeframes: number[] = [];

      const result = metricUtil.insertConstantValueIntoTimeframe(constantValue, timeframes);

      expect(result).toEqual([]);
    });
  });
});
