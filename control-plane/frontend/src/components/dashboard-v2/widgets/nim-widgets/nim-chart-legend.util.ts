import type { IRangeDates } from "@/utils/date.util";
import { LEGEND_DASHED_LINE_ICON, LEGEND_RECT_ICON } from "@/models/chart.model";

export const NIM_CIRCLE_LEGEND_ICON = "circle";
export const NIM_RECT_LEGEND_ICON = LEGEND_RECT_ICON;
export const NIM_LINE_LEGEND_ICON = LEGEND_RECT_ICON;
export const NIM_VERTICAL_BARS_LEGEND_ICON = LEGEND_DASHED_LINE_ICON;

/** NIM widgets use bold 12 px legend text to match prod Highcharts rendering. */
export const NIM_LEGEND_TEXT_STYLE: Record<string, unknown> = { fontSize: 12, fontWeight: "bold" };

export const NIM_METRIC_CHART_GRID = {
  containLabel: true,
  top: 6,
  right: -10,
  bottom: 50,
  left: 34,
};

export const NIM_HISTOGRAM_CHART_GRID = {
  ...NIM_METRIC_CHART_GRID,
  bottom: 82,
};

export function shouldHideFinalNimXAxisLabel(value: number | string, filterByDates: IRangeDates): boolean {
  const timestamp = typeof value === "number" ? value : new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return false;

  const date = new Date(timestamp);
  const startDate = filterByDates.dateStart;
  const endDate = filterByDates.dateEnd;
  const isMultiDayRange =
    startDate.getFullYear() !== endDate.getFullYear() ||
    startDate.getMonth() !== endDate.getMonth() ||
    startDate.getDate() !== endDate.getDate();

  if (!isMultiDayRange) return false;

  return (
    date.getFullYear() === endDate.getFullYear() &&
    date.getMonth() === endDate.getMonth() &&
    date.getDate() === endDate.getDate()
  );
}
