/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const formatUtil = {
  numberFormat,
  memoryFormat,
  convertToBytes,
  parseSizeString,
  roundToDecimal,
  toPercent,
  formatToFixedOrZero,
  numberWithPercentFormat,
  onlyPercentFormat,
  idleGpuDurationFormat,
  cpuParser,
  memoryParser,
  formatK8sResourceMemory,
  formatNumberWithCommas,
  convertBytesToGB,
  formatBytesRate,
  formatCompactNumber,
};

export function numberFormat(val: string | number, formatOnlyIfFloat = false, suffix?: string): string {
  const isInt = Number.isInteger(Number(val));
  if (formatOnlyIfFloat && isInt) {
    return val.toString();
  }
  const numberVal = Number(val);
  if (isNaN(numberVal) && val !== "NA") return "-";
  if (["-1", -1, "NA"].includes(val)) return "N/A";
  let formattedVal: string = numberVal.toFixed(2);
  if (suffix) formattedVal += suffix;
  return formattedVal;
}

export function memoryFormat(bytes: number | string | undefined): string {
  const k = 1000;
  const dm = 2;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  if (bytes == "-" || bytes === undefined) return "-";
  const val = Number(bytes);
  if (!val || val === 0) return "0 Bytes";
  if (val < 1) return "";

  const i = Math.floor(Math.log(val) / Math.log(k));

  return parseFloat((val / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

/**
 * Converts a numerical value from a specified unit to bytes.
 * @param {number} number - The numerical value to convert.
 * @param {string} unit - The unit of measurement for the input number.
 *                        Valid units: "B", "K", "KB", "M", "MB", "G", "GB",
 *                        "T", "TB", "P", "PB", "EB", "ZB", "YB".
 * @returns {number} The converted value in bytes.
 * @throws {Error} If an invalid unit is provided.
 */
export function convertToBytes(number: number, unit: string): number {
  const units: Record<string, number> = {
    B: 1,
    K: Math.pow(1000, 1),
    KB: Math.pow(1000, 1),
    M: Math.pow(1000, 2),
    MB: Math.pow(1000, 2),
    G: Math.pow(1000, 3),
    GB: Math.pow(1000, 3),
    T: Math.pow(1000, 4),
    TB: Math.pow(1000, 4),
    P: Math.pow(1000, 5),
    PB: Math.pow(1000, 5),
    EB: Math.pow(1000, 6),
    ZB: Math.pow(1000, 7),
    YB: Math.pow(1000, 8),
  };

  const unitMultiplier = units[unit.toUpperCase()];
  if (unitMultiplier === undefined) {
    throw new Error(`Invalid unit: ${unit}`);
  }

  return number * unitMultiplier;
}

/**
 * Parses a size string with unit into a tuple containing the size and unit.
 * @param sizeString The size string to parse.
 * @returns A tuple containing the size (number) and unit (string), or null if the input is invalid.
 */
export function parseSizeString(sizeString: string): [number, string] | null {
  if (!sizeString?.trim()) return null;

  // Regular expression to match size and unit
  const regex = /^(\d+(?:\.\d+)?)([KMGTPEZY]B?)$/i;
  const matches = sizeString.trim().match(regex);

  if (!matches) return null;

  // Extract size and unit from the matches
  const size = parseFloat(matches[1]);
  const unit = matches[2].trim().toUpperCase(); // Convert unit to uppercase for consistency

  if (isNaN(size) || unit === "") return null;

  // Return a tuple containing size and unit
  return [size, unit];
}

/**
 * Rounds a number to a specified number of decimal places.
 * @param num The number to round.
 * @param maxDecimals The maximum number of decimal places (default is 2).
 * @returns The rounded number.
 */
export function roundToDecimal(num: number, maxDecimals = 2): number {
  return num % 1 === 0 ? num : parseFloat(num.toFixed(maxDecimals));
}

export function toPercent(val: string | number, compareVal: string | number): string {
  if (+compareVal === 0) {
    return "-";
  }
  const calculatedVal: number = (+val / +compareVal) * 100;
  return `${calculatedVal.toFixed(2)}%`;
}

export function formatToFixedOrZero(num: number): number {
  const fixedNumber: string = num.toFixed(2);
  return fixedNumber === "0.00" || fixedNumber === "-0.00" ? 0 : parseFloat(fixedNumber);
}

/**
 *
 * @param val
 * @param compareVal
 * @param isMemory
 * @returns string format containt the size + percentage
 *
 * examples:
 * input: numberWithPercentFormat(2097152000, 4194304000, true)
 * output: "2.1 GB (50.00%)"
 * input: numberWithPercentFormat(1, 2)
 * output: "1.00 (50.00%)"
 */
export function numberWithPercentFormat(val: number, compareVal: number, isMemory = false): string {
  const formattedVal: string = isMemory ? memoryFormat(val) : numberFormat(val);
  if (["-", "0.00", "0 Bytes"].includes(formattedVal)) return formattedVal;
  const formattedCalcVal = toPercent(val, compareVal as number | string);
  return `${formattedVal} (${formattedCalcVal})`;
}

export function onlyPercentFormat(val: number | null, compareVal: number): string {
  if (val == null) {
    return "-";
  }
  const formattedCalcVal = toPercent(val, compareVal as number | string);
  return `${formattedCalcVal}`;
}

export function idleGpuDurationFormat(allSeconds: number | undefined): string {
  if (typeof allSeconds === "undefined" || allSeconds < 45) {
    return "-";
  }
  const seconds = Math.floor(allSeconds % 60);
  const allMinutes = allSeconds / 60;
  const minutes = Math.floor(allMinutes % 60);
  const hours = Math.floor(allMinutes / 60);
  return `${doubleDigit(hours)}:${doubleDigit(minutes)}:${doubleDigit(seconds)}`;
}

function doubleDigit(num: number): string {
  return num > 9 ? num.toString() : "0" + num;
}

//kubernetes-resource-parser
export function cpuParser(input: string): number {
  const milliMatch = input.match(/^([0-9]+)m$/);
  if (milliMatch) {
    const value = parseInt(milliMatch[1], 10);
    return value / 1000;
  }

  return parseFloat(input);
}

export function memoryParser(input: string): number {
  const memoryMultipliers: { [key: string]: number } = {
    k: 1000,
    M: 1000 ** 2,
    G: 1000 ** 3,
    T: 1000 ** 4,
    P: 1000 ** 5,
    E: 1000 ** 6,
    Ki: 1024,
    Mi: 1024 ** 2,
    Gi: 1024 ** 3,
    Ti: 1024 ** 4,
    Pi: 1024 ** 5,
    Ei: 1024 ** 6,
  };
  const unitMatch = input.match(/^([0-9]+)([A-Za-z]{1,2})$/);
  if (unitMatch) {
    return parseInt(unitMatch[1], 10) * memoryMultipliers[unitMatch[2]];
  }

  return parseInt(input, 10);
}

export function formatK8sResourceMemory(input: string | number): string {
  const memory = typeof input === "string" ? input : input.toString();
  return memoryFormat(memoryParser(memory));
}

/**
 * Formats a number by adding commas as thousand separators.
 * e.g: 1000000 -> 1,000,000
 * @param {number} val - The number to format.
 * @returns {string} The formatted number with commas as thousand separators.
 */
function formatNumberWithCommas(val: number): string {
  return val.toLocaleString();
}

/**
 * Converts a value in bytes to megabytes (GB) and formats it to two decimal places.
 * @param value
 */
function convertBytesToGB(value: string | number): string {
  const bytes = typeof value === "string" ? parseFloat(value) : value;
  return (bytes / (1024 * 1024 * 1024)).toFixed(2);
}

/**
 * Formats a bytes/sec value using the largest fitting unit (GB/s, MB/s, KB/s, B/s).
 * Mirrors the 1024-based scaling used by convertBytesToGB.
 */
function formatBytesRate(value: string | number): string {
  const bytes = typeof value === "string" ? parseFloat(value) : value;
  if (!isFinite(bytes)) return "0 B/s";
  const abs = Math.abs(bytes);
  const GB = 1024 ** 3;
  const MB = 1024 ** 2;
  const KB = 1024;
  if (abs >= GB) return `${(bytes / GB).toFixed(2)} GB/s`;
  if (abs >= MB) return `${(bytes / MB).toFixed(2)} MB/s`;
  if (abs >= KB) return `${(bytes / KB).toFixed(2)} KB/s`;
  return `${Math.round(bytes)} B/s`;
}

/**
 * Formats a number for display using social media style abbreviations (k, m)
 * e.g: 1500 -> "1.5k", 2300000 -> "2.3m"
 * @param {number} num - The number to format.
 * @returns {string} Formatted string with 'k' for thousands or 'm' for millions.
 */
export function formatCompactNumber(num?: number): string {
  if (!num) return "";
  const thousand = Math.pow(1000, 1);
  const million = Math.pow(1000, 2);

  if (num >= million) {
    return `${(num / million).toFixed(1)}m`;
  }
  if (num >= thousand) {
    return `${(num / thousand).toFixed(1)}k`;
  }
  return num.toString();
}
