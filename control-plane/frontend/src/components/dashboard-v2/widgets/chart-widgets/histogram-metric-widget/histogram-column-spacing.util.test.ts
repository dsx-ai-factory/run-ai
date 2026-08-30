import { describe, expect, it } from "vitest";
import {
  calculatePointInterval,
  millisecondsToMinutes,
  getColumnSpacingConfig,
  COLUMN_SPACING_CONFIGS,
  type ColumnSpacingConfig,
} from "./histogram-column-spacing.util";

describe("histogram-column-spacing.util", () => {
  describe("COLUMN_SPACING_CONFIGS", () => {
    it("should have all required configuration entries", () => {
      expect(COLUMN_SPACING_CONFIGS).toHaveLength(11);
    });

    it("should have configurations sorted by minIntervalMinutes in descending order", () => {
      for (let i = 0; i < COLUMN_SPACING_CONFIGS.length - 1; i++) {
        expect(COLUMN_SPACING_CONFIGS[i].minIntervalMinutes).toBeGreaterThan(
          COLUMN_SPACING_CONFIGS[i + 1].minIntervalMinutes,
        );
      }
    });

    it("should have valid configuration properties", () => {
      COLUMN_SPACING_CONFIGS.forEach((config: ColumnSpacingConfig) => {
        expect(config).toHaveProperty("minIntervalMinutes");
        expect(config).toHaveProperty("maxPointWidth");
        expect(config).toHaveProperty("pointPadding");
        expect(config).toHaveProperty("groupPadding");
        expect(config).toHaveProperty("description");

        expect(typeof config.minIntervalMinutes).toBe("number");
        expect(typeof config.maxPointWidth).toBe("number");
        expect(typeof config.pointPadding).toBe("number");
        expect(typeof config.groupPadding).toBe("number");
        expect(typeof config.description).toBe("string");

        expect(config.minIntervalMinutes).toBeGreaterThanOrEqual(0);
        expect(config.maxPointWidth).toBeGreaterThan(0);
        expect(config.pointPadding).toBeGreaterThanOrEqual(0);
        expect(config.groupPadding).toBeGreaterThanOrEqual(0);
      });
    });

    it("should have the last config with minIntervalMinutes of 0", () => {
      const lastConfig = COLUMN_SPACING_CONFIGS[COLUMN_SPACING_CONFIGS.length - 1];
      expect(lastConfig.minIntervalMinutes).toBe(0);
    });
  });

  describe("calculatePointInterval", () => {
    it("should return undefined when timeframes array is empty", () => {
      const result = calculatePointInterval([]);
      expect(result).toBeUndefined();
    });

    it("should return undefined when timeframes array has only one element", () => {
      const result = calculatePointInterval(["2024-03-19T09:00:00Z"]);
      expect(result).toBeUndefined();
    });

    it("should calculate interval between two consecutive timeframes", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T10:00:00Z"];
      const result = calculatePointInterval(timeframes);

      // 1 hour = 3600000 milliseconds
      expect(result).toBe(3600000);
    });

    it("should calculate interval for 30-minute intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:30:00Z"];
      const result = calculatePointInterval(timeframes);

      // 30 minutes = 1800000 milliseconds
      expect(result).toBe(1800000);
    });

    it("should calculate interval for 5-minute intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:05:00Z"];
      const result = calculatePointInterval(timeframes);

      // 5 minutes = 300000 milliseconds
      expect(result).toBe(300000);
    });

    it("should calculate interval for 1-second intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:00:01Z"];
      const result = calculatePointInterval(timeframes);

      // 1 second = 1000 milliseconds
      expect(result).toBe(1000);
    });

    it("should calculate interval for multi-day ranges", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-20T09:00:00Z"];
      const result = calculatePointInterval(timeframes);

      // 1 day = 86400000 milliseconds
      expect(result).toBe(86400000);
    });

    it("should only use the first two timeframes for calculation", () => {
      const timeframes = [
        "2024-03-19T09:00:00Z",
        "2024-03-19T10:00:00Z",
        "2024-03-19T11:00:00Z",
        "2024-03-19T12:00:00Z",
      ];
      const result = calculatePointInterval(timeframes);

      // Should still be 1 hour
      expect(result).toBe(3600000);
    });

    it("should handle millisecond precision timestamps", () => {
      const timeframes = ["2024-03-19T09:00:00.000Z", "2024-03-19T09:00:00.500Z"];
      const result = calculatePointInterval(timeframes);

      // 500 milliseconds
      expect(result).toBe(500);
    });
  });

  describe("millisecondsToMinutes", () => {
    it("should convert milliseconds to minutes correctly", () => {
      expect(millisecondsToMinutes(60000)).toBe(1); // 1 minute
      expect(millisecondsToMinutes(3600000)).toBe(60); // 1 hour = 60 minutes
      expect(millisecondsToMinutes(1800000)).toBe(30); // 30 minutes
      expect(millisecondsToMinutes(300000)).toBe(5); // 5 minutes
      expect(millisecondsToMinutes(120000)).toBe(2); // 2 minutes
    });

    it("should handle zero milliseconds", () => {
      expect(millisecondsToMinutes(0)).toBe(0);
    });

    it("should handle sub-minute intervals", () => {
      expect(millisecondsToMinutes(30000)).toBe(0.5); // 30 seconds = 0.5 minutes
      expect(millisecondsToMinutes(15000)).toBe(0.25); // 15 seconds = 0.25 minutes
      expect(millisecondsToMinutes(1000)).toBeCloseTo(0.0167, 3); // 1 second ≈ 0.0167 minutes
    });

    it("should handle large time intervals", () => {
      expect(millisecondsToMinutes(86400000)).toBe(1440); // 1 day = 1440 minutes
      expect(millisecondsToMinutes(604800000)).toBe(10080); // 1 week = 10080 minutes
    });

    it("should handle decimal results correctly", () => {
      expect(millisecondsToMinutes(90000)).toBe(1.5); // 1.5 minutes
      expect(millisecondsToMinutes(75000)).toBe(1.25); // 1.25 minutes
    });
  });

  describe("getColumnSpacingConfig", () => {
    it("should return config for extra long intervals (180+ minutes)", () => {
      const result = getColumnSpacingConfig(180);
      expect(result.maxPointWidth).toBe(80);
      expect(result.pointPadding).toBe(0.05);
      expect(result.groupPadding).toBe(0.05);
    });

    it("should return config for very long intervals (120-180 minutes)", () => {
      const result = getColumnSpacingConfig(120);
      expect(result.maxPointWidth).toBe(65);
      expect(result.pointPadding).toBe(0.2);
      expect(result.groupPadding).toBe(0.2);
    });

    it("should return config for long intervals (60-120 minutes)", () => {
      const result = getColumnSpacingConfig(60);
      expect(result.maxPointWidth).toBe(55);
      expect(result.pointPadding).toBe(0.15);
      expect(result.groupPadding).toBe(0.15);
    });

    it("should return config for medium-long intervals (30-60 minutes)", () => {
      const result = getColumnSpacingConfig(30);
      expect(result.maxPointWidth).toBe(50);
      expect(result.pointPadding).toBe(0.18);
      expect(result.groupPadding).toBe(0.18);
    });

    it("should return config for medium intervals (15-30 minutes)", () => {
      const result = getColumnSpacingConfig(15);
      expect(result.maxPointWidth).toBe(40);
      expect(result.pointPadding).toBe(0.2);
      expect(result.groupPadding).toBe(0.2);
    });

    it("should return config for medium-short intervals (10-15 minutes)", () => {
      const result = getColumnSpacingConfig(10);
      expect(result.maxPointWidth).toBe(35);
      expect(result.pointPadding).toBe(0.2);
      expect(result.groupPadding).toBe(0.2);
    });

    it("should return config for short intervals (5-10 minutes)", () => {
      const result = getColumnSpacingConfig(5);
      expect(result.maxPointWidth).toBe(30);
      expect(result.pointPadding).toBe(0.22);
      expect(result.groupPadding).toBe(0.22);
    });

    it("should return config for very short intervals (2-5 minutes)", () => {
      const result = getColumnSpacingConfig(2);
      expect(result.maxPointWidth).toBe(24);
      expect(result.pointPadding).toBe(0.24);
      expect(result.groupPadding).toBe(0.24);
    });

    it("should return config for ultra-short intervals (1-2 minutes)", () => {
      const result = getColumnSpacingConfig(1);
      expect(result.maxPointWidth).toBe(18);
      expect(result.pointPadding).toBe(0.26);
      expect(result.groupPadding).toBe(0.26);
    });

    it("should return config for extreme intervals (0.5-1 minute)", () => {
      const result = getColumnSpacingConfig(0.5);
      expect(result.maxPointWidth).toBe(12);
      expect(result.pointPadding).toBe(0.28);
      expect(result.groupPadding).toBe(0.28);
    });

    it("should return config for sub-30sec intervals", () => {
      const result = getColumnSpacingConfig(0.25);
      expect(result.maxPointWidth).toBe(8);
      expect(result.pointPadding).toBe(0.3);
      expect(result.groupPadding).toBe(0.3);
    });

    it("should return the smallest interval config for zero", () => {
      const result = getColumnSpacingConfig(0);
      expect(result.maxPointWidth).toBe(8);
      expect(result.pointPadding).toBe(0.3);
      expect(result.groupPadding).toBe(0.3);
    });

    it("should handle boundary values correctly - exactly 180 minutes", () => {
      const result = getColumnSpacingConfig(180);
      expect(result.maxPointWidth).toBe(80);
    });

    it("should handle boundary values correctly - just below 180 minutes", () => {
      const result = getColumnSpacingConfig(179.9);
      expect(result.maxPointWidth).toBe(65);
    });

    it("should handle very large intervals (months)", () => {
      const result = getColumnSpacingConfig(43200); // 30 days in minutes
      expect(result.maxPointWidth).toBe(80);
      expect(result.pointPadding).toBe(0.05);
      expect(result.groupPadding).toBe(0.05);
    });

    it("should return consistent results for the same input", () => {
      const result1 = getColumnSpacingConfig(45);
      const result2 = getColumnSpacingConfig(45);
      expect(result1).toEqual(result2);
    });

    it("should handle edge case: interval between thresholds", () => {
      // Test 25 minutes (between 15 and 30)
      const result = getColumnSpacingConfig(25);
      expect(result.maxPointWidth).toBe(40); // Should use 15-minute config since 25 >= 15 but < 30
    });

    it("should select config based on >= comparison", () => {
      // Exact match on threshold
      const result60 = getColumnSpacingConfig(60);
      expect(result60.maxPointWidth).toBe(55);

      // Just above threshold
      const result61 = getColumnSpacingConfig(61);
      expect(result61.maxPointWidth).toBe(55);

      // Just below threshold
      const result59 = getColumnSpacingConfig(59);
      expect(result59.maxPointWidth).toBe(50);
    });
  });

  describe("Integration tests: full workflow", () => {
    it("should work end-to-end for 1-hour intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T10:00:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBe(3600000);

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(60);

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(55);
      expect(config.pointPadding).toBe(0.15);
      expect(config.groupPadding).toBe(0.15);
    });

    it("should work end-to-end for 5-minute intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:05:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBe(300000);

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(5);

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(30);
      expect(config.pointPadding).toBe(0.22);
      expect(config.groupPadding).toBe(0.22);
    });

    it("should work end-to-end for 30-second intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:00:30Z"];
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBe(30000);

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(0.5);

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(12);
      expect(config.pointPadding).toBe(0.28);
      expect(config.groupPadding).toBe(0.28);
    });

    it("should work end-to-end for 3-hour intervals (22-day range)", () => {
      const timeframes = ["2024-03-01T00:00:00Z", "2024-03-01T03:00:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBe(10800000); // 3 hours in ms

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(180);

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(80);
      expect(config.pointPadding).toBe(0.05);
      expect(config.groupPadding).toBe(0.05);
    });

    it("should handle undefined pointRange gracefully", () => {
      const timeframes = ["2024-03-19T09:00:00Z"]; // Only one timeframe
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBeUndefined();

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(1); // Fallback value

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(18); // 1-minute config
    });

    it("should handle empty timeframes array gracefully", () => {
      const timeframes: string[] = [];
      const pointRange = calculatePointInterval(timeframes);
      expect(pointRange).toBeUndefined();

      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      expect(intervalInMinutes).toBe(1); // Fallback value

      const config = getColumnSpacingConfig(intervalInMinutes);
      expect(config.maxPointWidth).toBe(18);
    });
  });

  describe("Real-world scenarios", () => {
    it("should handle 7-day range with 1-hour intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T10:00:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      const config = getColumnSpacingConfig(intervalInMinutes);

      expect(intervalInMinutes).toBe(60);
      expect(config.maxPointWidth).toBe(55);
      expect(config.pointPadding).toBe(0.15);
    });

    it("should handle 12-hour range with 2-minute intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:02:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      const config = getColumnSpacingConfig(intervalInMinutes);

      expect(intervalInMinutes).toBe(2);
      expect(config.maxPointWidth).toBe(24);
      expect(config.pointPadding).toBe(0.24);
    });

    it("should handle 30-day range with 2-hour intervals", () => {
      const timeframes = ["2024-03-01T00:00:00Z", "2024-03-01T02:00:00Z"];
      const pointRange = calculatePointInterval(timeframes);
      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      const config = getColumnSpacingConfig(intervalInMinutes);

      expect(intervalInMinutes).toBe(120);
      expect(config.maxPointWidth).toBe(65);
      expect(config.pointPadding).toBe(0.2);
    });

    it("should handle sub-hour range with 10-second intervals", () => {
      const timeframes = ["2024-03-19T09:00:00Z", "2024-03-19T09:00:10Z"];
      const pointRange = calculatePointInterval(timeframes);
      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      const config = getColumnSpacingConfig(intervalInMinutes);

      expect(intervalInMinutes).toBeCloseTo(0.167, 2);
      expect(config.maxPointWidth).toBe(8); // Sub-30sec config
      expect(config.pointPadding).toBe(0.3);
    });
  });
});
