import { describe, expect, it } from "vitest";
import type { IColumnHistogramDataParser } from "@/models/chart.model";
import type { MetricsCompositeResponse } from "@/swagger-models/workloads-client";
import { mapHistogramToRanges, parseHistogramToColumnSeries } from "./column-histogram-process-metrics-data.util";

// Define range configuration for Time to First Token
const getRangeKey = (bucketValue: number): string => {
  if (bucketValue <= 0.25) return "0-0.25";
  if (bucketValue <= 0.5) return "0.25-0.5";
  if (bucketValue <= 1) return "0.5-1";
  if (bucketValue <= 2.5) return "1-2.5";
  if (bucketValue <= 5) return "2.5-5";
  return "5+";
};

const rangeKeys = ["0-0.25", "0.25-0.5", "0.5-1", "1-2.5", "2.5-5", "5+"];

describe("column-histogram-process-metrics-data.util", () => {
  describe("mapHistogramToRanges", () => {
    it("should map histogram data to predefined ranges correctly", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "5" },
            { timestamp: "2024-03-19T10:00:00Z", value: "3" },
          ],
        },
        {
          name: "0.3",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "10" },
            { timestamp: "2024-03-19T10:00:00Z", value: "8" },
          ],
        },
        {
          name: "0.7",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "15" },
            { timestamp: "2024-03-19T10:00:00Z", value: "12" },
          ],
        },
        {
          name: "1.5",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "20" },
            { timestamp: "2024-03-19T10:00:00Z", value: "18" },
          ],
        },
        {
          name: "3.0",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "25" },
            { timestamp: "2024-03-19T10:00:00Z", value: "22" },
          ],
        },
        {
          name: "6.0",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "30" },
            { timestamp: "2024-03-19T10:00:00Z", value: "28" },
          ],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      expect(result).toHaveLength(6);

      // Check 0-0.25 range (0.1 bucket)
      expect(result[0].name).toBe("0-0.25");
      expect(result[0].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "5" },
        { timestamp: "2024-03-19T10:00:00Z", value: "3" },
      ]);

      // Check 0.25-0.5 range (0.3 bucket)
      expect(result[1].name).toBe("0.25-0.5");
      expect(result[1].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "10" },
        { timestamp: "2024-03-19T10:00:00Z", value: "8" },
      ]);

      // Check 0.5-1 range (0.7 bucket)
      expect(result[2].name).toBe("0.5-1");
      expect(result[2].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "15" },
        { timestamp: "2024-03-19T10:00:00Z", value: "12" },
      ]);

      // Check 1-2.5 range (1.5 bucket)
      expect(result[3].name).toBe("1-2.5");
      expect(result[3].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "20" },
        { timestamp: "2024-03-19T10:00:00Z", value: "18" },
      ]);

      // Check 2.5-5 range (3.0 bucket)
      expect(result[4].name).toBe("2.5-5");
      expect(result[4].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "25" },
        { timestamp: "2024-03-19T10:00:00Z", value: "22" },
      ]);

      // Check 5+ range (6.0 bucket)
      expect(result[5].name).toBe("5+");
      expect(result[5].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "30" },
        { timestamp: "2024-03-19T10:00:00Z", value: "28" },
      ]);
    });

    it("should sum multiple buckets in the same range", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "5" }],
        },
        {
          name: "0.2",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "10" }],
        },
        {
          name: "0.25",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "15" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // All three buckets (0.1, 0.2, 0.25) should be in the 0-0.25 range
      expect(result[0].name).toBe("0-0.25");
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "30" }]);
    });

    it("should handle edge case values at range boundaries", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.25",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "10" }],
        },
        {
          name: "0.5",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "20" }],
        },
        {
          name: "1.0",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "30" }],
        },
        {
          name: "2.5",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "40" }],
        },
        {
          name: "5.0",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "50" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // 0.25 should be in 0-0.25 range
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "10" }]);

      // 0.5 should be in 0.25-0.5 range
      expect(result[1].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "20" }]);

      // 1.0 should be in 0.5-1 range
      expect(result[2].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "30" }]);

      // 2.5 should be in 1-2.5 range
      expect(result[3].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "40" }]);

      // 5.0 should be in 2.5-5 range
      expect(result[4].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "50" }]);
    });

    it("should handle +Inf and NaN values by placing them in 5+ range", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "5" }],
        },
        {
          name: "+Inf",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "100" }],
        },
        {
          name: "NaN",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "50" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // 0.1 should be in 0-0.25 range
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "5" }]);

      // +Inf and NaN should be summed in 5+ range
      expect(result[5].name).toBe("5+");
      expect(result[5].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "150" }]);
    });

    it("should handle empty column series array", () => {
      const result = mapHistogramToRanges([], getRangeKey, rangeKeys);

      expect(result).toHaveLength(6);
      result.forEach((range) => {
        expect(range.data).toEqual([]);
      });
    });

    it("should handle multiple timestamps and aggregate correctly", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.3",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "5" },
            { timestamp: "2024-03-19T10:00:00Z", value: "10" },
            { timestamp: "2024-03-19T11:00:00Z", value: "15" },
          ],
        },
        {
          name: "0.4",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "3" },
            { timestamp: "2024-03-19T10:00:00Z", value: "6" },
            { timestamp: "2024-03-19T11:00:00Z", value: "9" },
          ],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // Both 0.3 and 0.4 should be in 0.25-0.5 range and summed per timestamp
      expect(result[1].name).toBe("0.25-0.5");
      expect(result[1].data).toHaveLength(3);
      expect(result[1].data).toEqual(
        expect.arrayContaining([
          { timestamp: "2024-03-19T09:00:00Z", value: "8" },
          { timestamp: "2024-03-19T10:00:00Z", value: "16" },
          { timestamp: "2024-03-19T11:00:00Z", value: "24" },
        ]),
      );
    });

    it("should handle series with missing or empty values", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "" },
            { timestamp: "2024-03-19T10:00:00Z", value: "10" },
          ],
        },
        {
          name: "0.2",
          data: [
            { timestamp: "2024-03-19T09:00:00Z", value: "5" },
            {
              timestamp: "2024-03-19T10:00:00Z",
              value: "",
            },
          ],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      expect(result[0].name).toBe("0-0.25");
      expect(result[0].data).toEqual(
        expect.arrayContaining([
          { timestamp: "2024-03-19T09:00:00Z", value: "5" },
          { timestamp: "2024-03-19T10:00:00Z", value: "10" },
        ]),
      );
    });

    it("should preserve the order of ranges", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        { name: "10.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "1" }] },
        { name: "0.1", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "2" }] },
        { name: "1.5", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "3" }] },
        { name: "0.3", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "4" }] },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      expect(result.map((r) => r.name)).toEqual(["0-0.25", "0.25-0.5", "0.5-1", "1-2.5", "2.5-5", "5+"]);
    });

    it("should handle very large bucket values", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "100",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "50" }],
        },
        {
          name: "1000",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "25" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // Both should be in 5+ range
      expect(result[5].name).toBe("5+");
      expect(result[5].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "75" }]);
    });

    it("should handle decimal values correctly", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.001",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "100" }],
        },
        {
          name: "0.005",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "200" }],
        },
        {
          name: "0.01",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "300" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // All should be in 0-0.25 range
      expect(result[0].name).toBe("0-0.25");
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "600" }]);
    });

    it("should handle ranges with no data", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "10" }],
        },
        {
          name: "10.0",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "20" }],
        },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      expect(result).toHaveLength(6);

      // Check that empty ranges exist
      expect(result[1].name).toBe("0.25-0.5");
      expect(result[1].data).toEqual([]);

      expect(result[2].name).toBe("0.5-1");
      expect(result[2].data).toEqual([]);

      expect(result[3].name).toBe("1-2.5");
      expect(result[3].data).toEqual([]);

      expect(result[4].name).toBe("2.5-5");
      expect(result[4].data).toEqual([]);

      // Populated ranges
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "10" }]);
      expect(result[5].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "20" }]);
    });

    it("should handle complex real-world scenario with all ranges populated", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        { name: "0.001", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "10" }] },
        { name: "0.005", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "15" }] },
        { name: "0.01", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "20" }] },
        { name: "0.02", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "25" }] },
        { name: "0.04", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "30" }] },
        { name: "0.06", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "35" }] },
        { name: "0.08", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "40" }] },
        { name: "0.1", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "45" }] },
        { name: "0.3", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "50" }] },
        { name: "0.4", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "55" }] },
        { name: "0.6", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "60" }] },
        { name: "0.8", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "65" }] },
        { name: "1.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "70" }] },
        { name: "1.5", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "75" }] },
        { name: "2.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "80" }] },
        { name: "3.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "85" }] },
        { name: "4.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "90" }] },
        { name: "7.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "95" }] },
        { name: "10.0", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "100" }] },
        { name: "+Inf", data: [{ timestamp: "2024-03-19T09:00:00Z", value: "105" }] },
      ];

      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys);

      // 0-0.25: 0.001, 0.005, 0.01, 0.02, 0.04, 0.06, 0.08, 0.1 = 220
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "220" }]);

      // 0.25-0.5: 0.3, 0.4 = 105
      expect(result[1].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "105" }]);

      // 0.5-1: 0.6, 0.8, 1.0 = 195
      expect(result[2].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "195" }]);

      // 1-2.5: 1.5, 2.0 = 155
      expect(result[3].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "155" }]);

      // 2.5-5: 3.0, 4.0 = 175
      expect(result[4].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "175" }]);

      // 5+: 7.0, 10.0, +Inf = 300
      expect(result[5].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "300" }]);
    });

    it("should use custom defaultRangeForNaN parameter", () => {
      const mockColumnSeries: IColumnHistogramDataParser[] = [
        {
          name: "0.1",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "5" }],
        },
        {
          name: "+Inf",
          data: [{ timestamp: "2024-03-19T09:00:00Z", value: "100" }],
        },
      ];

      // Use first range for NaN values instead of last
      const result = mapHistogramToRanges(mockColumnSeries, getRangeKey, rangeKeys, "0-0.25");

      // Both 0.1 and +Inf should be in 0-0.25 range
      expect(result[0].name).toBe("0-0.25");
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "105" }]);
    });
  });

  describe("parseHistogramToColumnSeries", () => {
    it("should parse histogram metrics response correctly", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.001": "0",
                  "0.005": "0",
                  "0.1": "8",
                  "0.25": "15",
                  "0.5": "20",
                },
              },
              {
                timestamp: "2024-03-19T10:00:00Z",
                data: {
                  "0.001": "0",
                  "0.005": "5",
                  "0.1": "12",
                  "0.25": "18",
                  "0.5": "25",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(5);
      expect(result[0].name).toBe("0.001");
      expect(result[0].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "0" },
        { timestamp: "2024-03-19T10:00:00Z", value: "0" },
      ]);
      expect(result[1].name).toBe("0.005");
      expect(result[1].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "0" },
        { timestamp: "2024-03-19T10:00:00Z", value: "5" },
      ]);
      expect(result[2].name).toBe("0.1");
      expect(result[2].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "8" },
        { timestamp: "2024-03-19T10:00:00Z", value: "7" },
      ]);
    });

    it("should convert NaN bucket names to +Inf", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.1": "10",
                  "0.5": "20",
                  NaN: "30",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      // NaN should be converted to +Inf
      const infSeries = result.find((s) => s.name === "+Inf");
      expect(infSeries).toBeDefined();

      // Verify NaN was converted (no series named "NaN" should exist)
      const nanSeries = result.find((s) => s.name === "NaN");
      expect(nanSeries).toBeUndefined();
    });

    it("should handle empty metrics response", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toEqual([]);
    });

    it("should handle undefined histogram", () => {
      const metricsResponse: MetricsCompositeResponse = {
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toEqual([]);
    });

    it("should handle empty values array", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toEqual([]);
    });

    it("should process cumulative histogram values correctly by calculating differences", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.01": "8",
                  "0.02": "26",
                  "0.04": "374",
                  "0.06": "534",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(4);
      expect(result[0].name).toBe("0.01");
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "8" }]);
      expect(result[1].name).toBe("0.02");
      expect(result[1].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "18" }]);
      expect(result[2].name).toBe("0.04");
      expect(result[2].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "348" }]);
      expect(result[3].name).toBe("0.06");
      expect(result[3].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "160" }]);
    });

    it("should sort buckets by key value from min to max", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "10.0": "100",
                  "0.1": "10",
                  "1.0": "50",
                  "0.01": "5",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(4);
      expect(result[0].name).toBe("0.01");
      expect(result[1].name).toBe("0.1");
      expect(result[2].name).toBe("1");
      expect(result[3].name).toBe("10");
    });

    it("should handle +Inf bucket correctly", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.1": "10",
                  "0.5": "20",
                  "+Inf": "30",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      // +Inf bucket should be present
      const infSeries = result.find((s) => s.name === "+Inf");
      expect(infSeries).toBeDefined();
      expect(infSeries?.data).toHaveLength(1);
      expect(infSeries?.data[0].timestamp).toBe("2024-03-19T09:00:00Z");
    });

    it("should handle multiple timestamps with consistent buckets", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.1": "10",
                  "0.5": "25",
                },
              },
              {
                timestamp: "2024-03-19T10:00:00Z",
                data: {
                  "0.1": "15",
                  "0.5": "35",
                },
              },
              {
                timestamp: "2024-03-19T11:00:00Z",
                data: {
                  "0.1": "20",
                  "0.5": "45",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(2);
      expect(result[0].name).toBe("0.1");
      expect(result[0].data).toHaveLength(3);
      expect(result[0].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "10" },
        { timestamp: "2024-03-19T10:00:00Z", value: "15" },
        { timestamp: "2024-03-19T11:00:00Z", value: "20" },
      ]);
      expect(result[1].name).toBe("0.5");
      expect(result[1].data).toHaveLength(3);
      expect(result[1].data).toEqual([
        { timestamp: "2024-03-19T09:00:00Z", value: "15" },
        { timestamp: "2024-03-19T10:00:00Z", value: "20" },
        { timestamp: "2024-03-19T11:00:00Z", value: "25" },
      ]);
    });

    it("should handle single data point", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {
                  "0.1": "100",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(1);
      expect(result[0].name).toBe("0.1");
      expect(result[0].data).toEqual([{ timestamp: "2024-03-19T09:00:00Z", value: "100" }]);
    });

    it("should handle empty data object", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2024-03-19T09:00:00Z",
                data: {},
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toEqual([]);
    });

    it("should handle real-world prometheus histogram format", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                timestamp: "2025-09-28T18:02:37.5Z",
                data: {
                  "+Inf": "605",
                  "0.001": "0",
                  "0.005": "0",
                  "0.01": "8",
                  "0.02": "26",
                  "0.04": "374",
                  "0.06": "534",
                  "0.08": "605",
                  "0.1": "605",
                  "0.25": "605",
                  "0.5": "605",
                  "0.75": "605",
                  "1.0": "605",
                  "2.5": "605",
                  "5.0": "605",
                  "7.5": "605",
                  "10.0": "605",
                },
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result.length).toBeGreaterThan(0);

      // Check that buckets are sorted
      const bucketValues = result.map((s) => parseFloat(s.name)).filter((v) => !isNaN(v));
      for (let i = 1; i < bucketValues.length; i++) {
        expect(bucketValues[i]).toBeGreaterThanOrEqual(bucketValues[i - 1]);
      }

      // Check that +Inf is present
      const infSeries = result.find((s) => s.name === "+Inf");
      expect(infSeries).toBeDefined();

      // Check that cumulative values are converted to differences
      const firstBucket = result[0];
      expect(firstBucket.data[0].value).toBe("0");
    });

    it("should handle missing timestamp field", () => {
      const metricsResponse: MetricsCompositeResponse = {
        histogram: [
          {
            values: [
              {
                data: {
                  "0.1": "10",
                  "0.5": "20",
                },
                timestamp: null,
              },
            ],
            type: "",
          },
        ],
        measurements: [],
      };

      const result = parseHistogramToColumnSeries(metricsResponse);

      expect(result).toHaveLength(2);
      expect(result[0].data[0].timestamp).toBe("");
      expect(result[1].data[0].timestamp).toBe("");
    });
  });
});
