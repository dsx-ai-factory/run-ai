/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, test } from "vitest";
import {
  DRILLDOWN_TRANSITION_PRESET,
  INSIDE_DATAZOOM_CONFIG,
  LEGEND_DASHED_LINE_ICON,
  LEGEND_RECT_ITEM_STYLE,
} from "./chart.model";

describe("shared chart legend icons", () => {
  test("uses one deterministic four-stripe dashed-line icon", () => {
    expect(LEGEND_DASHED_LINE_ICON.startsWith("path://")).toBe(true);
    expect(LEGEND_DASHED_LINE_ICON.match(/\bM/g)).toHaveLength(4);
  });

  test("uses one 16x7 filled rectangle contract", () => {
    expect(LEGEND_RECT_ITEM_STYLE).toEqual({
      icon: "rect",
      itemWidth: 16,
      itemHeight: 7,
    });
  });
});

describe("shared drilldown animation", () => {
  test("uses the Highcharts-equivalent duration for entering and updating bars", () => {
    expect(DRILLDOWN_TRANSITION_PRESET).toMatchObject({
      animationDuration: 500,
      animationDurationUpdate: 500,
      animationEasing: "sinusoidalInOut",
      animationEasingUpdate: "sinusoidalInOut",
    });
  });
});

describe("shared inside data zoom", () => {
  test("does not consume drawer wheel or pointer scrolling when area zoom is inactive", () => {
    expect(INSIDE_DATAZOOM_CONFIG).toMatchObject({
      disabled: true,
      zoomOnMouseWheel: false,
      moveOnMouseWheel: false,
      moveOnMouseMove: false,
      preventDefaultMouseMove: false,
    });
  });
});
