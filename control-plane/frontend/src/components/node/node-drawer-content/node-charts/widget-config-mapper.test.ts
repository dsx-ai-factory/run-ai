import { describe, expect, it } from "vitest";

import { advancedWidgets, defaultWidgets } from "./widget-config-mapper";

describe("node chart connected-tooltip configuration", () => {
  it.each([
    ["default", defaultWidgets],
    ["advanced", advancedWidgets],
  ])("uses canonical buckets, an anchor, and a five-sample zoom limit for every %s chart", (_mode, widgets) => {
    widgets.forEach(({ config }) => {
      expect(config.useCanonicalTimeBuckets).toBe(true);
      expect(config.useConnectedTooltipAnchor).toBe(true);
      expect(config.chartOptions.minimumZoomSampleCount).toBe(5);
    });
  });

  it.each([
    "NVLink transmitted bandwidth",
    "NVLink received bandwidth",
    "PCIe transmitted bandwidth",
    "PCIe received bandwidth",
  ])("uses the Highcharts-like zero scale for %s", (chartTitle) => {
    const config = advancedWidgets.find((widget) => widget.config.chartTitle === chartTitle)?.config;

    expect(config?.chartOptions.highchartsLikeYAxis).toBe(true);
  });
});
