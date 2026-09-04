/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, test } from "vitest";
import { dateUtil, TimeUnit } from "@/utils/date.util";
import { PREDEFINED_DATES } from "@/models/date.model";
import { isDateString, hoursGranularityPredefinedOptions } from "@/utils/date.util/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

describe("date.util", () => {
  describe("dateFormat", () => {
    it("Should use default format", async () => {
      expect(dateUtil.dateFormat(new Date(100000000000))).toEqual("03/03/1973");
    });

    it("Should use centralized format constants", async () => {
      expect(dateUtil.dateFormat(new Date(100000000000), DATE_FORMATS.STANDARD_DATE)).toEqual("03/03/1973");
      expect(dateUtil.dateFormat(new Date(100000000000), DATE_FORMATS.SHORT_DATE)).toEqual("03/03");
    });
  });
  describe("adjustDates", () => {
    it("Should adjust to start date", async () => {
      const date = new Date(100000010010);
      expect(date.getUTCMinutes()).toEqual(46);
      expect(date.getUTCSeconds()).toEqual(50);
      dateUtil.adjustToStartOfTheDay(date);
      expect(date.getUTCMinutes()).toEqual(0);
      expect(date.getUTCSeconds()).toEqual(0);
    });

    it("Should adjust to end of the day", async () => {
      const date = new Date(100000000002);
      expect(date.getUTCMinutes()).toEqual(46);
      expect(date.getUTCSeconds()).toEqual(40);
      dateUtil.adjustToEndOfTheDay(date);
      expect(date.getUTCMinutes()).toEqual(59);
      expect(date.getUTCSeconds()).toEqual(59);
    });
  });
  describe("differenceBy", () => {
    it("should return the difference in hours when no time unit is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2022-01-01T01:00:00.000Z");
      expect(dateUtil.differenceBy(undefined, endDate, startDate)).toEqual(1);
    });

    it("should return the difference in minutes when TimeUnit.minute is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2022-01-01T00:15:00.000Z");
      expect(dateUtil.differenceBy(TimeUnit.minute, endDate, startDate)).toEqual(15);
    });

    it("should return the difference in days when TimeUnit.day is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2022-01-02T00:00:00.000Z");
      expect(dateUtil.differenceBy(TimeUnit.day, endDate, startDate)).toEqual(1);
    });

    it("should return the difference in weeks when TimeUnit.week is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2022-01-08T00:00:00.000Z");
      expect(dateUtil.differenceBy(TimeUnit.week, endDate, startDate)).toEqual(1);
    });

    it("should return the difference in months when TimeUnit.month is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2022-02-01T00:00:00.000Z");
      expect(dateUtil.differenceBy(TimeUnit.month, endDate, startDate)).toEqual(1);
    });

    it("should return the difference in years when TimeUnit.year is provided", () => {
      const startDate = new Date("2022-01-01T00:00:00.000Z");
      const endDate = new Date("2023-01-01T00:00:00.000Z");
      expect(dateUtil.differenceBy(TimeUnit.year, endDate, startDate)).toEqual(1);
    });
  });

  describe("adjustDateBy", () => {
    it("should add the specified amount of hours to the date when TimeUnit.hour is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.hour, date, 1).getTime()).toEqual(
        new Date("2022-01-01T01:00:00.000Z").getTime(),
      );
    });

    it("should add the specified amount of minutes to the date when TimeUnit.minute is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.minute, date, 15).getTime()).toEqual(
        new Date("2022-01-01T00:15:00.000Z").getTime(),
      );
    });

    it("should add the specified amount of days to the date when TimeUnit.day is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.day, date, 1).getTime()).toEqual(
        new Date("2022-01-02T00:00:00.000Z").getTime(),
      );
    });

    it("should add the specified amount of weeks to the date when TimeUnit.week is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.week, date, 1).getTime()).toEqual(
        new Date("2022-01-08T00:00:00.000Z").getTime(),
      );
    });

    it("should add the specified amount of months to the date when TimeUnit.month is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.month, date, 1).getTime()).toEqual(
        new Date("2022-02-01T00:00:00.000Z").getTime(),
      );
    });

    it("should add the specified amount of years to the date when TimeUnit.year is provided", () => {
      const date = new Date("2022-01-01T00:00:00.000Z");
      expect(dateUtil.adjustDateBy(TimeUnit.year, date, 1).getTime()).toEqual(
        new Date("2023-01-01T00:00:00.000Z").getTime(),
      );
    });
  });

  describe("formatDuration", () => {
    test("returns the expected string for 0 seconds", () => {
      expect(dateUtil.formatDuration(0)).toBe("0d-00h-00m");
    });

    test("returns the expected string for 60 seconds", () => {
      expect(dateUtil.formatDuration(60)).toBe("0d-00h-01m");
    });

    test("returns the expected string for 3600 seconds", () => {
      expect(dateUtil.formatDuration(3600)).toBe("0d-01h-00m");
    });

    test("returns the expected string for 86400 seconds", () => {
      expect(dateUtil.formatDuration(86400)).toBe("1d-00h-00m");
    });

    test("returns the expected string for 90061 seconds", () => {
      expect(dateUtil.formatDuration(90061)).toBe("1d-01h-01m");
    });
  });
  describe("isDurationMoreThanMinutes", () => {
    it("should return true if the duration is more than the specified minutes", () => {
      const duration = 90061; // 1d-1h-1m
      expect(dateUtil.isDurationMoreThanMinute(duration)).toBe(true);
    });

    it("should return false if the duration is less than the specified minutes", () => {
      const duration = 30;
      expect(dateUtil.isDurationMoreThanMinute(duration)).toBe(false);
    });
  });
  describe("isDateString", () => {
    it("should return true for a valid date string", () => {
      const validDateString = "2023-12-21";
      expect(isDateString(validDateString)).toBe(true);
    });

    it("should return false for an invalid date string", () => {
      const invalidDateString = "This is not a date";
      expect(isDateString(invalidDateString)).toBe(false);
    });

    it("should return false for an empty string", () => {
      const emptyString = "";
      expect(isDateString(emptyString)).toBe(false);
    });
  });

  describe("convertDateToISO", () => {
    test("should convert a Date object to ISO 8601 format", () => {
      const date = new Date(2024, 2, 3, 12, 30, 45, 678); // March 3rd, 2024, 12:30:45.678
      const expectedISO = date.toISOString();
      expect(dateUtil.convertDateToISO(date)).toBe(expectedISO);
    });
  });

  describe("getUpdatedTime", () => {
    it("refreshes both bounds of a stale Last 7 days range", () => {
      const updated = dateUtil.getUpdatedTime({
        dateStart: new Date("2026-05-24T00:00:00.000Z"),
        dateEnd: new Date("2026-05-31T00:00:00.000Z"),
        label: PREDEFINED_DATES.LAST_7_DAYS,
      });

      expect(updated.dateEnd.getTime()).toBeGreaterThan(new Date("2026-05-31T00:00:00.000Z").getTime());
      expect(updated.dateEnd.getTime() - updated.dateStart.getTime()).toBe(7 * 24 * 60 * 60 * 1000);
    });
  });

  describe("disablePredefinedOptionsBeforeDate", () => {
    function getDaysAgoDate(daysAgo: number): Date {
      const today = new Date();
      const daysAgoDate = new Date(today);
      daysAgoDate.setDate(today.getDate() - daysAgo);
      return daysAgoDate;
    }

    it("should disable options with endDate before disableBeforeDate", () => {
      const disableBeforeDate = new Date();
      const disabledOptions = dateUtil.disablePredefinedOptionsBeforeDate(
        hoursGranularityPredefinedOptions,
        disableBeforeDate,
      );

      expect(disabledOptions.find((option) => option.label === "Last hour")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 6 hours")?.disabled).toBe(true);
      expect(disabledOptions.find((option) => option.label === "Today")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Yesterday")?.disabled).toBe(true);
      expect(disabledOptions.find((option) => option.label === "Last 7 days")?.disabled).toBe(true);
      expect(disabledOptions.find((option) => option.label === "Last 30 days")?.disabled).toBe(true);
    });

    it("should disable options with endDate before 6 days", () => {
      const sixDaysAgo = getDaysAgoDate(6);
      const disabledOptions = dateUtil.disablePredefinedOptionsBeforeDate(hoursGranularityPredefinedOptions, sixDaysAgo);

      expect(disabledOptions.find((option) => option.label === "Today")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Yesterday")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 7 days")?.disabled).toBe(true);
      expect(disabledOptions.find((option) => option.label === "Last 30 days")?.disabled).toBe(true);
    });

    it("should disable options with endDate before 15 days", () => {
      const fifteenDaysAgo = getDaysAgoDate(15);
      const disabledOptions = dateUtil.disablePredefinedOptionsBeforeDate(
        hoursGranularityPredefinedOptions,
        fifteenDaysAgo,
      );

      expect(disabledOptions.find((option) => option.label === "Today")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Yesterday")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 7 days")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 30 days")?.disabled).toBe(true);
    });

    it("should disable options with endDate before 40 days", () => {
      const fortyDaysAgo = getDaysAgoDate(40);
      const disabledOptions = dateUtil.disablePredefinedOptionsBeforeDate(
        hoursGranularityPredefinedOptions,
        fortyDaysAgo,
      );

      expect(disabledOptions.find((option) => option.label === "Today")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Yesterday")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 7 days")?.disabled).toBe(false);
      expect(disabledOptions.find((option) => option.label === "Last 30 days")?.disabled).toBe(false);
    });

    it("should not disable options if disableBeforeDate is after all", () => {
      const disableBeforeDate = new Date("2023-03-20");
      const disabledOptions = dateUtil.disablePredefinedOptionsBeforeDate(
        hoursGranularityPredefinedOptions,
        disableBeforeDate,
      );

      expect(disabledOptions.every((option) => !option.disabled)).toBe(true);
    });
  });

  describe(" dateUtil.formatTimestamp", () => {
    it("returns '0min' for 0 seconds", () => {
      expect(dateUtil.formatTimestamp(0)).toBe("0min");
    });

    it("returns '1min' for 60 seconds", () => {
      expect(dateUtil.formatTimestamp(60)).toBe("1min");
    });

    it("returns '1h' for 3600 seconds", () => {
      expect(dateUtil.formatTimestamp(3600)).toBe("1h");
    });

    it("returns '1d' for 86400 seconds", () => {
      expect(dateUtil.formatTimestamp(86400)).toBe("1d");
    });

    it("returns '1d 1h 1min' for 90061 seconds", () => {
      expect(dateUtil.formatTimestamp(90061)).toBe("1d 1h 1min");
    });

    it("returns '2d' for 172800 seconds", () => {
      expect(dateUtil.formatTimestamp(172800)).toBe("2d");
    });

    it("returns '1h 1min' for 3660 seconds", () => {
      expect(dateUtil.formatTimestamp(3660)).toBe("1h 1min");
    });

    it("returns '1d 1min' for 86460 seconds", () => {
      expect(dateUtil.formatTimestamp(86460)).toBe("1d 1min");
    });

    it("returns '1d 1h' for 90000 seconds", () => {
      expect(dateUtil.formatTimestamp(90000)).toBe("1d 1h");
    });
  });

  describe("formatDateRange", () => {
    it("should format the date range correctly with default format", () => {
      const from = "2024-01-01";
      const to = "2024-12-31";
      const result = dateUtil.formatDateRange(from, to);
      expect(result).toBe("01/01/2024 - 31/12/2024");
    });

    it("should format the date range correctly with custom format", () => {
      const from = "2024-01-01";
      const to = "2024-12-31";
      const formatString = DATE_FORMATS.STANDARD_DATE;
      const result = dateUtil.formatDateRange(from, to, formatString);
      expect(result).toBe("01/01/2024 - 31/12/2024");
    });
  });

  describe("rangeOfDatesCallback", () => {
    const maxPeriodInMs = 90 * 24 * 60 * 60 * 1000; // 90 days in milliseconds

    it("returns true if the start date is within the range and period", () => {
      const date = "2024-01-01";
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = true;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(true);
    });

    it("returns false if the start date is not within the range", () => {
      const date = "2024-04-01";
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = true;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(false);
    });

    it("returns false if the start date exceeds the maximum period", () => {
      const date = "2023-01-01";
      const start = "2023-01-01";
      const end = "2024-03-01";
      const isStart = true;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(false);
    });

    it("returns true if the end date is within the range and period", () => {
      const date = "2024-03-01";
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = false;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(true);
    });

    it("returns false if the end date is not within the range", () => {
      const date = "2023-12-01";
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = false;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(false);
    });

    it("returns false if the end date exceeds the maximum period", () => {
      const date = "2024-04-01";
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = false;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(false);
    });

    it("returns false if the end date is after the current date", () => {
      const date = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().split("T")[0]; // Tomorrow
      const start = "2024-01-01";
      const end = "2024-03-01";
      const isStart = false;
      expect(dateUtil.rangeOfDatesCallback(date, start, end, isStart, maxPeriodInMs)).toBe(false);
    });
  });

  describe("formatSecondsToDuration", () => {
    it("should return empty string for 0 seconds", () => {
      expect(dateUtil.formatSecondsToDuration(0)).toBe("");
    });

    it("should format seconds only", () => {
      expect(dateUtil.formatSecondsToDuration(30)).toBe("30s");
      expect(dateUtil.formatSecondsToDuration(59)).toBe("59s");
    });

    it("should format minutes only", () => {
      expect(dateUtil.formatSecondsToDuration(60)).toBe("1m");
      expect(dateUtil.formatSecondsToDuration(120)).toBe("2m");
      expect(dateUtil.formatSecondsToDuration(3540)).toBe("59m");
    });

    it("should format hours only", () => {
      expect(dateUtil.formatSecondsToDuration(3600)).toBe("1h");
      expect(dateUtil.formatSecondsToDuration(7200)).toBe("2h");
      expect(dateUtil.formatSecondsToDuration(82800)).toBe("23h");
    });

    it("should format days only", () => {
      expect(dateUtil.formatSecondsToDuration(86400)).toBe("1d");
      expect(dateUtil.formatSecondsToDuration(172800)).toBe("2d");
    });

    it("should format minutes and seconds", () => {
      expect(dateUtil.formatSecondsToDuration(90)).toBe("1m30s");
      expect(dateUtil.formatSecondsToDuration(3599)).toBe("59m59s");
    });

    it("should format hours and minutes", () => {
      expect(dateUtil.formatSecondsToDuration(3660)).toBe("1h1m");
      expect(dateUtil.formatSecondsToDuration(7260)).toBe("2h1m");
    });

    it("should format hours, minutes and seconds", () => {
      expect(dateUtil.formatSecondsToDuration(3661)).toBe("1h1m1s");
      expect(dateUtil.formatSecondsToDuration(3723)).toBe("1h2m3s");
    });

    it("should format days and hours", () => {
      expect(dateUtil.formatSecondsToDuration(90000)).toBe("1d1h");
      expect(dateUtil.formatSecondsToDuration(176400)).toBe("2d1h");
    });

    it("should format days, hours and minutes", () => {
      expect(dateUtil.formatSecondsToDuration(90060)).toBe("1d1h1m");
      expect(dateUtil.formatSecondsToDuration(266460)).toBe("3d2h1m");
    });

    it("should format all components", () => {
      expect(dateUtil.formatSecondsToDuration(90061)).toBe("1d1h1m1s");
      expect(dateUtil.formatSecondsToDuration(266521)).toBe("3d2h2m1s");
    });

    it("should handle large durations", () => {
      expect(dateUtil.formatSecondsToDuration(31536000)).toBe("365d"); // 1 year
      expect(dateUtil.formatSecondsToDuration(31622461)).toBe("366d1m1s"); // 1 year + 1 day + 1 minute + 1 second
    });
  });

  describe("parseDurationStringToSeconds", () => {
    it("should return 0 for empty string", () => {
      expect(dateUtil.parseDurationStringToSeconds("")).toBe(0);
    });

    it("should parse seconds only", () => {
      expect(dateUtil.parseDurationStringToSeconds("30s")).toBe(30);
      expect(dateUtil.parseDurationStringToSeconds("59s")).toBe(59);
    });

    it("should parse minutes only", () => {
      expect(dateUtil.parseDurationStringToSeconds("1m")).toBe(60);
      expect(dateUtil.parseDurationStringToSeconds("2m")).toBe(120);
      expect(dateUtil.parseDurationStringToSeconds("59m")).toBe(3540);
    });

    it("should parse hours only", () => {
      expect(dateUtil.parseDurationStringToSeconds("1h")).toBe(3600);
      expect(dateUtil.parseDurationStringToSeconds("2h")).toBe(7200);
      expect(dateUtil.parseDurationStringToSeconds("23h")).toBe(82800);
    });

    it("should parse days only", () => {
      expect(dateUtil.parseDurationStringToSeconds("1d")).toBe(86400);
      expect(dateUtil.parseDurationStringToSeconds("2d")).toBe(172800);
    });

    it("should parse minutes and seconds", () => {
      expect(dateUtil.parseDurationStringToSeconds("1m30s")).toBe(90);
      expect(dateUtil.parseDurationStringToSeconds("59m59s")).toBe(3599);
    });

    it("should parse hours and minutes", () => {
      expect(dateUtil.parseDurationStringToSeconds("1h1m")).toBe(3660);
      expect(dateUtil.parseDurationStringToSeconds("2h1m")).toBe(7260);
    });

    it("should parse hours, minutes and seconds", () => {
      expect(dateUtil.parseDurationStringToSeconds("1h1m1s")).toBe(3661);
      expect(dateUtil.parseDurationStringToSeconds("1h2m3s")).toBe(3723);
    });

    it("should parse days and hours", () => {
      expect(dateUtil.parseDurationStringToSeconds("1d1h")).toBe(90000);
      expect(dateUtil.parseDurationStringToSeconds("2d1h")).toBe(176400);
    });

    it("should parse days, hours and minutes", () => {
      expect(dateUtil.parseDurationStringToSeconds("1d1h1m")).toBe(90060);
      expect(dateUtil.parseDurationStringToSeconds("3d2h1m")).toBe(266460);
    });

    it("should parse all components", () => {
      expect(dateUtil.parseDurationStringToSeconds("1d1h1m1s")).toBe(90061);
      expect(dateUtil.parseDurationStringToSeconds("3d2h2m1s")).toBe(266521);
    });

    it("should handle components in any order", () => {
      expect(dateUtil.parseDurationStringToSeconds("1s1m1h1d")).toBe(90061);
      expect(dateUtil.parseDurationStringToSeconds("30s5m2h")).toBe(7530);
    });

    it("should handle missing components", () => {
      expect(dateUtil.parseDurationStringToSeconds("1d30m")).toBe(88200); // 1 day + 30 minutes
      expect(dateUtil.parseDurationStringToSeconds("2h30s")).toBe(7230); // 2 hours + 30 seconds
    });

    it("should handle large durations", () => {
      expect(dateUtil.parseDurationStringToSeconds("365d")).toBe(31536000); // 1 year
      expect(dateUtil.parseDurationStringToSeconds("365d24h1m1s")).toBe(31622461);
    });
  });

  describe("formatSecondsToDuration and parseDurationStringToSeconds integration", () => {
    it("should be inverse operations", () => {
      const testCases = [0, 30, 60, 90, 3600, 3661, 86400, 90061, 266521, 31536000];

      testCases.forEach((seconds) => {
        const formatted = dateUtil.formatSecondsToDuration(seconds);
        const parsed = dateUtil.parseDurationStringToSeconds(formatted);
        expect(parsed).toBe(seconds);
      });
    });

    it("should handle edge cases consistently", () => {
      // Test that formatting 0 seconds and parsing empty string both return 0
      expect(dateUtil.formatSecondsToDuration(0)).toBe("");
      expect(dateUtil.parseDurationStringToSeconds("")).toBe(0);

      // Test round trip for minimum non-zero values
      expect(dateUtil.parseDurationStringToSeconds(dateUtil.formatSecondsToDuration(1))).toBe(1);
      expect(dateUtil.parseDurationStringToSeconds(dateUtil.formatSecondsToDuration(60))).toBe(60);
      expect(dateUtil.parseDurationStringToSeconds(dateUtil.formatSecondsToDuration(3600))).toBe(3600);
      expect(dateUtil.parseDurationStringToSeconds(dateUtil.formatSecondsToDuration(86400))).toBe(86400);
    });
  });
});
