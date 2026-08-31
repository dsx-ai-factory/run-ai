/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  format,
  parse,
  addMinutes,
  addHours,
  addDays,
  addWeeks,
  addMonths,
  addYears,
  differenceInMinutes,
  differenceInHours,
  differenceInDays,
  differenceInWeeks,
  differenceInMonths,
  differenceInYears,
  endOfDay,
  isToday,
  setSeconds,
  startOfDay,
} from "date-fns";
import {
  PREDEFINED_DATES,
  type DateTimeFormatOptions,
  type PredefinedOption,
  PREDEFINED_HOURS,
  PREDEFINED_MINUTES,
} from "@/models/date.model";
import { DATE_FORMATS, type DateFormatType } from "@/constants/date-formats";

export enum TimeUnit {
  minute = 0,
  hour = 1,
  day = 2,
  week = 3,
  month = 4,
  year = 5,
}

export const defaultPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_DATES.TODAY, daysCount: 0 },
  {
    label: PREDEFINED_DATES.YESTERDAY,
    daysCount: 1,
    hoursCount: 0,
    endDate: adjustDateBy(TimeUnit.day, new Date(), -1),
  },
  { label: PREDEFINED_DATES.LAST_7_DAYS, daysCount: 7 },
  { label: PREDEFINED_DATES.LAST_30_DAYS, daysCount: 30 },
];

export const hoursGranularityPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_HOURS.LAST_HOUR, hoursCount: 1 },
  { label: PREDEFINED_HOURS.LAST_6_HOURS, hoursCount: 6 },
  ...defaultPredefinedOptions,
];

export const extendedDefaultPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_DATES.TODAY, hoursCount: 0 },
  {
    label: PREDEFINED_DATES.YESTERDAY,
    daysCount: 1,
    endDate: adjustDateBy(TimeUnit.day, new Date(), -1),
  },
  { label: PREDEFINED_DATES.LAST_7_DAYS, daysCount: 7 },
  { label: PREDEFINED_DATES.LAST_14_DAYS, daysCount: 14 },
  { label: PREDEFINED_DATES.LAST_30_DAYS, daysCount: 30 },
];

export const extendedHoursPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_MINUTES.LAST_15_MINUTES, minutesCount: 15 },
  { label: PREDEFINED_HOURS.LAST_3_HOURS, hoursCount: 3 },
  { label: PREDEFINED_HOURS.LAST_6_HOURS, hoursCount: 6 },
  { label: PREDEFINED_HOURS.LAST_24_HOURS, hoursCount: 24 },
  ...extendedDefaultPredefinedOptions,
];

export const minutesGranularityPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_MINUTES.LAST_MINUTE, minutesCount: 1 },
  { label: PREDEFINED_MINUTES.LAST_5_MINUTES, minutesCount: 5 },
  { label: PREDEFINED_MINUTES.LAST_15_MINUTES, minutesCount: 15 },
];

export const workloadsLifespanPredefinedOptions: Array<PredefinedOption> = [
  { label: PREDEFINED_DATES.WORKLOADS_LIFESPAN },
];

export const sinceFirstRunPredefinedOptions: Array<PredefinedOption> = [{ label: PREDEFINED_DATES.SINCE_FIRST_RUN }];

const MS_IN_SECOND = 1000;

export const dateUtil = {
  adjustDateBy,
  adjustToEndOfTheDay,
  adjustToStartOfTheDay,
  dateAndTimeFormat,
  dateFormat,
  differenceBy,
  parseToDate,
  timeAgo,
  moreThanHourFromNow,
  isSameDay,
  formatDuration,
  getEpochDates,
  dateToTimestamp,
  convertDateToISO,
  dateSevenDaysAgo,
  getEndOfDay,
  getStartOfDay,
  isDateIsToday,
  setSecondsToDate,
  disablePredefinedOptionsBeforeDate,
  getUpdatedTime,
  formatTimestamp,
  isDurationMoreThanMinute,
  formatDateRange,
  rangeOfDatesCallback,
  formatSecondsToDuration,
  parseDurationStringToSeconds,
  formatTimeRemaining,
};

// TODO: need to move it to date.model.
// now I can do that becuase date.model have date.util import and we
// need to import this type to date.util (from the date.model) - will
// have circular dependencies
export interface IRangeDates {
  dateStart: Date;
  dateEnd: Date;
  label: string;
}
// add the type for formatString to prevent other not defined formats
function dateFormat(date: Date, formatString: DateFormatType = DATE_FORMATS.STANDARD_DATE): string {
  if (!date) return "";
  return format(date, formatString);
}

function convertDateToISO(date: Date): string {
  return date.toISOString();
}

function dateAndTimeFormat(date: Date, options: DateTimeFormatOptions = {}): string {
  if (!date) return "";

  const { includeSeconds = false, includeMilliseconds = false } = options;

  if (includeMilliseconds) {
    return format(date, DATE_FORMATS.STANDARD_DATE_TIME_MILLISECONDS);
  } else if (includeSeconds) {
    return format(date, DATE_FORMATS.STANDARD_DATE_TIME_SECONDS);
  }

  return format(date, DATE_FORMATS.STANDARD_DATE_TIME);
}
function adjustToStartOfTheDay(date: Date): void {
  date.setHours(0, 0, 0, 0);
}

function adjustToEndOfTheDay(date: Date): void {
  date.setHours(23, 59, 59, 999);
}

function parseToDate(dateToParse: string, format: DateFormatType): Date | null {
  if (!dateToParse) return null;
  return parse(dateToParse, format, new Date());
}

function differenceBy(timeUnit: TimeUnit = TimeUnit.hour, endDate: Date | number, startDate: Date | number): number {
  if (!startDate || !endDate) return 0;
  switch (timeUnit) {
    case TimeUnit.minute:
      return differenceInMinutes(endDate, startDate);
    case TimeUnit.day:
      return differenceInDays(endDate, startDate);
    case TimeUnit.week:
      return differenceInWeeks(endDate, startDate);
    case TimeUnit.month:
      return differenceInMonths(endDate, startDate);
    case TimeUnit.year:
      return differenceInYears(endDate, startDate);
    default:
      return differenceInHours(endDate, startDate);
  }
}

function adjustDateBy(timeUnit: TimeUnit, date: Date | number, amount: number): Date {
  switch (timeUnit) {
    case TimeUnit.minute:
      return addMinutes(date, amount);
    case TimeUnit.hour:
      return addHours(date, amount);
    case TimeUnit.day:
      return addDays(date, amount);
    case TimeUnit.week:
      return addWeeks(date, amount);
    case TimeUnit.month:
      return addMonths(date, amount);
    case TimeUnit.year:
      return addYears(date, amount);
    default:
      return addHours(date, amount);
  }
}

// TODO: Convert this to date-fns
function timeAgo(timestamp: number | string | Date): string {
  const date = new Date(timestamp);
  let seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000);
  let minutes = Math.floor(seconds / 60);
  let hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  // compute residual after removing whole days, hours and minutes
  seconds = seconds % 60;
  minutes = minutes % 60;
  hours = hours % 24;

  let ret = "";
  if (days > 0) {
    ret += days + "d";
  }
  if (hours > 0 || (days > 0 && minutes > 0)) {
    if (ret) {
      ret += "-";
    }
    ret += hours + "h";
  }
  if (minutes > 0) {
    if (ret) {
      ret += "-";
    }
    ret += minutes + "m";
  }
  if (!ret) {
    ret += seconds + "s";
  }
  return ret;
}

function moreThanHourFromNow(creationTimestamp: number): boolean {
  if (!creationTimestamp) return false;

  const oneHourInMs = 3600000;
  const dateNow: number = new Date().getTime();
  return dateNow - Number(creationTimestamp) > oneHourInMs;
}

function isDurationMoreThanMinute(durationInSeconds: number): boolean {
  return durationInSeconds > 60;
}

function isSameDay(timestamp1: number, timestamp2: number): boolean {
  const date1: Date = new Date(timestamp1);
  const date2: Date = new Date(timestamp2);
  return (
    date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
  );
}

function formatDuration(seconds: number, includeSeconds = false): string {
  const date = new Date(seconds * 1000);
  const days = Math.floor(seconds / (3600 * 24));
  const hours = date.getUTCHours().toString().padStart(2, "0");
  const minutes = date.getUTCMinutes().toString().padStart(2, "0");
  if (!includeSeconds) return `${days}d-${hours}h-${minutes}m`;

  const secondsPart = `-${date.getUTCSeconds().toString().padStart(2, "0")}s`;
  return `${days}d-${hours}h-${minutes}m${secondsPart}`;
}

/**
 * Formats a given duration in seconds into a human-readable string.
 *
 * The format will include days, hours, and minutes if applicable.
 * For example, 90061 seconds will be formatted as "1d 1h 1min".
 *
 * @param {number} seconds - The duration in seconds to format.
 * @returns {string} - The formatted duration string.
 */
function formatTimestamp(seconds: number): string {
  const days = Math.floor(seconds / (24 * 60 * 60));
  seconds %= 24 * 60 * 60;
  const hours = Math.floor(seconds / (60 * 60));
  seconds %= 60 * 60;
  const minutes = Math.floor(seconds / 60);

  let result = "";
  if (days > 0) result += `${days}d `;
  if (hours > 0) result += `${hours}h `;
  if (minutes > 0) result += `${minutes}min`;
  if (result === "") result = "0min";
  return result.trim();
}

function getEpochDates(start: string, end: string): { startDate: number; endDate: number } {
  const startDateObj = new Date(start);
  const endDateObj = new Date(end);

  // Normalize to UTC
  const startDateUTC = new Date(
    Date.UTC(
      startDateObj.getUTCFullYear(),
      startDateObj.getUTCMonth(),
      startDateObj.getUTCDate(),
      startDateObj.getUTCHours(),
      startDateObj.getUTCMinutes(),
      startDateObj.getUTCSeconds(),
    ),
  );
  const endDateUTC = new Date(
    Date.UTC(
      endDateObj.getUTCFullYear(),
      endDateObj.getUTCMonth(),
      endDateObj.getUTCDate(),
      endDateObj.getUTCHours(),
      endDateObj.getUTCMinutes(),
      endDateObj.getUTCSeconds(),
    ),
  );

  const startDate: number = Math.floor(startDateUTC.getTime() / 1000);
  const endDate: number = Math.floor(endDateUTC.getTime() / 1000);

  return { startDate, endDate };
}

export function isDateString(str: string): boolean {
  const date: Date = new Date(str);
  return !isNaN(date.getTime());
}
function dateToTimestamp(date?: string): number {
  return +new Date(date || 0);
}

function dateSevenDaysAgo(): Date {
  return dateUtil.adjustDateBy(TimeUnit.day, new Date(), -7);
}

function getEndOfDay(date: Date): Date {
  return endOfDay(date);
}

function getStartOfDay(date: Date): Date {
  return startOfDay(date);
}

function isDateIsToday(date: Date): boolean {
  return isToday(date);
}

function setSecondsToDate(date: Date, seconds: number): Date {
  return setSeconds(date, seconds);
}

function disablePredefinedOptionsBeforeDate(options: PredefinedOption[], creationDate: Date): PredefinedOption[] {
  const now = new Date();
  return options.map((option: PredefinedOption) => {
    const hoursDiff = Math.abs(differenceBy(TimeUnit.hour, now, creationDate));
    const daysDiff = Math.abs(differenceBy(TimeUnit.day, now, creationDate));
    const midnight = getEndOfDay(new Date(now.getDate() - 1));

    if (option.hoursCount === 1 || option.daysCount === 0) {
      //LAST_HOUR, TODAY
      return { ...option, disabled: false };
    } else if (option.daysCount === 1 && creationDate < midnight) {
      //YESTERDAY
      return { ...option, disabled: false };
    } else if (option.hoursCount != undefined && option.hoursCount > 0) {
      //LAST_3_HOURS, LAST_6_HOURS, LAST_24_HOURS
      return { ...option, disabled: hoursDiff < option.hoursCount };
    } else if (option.daysCount != undefined && option.daysCount > 0) {
      //LAST_7_DAYS, LAST_14_DAYS, LAST_30_DAYS
      return { ...option, disabled: daysDiff < option.daysCount };
    } else {
      return { ...option, disabled: false };
    }
  });
}

function getUpdatedTime(dateRange: IRangeDates): IRangeDates {
  switch (dateRange.label) {
    case PREDEFINED_DATES.LAST_14_DAYS:
    case PREDEFINED_DATES.LAST_7_DAYS:
    case PREDEFINED_DATES.LAST_30_DAYS: {
      const daysCount =
        dateRange.label === PREDEFINED_DATES.LAST_7_DAYS
          ? 7
          : dateRange.label === PREDEFINED_DATES.LAST_14_DAYS
            ? 14
            : 30;
      const dateEnd = new Date();
      return {
        ...dateRange,
        dateStart: adjustDateBy(TimeUnit.day, dateEnd, -daysCount),
        dateEnd,
      };
    }
    case PREDEFINED_DATES.TODAY: {
      return { ...dateRange, dateEnd: new Date() };
    }
    case PREDEFINED_DATES.WORKLOADS_LIFESPAN:
    case PREDEFINED_DATES.SINCE_FIRST_RUN: {
      return dateRange;
    }

    case PREDEFINED_DATES.CUSTOM: {
      if (dateUtil.isDateIsToday(dateRange.dateEnd)) {
        return { ...dateRange, dateEnd: new Date() };
      }
      return dateRange;
    }

    case PREDEFINED_MINUTES.LAST_15_MINUTES:
    case PREDEFINED_HOURS.LAST_24_HOURS:
    case PREDEFINED_HOURS.LAST_3_HOURS:
    case PREDEFINED_HOURS.LAST_6_HOURS:
    case PREDEFINED_HOURS.LAST_HOUR: {
      const sub: number = dateRange.dateEnd.getTime() - dateRange.dateStart.getTime();
      const dateEnd: Date = new Date();
      return { ...dateRange, dateEnd, dateStart: new Date(dateEnd.getTime() - sub) };
    }

    case PREDEFINED_DATES.YESTERDAY:
    default: {
      return dateRange;
    }
  }
}

/**
 * Formats a date range into a string.
 *
 * This function takes two date strings and a format string, and returns a formatted string
 * representing the date range. The default format is "dd/MM/yyyy".
 *
 * @param {string} from - The start date of the range.
 * @param {string} to - The end date of the range.
 * @param {string} [formatString=DATE_FORMATS.STANDARD_DATE] - The format string to use for formatting the dates.
 * @returns {string} - The formatted date range string.
 */
function formatDateRange(from: string, to: string, formatString: DateFormatType = DATE_FORMATS.STANDARD_DATE): string {
  const startDate = format(new Date(from), formatString);
  const endDate = format(new Date(to), formatString);
  return `${startDate} - ${endDate}`;
}

/**
 * Callback function to validate a date range.
 *
 * This function checks if a given date falls within a specified date range.
 * It also ensures that the date range does not exceed a maximum period.
 *
 * @param {string} date - The date to validate.
 * @param {string} start - The start date of the range.
 * @param {string} end - The end date of the range.
 * @param {boolean} isStart - A flag indicating if the date is the start date of the range.
 * @param {number} maxPeriodInMs - The maximum allowed period in milliseconds.
 * @returns {boolean} - Returns true if the date is valid within the range and period, otherwise false.
 */
function rangeOfDatesCallback(
  date: string,
  start: string,
  end: string,
  isStart: boolean,
  maxPeriodInMs: number,
): boolean {
  const selectedDate = new Date(date);
  const startDate = new Date(start);
  const endDate = new Date(end);
  const now = new Date();

  if (isStart) {
    const isBeforeEndDate = selectedDate < endDate;
    const isWithinNinetyDays = endDate.getTime() - selectedDate.getTime() < maxPeriodInMs;
    return isBeforeEndDate && isWithinNinetyDays;
  } else {
    const isAfterStartDate = selectedDate > startDate;
    const isBeforeOrEqualNow = selectedDate <= now;
    const isWithinNinetyDays = selectedDate.getTime() - startDate.getTime() < maxPeriodInMs;
    return isAfterStartDate && isBeforeOrEqualNow && isWithinNinetyDays;
  }
}

/**
 * Formats seconds into a compact duration string (e.g., "2d5h35m")
 * @param {number} totalSeconds - The total number of seconds to format.
 */
function formatSecondsToDuration(totalSeconds: number): string {
  const SECONDS_IN_A_MINUTE = 60;
  const SECONDS_IN_AN_HOUR = SECONDS_IN_A_MINUTE * 60;
  const SECONDS_IN_A_DAY = SECONDS_IN_AN_HOUR * 24;

  if (totalSeconds === 0) return "";

  const days = Math.floor(totalSeconds / SECONDS_IN_A_DAY);
  totalSeconds %= SECONDS_IN_A_DAY;
  const hours = Math.floor(totalSeconds / SECONDS_IN_AN_HOUR);
  totalSeconds %= SECONDS_IN_AN_HOUR;
  const minutes = Math.floor(totalSeconds / SECONDS_IN_A_MINUTE);
  const seconds = totalSeconds % SECONDS_IN_A_MINUTE;

  let result = "";
  if (days > 0) result += `${days}d`;
  if (hours > 0) result += `${hours}h`;
  if (minutes > 0) result += `${minutes}m`;
  if (seconds > 0) result += `${seconds}s`;

  return result;
}

/**
 * Parses a compact duration string (e.g., "2d5h35m") to total seconds
 * @param {string} durationString - The duration string to parse.
 */
function parseDurationStringToSeconds(durationString: string): number {
  const SECONDS_IN_A_MINUTE = 60;
  const SECONDS_IN_AN_HOUR = SECONDS_IN_A_MINUTE * 60;
  const SECONDS_IN_A_DAY = SECONDS_IN_AN_HOUR * 24;

  if (!durationString) return 0;

  let totalSeconds = 0;

  // Match days (e.g., "2d")
  const daysMatch = durationString.match(/(\d+)d/);
  if (daysMatch) {
    totalSeconds += parseInt(daysMatch[1]) * SECONDS_IN_A_DAY;
  }

  // Match hours (e.g., "5h")
  const hoursMatch = durationString.match(/(\d+)h/);
  if (hoursMatch) {
    totalSeconds += parseInt(hoursMatch[1]) * SECONDS_IN_AN_HOUR;
  }

  // Match minutes (e.g., "35m")
  const minutesMatch = durationString.match(/(\d+)m/);
  if (minutesMatch) {
    totalSeconds += parseInt(minutesMatch[1]) * SECONDS_IN_A_MINUTE;
  }

  // Match seconds (e.g., "10s")
  const secondsMatch = durationString.match(/(\d+)s/);
  if (secondsMatch) {
    totalSeconds += parseInt(secondsMatch[1]);
  }

  return totalSeconds;
}

/**
 * Calculates and formats the time remaining until a future timestamp.
 * Returns "-" if the timestamp is null, in the past, or less than a minute away.
 *
 * @param endTimestamp - ISO timestamp string or null representing the end time
 * @returns Formatted duration string (e.g., "1d-02h-30m") or "-"
 */
function formatTimeRemaining(endTimestamp: string): string {
  const now = new Date();
  const endsAt = new Date(endTimestamp);
  const diffInSeconds = Math.max(0, Math.floor((endsAt.getTime() - now.getTime()) / MS_IN_SECOND));

  if (!isDurationMoreThanMinute(diffInSeconds)) return "-";
  return formatDuration(diffInSeconds);
}
