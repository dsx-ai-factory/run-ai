/**
 * Centralized date format constants for the application
 * These formats follow the date-fns format string convention
 *
 * Using European format (dd/MM/yyyy) as the standard
 */

export const DATE_FORMATS = {
  // Standard date formats (eg. 10/02/2025 or 10/02/2025 12:00 etc...)
  STANDARD_DATE: "dd/MM/yyyy",
  STANDARD_DATE_TIME: "dd/MM/yyyy, HH:mm",
  STANDARD_DATE_TIME_SECONDS: "dd/MM/yyyy, HH:mm:ss",
  STANDARD_DATE_TIME_MILLISECONDS: "dd/MM/yyyy, HH:mm:ss:SSS",

  // Short formats (eg. 10/02 or 10/02 12:00)
  SHORT_DATE: "dd/MM",
  SHORT_DATE_TIME: "dd/MM HH:mm",

  // Full date with month name (eg. Feb 10, 2025)
  FULL_DATE_MONTH_NAME: "MMM dd, yyyy",
} as const;

export type DateFormatType = (typeof DATE_FORMATS)[keyof typeof DATE_FORMATS];
