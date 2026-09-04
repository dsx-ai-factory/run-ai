/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { HistoryRecord, PhaseUpdate } from "@/swagger-models/workloads-client";
import { Phase } from "@/swagger-models/workloads-client";
import type { EChartsOption, CustomSeriesRenderItemAPI, CustomSeriesRenderItemParams } from "echarts";
import * as echarts from "echarts/core";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { chartUtil } from "@/utils/chart.util";

export const workloadChartUtil = { getWorkloadXRangeOptions };

const categories: Phase[] = [
  Phase.Running,
  Phase.Initializing,
  Phase.Pending,
  Phase.Degraded,
  Phase.Stopped,
  Phase.Completed,
  Phase.Failed,
];
// Phases that should display in white color in tooltips for better visibility (they have dark colors that don't contrast well)
const PHASES_WITH_WHITE_TOOLTIP_COLOR: Phase[] = [Phase.Initializing, Phase.Stopped, Phase.Completed];
const WORKLOAD_HISTORY_X_AXIS_SPLIT_NUMBER = 6;
const WORKLOAD_HISTORY_CONTENT_LEFT = 22;
const WORKLOAD_HISTORY_Y_AXIS_LABEL_MARGIN = 108;
const getFilteredCategories = (historyRecords: HistoryRecord[]) =>
  categories.filter((phase: Phase) => {
    const completed = historyRecords.some((record: HistoryRecord) => record.spec.phaseUpdate?.phase === Phase.Completed);
    return completed ? phase !== Phase.Failed : phase !== Phase.Completed;
  });

interface IWorkloadXRangeDataPoint {
  startTime: number;
  endTime: number;
  categoryIndex: number;
  phase: Phase;
}

function getWorkloadXRangeOptions(historyRecords: HistoryRecord[]): EChartsOption {
  // Avoid rendering Completed and Failed terminal phases on the same chart, matching the old Highcharts view.
  const filteredCategories: Phase[] = getFilteredCategories(historyRecords);
  const dataPoints = _buildXRangeDataPoints(historyRecords, filteredCategories);
  const visibleDataPoints = dataPoints.filter((point) => point.categoryIndex >= 0);

  const renderItem = (params: CustomSeriesRenderItemParams, api: CustomSeriesRenderItemAPI) => {
    const dataIndex = params.dataIndex;
    const point = visibleDataPoints[dataIndex];
    if (!point) return;

    const startCoord = api.coord([point.startTime, point.categoryIndex]);
    const endCoord = api.coord([point.endTime, point.categoryIndex]);
    const barHeight = 20;

    const primaryColor = phaseColors[point.phase];
    const secondaryColor = phaseColorsSecondary[point.phase];
    const coordSys = params.coordSys as unknown as { x: number; y: number; width: number; height: number };
    const plotRect = {
      x: coordSys.x,
      y: coordSys.y,
      width: coordSys.width,
      height: coordSys.height,
    };
    const fullBarShape = echarts.graphic.clipRectByRect(
      {
        x: startCoord[0],
        y: startCoord[1] - barHeight / 2,
        width: endCoord[0] - startCoord[0],
        height: barHeight,
      },
      plotRect,
    );
    const centerBarShape = echarts.graphic.clipRectByRect(
      {
        x: startCoord[0],
        y: startCoord[1] - barHeight * 0.1,
        width: endCoord[0] - startCoord[0],
        height: barHeight * 0.2,
      },
      plotRect,
    );

    const children = [];

    if (fullBarShape) {
      children.push({
        type: "rect",
        shape: fullBarShape,
        style: {
          fill: secondaryColor,
        },
      });
    }

    if (centerBarShape) {
      children.push({
        type: "rect",
        shape: centerBarShape,
        style: {
          fill: primaryColor,
        },
      });
    }

    return {
      type: "group",
      children,
    };
  };

  return {
    textStyle: {
      fontFamily: "Roboto",
    },
    grid: {
      top: 40,
      bottom: 44,
      right: 50,
      left: WORKLOAD_HISTORY_CONTENT_LEFT + WORKLOAD_HISTORY_Y_AXIS_LABEL_MARGIN,
    },
    tooltip: chartUtil.buildStandardTooltip({
      trigger: "item",
      formatter: (params: Record<string, unknown>) => {
        const dataIndex = params.dataIndex as number;
        const point = visibleDataPoints[dataIndex];
        if (!point) return "";
        const phase = filteredCategories[point.categoryIndex];
        const startTime = dateUtil.dateFormat(new Date(point.startTime), DATE_FORMATS.STANDARD_DATE_TIME);
        const endTime = dateUtil.dateFormat(new Date(point.endTime), DATE_FORMATS.STANDARD_DATE_TIME);
        const statusColor = PHASES_WITH_WHITE_TOOLTIP_COLOR.includes(phase) ? "white" : phaseColors[phase];
        return `<span style="text-align: start;">Status: <b style="color: ${statusColor};">${phase}</b></span><br/><span style="text-align: start;">${startTime} - ${endTime}</span>`;
      },
    }),
    title: {
      text: "Status over time",
      padding: 0,
      textStyle: {
        fontSize: 14,
        fontWeight: 400,
        color: "#000000B2",
      },
      left: WORKLOAD_HISTORY_CONTENT_LEFT,
    },
    animation: false,
    xAxis: {
      type: "time",
      minInterval: 10 * 60 * 1000,
      splitNumber: WORKLOAD_HISTORY_X_AXIS_SPLIT_NUMBER,
      axisLine: {
        show: true,
        lineStyle: { color: "#ccd6eb", width: 1 },
      },
      axisTick: {
        show: true,
        length: 5,
        lineStyle: { color: "#ccd6eb", width: 1 },
      },
      axisLabel: {
        ...chartUtil.getDatetimeAxisLabelOptions<number>((value: number) => {
          return dateUtil.dateFormat(new Date(value), DATE_FORMATS.STANDARD_DATE_TIME).replace(", ", ",\n");
        }),
        fontSize: 11,
        lineHeight: 14,
      },
    },
    yAxis: {
      type: "category",
      data: filteredCategories,
      inverse: true,
      min: 0,
      max: filteredCategories.length - 1,
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: {
        align: "left",
        margin: WORKLOAD_HISTORY_Y_AXIS_LABEL_MARGIN,
        fontSize: 12,
        formatter: (value: string) => {
          const phase = value as Phase;
          return `{${phase}|${_getPhaseText(phase)}}`;
        },
        rich: filteredCategories.reduce(
          (acc, phase) => {
            acc[phase] = { color: phaseColors[phase], fontSize: 12 };
            return acc;
          },
          {} as Record<string, { color: string; fontSize: number }>,
        ),
      },
    },
    series: [
      {
        type: "custom",
        renderItem: renderItem,
        encode: {
          x: [0, 1],
          y: 2,
        },
        data: visibleDataPoints.map((point) => [point.startTime, point.endTime, point.categoryIndex]),
      },
    ] as unknown as EChartsOption["series"],
  };
}

function _buildXRangeDataPoints(historyRecords: HistoryRecord[], categories: Phase[]): IWorkloadXRangeDataPoint[] {
  return historyRecords.map((historyRecord: HistoryRecord, index: number) => {
    const nextEvent: HistoryRecord | undefined = historyRecords[index + 1];

    const finished: number = nextEvent?.meta.creationTimestamp
      ? new Date(nextEvent.meta.creationTimestamp).getTime()
      : Date.now();

    const phase = (historyRecord.spec.phaseUpdate as PhaseUpdate).phase;
    return {
      startTime: new Date(historyRecord.meta.creationTimestamp).getTime(),
      endTime: finished,
      categoryIndex: categories.indexOf(phase),
      phase,
    };
  });
}

function _getPhaseText(phase: Phase): string {
  switch (phase) {
    case Phase.Degraded:
      return "Running with issues";
    default:
      return phase;
  }
}

export type PhaseColorMap = {
  [key in (typeof Phase)[keyof typeof Phase]]: string;
};

export const phaseColors: PhaseColorMap = {
  [Phase.Initializing]: "#0000008A",
  [Phase.Creating]: "#4CAF50",
  [Phase.Running]: "#4CAF50",
  [Phase.Pending]: "#FFAB00",
  [Phase.Degraded]: "#FFAB00",
  [Phase.Deleting]: "#000000",
  [Phase.Updating]: "#000000",
  [Phase.Stopped]: "#0000008A",
  [Phase.Stopping]: "#0000008A",
  [Phase.Completed]: "#0000008A",
  [Phase.Terminating]: "#0000008A",
  [Phase.Unknown]: "#0000008A",
  [Phase.Resuming]: "#0000008A",
  [Phase.Failed]: "#EF5350",
};

export const phaseColorsSecondary: PhaseColorMap = {
  [Phase.Initializing]: "#FFAB001A",
  [Phase.Resuming]: "#FFAB001A",
  [Phase.Creating]: "#4CAF501A",
  [Phase.Pending]: "#FFAB001A",
  [Phase.Deleting]: "#000000",
  [Phase.Running]: "#4CAF501A",
  [Phase.Updating]: "#000000",
  [Phase.Stopped]: "#0000001A",
  [Phase.Stopping]: "#0000008A",
  [Phase.Degraded]: "#0000008A",
  [Phase.Failed]: "#fbd1d0",
  [Phase.Completed]: "#0000001A",
  [Phase.Terminating]: "#0000001A",
  [Phase.Unknown]: "#0000008A",
};
