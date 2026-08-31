/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, test } from "vitest";
import type { XAXisComponentOption } from "echarts";
import { Phase, type HistoryRecord } from "@/swagger-models/workloads-client";
import { workloadChartUtil } from "./workload-chart.util";

function getTooltipHtml(phase: Phase): string {
  const options = workloadChartUtil.getWorkloadXRangeOptions([
    {
      meta: { creationTimestamp: "2026-07-08T09:00:00Z" },
      spec: { phaseUpdate: { phase } },
    } as HistoryRecord,
  ]);
  const formatter = (options.tooltip as { formatter: (params: { dataIndex: number }) => string }).formatter;

  return formatter({ dataIndex: 0 });
}

describe("workloadChartUtil", () => {
  test("keeps workload history time labels readable at responsive drawer widths", () => {
    const options = workloadChartUtil.getWorkloadXRangeOptions([]);
    const xAxis = options.xAxis as XAXisComponentOption;

    expect(xAxis).toMatchObject({
      type: "time",
      splitNumber: 6,
      axisLabel: {
        fontSize: 11,
        lineHeight: 14,
        hideOverlap: true,
      },
    });
  });

  test("keeps the title and status labels aligned with legacy chart spacing", () => {
    const options = workloadChartUtil.getWorkloadXRangeOptions([]);

    expect(options.grid).toMatchObject({ left: 130 });
    expect(options.title).toMatchObject({ left: 22, padding: 0 });
    expect(options.yAxis).toMatchObject({
      axisLabel: {
        fontSize: 12,
        margin: 108,
      },
    });
  });

  test("preserves the two-line date and time label format", () => {
    const options = workloadChartUtil.getWorkloadXRangeOptions([]);
    const xAxis = options.xAxis as XAXisComponentOption;
    const formatter = (xAxis.axisLabel as { formatter: (value: number) => string }).formatter;
    const formattedLabel = formatter(new Date(2026, 6, 2, 12, 0).getTime());

    expect(formattedLabel).toMatch(/^02\/07\/2026,\n12:00$/);
  });

  test("uses the shared chart tooltip styling", () => {
    const options = workloadChartUtil.getWorkloadXRangeOptions([]);

    expect(options.tooltip).toMatchObject({
      trigger: "item",
      backgroundColor: "#121212",
      borderColor: "#121212",
      textStyle: {
        color: "#F0F0F0",
        fontFamily: "Roboto, sans-serif",
        fontSize: 12,
      },
    });
  });

  test.each([Phase.Initializing, Phase.Stopped, Phase.Completed])(
    "renders the gray %s status name in white inside the tooltip",
    (phase) => {
      expect(getTooltipHtml(phase)).toContain(`<b style="color: white;">${phase}</b>`);
    },
  );

  test("preserves the phase color for a non-gray status tooltip", () => {
    expect(getTooltipHtml(Phase.Running)).toContain('<b style="color: #4CAF50;">Running</b>');
  });
});
