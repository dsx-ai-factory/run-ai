/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents a single segment in the stacked bar chart
 */
export interface IStackedBarSegment {
  /** The numeric value of this segment */
  value: number;
  /** The color for this segment (hex, rgb, or named color) */
  color: string;
  /** Optional label to display on hover or in tooltips */
  label?: string;
  /** Optional custom text to display on the segment instead of numeric value */
  displayValue?: string;
  /** If true, this segment will not appear in the tooltip (useful for placeholder bars) */
  hideFromTooltip?: boolean;
}

/**
 * Props interface for the StackedBarChart component
 */
export interface IStackedBarChartProps {
  /** Array of segments to display in the chart */
  segments: IStackedBarSegment[];
  /** Height of the chart in pixels (default: 40) */
  height?: number;
  /** Width of the chart in pixels (default: auto/100%) */
  width?: number;
  /** Whether to show value labels on segments (default: true) */
  showLabels?: boolean;
  /** Minimum value for the scale (default: 0) */
  minValue?: number;
  /** Maximum value for the scale (default: 100) */
  maxValue?: number;
  /** Whether to show the scale labels below the chart (default: true) */
  showScale?: boolean;
  /** Entity name to display as tooltip title */
  entityName?: string;
  /** Whether to show percentage alongside values (default: false) */
  showPercentage?: boolean;
  /** Optional summary line to display at the bottom of the tooltip (without bullet point) */
  tooltipSummary?: string;
}
