/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Utility functions and configuration for histogram column spacing.
 * Handles dynamic column width and padding based on time intervals.
 */

// Configuration for column spacing based on interval duration
export interface ColumnSpacingConfig {
  minIntervalMinutes: number;
  maxPointWidth: number;
  pointPadding: number;
  groupPadding: number;
  description: string;
}

export const COLUMN_SPACING_CONFIGS: ColumnSpacingConfig[] = [
  {
    minIntervalMinutes: 180,
    maxPointWidth: 80,
    pointPadding: 0.05,
    groupPadding: 0.05,
    description: "Extra long intervals (3+ hours): for 25-30 day ranges",
  },
  {
    minIntervalMinutes: 120,
    maxPointWidth: 65,
    pointPadding: 0.2,
    groupPadding: 0.2,
    description: "Very long intervals (2-3 hours): for 20-24 day ranges",
  },
  {
    minIntervalMinutes: 60,
    maxPointWidth: 55,
    pointPadding: 0.15,
    groupPadding: 0.15,
    description: "Long intervals (1-2 hours): for 7-20 day ranges",
  },
  {
    minIntervalMinutes: 30,
    maxPointWidth: 50,
    pointPadding: 0.18,
    groupPadding: 0.18,
    description: "Medium-long intervals (30-60 min): for 4-7 day ranges",
  },
  {
    minIntervalMinutes: 15,
    maxPointWidth: 40,
    pointPadding: 0.2,
    groupPadding: 0.2,
    description: "Medium intervals (15-30 min): for 3-6 day ranges",
  },
  {
    minIntervalMinutes: 10,
    maxPointWidth: 35,
    pointPadding: 0.2,
    groupPadding: 0.2,
    description: "Medium-short intervals (10-15 min): for 2-3 day ranges",
  },
  {
    minIntervalMinutes: 5,
    maxPointWidth: 30,
    pointPadding: 0.22,
    groupPadding: 0.22,
    description: "Short intervals (5-10 min): for 1-2 day ranges",
  },
  {
    minIntervalMinutes: 2,
    maxPointWidth: 24,
    pointPadding: 0.24,
    groupPadding: 0.24,
    description: "Very short intervals (2-5 min): for 12-24 hour ranges",
  },
  {
    minIntervalMinutes: 1,
    maxPointWidth: 18,
    pointPadding: 0.26,
    groupPadding: 0.26,
    description: "Ultra-short intervals (1-2 min): for 6-12 hour ranges",
  },
  {
    minIntervalMinutes: 0.5,
    maxPointWidth: 12,
    pointPadding: 0.28,
    groupPadding: 0.28,
    description: "Extreme intervals (30sec-1min): for 2-6 hour ranges",
  },
  {
    minIntervalMinutes: 0,
    maxPointWidth: 8,
    pointPadding: 0.3,
    groupPadding: 0.3,
    description: "Sub-30sec intervals: for sub-hour ranges",
  },
];

/**
 * Calculate time interval between data points in milliseconds
 * @param timeframes Array of timestamp strings
 * @returns The interval in milliseconds between the first two timeframes, or undefined if less than 2 timeframes
 */
export function calculatePointInterval(timeframes: Array<string>): number | undefined {
  if (timeframes.length < 2) return undefined;

  const firstTime = new Date(timeframes[0]).getTime();
  const secondTime = new Date(timeframes[1]).getTime();
  return secondTime - firstTime;
}

/**
 * Convert milliseconds to minutes
 * @param milliseconds Time in milliseconds
 * @returns Time in minutes
 */
export function millisecondsToMinutes(milliseconds: number): number {
  return milliseconds / (1000 * 60);
}

/**
 * Get spacing configuration based on time interval
 * @param intervalInMinutes Time interval in minutes
 * @returns Configuration object with maxPointWidth, pointPadding, and groupPadding
 */
export function getColumnSpacingConfig(intervalInMinutes: number): {
  maxPointWidth: number;
  pointPadding: number;
  groupPadding: number;
} {
  const config = COLUMN_SPACING_CONFIGS.find((cfg) => intervalInMinutes >= cfg.minIntervalMinutes);

  // Fallback to the last config (smallest interval)
  return config || COLUMN_SPACING_CONFIGS[COLUMN_SPACING_CONFIGS.length - 1];
}
