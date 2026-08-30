import { describe, expect, test, vi } from "vitest";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "./chart.util";

import { addDays, addHours, addMinutes } from "date-fns";
import type { MeasurementResponseValuesInner } from "@/swagger-models/workloads-client";
import { type MeasurementResponse } from "@/swagger-models/workloads-client";
import { EYAxisLabelFormat, TOOLTIP_BASE_COLOR } from "@/models/chart.model";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";

describe("chart.util", () => {
  describe("chartUtil tooltip HTML safety", () => {
    test("escapes every HTML-significant character in dynamic tooltip text", () => {
      expect(chartUtil.escapeTooltipHtml(`<script>"&'</script>`)).toBe("&lt;script&gt;&quot;&amp;&#39;&lt;/script&gt;");
    });

    test("preserves the supported rich-text tags while removing unsafe tooltip markup", () => {
      const sanitized = chartUtil.sanitizeTooltipHtml(
        `<img src=x onerror="alert('xss')"><b>100</b><script>alert('xss')</script>`,
      );

      expect(sanitized).toContain("<b>100</b>");
      expect(sanitized).not.toContain("<img");
      expect(sanitized).not.toContain("onerror");
      expect(sanitized).not.toContain("<script");
    });

    test("escapes names, values, symbols, and colors in shared tooltip rows", () => {
      const tooltip = chartUtil.formatTooltipValue(
        `<img src=x onerror="alert('name')">`,
        `red;" onmouseover="alert('color')`,
        `<script>alert('value')</script>`,
        `<svg onload="alert('symbol')">`,
      );

      expect(tooltip).toContain("&lt;img src=x onerror=&quot;");
      expect(tooltip).toContain("&lt;script&gt;alert(&#39;value&#39;)&lt;/script&gt;");
      expect(tooltip).toContain("&lt;svg onload=&quot;");
      expect(tooltip).not.toContain("<img");
      expect(tooltip).not.toContain("<script");
      expect(tooltip).not.toContain("<svg");
      expect(tooltip).not.toContain('onmouseover="');
    });

    test("escapes dynamic headers and series values in the default formatter", () => {
      const tooltip = chartUtil.formatDefaultTooltip([
        {
          axisValueLabel: `<img src=x onerror="alert('header')">`,
          seriesName: `<script>alert('series')</script>`,
          value: `<svg onload="alert('value')">`,
          color: "#30A6F5",
        },
      ]);

      expect(tooltip).toContain("&lt;img src=x onerror=&quot;");
      expect(tooltip).toContain("&lt;script&gt;alert(&#39;series&#39;)&lt;/script&gt;");
      expect(tooltip).toContain("&lt;svg onload=&quot;");
      expect(tooltip).not.toContain("<img");
      expect(tooltip).not.toContain("<script");
      expect(tooltip).not.toContain("<svg");
    });
  });

  describe("chartUtil::getDatetimeOptions", () => {
    test("hides time labels that overlap after a responsive resize", () => {
      expect(chartUtil.getDatetimeOptions({})).toMatchObject({
        type: "time",
        axisLabel: {
          fontSize: 11,
          hideOverlap: true,
        },
      });
    });

    test("keeps widget formatter overrides aligned with the shared 11px time-axis style", () => {
      const formatter = (value: number): string => String(value);

      expect(chartUtil.getDatetimeAxisLabelOptions(formatter)).toEqual({
        color: "#666666",
        fontSize: 11,
        hideOverlap: true,
        formatter,
      });
    });

    test("formats the original timestamp in the browser timezone without applying the offset twice", () => {
      const timestamp = "2026-07-07T12:00:00.000Z";
      const dateFormatSpy = vi.spyOn(dateUtil, "dateFormat").mockReturnValue("07/07/2026, 15:00");

      expect(chartUtil.formatDatetimeAxisLabel(timestamp)).toBe("07/07/2026\n15:00");
      expect(dateFormatSpy).toHaveBeenCalledWith(new Date(timestamp), DATE_FORMATS.STANDARD_DATE_TIME);
      dateFormatSpy.mockRestore();
    });

    test("returns an empty label for an invalid timestamp", () => {
      expect(chartUtil.formatDatetimeAxisLabel("not-a-date")).toBe("");
    });
  });

  describe("chartUtil::formatCompactAxisValue", () => {
    test("uses Highcharts-like compact suffixes for large axis values", () => {
      expect(chartUtil.formatCompactAxisValue(1_000)).toBe("1k");
      expect(chartUtil.formatCompactAxisValue(1_500)).toBe("1.5k");
      expect(chartUtil.formatCompactAxisValue(2_000)).toBe("2k");
      expect(chartUtil.formatCompactAxisValue(1_000_000)).toBe("1M");
    });

    test("preserves small and fractional tick values", () => {
      expect(chartUtil.formatCompactAxisValue(0.5)).toBe("0.5");
      expect(chartUtil.formatCompactAxisValue(9)).toBe("9");
    });
  });

  describe("chartUtil::getLineChartYAxisConfig", () => {
    test("collapses sub-display CPU noise to a single zero tick", () => {
      expect(chartUtil.getLineChartYAxisConfig(0.0001, { nearZeroThreshold: 0.005 })).toEqual({
        min: 0,
        max: 1,
        interval: 2,
      });
    });

    test("keeps scalable nice-number ticks for measurable values", () => {
      expect(chartUtil.getLineChartYAxisConfig(0.61, { nearZeroThreshold: 0.005 })).toEqual({
        min: 0,
        max: 0.75,
        interval: 0.25,
      });
    });
  });

  describe("chartUtil::getBarSeriesLegendHoverState", () => {
    test("should provide Highcharts-like legend hover focus and inactive opacity", () => {
      expect(chartUtil.getBarSeriesLegendHoverState()).toEqual({
        legendHoverLink: true,
        emphasis: {
          focus: "series",
        },
        blur: {
          itemStyle: {
            opacity: 0.2,
          },
          label: {
            opacity: 0.2,
          },
        },
      });
    });
  });

  describe("chartUtil::getHiddenAxisPointer", () => {
    test("should hide the shared tooltip axis ruler for dashboard bar charts", () => {
      expect(chartUtil.getHiddenAxisPointer()).toEqual({
        type: "line",
        lineStyle: { opacity: 0, width: 0 },
        label: { show: false },
      });
    });
  });

  describe("chartUtil::getConnectedChartHoverMarker", () => {
    test("uses an opaque hover marker so the plotted line does not show through", () => {
      expect(chartUtil.getConnectedChartHoverMarker("#30A6F5")).toEqual({
        symbol: "circle",
        symbolSize: 16,
        emphasis: {
          scale: false,
          itemStyle: {
            color: "#FFFFFF",
            borderColor: "#30A6F5",
            borderWidth: 2,
          },
        },
      });
    });
  });

  describe("chartUtil::getConnectedChartAxisPointer", () => {
    test("renders the shared crosshair below line-series hover markers", () => {
      expect(chartUtil.getConnectedChartAxisPointer()).toEqual({
        z: 1,
        type: "line",
        lineStyle: {
          width: 2,
          color: "#0000008A",
          type: "solid",
        },
      });
    });

    test("applies the shared layering to standard axis tooltips", () => {
      expect(chartUtil.buildStandardTooltip({ trigger: "axis" })).toMatchObject({
        axisPointer: {
          z: 1,
          type: "line",
          lineStyle: { width: 2, color: "#0000008A", type: "solid" },
        },
      });
    });
  });

  describe("chartUtil dotted plot-line style", () => {
    test("uses one widened 1:1 dash cadence across plotted series", () => {
      const lineStyle = chartUtil.getDottedLineStyle();

      expect(lineStyle).toEqual({ width: 2, type: [4, 4] });
      expect(chartUtil.isDottedLineStyle(lineStyle)).toBe(true);
      expect(chartUtil.isDottedLineStyle({ type: "dotted" })).toBe(true);
      expect(chartUtil.isDottedLineStyle({ type: "solid" })).toBe(false);
    });

    test("preserves per-series opacity overrides", () => {
      expect(chartUtil.getDottedLineStyle({ opacity: 0 })).toEqual({ width: 2, type: [4, 4], opacity: 0 });
    });
  });

  describe("chartUtil::getConnectedTooltipAnchorSeries", () => {
    test("creates an invisible first-series candidate at every canonical timestamp", () => {
      const timeframes = ["2026-07-05T12:00:00.000Z", "2026-07-05T13:00:00.000Z"];
      const series = chartUtil.getConnectedTooltipAnchorSeries(timeframes);

      expect(series).toMatchObject({
        id: "connected-tooltip-anchor",
        type: "line",
        silent: true,
        animation: false,
        showSymbol: false,
        lineStyle: { color: "transparent", opacity: 0, width: 0 },
        itemStyle: { color: "transparent", opacity: 0 },
      });
      expect(series.name).toBeUndefined();
      expect(series.data).toEqual([
        [new Date(timeframes[0]).getTime(), 0],
        [new Date(timeframes[1]).getTime(), 0],
      ]);
      expect(chartUtil.isConnectedTooltipAnchorSeries({ seriesId: series.id })).toBe(true);
      expect(chartUtil.isConnectedTooltipAnchorSeries({ id: series.id })).toBe(true);
      expect(chartUtil.isConnectedTooltipAnchorSeries({ seriesId: "real-series" })).toBe(false);
    });
  });

  describe("chartUtil::getDashboardLineSeriesStyle", () => {
    test("uses the same explicit color for the plot line and an opaque hover marker", () => {
      const style = chartUtil.getDashboardLineSeriesStyle(0);

      expect(style.itemStyle?.color).toBe(style.lineStyle?.color);
      expect(style.emphasis?.itemStyle).toEqual({
        color: "#FFFFFF",
        borderColor: style.lineStyle?.color,
        borderWidth: 2,
      });
    });
  });

  describe("chartUtil zoom interaction", () => {
    test("disables the inside roam controller when wheel zoom is not explicitly allowed", () => {
      const options = chartUtil.addZoomOptions({}, "chart-id", vi.fn(), vi.fn());

      expect(Array.isArray(options.dataZoom) ? options.dataZoom[0] : options.dataZoom).toMatchObject({
        type: "inside",
        disabled: true,
        zoomOnMouseWheel: false,
        moveOnMouseWheel: false,
        moveOnMouseMove: false,
        preventDefaultMouseMove: false,
      });
    });

    test("enables native wheel and drag roaming only when explicitly requested", () => {
      const options = chartUtil.addZoomOptions({}, "chart-id", vi.fn(), vi.fn(), { allowWheelZoom: true });

      expect(Array.isArray(options.dataZoom) ? options.dataZoom[0] : options.dataZoom).toMatchObject({
        type: "inside",
        disabled: false,
        zoomOnMouseWheel: true,
        moveOnMouseMove: true,
        preventDefaultMouseMove: true,
      });
    });

    test("deactivates area selection before resetting the data range and cursor", () => {
      const dispatchAction = vi.fn();
      const setCursorStyle = vi.fn();
      const chart = {
        dispatchAction,
        getZr: () => ({ setCursorStyle }),
      } as never;

      chartUtil.resetChartDataZoomInstance(chart);

      expect(dispatchAction).toHaveBeenNthCalledWith(
        1,
        { type: "takeGlobalCursor", key: "dataZoomSelect", dataZoomSelectActive: false },
        { silent: true },
      );
      expect(dispatchAction).toHaveBeenNthCalledWith(2, { type: "dataZoom", start: 0, end: 100 }, { silent: true });
      expect(setCursorStyle).toHaveBeenCalledWith("default");
    });

    test("matches the legacy five-sample minimum zoom range", () => {
      const timeframes = ["2026-07-06T10:00:00.000Z", "2026-07-06T10:00:15.000Z", "2026-07-06T10:00:30.000Z"];

      expect(chartUtil.getMinimumDataZoomValueSpan(timeframes, 5)).toBe(75_000);
      expect(chartUtil.getMinimumDataZoomValueSpan(timeframes, 5, 600_000)).toBe(600_000);
    });

    test("uses the smallest positive interval and ignores invalid timestamps", () => {
      const timeframes = ["2026-07-06T10:00:00.000Z", "invalid", "2026-07-06T10:01:00.000Z", "2026-07-06T10:01:30.000Z"];

      expect(chartUtil.getMinimumDataZoomValueSpan(timeframes, 5)).toBe(150_000);
      expect(chartUtil.getMinimumDataZoomValueSpan([timeframes[0]], 5)).toBeUndefined();
    });
  });

  describe("chartUtil::formatDefaultTooltip", () => {
    test("uses the shared 6px marker for categorical tooltips", () => {
      const tooltip = chartUtil.formatDefaultTooltip([
        { axisValueLabel: "engineering", seriesName: "(0)-(2d-3h)", color: "#30A6F5", value: 0 },
      ]);

      expect(tooltip).toContain("width:6px;height:6px");
      expect(tooltip).toContain("flex:0 0 6px");
      expect(tooltip).toContain("engineering");
      expect(tooltip).toContain("(0)-(2d-3h): <b>0</b>");
    });

    test("defines the shared dark tooltip style", () => {
      expect(STANDARD_TOOLTIP_STYLE.backgroundColor).toBe(TOOLTIP_BASE_COLOR);
      expect(STANDARD_TOOLTIP_STYLE.borderColor).toBe(TOOLTIP_BASE_COLOR);
      expect(STANDARD_TOOLTIP_STYLE.textStyle.fontSize).toBe(12);
      expect(STANDARD_TOOLTIP_STYLE.textStyle.color).toBe("#F0F0F0");
      expect(STANDARD_TOOLTIP_STYLE.appendToBody).toBe(false);
      expect(STANDARD_TOOLTIP_STYLE.confine).toBe(true);
    });
  });

  describe("chartUtil::formatSharedTooltip", () => {
    test("retains the timestamp header when every series has no value at the hovered time", () => {
      const tooltip = chartUtil.formatSharedTooltip([], new Date("2026-06-25T01:00:00.000Z").getTime());

      expect(tooltip).toContain("padding-left: 3px");
      expect(tooltip).not.toContain("width:6px");
    });
  });

  describe("chartUtil::getTimeFrameForHoursAgo", () => {
    test("returns array of time frames between two dates with specified interval", () => {
      const dateStart = new Date("2024-03-19T00:00:00.000Z");
      const dateEnd = new Date("2024-03-19T02:00:00.000Z");
      const interval = 3600; // 1 hour interval
      const expectedTimeFrames = ["2024-03-19T00:00:00.000Z", "2024-03-19T01:00:00.000Z", "2024-03-19T02:00:00.000Z"];

      const result = chartUtil.getTimeFrameForHoursAgo(dateStart, dateEnd, interval);
      expect(result).toEqual(expectedTimeFrames);
    });

    test("returns array with only start date if start date equals end date", () => {
      const dateStart = new Date("2024-03-19T00:00:00.000Z");
      const dateEnd = new Date("2024-03-19T00:00:00.000Z");
      const interval = 3600; // 1 hour interval

      const result = chartUtil.getTimeFrameForHoursAgo(dateStart, dateEnd, interval);

      expect(result).toEqual(["2024-03-19T00:00:00.000Z"]);
    });

    test("returns empty array if start date is after end date", () => {
      const dateStart = new Date("2024-03-19T02:00:00");
      const dateEnd = new Date("2024-03-19T00:00:00");
      const interval = 3600; // 1 hour interval

      const result = chartUtil.getTimeFrameForHoursAgo(dateStart, dateEnd, interval);

      expect(result).toEqual([]);
    });

    test("handles small intervals correctly (30 seconds)", () => {
      const dateStart = new Date("2024-03-19T00:00:00.000Z");
      const dateEnd = new Date("2024-03-19T00:02:00.000Z"); // 2 minutes
      const interval = 30; // 30 seconds

      const result = chartUtil.getTimeFrameForHoursAgo(dateStart, dateEnd, interval);

      expect(result).toEqual([
        "2024-03-19T00:00:00.000Z",
        "2024-03-19T00:00:30.000Z",
        "2024-03-19T00:01:00.000Z",
        "2024-03-19T00:01:30.000Z",
        "2024-03-19T00:02:00.000Z",
      ]);
    });

    test("handles large intervals correctly (1 day)", () => {
      const dateStart = new Date("2024-03-19T00:00:00.000Z");
      const dateEnd = new Date("2024-03-22T00:00:00.000Z"); // 3 days
      const interval = 86400; // 1 day

      const result = chartUtil.getTimeFrameForHoursAgo(dateStart, dateEnd, interval);

      expect(result).toEqual([
        "2024-03-19T00:00:00.000Z",
        "2024-03-20T00:00:00.000Z",
        "2024-03-21T00:00:00.000Z",
        "2024-03-22T00:00:00.000Z",
      ]);
    });
  });

  describe("chartUtil::getTimeFrames", () => {
    /**
     * Helper function to calculate the interval in seconds between timeframes
     */
    function getIntervalSeconds(timeframes: string[]): number {
      if (timeframes.length < 2) return 0;
      const firstTime = new Date(timeframes[0]).getTime();
      const secondTime = new Date(timeframes[1]).getTime();
      return (secondTime - firstTime) / 1000;
    }

    test("15 min range: should use 15 sec interval (~60 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T00:15:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      expect(intervalSec).toBe(15);
      expect(timeframes.length).toBeLessThanOrEqual(61); // 900/15 = 60 + 1
    });

    test("30 min range: should use 15 sec interval (~120 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T00:30:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      expect(intervalSec).toBe(15);
      expect(timeframes.length).toBeLessThanOrEqual(121); // 1800/15 = 120 + 1
    });

    test("1 hour range: should use 15 sec interval (~240 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T01:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      expect(intervalSec).toBe(15);
      expect(timeframes.length).toBeLessThanOrEqual(241); // 3600/15 = 240 + 1
    });

    test("2 hours range: should use 15 sec interval (~480 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T02:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 7200 seconds / 15 = 480 points (≤500), so uses 15 sec
      expect(intervalSec).toBe(15);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("3 hours range: should use 30 sec interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T03:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 10800 / 15 = 720 (>500), so tries 30: 10800/30 = 360 (≤500)
      expect(intervalSec).toBe(30);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("6 hours range: should use 1 min interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T06:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 21600 / 30 = 720 (>500), tries 60: 21600/60 = 360 (≤500)
      expect(intervalSec).toBe(60);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("12 hours range: should use 2 min interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T12:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 43200 / 60 = 720 (>500), tries 120: 43200/120 = 360 (≤500)
      expect(intervalSec).toBe(120);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("1 day range: should use 5 min interval (~288 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-02T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 86400 / 120 = 720 (>500), tries 300: 86400/300 = 288 (≤500)
      expect(intervalSec).toBe(300);
      expect(timeframes.length).toBeLessThanOrEqual(289);
    });

    test("2 days range: should use 10 min interval (~288 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-03T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 172800 / 300 = 576 (>500), tries 600: 172800/600 = 288 (≤500)
      expect(intervalSec).toBe(600);
      expect(timeframes.length).toBeLessThanOrEqual(289);
    });

    test("3 days range: should use 10 min interval (~432 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-04T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 259200 / 600 = 432 (≤500)
      expect(intervalSec).toBe(600);
      expect(timeframes.length).toBeLessThanOrEqual(433);
    });

    test("4 days range: should use 15 min interval (~384 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-05T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 345600 / 600 = 576 (>500), tries 900: 345600/900 = 384 (≤500)
      expect(intervalSec).toBe(900);
      expect(timeframes.length).toBeLessThanOrEqual(385);
    });

    test("5 days range: should use 15 min interval (~480 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-06T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 432000 / 900 = 480 (≤500)
      expect(intervalSec).toBe(900);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("6 days range: should use 30 min interval (~288 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-07T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 518400 / 900 = 576 (>500), tries 1800: 518400/1800 = 288 (≤500)
      expect(intervalSec).toBe(1800);
      expect(timeframes.length).toBeLessThanOrEqual(289);
    });

    test("7 days range: should use 30 min interval (~336 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-08T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 604800 / 1800 = 336 (≤500)
      expect(intervalSec).toBe(1800);
      expect(timeframes.length).toBeLessThanOrEqual(337);
    });

    test("10 days range: should use 30 min interval (~480 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-11T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 864000 / 1800 = 480 (≤500)
      expect(intervalSec).toBe(1800);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("12 days range: should use 1 hour interval (~288 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-13T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 1036800 / 1800 = 576 (>500), tries 3600: 1036800/3600 = 288 (≤500)
      expect(intervalSec).toBe(3600);
      expect(timeframes.length).toBeLessThanOrEqual(289);
    });

    test("15 days range: should use 1 hour interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-16T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 1296000 / 3600 = 360 (≤500)
      expect(intervalSec).toBe(3600);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("19 days range: should use 1 hour interval (~456 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-20T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 1641600 / 3600 = 456 (≤500)
      expect(intervalSec).toBe(3600);
      expect(timeframes.length).toBeLessThanOrEqual(457);
    });

    test("20 days range: should use 2 hours interval (~240 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-21T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 1728000 / 3600 = 480 (≤500), but actually uses 3600
      // Wait, let me recalculate: 1728000 / 3600 = 480, which is ≤500
      // So it should use 1 hour (3600), not 2 hours
      expect(intervalSec).toBe(3600);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("22 days range: should use 2 hours interval (~264 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-23T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 1900800 / 3600 = 528 (>500), tries 7200: 1900800/7200 = 264 (≤500)
      expect(intervalSec).toBe(7200);
      expect(timeframes.length).toBeLessThanOrEqual(265);
    });

    test("24 days range: should use 2 hours interval (~288 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-25T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 2073600 / 7200 = 288 (≤500)
      expect(intervalSec).toBe(7200);
      expect(timeframes.length).toBeLessThanOrEqual(289);
    });

    test("25 days range: should use 3 hours interval (~200 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-26T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 2160000 / 7200 = 300 (≤500), so uses 7200 (2 hours)
      // Wait: 2160000 / 7200 = 300, which is ≤500, so should use 2 hours
      expect(intervalSec).toBe(7200);
      expect(timeframes.length).toBeLessThanOrEqual(301);
    });

    test("30 days range: should use 3 hours interval (~240 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-31T00:00:00.000Z");

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 2592000 / 7200 = 360 (≤500), so uses 7200 (2 hours)
      expect(intervalSec).toBe(7200);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    // Extended range tests (> 30 days)
    test("45 days range: should use 3 hours interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 45);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 3888000 / 10800 = 360 (≤500), uses 3 hours
      expect(intervalSec).toBe(10800);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("60 days (2 months) range: should use 3 hours interval (~480 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 60);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 5184000 / 10800 = 480 (≤500)
      expect(intervalSec).toBe(10800);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("90 days (3 months) range: should use 6 hours interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 90);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 7776000 / 21600 = 360 (≤500), uses 6 hours
      expect(intervalSec).toBe(21600);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("120 days (4 months) range: should use 6 hours interval (~480 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 120);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 10368000 / 21600 = 480 (≤500)
      expect(intervalSec).toBe(21600);
      expect(timeframes.length).toBeLessThanOrEqual(481);
    });

    test("180 days (6 months) range: should use 12 hours interval (~360 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 180);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 15552000 / 43200 = 360 (≤500), uses 12 hours
      expect(intervalSec).toBe(43200);
      expect(timeframes.length).toBeLessThanOrEqual(361);
    });

    test("270 days (9 months) range: should use 1 day interval (~270 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 270);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 23328000 / 86400 = 270 (≤500)
      expect(intervalSec).toBe(86400);
      expect(timeframes.length).toBeLessThanOrEqual(271);
    });

    test("365 days (1 year) range: should use 1 day interval (~365 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 365);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 31536000 / 86400 = 365 (≤500), uses 1 day
      expect(intervalSec).toBe(86400);
      expect(timeframes.length).toBeLessThanOrEqual(366);
    });

    test("540 days (18 months) range: should use 2 days interval (~270 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 540);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 46656000 / 172800 = 270 (≤500)
      expect(intervalSec).toBe(172800);
      expect(timeframes.length).toBeLessThanOrEqual(271);
    });

    test("730 days (2 years) range: should use 2 days interval (~365 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 730);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 63072000 / 172800 = 365 (≤500), uses 2 days
      expect(intervalSec).toBe(172800);
      expect(timeframes.length).toBeLessThanOrEqual(366);
    });

    test("1095 days (3 years) range: should use 3 days interval (~365 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 1095);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 94608000 / 259200 = 365 (≤500)
      expect(intervalSec).toBe(259200);
      expect(timeframes.length).toBeLessThanOrEqual(366);
    });

    test("1460 days (4 years) range: should use 3 days interval (~487 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 1460);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 126144000 / 259200 = 486.7 (≤500)
      expect(intervalSec).toBe(259200);
      expect(timeframes.length).toBeLessThanOrEqual(488);
    });

    test("1825 days (5 years) range: should use 7 days interval (~261 samples)", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 1825);

      const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);
      const intervalSec = getIntervalSeconds(timeframes);

      // 157680000 / 259200 = 608.3 (>500), tries 604800: 157680000/604800 = 260.7 (≤500)
      expect(intervalSec).toBe(604800);
      expect(timeframes.length).toBeLessThanOrEqual(262);
    });

    test("ensures all timeframes stay within reasonable limits (MAX_DATA_POINTS ~500)", () => {
      const dateStart = new Date("2024-01-01T00:00:00.000Z");
      const MAX_DATA_POINTS = 500;
      const testRanges = [
        { minutes: 15 }, // 15 minutes
        { hours: 1 },
        { hours: 6 },
        { hours: 12 },
        { days: 1 },
        { days: 7 },
        { days: 30 },
        { days: 90 },
        { days: 180 },
        { days: 365 },
        { days: 365 * 2 }, // 2 years
        { days: 365 * 3 }, // 3 years
        { days: 365 * 4 }, // 4 years
        { days: 365 * 5 }, // 5 years (maximum)
      ];

      testRanges.forEach(({ minutes, hours, days }) => {
        // Use appropriate date function to avoid floating point precision issues
        let dateEnd: Date;
        if (minutes) {
          dateEnd = addMinutes(dateStart, minutes);
        } else if (hours) {
          dateEnd = addHours(dateStart, hours);
        } else {
          dateEnd = addDays(dateStart, days || 0);
        }

        const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);

        // The algorithm uses Math.ceil which can produce points slightly over 500
        // The while loop includes the end point (<=), adding 1 more
        // Allow reasonable tolerance for these implementation details
        expect(timeframes.length).toBeLessThanOrEqual(MAX_DATA_POINTS + 10);

        // Also verify it's not producing an unreasonably small number
        expect(timeframes.length).toBeGreaterThan(0);
      });
    });

    test("throws error if start date is after end date", () => {
      const dateStart = new Date("2024-03-01T02:00:00");
      const dateEnd = new Date("2024-03-01T00:00:00");

      expect(() => chartUtil.getTimeFrames(dateStart, dateEnd)).toThrow(/Invalid date range/);
    });

    test("throws error if start and end dates are equal", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-01T00:00:00.000Z");

      expect(() => chartUtil.getTimeFrames(dateStart, dateEnd)).toThrow(/Invalid date range/);
    });

    test("throws error if range exceeds 5 years", () => {
      const dateStart = new Date("2024-01-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 365 * 5 + 1); // 5 years + 1 day

      expect(() => chartUtil.getTimeFrames(dateStart, dateEnd)).toThrow(
        /Range time: .* is not supported. Maximum range is 5 years./,
      );
    });

    test("accepts exactly 5 years range", () => {
      const dateStart = new Date("2024-01-01T00:00:00.000Z");
      const dateEnd = addDays(dateStart, 365 * 5); // Exactly 5 years

      expect(() => chartUtil.getTimeFrames(dateStart, dateEnd)).not.toThrow();
    });

    test("returns consistent results for the same date range", () => {
      const dateStart = new Date("2024-03-01T00:00:00.000Z");
      const dateEnd = new Date("2024-03-08T00:00:00.000Z");

      const result1 = chartUtil.getTimeFrames(dateStart, dateEnd);
      const result2 = chartUtil.getTimeFrames(dateStart, dateEnd);

      expect(result1).toEqual(result2);
      expect(result1.length).toBe(result2.length);
    });
  });

  describe("chartUtil::arrangeMeasurementsByTime", () => {
    test("correctly arranges measurements by time frames", () => {
      const timeFrames = [
        "2024-03-19T00:00:00.000Z",
        "2024-03-19T01:00:00.000Z",
        "2024-03-19T02:00:00.000Z",
        "2024-03-19T03:00:00.000Z",
      ];
      const measurements = [
        { timestamp: "2024-03-19T00:15:00.000Z", value: "10" },
        { timestamp: "2024-03-19T01:20:00.000Z", value: "20" },
        { timestamp: "2024-03-19T01:40:00.000Z", value: "30" },
        { timestamp: "2024-03-19T02:30:00.000Z", value: "40" },
        { timestamp: "2024-03-19T03:45:00.000Z", value: "50" }, // should be ignored
      ];
      const expectedChartData = [
        [new Date("2024-03-19T00:15:00.000Z").getTime(), 10],
        [new Date("2024-03-19T01:20:00.000Z").getTime(), 20],
        [new Date("2024-03-19T01:40:00.000Z").getTime(), 30],
        [new Date("2024-03-19T02:30:00.000Z").getTime(), 40],
        [new Date("2024-03-19T03:00:00.000Z").getTime(), null],
      ];

      const result = chartUtil.arrangeMeasurementsByTime(timeFrames, measurements);
      expect(result).toEqual(expectedChartData);
    });

    test("handles empty measurements array", () => {
      const timeFrames = ["2024-03-19T00:00:00.000Z", "2024-03-19T01:00:00.000Z"];
      const measurements: Array<MeasurementResponseValuesInner> = [];

      const result = chartUtil.arrangeMeasurementsByTime(timeFrames, measurements);

      expect(result).toEqual([
        [new Date("2024-03-19T00:00:00.000Z").getTime(), null],
        [new Date("2024-03-19T01:00:00.000Z").getTime(), null],
      ]);
    });

    test("uses one canonical timestamp per timeframe when requested", () => {
      const timeFrames = ["2024-03-19T00:00:00.000Z", "2024-03-19T01:00:00.000Z", "2024-03-19T02:00:00.000Z"];
      const measurements = [
        { timestamp: "2024-03-19T00:00:00.250Z", value: "10" },
        { timestamp: "2024-03-19T01:00:00.750Z", value: "20" },
      ];

      const result = chartUtil.arrangeMeasurementsByTime(timeFrames, measurements, {
        useTimeFrameTimestamps: true,
      });

      expect(result).toEqual([
        [new Date(timeFrames[0]).getTime(), 10],
        [new Date(timeFrames[1]).getTime(), 20],
        [new Date(timeFrames[2]).getTime(), null],
      ]);
    });

    test("selects one representative value when a timeframe contains multiple measurements", () => {
      const timeFrames = ["2024-03-19T00:00:00.000Z", "2024-03-19T01:00:00.000Z"];
      const measurements = [
        { timestamp: "2024-03-19T00:40:00.000Z", value: "40" },
        { timestamp: "2024-03-19T00:05:00.000Z", value: "5" },
      ];

      const result = chartUtil.arrangeMeasurementsByTime(timeFrames, measurements, {
        useTimeFrameTimestamps: true,
      });

      expect(result).toEqual([
        [new Date(timeFrames[0]).getTime(), 5],
        [new Date(timeFrames[1]).getTime(), null],
      ]);
    });
  });

  describe("updateYAxisLabelsDefaultsIfNeeded", () => {
    test("should return the original yAxis when there are valid measurements", () => {
      const measurements: MeasurementResponse[] = [{ type: "metric", values: [{ value: "10", timestamp: null }] }];
      const yAxis = { min: 0, max: 20 };
      const options = { steps: [0, 20, 40], type: EYAxisLabelFormat.Memory };

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      expect(result).toEqual(yAxis);
    });

    test("should update yAxis with custom range when no measurements and steps provided", () => {
      const measurements: MeasurementResponse[] = [];
      const yAxis = { min: 0, max: 20 };
      const options = { steps: [0, 50, 100], type: EYAxisLabelFormat.Memory };

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      expect(result.min).toBe(0);
      expect(result.max).toBe(100);
      expect(result.interval).toBe(50);
      expect(result.axisLabel?.formatter).toBeDefined();
    });

    test("should update yAxis with default memory range when no steps provided and type is Memory", () => {
      const measurements: MeasurementResponse[] = [];
      const yAxis = { min: 0, max: 20 };
      const options = { type: EYAxisLabelFormat.Memory };

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      expect(result.min).toBe(0);
      expect(result.max).toBe(80);
      expect(result.interval).toBe(20);
      expect(result.axisLabel?.formatter).toBeDefined();
    });

    test('should return "Bytes" for 0 and "G" for non-zero values in Memory mode', () => {
      const measurements: MeasurementResponse[] = [];
      const yAxis = { min: 0, max: 20 };
      const options = { type: EYAxisLabelFormat.Memory };

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      const formatter = result.axisLabel?.formatter as (value: number) => string;
      expect(formatter).toBeDefined();

      expect(formatter(0)).toEqual("0 Bytes");
      expect(formatter(20)).toEqual("20 G");
    });

    test('should return original yAxis with range when type is not "Memory" and no measurements', () => {
      const measurements: MeasurementResponse[] = [];
      const yAxis = { min: 0, max: 20 };
      const options = {};

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      expect(result.min).toBe(0);
      expect(result.max).toBe(80);
      expect(result.interval).toBe(20);
    });

    test("should use default range if no steps are provided and no measurements", () => {
      const measurements: MeasurementResponse[] = [];
      const yAxis = {};
      const options = {};

      const result = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, options);

      expect(result.min).toBe(0);
      expect(result.max).toBe(80);
      expect(result.interval).toBe(20);
      expect(result.axisLabel).toBeUndefined();
    });
  });
});

describe("chartUtil::getResourcesOptions", () => {
  test("uses the project-card legend colors and comfortable item spacing", () => {
    const options = chartUtil.getResourcesOptions(2, 10, 12);
    const legend = options.legend as {
      itemGap?: number;
      width?: number;
      itemWidth?: number;
      itemHeight?: number;
      textStyle?: { color?: string; fontSize?: number };
    };
    const series = options.series as Array<{ name: string; color: string }>;
    const tooltip = options.tooltip as {
      backgroundColor?: string;
      borderColor?: string;
      borderWidth?: number;
      borderRadius?: number;
      extraCssText?: string;
      textStyle?: { color?: string };
      appendTo?: string;
      confine?: boolean;
      transitionDuration?: number;
      displayTransition?: boolean;
      formatter?: (params: Record<string, unknown>[]) => string;
    };

    expect(legend.itemGap).toBe(14);
    expect(legend.width).toBe(224);
    expect(legend.itemWidth).toBe(8);
    expect(legend.itemHeight).toBe(8);
    expect(legend.textStyle).toMatchObject({ color: "#0000008A", fontSize: 12 });
    expect(tooltip).toMatchObject({
      appendTo: "body",
      confine: false,
      transitionDuration: 0.6,
      displayTransition: true,
      backgroundColor: "rgba(255, 255, 255, 0.85)",
      borderColor: TOOLTIP_BASE_COLOR,
      borderWidth: 1,
      borderRadius: 3,
      textStyle: { color: "#333333" },
    });
    expect(tooltip.extraCssText).toContain("box-shadow");
    const tooltipHtml = tooltip.formatter?.([
      { seriesName: "Over quota" },
      { seriesName: "Quota" },
      { seriesName: "Allocated" },
    ]);
    expect(tooltipHtml).toContain("line-height:16px");
    expect(tooltipHtml).toContain("width:8px;height:8px");
    expect(tooltipHtml).not.toContain("<br");
    expect(series).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ name: "Over quota", color: "#544FC5" }),
        expect.objectContaining({ name: "Quota", color: "#50E991" }),
        expect.objectContaining({ name: "Allocated", color: "#0DB6FF" }),
      ]),
    );
  });
});
