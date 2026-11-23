/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { roundToDecimal } from "@/utils/format.util/index";
export type SchedulingDurationFields = {
  Days: number;
  Hours: number;
  Minutes: number;
  Seconds: number;
};

const secondsPerUnit: Record<string, number> = {
  w: 604800,
  d: 86400,
  h: 3600,
  m: 60,
  s: 1,
  ms: 0.001,
  us: 0.000001,
  µs: 0.000001,
  μs: 0.000001,
  ns: 0.000000001,
};
// Match the API's duration units as complete tokens, including fractional values.
const durationToken = /(\d[\d.]*)(ns|us|µs|μs|ms|s|m|h|d|w)/guy;
const roundSeconds = (value: number) => roundToDecimal(value, 9);

export function schedulingDurationSeconds(value: string | null | undefined): number {
  if (!value || value === "0") return 0;
  const source = value.startsWith("+") ? value.slice(1) : value;
  let offset = 0;
  let seconds = 0;
  for (const match of source.matchAll(durationToken)) {
    if (match.index !== offset || match[1].endsWith(".")) throw new Error("Enter a valid nonnegative duration.");
    seconds += Number(match[1]) * secondsPerUnit[match[2]];
    offset += match[0].length;
  }
  if (offset !== source.length || !offset || !Number.isFinite(seconds))
    throw new Error("Enter a valid nonnegative duration.");
  return roundSeconds(seconds);
}

export function schedulingDurationToFields(
  value: string | null | undefined,
  includeSeconds = true,
): SchedulingDurationFields {
  const seconds = schedulingDurationSeconds(value);
  return {
    Days: Math.floor(seconds / 86400),
    Hours: Math.floor((seconds % 86400) / 3600),
    Minutes: includeSeconds ? Math.floor((seconds % 3600) / 60) : roundSeconds(seconds % 3600) / 60,
    Seconds: includeSeconds ? roundSeconds(seconds % 60) : 0,
  };
}

function decimal(value: number) {
  const [coefficient, exponent] = String(value).split("e");
  if (exponent === undefined) return coefficient;
  const [whole, fraction = ""] = coefficient.split(".");
  const digits = whole + fraction;
  const position = whole.length + Number(exponent);
  if (position <= 0) return `0.${"0".repeat(-position)}${digits}`;
  if (position >= digits.length) return digits + "0".repeat(position - digits.length);
  return `${digits.slice(0, position)}.${digits.slice(position)}`;
}

export function serializeSchedulingDuration(
  fields: Record<string, number | null> | undefined,
  includeSeconds = true,
): string | undefined {
  if (!fields) return undefined;
  const units: [number | null | undefined, string, number][] = [
    [fields.Days, "d", 86400],
    [fields.Hours, "h", 3600],
    [fields.Minutes, "m", 60],
    [fields.Seconds, "s", 1],
  ];
  let seconds = 0;
  let result = "";
  for (const [input, unit, multiplier] of units) {
    const value = input ?? 0;
    if (!Number.isFinite(value) || value < 0) throw new Error("Enter a valid nonnegative duration.");
    if (!includeSeconds && unit === "s" && value !== 0)
      throw new Error("Use days, hours, or minutes for the historical usage window; seconds are not supported.");
    seconds += value * multiplier;
    if (value && (includeSeconds || unit !== "s")) result += `${decimal(value)}${unit}`;
  }
  if (!Number.isFinite(seconds) || seconds > Number.MAX_SAFE_INTEGER) throw new Error("The duration is too large.");
  return result || "0";
}
